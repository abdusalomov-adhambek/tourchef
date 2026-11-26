-- Dish gallery: bir nechta rasm. Eski image_url birinchi rasm sifatida saqlanadi.
ALTER TABLE dishes ADD COLUMN IF NOT EXISTS multiimage jsonb NOT NULL DEFAULT '[]'::jsonb;

UPDATE dishes
SET multiimage = jsonb_build_array(image_url)
WHERE image_url <> '' AND jsonb_array_length(multiimage) = 0;

CREATE INDEX IF NOT EXISTS idx_dishes_multiimage ON dishes USING gin (multiimage);
