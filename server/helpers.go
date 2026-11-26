package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
var emailPattern = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

// scanner is implemented by both pgx.Row and pgx.Rows.
type scanner interface {
	Scan(dest ...any) error
}

// ---------------------------------------------------------------------------
// Shared JSON shapes (mirrors docs/CONTRACT.md exactly, with a few additive
// convenience fields that are harmless for API consumers).
// ---------------------------------------------------------------------------

type UserJSON struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	FullName  string    `json:"full_name"`
	Role      string    `json:"role"`
	Status    string    `json:"status"`
	Phone     *string   `json:"phone,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type ChefJSON struct {
	ID          string  `json:"id"`
	UserID      string  `json:"user_id"`
	DisplayName string  `json:"display_name"`
	City        string  `json:"city"`
	Specialty   string  `json:"specialty"`
	Bio         string  `json:"bio"`
	AvatarURL   string  `json:"avatar_url"`
	Rating      float64 `json:"rating"`
	Status      string  `json:"status"`
	Email       string  `json:"email,omitempty"`
}

type RestaurantJSON struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	City        string  `json:"city"`
	Cuisine     string  `json:"cuisine"`
	Description string  `json:"description"`
	ImageURL    string  `json:"image_url"`
	Rating      float64 `json:"rating"`
	Status      string  `json:"status"`
}

type DishJSON struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Category    string    `json:"category"`
	Region      string    `json:"region"`
	Description string    `json:"description"`
	Ingredients string    `json:"ingredients"`
	ImageURL    string    `json:"image_url"`
	Rating      float64   `json:"rating"`
	Status      string    `json:"status"`
	CreatedBy   *string   `json:"created_by,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	MultiImage  []string  `json:"multiimage"`
}

type SessionJSON struct {
	ID            string    `json:"id"`
	MasterclassID string    `json:"masterclass_id"`
	StartsAt      time.Time `json:"starts_at"`
	Capacity      int       `json:"capacity"`
	Booked        int       `json:"booked"`
	Available     int       `json:"available"`
	Price         float64   `json:"price"`
	Status        string    `json:"status"`
}

type BookingJSON struct {
	ID            string         `json:"id"`
	BookingNumber string         `json:"booking_number"`
	UserID        string         `json:"user_id"`
	SessionID     string         `json:"session_id"`
	MasterclassID *string        `json:"masterclass_id,omitempty"`
	Guests        int            `json:"guests"`
	TotalPrice    float64        `json:"total_price"`
	Status        string         `json:"status"`
	PaymentStatus string         `json:"payment_status"`
	QRCodeURL     string         `json:"qr_code_url"`
	CreatedAt     time.Time      `json:"created_at"`
	CancelledAt   *time.Time     `json:"cancelled_at,omitempty"`
	Masterclass   map[string]any `json:"masterclass,omitempty"`
	Session       map[string]any `json:"session,omitempty"`
}

