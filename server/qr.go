package main

import (
	"image/color"
	"os"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/skip2/go-qrcode"
)

// TourChef accent (#C65D3A) — QR chipta rangi.
var qrForeground = color.RGBA{R: 0xC6, G: 0x5D, B: 0x3A, A: 0xFF}
var qrBackground = color.RGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}

// generateBookingQR — bron uchun QR tasvirini yaratadi va /uploads ga saqlaydi.
// QR ichida foydalanuvchi chiptasi ochiladigan havola yoziladi.
func generateBookingQR(bookingID, bookingNumber string) (string, error) {
	_ = bookingNumber
	content := cfg.SiteURL + "/client/bookings/" + bookingID
	q, err := qrcode.New(content, qrcode.Medium)
	if err != nil {
		return "", err
	}
	q.ForegroundColor = qrForeground
	q.BackgroundColor = qrBackground
	png, err := q.PNG(512)
	if err != nil {
		return "", err
	}
	name := "qr-" + strings.ReplaceAll(bookingID, "-", "") + ".png"
	if err := os.MkdirAll(cfg.UploadDir, 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(filepath.Join(cfg.UploadDir, name), png, 0o644); err != nil {
		return "", err
	}
	return "/uploads/" + name, nil
}

// handleBookingQR — eski bronlar uchun QR ni yaratib/qaytarib beradi.
func handleBookingQR(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return fail(c, fiber.StatusBadRequest, err.Error())
	}
	var number, url string
	err = db.QueryRow(c.Context(), `SELECT booking_number, qr_code_url FROM bookings
		WHERE id=$1::uuid AND user_id=$2::uuid`, id, currentUser(c).ID).Scan(&number, &url)
	if err == pgx.ErrNoRows {
		return fail(c, fiber.StatusNotFound, "booking not found")
	}
	if err != nil {
		return failErr(c, fiber.StatusInternalServerError, err)
	}
	if url == "" || c.Query("refresh") == "1" {
		url, err = generateBookingQR(id, number)
		if err != nil {
			return failErr(c, fiber.StatusInternalServerError, err)
		}
		if _, err = db.Exec(c.Context(), `UPDATE bookings SET qr_code_url=$1 WHERE id=$2::uuid`, url, id); err != nil {
			return failErr(c, fiber.StatusInternalServerError, err)
		}
	}
	return c.JSON(fiber.Map{"url": url, "booking_number": number})
}
