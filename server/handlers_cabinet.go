package main

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
)

func notify(ctx context.Context, userID, title, body, kind string) {
	_, _ = db.Exec(ctx, `INSERT INTO notifications(user_id, title, body, type)
		VALUES($1::uuid, $2, $3, $4)`, userID, title, body, kind)
}

type bookingCreateBody struct {
	SessionID string `json:"session_id"`
	Guests    int    `json:"guests"`
}

// handleCreateBooking reserves seats inside a single transaction with a
// row-level lock on the session, which prevents overbooking under concurrency.
func handleCreateBooking(c *fiber.Ctx) error {
	var body bookingCreateBody
	if err := c.BodyParser(&body); err != nil {
		return fail(c, fiber.StatusBadRequest, "invalid JSON body")
	}
	if !uuidPattern.MatchString(body.SessionID) {
		return fail(c, fiber.StatusBadRequest, "valid session_id is required")
	}
	if body.Guests < 1 {
		return fail(c, fiber.StatusBadRequest, "guests must be at least 1")
	}

	ctx := c.Context()
	user := currentUser(c)
	tx, err := db.Begin(ctx)
	if err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var capacity, booked int
	var price float64
	var sessionStatus string
	var mcID sql.NullString
	err = tx.QueryRow(ctx, `
		SELECT s.capacity, s.booked, s.price::float8, s.status,
		       CASE WHEN s.masterclass_id IS NULL THEN NULL ELSE s.masterclass_id::text END
		FROM sessions s WHERE s.id=$1::uuid FOR UPDATE`, body.SessionID).Scan(
		&capacity, &booked, &price, &sessionStatus, &mcID)
	if err == pgx.ErrNoRows {
		return fail(c, fiber.StatusNotFound, "session not found")
	}
	if err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}
	if sessionStatus == "cancelled" || sessionStatus == "completed" {
		return fail(c, fiber.StatusConflict, "session is not bookable")
	}
	if booked+body.Guests > capacity {
		return fail(c, fiber.StatusConflict, "not enough available seats")
	}

	number := "TC-" + time.Now().UTC().Format("20060102") + "-" + strings.ToUpper(randomHex(4))
	total := price * float64(body.Guests)
	var bookingID string
	err = tx.QueryRow(ctx, `
		INSERT INTO bookings(booking_number, user_id, session_id, masterclass_id, guests,
			total_price, status, payment_status)
		VALUES($1, $2::uuid, $3::uuid, $4::uuid, $5, $6, 'confirmed', 'paid_dev')
		RETURNING id::text`, number, user.ID, body.SessionID, mcID, body.Guests, total).Scan(&bookingID)
	if err != nil {
		if isUniqueViolation(err) {
			return fail(c, fiber.StatusConflict, "duplicate booking number, please retry")
		}
		return failErr(c, fiber.StatusInternalServerError, err)
	}

	// Har bir bron uchun QR chipta avtomatik yaratiladi.
	qrURL, err := generateBookingQR(bookingID, number)
	if err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}
	if _, err = tx.Exec(ctx, `UPDATE bookings SET qr_code_url=$1 WHERE id=$2::uuid`, qrURL, bookingID); err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}

	newBooked := booked + body.Guests
	newStatus := sessionStatus
	if newBooked >= capacity {
		newStatus = "full"
	}
	if _, err = tx.Exec(ctx, `UPDATE sessions SET booked=$1, status=$2 WHERE id=$3::uuid`,
		newBooked, newStatus, body.SessionID); err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}

	title := "Booking confirmed"
	text := "Your booking " + number + " is confirmed. Payment status: paid_dev (development mode)."
	if _, err = tx.Exec(ctx, `INSERT INTO notifications(user_id, title, body, type)
		VALUES($1::uuid, $2, $3, 'booking')`, user.ID, title, text); err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}
	if err = tx.Commit(ctx); err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}

	b, err := scanBooking(db.QueryRow(ctx, bookingSelect+` WHERE b.id=$1::uuid`, bookingID))
	if err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}
	return c.Status(fiber.StatusCreated).JSON(b)
}

