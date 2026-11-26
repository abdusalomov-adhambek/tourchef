package main

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

// handleAdminMasterclassDetail — admin masterclass sahifasi uchun to'liq ma'lumot:
// sessionlar, bron qilganlar va sig'im statistikasi.
func handleAdminMasterclassDetail(c *fiber.Ctx) error {
	key := c.Params("slug")
	where := `m.slug=$1`
	if uuidPattern.MatchString(key) {
		where = `m.id=$1::uuid`
	}
	mc, err := scanMasterclass(db.QueryRow(c.Context(), mcSelect+` WHERE `+where, key))
	if err != nil {
		return fail(c, fiber.StatusNotFound, "masterclass not found")
	}

	sessionRows, err := db.Query(c.Context(), `
		SELECT s.id::text, s.starts_at, s.capacity, s.booked, s.price::float8, s.status
		FROM sessions s WHERE s.masterclass_id=$1::uuid ORDER BY s.starts_at`, mc.ID)
	if err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}
	sessions := make([]fiber.Map, 0)
	var totalCapacity, totalBooked int
	for sessionRows.Next() {
		var id, status string
		var starts time.Time
		var capacity, booked int
		var price float64
		if err := sessionRows.Scan(&id, &starts, &capacity, &booked, &price, &status); err != nil {
			sessionRows.Close()
			return failErr(c, fiber.StatusInternalServerError, err)
		}
		totalCapacity += capacity
		totalBooked += booked
		sessions = append(sessions, fiber.Map{
			"id": id, "starts_at": starts, "capacity": capacity, "booked": booked,
			"price": price, "status": status,
		})
	}
	sessionRows.Close()

	bookingRows, err := db.Query(c.Context(), `
		SELECT b.id::text, b.booking_number, b.guests, b.status, b.payment_status, b.created_at,
		       s.starts_at, COALESCE(u.full_name, ''), COALESCE(u.email, ''), b.user_id::text
		FROM bookings b
		JOIN sessions s ON s.id = b.session_id
		LEFT JOIN users u ON u.id = b.user_id
		WHERE b.masterclass_id=$1::uuid
		ORDER BY b.created_at DESC`, mc.ID)
	if err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}
	bookings := make([]fiber.Map, 0)
	var activeGuests int
	for bookingRows.Next() {
		var id, number, status, payment, name, email, userID string
		var guests int
		var created, starts time.Time
		if err := bookingRows.Scan(&id, &number, &guests, &status, &payment, &created,
			&starts, &name, &email, &userID); err != nil {
			bookingRows.Close()
			return failErr(c, fiber.StatusInternalServerError, err)
		}
		if status != "cancelled" && status != "refunded" {
			activeGuests += guests
		}
		bookings = append(bookings, fiber.Map{
			"id": id, "booking_number": number, "guests": guests,
			"status": status, "payment_status": payment, "created_at": created,
			"starts_at": starts, "guest_name": name, "guest_email": email, "user_id": userID,
		})
	}
	bookingRows.Close()

	seatsLeft := totalCapacity - totalBooked
	if seatsLeft < 0 {
		seatsLeft = 0
	}
	return c.JSON(fiber.Map{
		"masterclass": mc,
		"sessions":    sessions,
		"bookings":    bookings,
		"stats": fiber.Map{
			"capacity": totalCapacity, "booked": totalBooked, "seats_left": seatsLeft,
			"active_guests": activeGuests, "bookings_count": len(bookings),
		},
	})
}
