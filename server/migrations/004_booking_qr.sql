-- QR-kod bron qilinganda avtomatik yaratiladi va shu yerga saqlanadi.
ALTER TABLE bookings ADD COLUMN IF NOT EXISTS qr_code_url text NOT NULL DEFAULT '';
CREATE INDEX IF NOT EXISTS idx_bookings_qr ON bookings (qr_code_url) WHERE qr_code_url <> '';
