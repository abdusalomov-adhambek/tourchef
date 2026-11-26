package main

import (
	"database/sql"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
)

const bookingSelect = `
SELECT b.id::text, b.booking_number, b.user_id::text, b.session_id::text,
       CASE WHEN b.masterclass_id IS NULL THEN NULL ELSE b.masterclass_id::text END,
       b.guests, b.total_price::float8, b.status, b.payment_status, b.created_at, b.cancelled_at,
       COALESCE(b.qr_code_url, ''),
       COALESCE(m.title, ''), COALESCE(m.slug, ''), COALESCE(m.image_url, ''),
       s.starts_at, s.capacity, s.booked, s.price::float8
FROM bookings b
LEFT JOIN masterclasses m ON m.id = b.masterclass_id
LEFT JOIN sessions s ON s.id = b.session_id`

func scanBooking(s scanner) (BookingJSON, error) {
	var b BookingJSON
	var mcID sql.NullString
	var cancelled sql.NullTime
	var mcTitle, mcSlug, mcImage string
	var starts time.Time
	var capacity, booked int
	var sessionPrice float64

	err := s.Scan(&b.ID, &b.BookingNumber, &b.UserID, &b.SessionID, &mcID,
		&b.Guests, &b.TotalPrice, &b.Status, &b.PaymentStatus, &b.CreatedAt, &cancelled,
		&b.QRCodeURL, &mcTitle, &mcSlug, &mcImage,
		&starts, &capacity, &booked, &sessionPrice)
	if err != nil {
		return b, err
	}
	if mcID.Valid {
		b.MasterclassID = strPtr(mcID.String)
		b.Masterclass = map[string]any{
			"id": mcID.String, "title": mcTitle, "slug": mcSlug, "image_url": mcImage,
		}
	}
	if cancelled.Valid {
		t := cancelled.Time
		b.CancelledAt = &t
	}
	b.Session = map[string]any{
		"starts_at": starts, "capacity": capacity, "booked": booked, "price": sessionPrice,
	}
	return b, nil
}

func listBookings(c *fiber.Ctx, f *sqlFilter, order string) error {
	query := bookingSelect + f.where() + order
	rows, err := db.Query(c.Context(), query, f.args...)
	if err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}
	defer rows.Close()
	out := make([]BookingJSON, 0)
	for rows.Next() {
		b, err := scanBooking(rows)
		if err != nil {
			return failErr(c, fiber.StatusInternalServerError, err)
		}
		out = append(out, b)
	}
	if err := rows.Err(); err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}
	return c.JSON(out)
}

// ---------------------------------------------------------------------------
// Admin chefs & users
// ---------------------------------------------------------------------------

func handleAdminChefs(c *fiber.Ctx) error {
	f := &sqlFilter{}
	if city := strings.TrimSpace(c.Query("city")); city != "" {
		f.add(`c.city ILIKE ?`, "%"+city+"%")
	}
	rows, err := db.Query(c.Context(), `SELECT `+chefSelectCols+` FROM chefs c
		LEFT JOIN users u ON u.id = c.user_id`+f.where()+`
		ORDER BY CASE c.status WHEN 'pending' THEN 0 WHEN 'approved' THEN 1 ELSE 2 END, c.created_at DESC`, f.args...)
	if err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}
	defer rows.Close()
	out := make([]ChefJSON, 0)
	for rows.Next() {
		chef, err := scanChef(rows)
		if err != nil {
			return failErr(c, fiber.StatusInternalServerError, err)
		}
		out = append(out, chef)
	}
	return c.JSON(out)
}

type statusBody struct {
	Status string `json:"status"`
}

