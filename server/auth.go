package main

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

const userSelectCols = `u.id::text, u.email, u.full_name, u.role, u.status, u.phone, u.created_at`

// Same columns without a table alias, for INSERT/UPDATE ... RETURNING.
const userReturnCols = `id::text, email, full_name, role, status, phone, created_at`

type authUser struct {
	ID       string
	Role     string
	Email    string
	FullName string
}

type jwtClaims struct {
	UserID string `json:"uid"`
	Role   string `json:"role"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

func checkPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func generateToken(u UserJSON) (string, error) {
	ttl, err := time.ParseDuration(cfg.JWTExpiry)
	if err != nil || ttl <= 0 {
		ttl = 168 * time.Hour
	}
	now := time.Now()
	claims := jwtClaims{
		UserID: u.ID,
		Role:   u.Role,
		Email:  u.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   u.ID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(cfg.JWTSecret))
}

func currentUser(c *fiber.Ctx) *authUser {
	v := c.Locals("user")
	if v == nil {
		return nil
	}
	u, ok := v.(*authUser)
	if !ok {
		return nil
	}
	return u
}

// requireAuth validates the Bearer JWT and stores the caller in c.Locals.
func requireAuth(c *fiber.Ctx) error {
	header := c.Get("Authorization")
	if !strings.HasPrefix(header, "Bearer ") {
		return fail(c, fiber.StatusUnauthorized, "missing or invalid authorization token")
	}
	raw := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
	if raw == "" {
		return fail(c, fiber.StatusUnauthorized, "missing authorization token")
	}

	claims := &jwtClaims{}
	token, err := jwt.ParseWithClaims(raw, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(cfg.JWTSecret), nil
	})
	if err != nil || !token.Valid || claims.UserID == "" {
		return fail(c, fiber.StatusUnauthorized, "invalid or expired token")
	}

	c.Locals("user", &authUser{
		ID:       claims.UserID,
		Role:     claims.Role,
		Email:    claims.Email,
		FullName: "",
	})
	return c.Next()
}

// requireRole must be registered after requireAuth in the route chain.
func requireRole(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		u := currentUser(c)
		if u == nil {
			return fail(c, fiber.StatusUnauthorized, "authentication required")
		}
		for _, r := range roles {
			if u.Role == r {
				return c.Next()
			}
		}
		return fail(c, fiber.StatusForbidden, "insufficient permissions")
	}
}

func findUserByID(ctx context.Context, id string) (UserJSON, error) {
	return scanUser(db.QueryRow(ctx, `SELECT `+userSelectCols+` FROM users u WHERE u.id=$1::uuid`, id))
}

func findUserByEmail(ctx context.Context, email string) (UserJSON, error) {
	return scanUser(db.QueryRow(ctx, `SELECT `+userSelectCols+` FROM users u WHERE lower(u.email)=lower($1)`, email))
}

type registerBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
	Role     string `json:"role"`
	Phone    string `json:"phone"`
}

type loginBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func handleRegister(c *fiber.Ctx) error {
	var body registerBody
	if err := c.BodyParser(&body); err != nil {
		return fail(c, fiber.StatusBadRequest, "invalid JSON body")
	}
	body.Email = strings.ToLower(strings.TrimSpace(body.Email))
	body.FullName = strings.TrimSpace(body.FullName)
	if body.Email == "" || !strings.Contains(body.Email, "@") {
		return fail(c, fiber.StatusBadRequest, "valid email is required")
	}
	if len(body.Password) < 6 {
		return fail(c, fiber.StatusBadRequest, "password must be at least 6 characters")
	}
	if body.FullName == "" {
		return fail(c, fiber.StatusBadRequest, "full_name is required")
	}
	if body.Role == "" {
		body.Role = "client"
	}
	if body.Role != "client" && body.Role != "chef" {
		return fail(c, fiber.StatusBadRequest, "role must be client or chef")
	}

	ctx := c.Context()
	var exists bool
	if err := db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE lower(email)=lower($1))`, body.Email).Scan(&exists); err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}
	if exists {
		return fail(c, fiber.StatusConflict, "email already registered")
	}

	hash, err := hashPassword(body.Password)
	if err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}

	// Chef accounts are active so they can log in, while their chef profile
	// enters the admin moderation queue with status='pending'.
	tx, err := db.Begin(ctx)
	if err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	u, err := scanUser(tx.QueryRow(ctx, `
		INSERT INTO users(email, password_hash, full_name, role, status, phone)
		VALUES($1, $2, $3, $4, 'active', NULLIF($5, ''))
		RETURNING `+userReturnCols, body.Email, hash, body.FullName, body.Role, body.Phone))
	if err != nil {
		if isUniqueViolation(err) {
			return fail(c, fiber.StatusConflict, "email already registered")
		}
		return failErr(c, fiber.StatusInternalServerError, err)
	}

	if body.Role == "chef" {
		_, err = tx.Exec(ctx, `
			INSERT INTO chefs(user_id, display_name, city, specialty, bio, avatar_url, status)
			VALUES($1, $2, '', '', '', '', 'pending')`, u.ID, body.FullName)
		if err != nil {
			return failErr(c, fiber.StatusInternalServerError, err)
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}

	token, err := generateToken(u)
	if err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}
	return c.JSON(fiber.Map{"token": token, "user": u})
}

func handleLogin(c *fiber.Ctx) error {
	var body loginBody
	if err := c.BodyParser(&body); err != nil {
		return fail(c, fiber.StatusBadRequest, "invalid JSON body")
	}
	body.Email = strings.TrimSpace(body.Email)
	if body.Email == "" || body.Password == "" {
		return fail(c, fiber.StatusBadRequest, "email and password are required")
	}

	u, err := findUserByEmail(c.Context(), body.Email)
	if err != nil {
		return fail(c, fiber.StatusUnauthorized, "invalid email or password")
	}

	var hash string
	if err := db.QueryRow(c.Context(), `SELECT password_hash FROM users WHERE id=$1::uuid`, u.ID).Scan(&hash); err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}
	if !checkPassword(hash, body.Password) {
		return fail(c, fiber.StatusUnauthorized, "invalid email or password")
	}
	if u.Status == "suspended" {
		return fail(c, fiber.StatusForbidden, "account is suspended")
	}

	token, err := generateToken(u)
	if err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}
	return c.JSON(fiber.Map{"token": token, "user": u})
}

func handleMe(c *fiber.Ctx) error {
	u := currentUser(c)
	if u == nil {
		return fail(c, fiber.StatusUnauthorized, "authentication required")
	}
	fresh, err := findUserByID(c.Context(), u.ID)
	if err != nil {
		return fail(c, fiber.StatusUnauthorized, "user not found")
	}
	if fresh.Status == "suspended" {
		return fail(c, fiber.StatusForbidden, "account is suspended")
	}
	return c.JSON(fresh)
}
