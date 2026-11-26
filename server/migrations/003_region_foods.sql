-- "Taste of Uzbekistan" interaktiv xarita uchun mintaqa-taom katalogi.
CREATE TABLE IF NOT EXISTS region_foods (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    slug          text NOT NULL UNIQUE,
    geo_id        text NOT NULL UNIQUE,
    name          text NOT NULL,
    name_uz       text NOT NULL DEFAULT '',
    city          text NOT NULL DEFAULT '',
    dish_name     text NOT NULL,
    dish_slug     text NOT NULL DEFAULT '',
    dish_image_url text NOT NULL DEFAULT '',
    description   text NOT NULL DEFAULT '',
    ingredients   jsonb NOT NULL DEFAULT '[]'::jsonb,
    cultural_info text NOT NULL DEFAULT '',
    latitude      double precision NOT NULL DEFAULT 0,
    longitude     double precision NOT NULL DEFAULT 0,
    status        text NOT NULL DEFAULT 'published',
    created_by    uuid,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_region_foods_status ON region_foods (status);
CREATE INDEX IF NOT EXISTS idx_region_foods_city ON region_foods (city);