func handleAdminUpdateChef(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err.Error())
	}
	var body statusBody
	if err := c.BodyParser(&body); err != nil {
		return fail(c, fiber.StatusBadRequest, "invalid JSON body")
	}
	switch body.Status {
	case "pending", "approved", "suspended":
	default:
		return fail(c, fiber.StatusBadRequest, "status must be pending, approved or suspended")
	}
	tag, err := db.Exec(c.Context(), `UPDATE chefs SET status=$1, updated_at=now() WHERE id=$2::uuid`, body.Status, id)
	if err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}
	if tag.RowsAffected() == 0 {
		return fail(c, fiber.StatusNotFound, "chef not found")
	}
	chef, err := scanChef(db.QueryRow(c.Context(), `SELECT `+chefSelectCols+` FROM chefs c
		LEFT JOIN users u ON u.id = c.user_id WHERE c.id=$1::uuid`, id))
	if err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}
	return c.JSON(chef)
}

func handleAdminUsers(c *fiber.Ctx) error {
	f := &sqlFilter{}
	if q := strings.TrimSpace(c.Query("q")); q != "" {
		f.add(`(u.full_name ILIKE ? OR u.email ILIKE ?)`, "%"+q+"%", "%"+q+"%")
	}
	if role := c.Query("role"); role != "" {
		f.add(`u.role = ?`, role)
	}
	if status := c.Query("status"); status != "" {
		f.add(`u.status = ?`, status)
	}
	rows, err := db.Query(c.Context(), `SELECT `+userSelectCols+` FROM users u`+f.where()+` ORDER BY u.created_at DESC`, f.args...)
	if err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}
	defer rows.Close()
	out := make([]UserJSON, 0)
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return failErr(c, fiber.StatusInternalServerError, err)
		}
		out = append(out, u)
	}
	return c.JSON(out)
}

func handleAdminUpdateUser(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err.Error())
	}
	var body statusBody
	if err := c.BodyParser(&body); err != nil {
		return fail(c, fiber.StatusBadRequest, "invalid JSON body")
	}
	switch body.Status {
	case "active", "pending", "suspended":
	default:
		return fail(c, fiber.StatusBadRequest, "status must be active, pending or suspended")
	}
	me := currentUser(c)
	if id == me.ID && body.Status != "active" {
		return fail(c, fiber.StatusBadRequest, "cannot suspend your own account")
	}
	u, err := scanUser(db.QueryRow(c.Context(), `UPDATE users SET status=$1, updated_at=now()
		WHERE id=$2::uuid RETURNING `+userReturnCols, body.Status, id))
	if err == pgx.ErrNoRows {
		return fail(c, fiber.StatusNotFound, "user not found")
	}
	if err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}
	return c.JSON(u)
}

// ---------------------------------------------------------------------------
// Admin masterclasses, sessions, bookings
// ---------------------------------------------------------------------------

func handleAdminMasterclasses(c *fiber.Ctx) error {
	f := &sqlFilter{}
	items, _, err := listMasterclasses(c.Context(), f, ` ORDER BY m.created_at DESC`, 200, 0)
	if err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}
	return c.JSON(items)
}

func handleAdminUpdateMasterclass(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err.Error())
	}
	var body statusBody
	if err := c.BodyParser(&body); err != nil {
		return fail(c, fiber.StatusBadRequest, "invalid JSON body")
	}
	switch body.Status {
	case "draft", "pending", "approved", "suspended":
	default:
		return fail(c, fiber.StatusBadRequest, "invalid status")
	}
	tag, err := db.Exec(c.Context(), `UPDATE masterclasses SET status=$1, updated_at=now()
		WHERE id=$2::uuid`, body.Status, id)
	if err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}
	if tag.RowsAffected() == 0 {
		return fail(c, fiber.StatusNotFound, "masterclass not found")
	}
	mc, err := fetchMasterclassByID(c, id)
	if err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}
	return c.JSON(mc)
}

func handleAdminSessions(c *fiber.Ctx) error {
	return listSessionsBy(c, `WHERE 1=1`)
}