func handleClientBookings(c *fiber.Ctx) error {
	f := &sqlFilter{}
	f.add(`b.user_id = ?`, currentUser(c).ID)
	if status := c.Query("status"); status != "" {
		f.add(`b.status = ?`, status)
	}
	return listBookings(c, f, ` ORDER BY b.created_at DESC`)
}

func handleCancelBooking(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err.Error())
	}
	ctx := c.Context()
	user := currentUser(c)
	tx, err := db.Begin(ctx)
	if err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var status string
	var guests int
	var sessionID string
	err = tx.QueryRow(ctx, `SELECT status, guests, session_id::text
		FROM bookings WHERE id=$1::uuid AND user_id=$2::uuid FOR UPDATE`,
		id, user.ID).Scan(&status, &guests, &sessionID)
	if err == pgx.ErrNoRows {
		return fail(c, fiber.StatusNotFound, "booking not found")
	}
	if err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}
	if status == "cancelled" {
		return fail(c, fiber.StatusConflict, "booking is already cancelled")
	}
	if status == "completed" || status == "refunded" {
		return fail(c, fiber.StatusConflict, "booking can no longer be cancelled")
	}

	if _, err = tx.Exec(ctx, `UPDATE bookings SET status='cancelled', cancelled_at=now(),
		payment_status='refunded' WHERE id=$1::uuid`, id); err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}
	if _, err = tx.Exec(ctx, `UPDATE sessions SET booked=GREATEST(0, booked-$1),
		status=CASE WHEN booked-$1 < capacity THEN 'scheduled' ELSE status END
		WHERE id=$2::uuid`, guests, sessionID); err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}
	if _, err = tx.Exec(ctx, `INSERT INTO notifications(user_id, title, body, type)
		VALUES($1::uuid, 'Booking cancelled', 'Your booking has been cancelled. Refund status: refunded.', 'booking')`,
		user.ID); err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}
	if err = tx.Commit(ctx); err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}
	b, err := scanBooking(db.QueryRow(ctx, bookingSelect+` WHERE b.id=$1::uuid`, id))
	if err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}
	return c.JSON(b)
}

// ---------------------------------------------------------------------------
// Favorites
// ---------------------------------------------------------------------------

func handleClientFavorites(c *fiber.Ctx) error {
	rows, err := db.Query(c.Context(), `SELECT f.id::text, f.masterclass_id::text, f.created_at
		FROM favorites f WHERE f.user_id=$1::uuid ORDER BY f.created_at DESC`, currentUser(c).ID)
	if err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}
	defer rows.Close()
	type favRow struct {
		id, mcID string
		created  time.Time
	}
	collected := make([]favRow, 0)
	for rows.Next() {
		var row favRow
		if err := rows.Scan(&row.id, &row.mcID, &row.created); err != nil {
			return failErr(c, fiber.StatusInternalServerError, err)
		}
		collected = append(collected, row)
	}
	if err := rows.Err(); err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}

	out := make([]fiber.Map, 0, len(collected))
	for _, row := range collected {
		mc, err := fetchMasterclassByID(c, row.mcID)
		if err != nil {
			continue // masterclass was removed; skip stale favorite
		}
		out = append(out, fiber.Map{
			"id": row.id, "masterclass_id": row.mcID,
			"created_at": row.created, "masterclass": mc,
		})
	}
	return c.JSON(out)
}

type favoriteBody struct {
	MasterclassID string `json:"masterclass_id"`
}

