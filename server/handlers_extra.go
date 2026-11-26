package main

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// handleUpdateMe — profil ma'lumotlarini tahrirlash (client/chef/admin).
func handleUpdateMe(c *fiber.Ctx) error {
	u := currentUser(c)
	if u == nil {
		return fail(c, fiber.StatusUnauthorized, "authentication required")
	}
	var body struct {
		FullName *string `json:"full_name"`
		Email    *string `json:"email"`
		Phone    *string `json:"phone"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fail(c, fiber.StatusBadRequest, "invalid JSON body")
	}

	set := []string{}
	args := []any{}
	n := 1
	if body.FullName != nil {
		v := strings.TrimSpace(*body.FullName)
		if v == "" {
			return fail(c, fiber.StatusBadRequest, "full_name cannot be empty")
		}
		set = append(set, fmt.Sprintf("full_name=$%d", n))
		args = append(args, v)
		n++
	}
	if body.Email != nil {
		v := strings.ToLower(strings.TrimSpace(*body.Email))
		if !emailPattern.MatchString(v) {
			return fail(c, fiber.StatusBadRequest, "invalid email address")
		}
		var exists int
		err := db.QueryRow(c.Context(), `SELECT COUNT(*) FROM users WHERE lower(email)=$1 AND id<>$2::uuid`, v, u.ID).Scan(&exists)
		if err != nil {
			return failErr(c, fiber.StatusInternalServerError, err)
		}
		if exists > 0 {
			return fail(c, fiber.StatusConflict, "email already in use")
		}
		set = append(set, fmt.Sprintf("email=$%d", n))
		args = append(args, v)
		n++
	}
	if body.Phone != nil {
		set = append(set, fmt.Sprintf("phone=$%d", n))
		args = append(args, strings.TrimSpace(*body.Phone))
		n++
	}
	if len(set) == 0 {
		return fail(c, fiber.StatusBadRequest, "no fields to update")
	}
	args = append(args, u.ID)
	query := `UPDATE users SET ` + strings.Join(set, ", ") + ` WHERE id=$` + itoa(n) + `::uuid RETURNING ` + userReturnCols
	updated, err := scanUser(db.QueryRow(c.Context(), query, args...))
	if err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}
	return c.JSON(updated)
}

// handleForgotPassword — development rejimi: xat yuborilmaydi, javob har doim muvaffaqiyatli.
func handleForgotPassword(c *fiber.Ctx) error {
	var body struct {
		Email string `json:"email"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fail(c, fiber.StatusBadRequest, "invalid JSON body")
	}
	email := strings.ToLower(strings.TrimSpace(body.Email))
	if !emailPattern.MatchString(email) {
		return fail(c, fiber.StatusBadRequest, "valid email is required")
	}
	var id string
	err := db.QueryRow(c.Context(), `SELECT id::text FROM users WHERE lower(email)=$1`, email).Scan(&id)
	if err == nil {
		fmt.Printf("[dev] password reset requested for %s (user %s)\n", email, id)
	}
	return c.JSON(fiber.Map{
		"ok":      true,
		"message": "If this email is registered, recovery instructions have been sent.",
		"dev":     true,
	})
}

// handleAdminReels — barcha reels (moderatsiya navbati ham).
func handleAdminReels(c *fiber.Ctx) error {
	rows, err := db.Query(c.Context(), reelSelect+` ORDER BY
		CASE rl.status WHEN 'pending' THEN 0 WHEN 'published' THEN 1 ELSE 2 END, rl.created_at DESC`)
	if err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}
	defer rows.Close()
	out := make([]ReelJSON, 0)
	for rows.Next() {
		reel, err := scanReel(rows)
		if err != nil {
			return failErr(c, fiber.StatusInternalServerError, err)
		}
		out = append(out, reel)
	}
	return c.JSON(out)
}

// handleAdminUpdateReel — status o'zgartirish.
func handleAdminUpdateReel(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err.Error())
	}
	var body struct {
		Status string `json:"status"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fail(c, fiber.StatusBadRequest, "invalid JSON body")
	}
	switch body.Status {
	case "published", "rejected", "pending":
	default:
		return fail(c, fiber.StatusBadRequest, "status must be published, rejected or pending")
	}
	rows, err := db.Query(c.Context(), reelSelect+` WHERE rl.id=$1::uuid`, id)
	if err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}
	defer rows.Close()
	if !rows.Next() {
		return fail(c, fiber.StatusNotFound, "reel not found")
	}
	reel, err := scanReel(rows)
	if err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}
	if _, err := db.Exec(c.Context(), `UPDATE reels SET status=$1 WHERE id=$2::uuid`, body.Status, id); err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}
	reel.Status = body.Status
	return c.JSON(reel)
}