type NotificationJSON struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	Type      string    `json:"type"`
	IsRead    bool      `json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
}

type ReelJSON struct {
	ID          string         `json:"id"`
	Title       string         `json:"title"`
	Caption     string         `json:"caption"`
	ImageURL    string         `json:"image_url"`
	VideoURL    string         `json:"video_url"`
	Views       int            `json:"views"`
	Likes       int            `json:"likes"`
	Chef        map[string]any `json:"chef,omitempty"`
	Restaurant  map[string]any `json:"restaurant,omitempty"`
	Masterclass map[string]any `json:"masterclass,omitempty"`
	Status      string         `json:"status"`
	CreatedAt   time.Time      `json:"created_at"`
}

type MasterclassJSON struct {
	ID           string         `json:"id"`
	Title        string         `json:"title"`
	Slug         string         `json:"slug"`
	Description  string         `json:"description"`
	DishID       *string        `json:"dish_id,omitempty"`
	DishName     string         `json:"dish_name,omitempty"`
	Dish         map[string]any `json:"dish,omitempty"`
	ChefID       *string        `json:"chef_id,omitempty"`
	ChefName     string         `json:"chef_name,omitempty"`
	Chef         map[string]any `json:"chef,omitempty"`
	RestaurantID *string        `json:"restaurant_id,omitempty"`
	RestName     string         `json:"restaurant_name,omitempty"`
	Restaurant   map[string]any `json:"restaurant,omitempty"`
	City         string         `json:"city"`
	DurationMin  int            `json:"duration_min"`
	Price        float64        `json:"price"`
	Difficulty   string         `json:"difficulty"`
	Language     string         `json:"language"`
	Rating       float64        `json:"rating"`
	ReviewsCount int            `json:"reviews_count"`
	ImageURL     string         `json:"image_url"`
	Status       string         `json:"status"`
	CreatedAt    time.Time      `json:"created_at"`
	Sessions     []SessionJSON  `json:"sessions,omitempty"`
}

// ---------------------------------------------------------------------------
// Scan helpers
// ---------------------------------------------------------------------------

func scanUser(s scanner) (UserJSON, error) {
	var u UserJSON
	var phone sql.NullString
	err := s.Scan(&u.ID, &u.Email, &u.FullName, &u.Role, &u.Status, &phone, &u.CreatedAt)
	if err != nil {
		return u, err
	}
	if phone.Valid {
		v := phone.String
		u.Phone = &v
	}
	return u, nil
}

func scanChef(s scanner) (ChefJSON, error) {
	var c ChefJSON
	err := s.Scan(&c.ID, &c.UserID, &c.DisplayName, &c.City, &c.Specialty, &c.Bio,
		&c.AvatarURL, &c.Rating, &c.Status, &c.Email)
	if err != nil {
		return c, err
	}
	return c, nil
}

const chefSelectCols = `c.id::text, c.user_id::text, c.display_name, c.city, c.specialty,
	c.bio, c.avatar_url, c.rating::float8, c.status, COALESCE(u.email, '')`

// Same columns without a table alias, for UPDATE ... RETURNING.
const chefReturnCols = `id::text, user_id::text, display_name, city, specialty,
	bio, avatar_url, rating::float8, status`

func scanRestaurant(s scanner) (RestaurantJSON, error) {
	var r RestaurantJSON
	err := s.Scan(&r.ID, &r.Name, &r.Slug, &r.City, &r.Cuisine, &r.Description,
		&r.ImageURL, &r.Rating, &r.Status)
	if err != nil {
		return r, err
	}
	return r, nil
}

const restaurantSelectCols = `r.id::text, r.name, r.slug, r.city, r.cuisine, r.description,
	r.image_url, r.rating::float8, r.status`

func scanDish(s scanner) (DishJSON, error) {
	var d DishJSON
	var createdBy sql.NullString
	var multiRaw string
	err := s.Scan(&d.ID, &d.Name, &d.Slug, &d.Category, &d.Region, &d.Description,
		&d.Ingredients, &d.ImageURL, &d.Rating, &d.Status, &createdBy, &d.CreatedAt, &multiRaw)
	if err != nil {
		return d, err
	}
	if createdBy.Valid {
		v := createdBy.String
		d.CreatedBy = &v
	}
	d.MultiImage = []string{}
	if multiRaw != "" {
		_ = json.Unmarshal([]byte(multiRaw), &d.MultiImage)
	}
	return d, nil
}

const dishSelectCols = `d.id::text, d.name, d.slug, d.category, d.region, d.description,
	d.ingredients, d.image_url, d.rating::float8, d.status,
	CASE WHEN d.created_by IS NULL THEN NULL ELSE d.created_by::text END, d.created_at,
	COALESCE(d.multiimage, '[]'::jsonb)::text`

func scanSession(s scanner) (SessionJSON, error) {
	var sess SessionJSON
	err := s.Scan(&sess.ID, &sess.MasterclassID, &sess.StartsAt, &sess.Capacity,
		&sess.Booked, &sess.Price, &sess.Status)
	if err != nil {
		return sess, err
	}
	sess.Available = sess.Capacity - sess.Booked
	return sess, nil
}

const sessionSelectCols = `s.id::text, s.masterclass_id::text, s.starts_at, s.capacity,
	s.booked, s.price::float8, s.status`

func scanNotification(s scanner) (NotificationJSON, error) {
	var n NotificationJSON
	err := s.Scan(&n.ID, &n.UserID, &n.Title, &n.Body, &n.Type, &n.IsRead, &n.CreatedAt)
	if err != nil {
		return n, err
	}
	return n, nil
}

// ---------------------------------------------------------------------------
// HTTP helpers
// ---------------------------------------------------------------------------

func fail(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(fiber.Map{"error": message})
}

func failErr(c *fiber.Ctx, status int, err error) error {
	return fail(c, status, err.Error())
}

func parseID(c *fiber.Ctx, name string) (string, error) {
	id := c.Params(name)
	if !uuidPattern.MatchString(id) {
		return "", fmt.Errorf("invalid %s", name)
	}
	return id, nil
}

func queryInt(c *fiber.Ctx, key string, def int) int {
	raw := c.Query(key)
	if raw == "" {
		return def
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return def
	}
	return n
}

func queryFloat(c *fiber.Ctx, key string, def float64) float64 {
	raw := c.Query(key)
	if raw == "" {
		return def
	}
	f, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return def
	}
	return f
}

// slugify creates a URL-safe slug while preserving unicode letters.
func slugify(input string) string {
	var b strings.Builder
	dash := false
	for _, r := range strings.ToLower(strings.TrimSpace(input)) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			dash = false
		} else if !dash {
			b.WriteByte('-')
			dash = true
		}
	}
	return strings.Trim(b.String(), "-")
}

func randomHex(bytes int) string {
	buf := make([]byte, bytes)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(buf)
}

// uniqueSlug guarantees a slug is unique within the given table.
func uniqueSlug(ctx context.Context, pool *pgxpool.Pool, table, base string) string {
	slug := slugify(base)
	if slug == "" {
		slug = "item"
	}
	for i := 0; i < 5; i++ {
		var exists bool
		err := pool.QueryRow(ctx, fmt.Sprintf(`SELECT EXISTS(SELECT 1 FROM %s WHERE slug=$1)`, table), slug).Scan(&exists)
		if err != nil {
			return slug + "-" + randomHex(4)
		}
		if !exists {
			return slug
		}
		slug = fmt.Sprintf("%s-%s", slugify(base), randomHex(4))
	}
	return slug + "-" + randomHex(8)
}

// sqlFilter builds parameterized WHERE clauses. Each "?" in a clause is
// replaced by the next sequential $N placeholder, so user input is never
// interpolated into SQL text.
type sqlFilter struct {
	clauses []string
	args    []any
}

func (f *sqlFilter) add(clause string, args ...any) {
	for _, arg := range args {
		clause = strings.Replace(clause, "?", fmt.Sprintf("$%d", len(f.args)+1), 1)
		f.args = append(f.args, arg)
	}
	f.clauses = append(f.clauses, clause)
}

func (f *sqlFilter) where() string {
	if len(f.clauses) == 0 {
		return ""
	}
	return " WHERE " + strings.Join(f.clauses, " AND ")
}

func strPtr(s string) *string { return &s }

// isUniqueViolation detects Postgres 23505 errors.
func isUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "23505")
}