func handleAddFavorite(c *fiber.Ctx) error {
	var body favoriteBody
	if err := c.BodyParser(&body); err != nil {
		return fail(c, fiber.StatusBadRequest, "invalid JSON body")
	}
	if !uuidPattern.MatchString(body.MasterclassID) {
		return fail(c, fiber.StatusBadRequest, "valid masterclass_id is required")
	}
	user := currentUser(c)

	var id string
	var created time.Time
	err := db.QueryRow(c.Context(), `SELECT id::text, created_at FROM favorites
		WHERE user_id=$1::uuid AND masterclass_id=$2::uuid`, user.ID, body.MasterclassID).Scan(&id, &created)
	if err == pgx.ErrNoRows {
		err = db.QueryRow(c.Context(), `INSERT INTO favorites(user_id, masterclass_id)
			VALUES($1::uuid, $2::uuid) ON CONFLICT (user_id, masterclass_id) DO NOTHING
			RETURNING id::text, created_at`, user.ID, body.MasterclassID).Scan(&id, &created)
		if err == pgx.ErrNoRows {
			// Another request won the race; fetch the existing row.
			err = db.QueryRow(c.Context(), `SELECT id::text, created_at FROM favorites
				WHERE user_id=$1::uuid AND masterclass_id=$2::uuid`, user.ID, body.MasterclassID).Scan(&id, &created)
		}
	}
	if err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}
	mc, err := fetchMasterclassByID(c, body.MasterclassID)
	if err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id": id, "masterclass_id": body.MasterclassID,
		"created_at": created, "masterclass": mc,
	})
}

func handleDeleteFavorite(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err.Error())
	}
	tag, err := db.Exec(c.Context(), `DELETE FROM favorites WHERE id=$1::uuid AND user_id=$2::uuid`,
		id, currentUser(c).ID)
	if err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}
	if tag.RowsAffected() == 0 {
		return fail(c, fiber.StatusNotFound, "favorite not found")
	}
	return c.JSON(fiber.Map{"ok": true})
}

// ---------------------------------------------------------------------------
// Reviews & notifications
// ---------------------------------------------------------------------------

type reviewBody struct {
	MasterclassID string `json:"masterclass_id"`
	Rating        int    `json:"rating"`
	Comment       string `json:"comment"`
}

func handleCreateReview(c *fiber.Ctx) error {
	var body reviewBody
	if err := c.BodyParser(&body); err != nil {
		return fail(c, fiber.StatusBadRequest, "invalid JSON body")
	}
	if !uuidPattern.MatchString(body.MasterclassID) {
		return fail(c, fiber.StatusBadRequest, "valid masterclass_id is required")
	}
	if body.Rating < 1 || body.Rating > 5 {
		return fail(c, fiber.StatusBadRequest, "rating must be between 1 and 5")
	}
	user := currentUser(c)
	var exists bool
	if err := db.QueryRow(c.Context(), `SELECT EXISTS(SELECT 1 FROM masterclasses WHERE id=$1::uuid)`,
		body.MasterclassID).Scan(&exists); err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}
	if !exists {
		return fail(c, fiber.StatusNotFound, "masterclass not found")
	}

	var id string
	err := db.QueryRow(c.Context(), `
		INSERT INTO reviews(user_id, masterclass_id, rating, comment)
		VALUES($1::uuid, $2::uuid, $3, $4)
		ON CONFLICT (user_id, masterclass_id)
		DO UPDATE SET rating=EXCLUDED.rating, comment=EXCLUDED.comment, updated_at=now()
		RETURNING id::text`, user.ID, body.MasterclassID, body.Rating, body.Comment).Scan(&id)
	if err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id": id, "masterclass_id": body.MasterclassID, "user_id": user.ID,
		"rating": body.Rating, "comment": body.Comment,
	})
}

func handleClientNotifications(c *fiber.Ctx) error {
	rows, err := db.Query(c.Context(), `SELECT id::text, user_id::text, title, body, type,
		is_read, created_at FROM notifications WHERE user_id=$1::uuid
		ORDER BY created_at DESC`, currentUser(c).ID)
	if err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}
	defer rows.Close()
	out := make([]NotificationJSON, 0)
	for rows.Next() {
		n, err := scanNotification(rows)
		if err != nil {
			return failErr(c, fiber.StatusInternalServerError, err)
		}
		out = append(out, n)
	}
	return c.JSON(out)
}