func handleAdminBookings(c *fiber.Ctx) error {
	f := &sqlFilter{}
	if q := strings.TrimSpace(c.Query("q")); q != "" {
		f.add(`b.booking_number ILIKE ?`, "%"+q+"%")
	}
	if status := c.Query("status"); status != "" {
		f.add(`b.status = ?`, status)
	}
	if userID := c.Query("user_id"); userID != "" {
		f.add(`b.user_id::text = ?`, userID)
	}
	return listBookings(c, f, ` ORDER BY b.created_at DESC`)
}

type bookingPatchBody struct {
	Status        *string `json:"status"`
	PaymentStatus *string `json:"payment_status"`
}

func handleAdminUpdateBooking(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err.Error())
	}
	var body bookingPatchBody
	if err := c.BodyParser(&body); err != nil {
		return fail(c, fiber.StatusBadRequest, "invalid JSON body")
	}
	if body.Status == nil && body.PaymentStatus == nil {
		return fail(c, fiber.StatusBadRequest, "no fields to update")
	}
	if body.Status != nil {
		switch *body.Status {
		case "pending", "confirmed", "completed", "cancelled", "refunded":
		default:
			return fail(c, fiber.StatusBadRequest, "invalid status")
		}
	}
	if body.PaymentStatus != nil {
		switch *body.PaymentStatus {
		case "pending", "paid_dev", "paid", "failed", "refunded":
		default:
			return fail(c, fiber.StatusBadRequest, "invalid payment_status")
		}
	}

	sets := make([]string, 0, 2)
	args := make([]any, 0, 3)
	if body.Status != nil {
		sets = append(sets, "status=$"+itoa(len(args)+1))
		args = append(args, *body.Status)
	}
	if body.PaymentStatus != nil {
		sets = append(sets, "payment_status=$"+itoa(len(args)+1))
		args = append(args, *body.PaymentStatus)
	}
	args = append(args, id)
	query := `UPDATE bookings SET ` + strings.Join(sets, ", ") + ` WHERE id=$` + itoa(len(args)) + `::uuid`
	tag, err := db.Exec(c.Context(), query, args...)
	if err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}
	if tag.RowsAffected() == 0 {
		return fail(c, fiber.StatusNotFound, "booking not found")
	}
	b, err := scanBooking(db.QueryRow(c.Context(), bookingSelect+` WHERE b.id=$1::uuid`, id))
	if err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}
	return c.JSON(b)
}

// ---------------------------------------------------------------------------
// Admin stats
// ---------------------------------------------------------------------------

func handleAdminStats(c *fiber.Ctx) error {
	var users, chefs, restaurants, dishes, masterclasses, bookings int
	var revenue float64
	var pendingChefs, pendingMC int
	err := db.QueryRow(c.Context(), `
		SELECT
			(SELECT COUNT(*) FROM users),
			(SELECT COUNT(*) FROM chefs),
			(SELECT COUNT(*) FROM restaurants),
			(SELECT COUNT(*) FROM dishes WHERE status <> 'deleted'),
			(SELECT COUNT(*) FROM masterclasses),
			(SELECT COUNT(*) FROM bookings),
			(SELECT COALESCE(SUM(total_price::float8), 0) FROM bookings
				WHERE payment_status IN ('paid_dev', 'paid') AND status <> 'cancelled'),
			(SELECT COUNT(*) FROM chefs WHERE status='pending'),
			(SELECT COUNT(*) FROM masterclasses WHERE status='pending')`).Scan(
		&users, &chefs, &restaurants, &dishes, &masterclasses, &bookings,
		&revenue, &pendingChefs, &pendingMC)
	if err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}
	return c.JSON(fiber.Map{
		"users": users, "chefs": chefs, "restaurants": restaurants,
		"dishes": dishes, "masterclasses": masterclasses, "bookings": bookings,
		"revenue": revenue, "pending_chefs": pendingChefs, "pending_mc": pendingMC,
	})
}
