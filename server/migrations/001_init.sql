-- TourChef initial schema. Safe to run once; tracked by schema_migrations.
-- gen_random_uuid() is built into PostgreSQL 13+, no extension required.

CREATE TABLE IF NOT EXISTS users (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email         text NOT NULL UNIQUE,
    password_hash text NOT NULL,
    full_name     text NOT NULL,
    role          text NOT NULL DEFAULT 'client' CHECK (role IN ('client', 'chef', 'admin')),
    status        text NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'pending', 'suspended')),
    phone         text,
    avatar_url    text NOT NULL DEFAULT '',
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS chefs (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      uuid NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    display_name text NOT NULL,
    city         text NOT NULL DEFAULT '',
    specialty    text NOT NULL DEFAULT '',
    bio          text NOT NULL DEFAULT '',
    avatar_url   text NOT NULL DEFAULT '',
    rating       numeric(2,1) NOT NULL DEFAULT 0,
    status       text NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'suspended')),
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS restaurants (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name        text NOT NULL,
    slug        text NOT NULL UNIQUE,
    city        text NOT NULL DEFAULT '',
    cuisine     text NOT NULL DEFAULT '',
    description text NOT NULL DEFAULT '',
    image_url   text NOT NULL DEFAULT '',
    rating      numeric(2,1) NOT NULL DEFAULT 0,
    status      text NOT NULL DEFAULT 'approved' CHECK (status IN ('pending', 'approved', 'suspended')),
    phone       text NOT NULL DEFAULT '',
    email       text NOT NULL DEFAULT '',
    address     text NOT NULL DEFAULT '',
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS dishes (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name        text NOT NULL,
    slug        text NOT NULL UNIQUE,
    category    text NOT NULL DEFAULT '',
    region      text NOT NULL DEFAULT '',
    description text NOT NULL DEFAULT '',
    ingredients text NOT NULL DEFAULT '',
    image_url   text NOT NULL DEFAULT '',
    rating      numeric(2,1) NOT NULL DEFAULT 0,
    status      text NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'hidden', 'deleted')),
    created_by  uuid REFERENCES users(id) ON DELETE SET NULL,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS masterclasses (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    title         text NOT NULL,
    slug          text NOT NULL UNIQUE,
    description   text NOT NULL DEFAULT '',
    dish_id       uuid REFERENCES dishes(id) ON DELETE SET NULL,
    chef_id       uuid REFERENCES chefs(id) ON DELETE SET NULL,
    restaurant_id uuid REFERENCES restaurants(id) ON DELETE SET NULL,
    city          text NOT NULL DEFAULT '',
    duration_min  integer NOT NULL DEFAULT 120 CHECK (duration_min > 0),
    price         numeric(10,2) NOT NULL DEFAULT 0,
    difficulty    text NOT NULL DEFAULT 'beginner',
    language      text NOT NULL DEFAULT 'uz',
    rating        numeric(2,1) NOT NULL DEFAULT 0,
    image_url     text NOT NULL DEFAULT '',
    status        text NOT NULL DEFAULT 'pending' CHECK (status IN ('draft', 'pending', 'approved', 'suspended')),
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS sessions (
    id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    masterclass_id uuid NOT NULL REFERENCES masterclasses(id) ON DELETE CASCADE,
    starts_at      timestamptz NOT NULL,
    capacity       integer NOT NULL DEFAULT 10 CHECK (capacity > 0),
    booked         integer NOT NULL DEFAULT 0 CHECK (booked >= 0),
    price          numeric(10,2) NOT NULL DEFAULT 0,
    status         text NOT NULL DEFAULT 'scheduled' CHECK (status IN ('scheduled', 'full', 'completed', 'cancelled')),
    created_at     timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS bookings (
    id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    booking_number text NOT NULL UNIQUE,
    user_id        uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    session_id     uuid NOT NULL REFERENCES sessions(id) ON DELETE RESTRICT,
    masterclass_id uuid REFERENCES masterclasses(id) ON DELETE SET NULL,
    guests         integer NOT NULL CHECK (guests > 0),
    total_price    numeric(10,2) NOT NULL DEFAULT 0,
    status         text NOT NULL DEFAULT 'confirmed' CHECK (status IN ('pending', 'confirmed', 'completed', 'cancelled', 'refunded')),
    payment_status text NOT NULL DEFAULT 'paid_dev' CHECK (payment_status IN ('pending', 'paid_dev', 'paid', 'failed', 'refunded')),
    created_at     timestamptz NOT NULL DEFAULT now(),
    cancelled_at   timestamptz
);

CREATE TABLE IF NOT EXISTS favorites (
    id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id        uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    masterclass_id uuid NOT NULL REFERENCES masterclasses(id) ON DELETE CASCADE,
    created_at     timestamptz NOT NULL DEFAULT now(),
    UNIQUE (user_id, masterclass_id)
);

CREATE TABLE IF NOT EXISTS reviews (
    id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id        uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    masterclass_id uuid NOT NULL REFERENCES masterclasses(id) ON DELETE CASCADE,
    rating         integer NOT NULL CHECK (rating BETWEEN 1 AND 5),
    comment        text NOT NULL DEFAULT '',
    created_at     timestamptz NOT NULL DEFAULT now(),
    updated_at     timestamptz NOT NULL DEFAULT now(),
    UNIQUE (user_id, masterclass_id)
);

CREATE TABLE IF NOT EXISTS notifications (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title      text NOT NULL,
    body       text NOT NULL DEFAULT '',
    type       text NOT NULL DEFAULT 'info',
    is_read    boolean NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS reels (
    id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    title          text NOT NULL DEFAULT '',
    caption        text NOT NULL DEFAULT '',
    image_url      text NOT NULL DEFAULT '',
    video_url      text NOT NULL DEFAULT '',
    views          integer NOT NULL DEFAULT 0,
    likes          integer NOT NULL DEFAULT 0,
    chef_id        uuid REFERENCES chefs(id) ON DELETE SET NULL,
    restaurant_id  uuid REFERENCES restaurants(id) ON DELETE SET NULL,
    masterclass_id uuid REFERENCES masterclasses(id) ON DELETE SET NULL,
    status         text NOT NULL DEFAULT 'published' CHECK (status IN ('draft', 'pending', 'published', 'hidden', 'rejected')),
    created_at     timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS audit_logs (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_id   uuid REFERENCES users(id) ON DELETE SET NULL,
    action     text NOT NULL,
    entity     text NOT NULL DEFAULT '',
    entity_id  text NOT NULL DEFAULT '',
    details    jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS platform_settings (
    key        text PRIMARY KEY,
    value      jsonb NOT NULL,
    updated_at timestamptz NOT NULL DEFAULT now()
);

-- Indexes for filtering / sorting paths used by the public API.
CREATE INDEX IF NOT EXISTS idx_users_role_status ON users (role, status);
CREATE INDEX IF NOT EXISTS idx_chefs_status ON chefs (status);
CREATE INDEX IF NOT EXISTS idx_restaurants_city ON restaurants (city);
CREATE INDEX IF NOT EXISTS idx_dishes_category ON dishes (category);
CREATE INDEX IF NOT EXISTS idx_dishes_region ON dishes (region);
CREATE INDEX IF NOT EXISTS idx_dishes_status ON dishes (status);
CREATE INDEX IF NOT EXISTS idx_mc_status_city ON masterclasses (status, city);
CREATE INDEX IF NOT EXISTS idx_mc_dish ON masterclasses (dish_id);
CREATE INDEX IF NOT EXISTS idx_mc_chef ON masterclasses (chef_id);
CREATE INDEX IF NOT EXISTS idx_mc_price ON masterclasses (price);
CREATE INDEX IF NOT EXISTS idx_mc_created ON masterclasses (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_sessions_mc ON sessions (masterclass_id);
CREATE INDEX IF NOT EXISTS idx_sessions_starts ON sessions (starts_at);
CREATE INDEX IF NOT EXISTS idx_bookings_user ON bookings (user_id);
CREATE INDEX IF NOT EXISTS idx_bookings_session ON bookings (session_id);
CREATE INDEX IF NOT EXISTS idx_favorites_user ON favorites (user_id);
CREATE INDEX IF NOT EXISTS idx_reviews_mc ON reviews (masterclass_id);
CREATE INDEX IF NOT EXISTS idx_notifications_user ON notifications (user_id, is_read);
CREATE INDEX IF NOT EXISTS idx_reels_status ON reels (status);
