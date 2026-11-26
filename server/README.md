# TourChef API (Go + Fiber + PostgreSQL)

Backend implementing `docs/CONTRACT.md`. All endpoints query PostgreSQL through
pgx with parameterized SQL. Migrations are embedded in the binary and applied
idempotently on startup.

## Quick start

```bash
cd server
cp .env.example .env
# create a database named "tourchef" (or change DATABASE_URL)
go mod tidy
go run .
```

Docker:

```bash
docker build -t tourchef-api ./server
docker run --rm -p 8080:8080 \
  -e DATABASE_URL='postgres://postgres:postgres@host.docker.internal:5432/tourchef?sslmode=disable' \
  -e JWT_SECRET='your-long-random-secret' \
  tourchef-api
```

Health check: `GET http://localhost:8080/health` → `{"status":"ok"}`.
All API routes are prefixed with `/api`.

## Environment variables

| Variable | Default | Description |
|---|---|---|
| `APP_ENV` | `dev` | Environment label (logging only) |
| `PORT` | `8080` | HTTP listen port |
| `DATABASE_URL` | `postgres://postgres:postgres@localhost:5432/tourchef?sslmode=disable` | PostgreSQL DSN |
| `JWT_SECRET` | `tourchef-dev-secret-change-me` | HMAC secret for JWTs. **Set in production.** |
| `JWT_EXPIRY` | `168h` | Access-token lifetime (Go duration) |
| `SEED` | `true` | Insert development seed data when the DB is empty |
| `CORS_ORIGIN` | `*` | Allowed CORS origin(s) |

`.env` is loaded automatically if present (optional).

## Seed accounts (`SEED=true`)

| Email | Password | Role | Chef status |
|---|---|---|---|
| `admin@tourchef.uz` | `admin123` | admin | — |
| `family@tourchef.uz` | `family123` | family host (role: chef) | approved |
| `client@tourchef.uz` | `client123` | client | — |
| `pending1@tourchef.uz` | `family123` | family host (role: chef) | pending |
| `pending2@tourchef.uz` | `family123` | family host (role: chef) | pending |

Seed also creates 3 restaurants, 8 Uzbek dishes (Wikimedia Commons images),
5 approved masterclasses, 10 sessions, 3 reels, sample favorites/reviews and
notifications. Seeding is skipped when `admin@tourchef.uz` already exists.

## Authentication

Send `Authorization: Bearer <JWT>`. Roles: `client`, `chef`, `admin`.
Errors always have the shape `{"error":"message"}` with status 400/401/403/404/409/500.

## Endpoints

### Auth
| Method | Path | Auth | Description |
|---|---|---|---|
| POST | `/api/auth/register` | – | `{email,password,full_name,role:"client"\|"chef",phone?}` → `{token,user}` |
| POST | `/api/auth/login` | – | `{email,password}` → `{token,user}` |
| GET | `/api/auth/me` | yes | current `user` |

### Public
| Method | Path | Description |
|---|---|---|
| GET | `/api/chefs` | approved chefs |
| GET | `/api/dishes` | filters `q,category,region,sort=new\|name\|rating,limit,offset` → `{items,total}` |
| GET | `/api/dishes/:slug` | dish detail |
| GET | `/api/masterclasses` | filters `q,city,dish,category,min_price,max_price,duration=1-2\|2-4\|4+,rating,sort=rec\|price_asc\|price_desc\|rating\|new,page,limit` → `{items,total,page,pages}` |
| GET | `/api/masterclasses/:slug` | masterclass with `sessions`, `chef`, `restaurant`, `dish` |
| GET | `/api/masterclasses/:slug/sessions` | sessions array |
| GET | `/api/restaurants` | restaurants array |
| GET | `/api/restaurants/:slug` | restaurant with `masterclasses` |
| GET | `/api/reels` | published reels with chef/restaurant/masterclass |
| GET | `/api/reviews?masterclass_id=` | reviews array |

All masterclass catalog filters are executed in SQL server-side.

### Dishes (admin)
`POST /api/dishes`, `PATCH /api/dishes/:id` (partial), `DELETE /api/dishes/:id` → `{ok:true}`.

### Chef
`POST /api/masterclasses`, `GET /api/chef/masterclasses`,
`POST /api/chef/sessions`, `GET /api/chef/sessions`.
Chef-created masterclasses get status `pending` until an admin approves them.

### Client cabinet
| Method | Path | Description |
|---|---|---|
| POST | `/api/client/bookings` | `{session_id,guests}` → booking. Transaction + `FOR UPDATE` row lock prevents overbooking; `payment_status='paid_dev'` |
| GET | `/api/client/bookings` | own bookings (`?status=` optional) |
| PATCH | `/api/client/bookings/:id/cancel` | cancel and release seats |
| GET/POST | `/api/client/favorites` | list / add `{masterclass_id}` |
| DELETE | `/api/client/favorites/:id` | remove |
| POST | `/api/client/reviews` | `{masterclass_id,rating,comment}` |
| GET | `/api/client/notifications` | own notifications |

`POST /api/bookings` is also available to clients as an alias of the cabinet route.

### Admin
| Method | Path | Description |
|---|---|---|
| GET | `/api/admin/chefs` | all chefs including `pending` |
| PATCH | `/api/admin/chefs/:id` | `{status:"approved"\|"suspended"}` |
| GET | `/api/admin/users` | users (`?q=&role=&status=`) |
| PATCH | `/api/admin/users/:id` | `{status}` |
| GET | `/api/admin/masterclasses` | all masterclasses |
| PATCH | `/api/admin/masterclasses/:id` | `{status}` |
| GET | `/api/admin/sessions` | all sessions |
| GET | `/api/admin/bookings` | all bookings (`?q=&status=&user_id=`) |
| PATCH | `/api/admin/bookings/:id` | `{status,payment_status?}` |
| GET | `/api/admin/stats` | users, chefs, restaurants, dishes, masterclasses, bookings, revenue, pending_chefs, pending_mc |

## Notes / deviations

- UUIDs are stored as PostgreSQL `uuid` and serialized as strings.
- Payments are a clearly labeled **development mode** (`paid_dev`); no real
  payment provider credentials are used.
- Chef account registration creates a `users` row with status `active` so the
  chef can log in, plus a `chefs` row with status `pending` so it appears in
  `/api/admin/chefs`.
- A chef-created masterclass is submitted as `pending`; only an admin can set it
  to `approved`, matching the plan's moderation rules.
- Admin dish creation immediately affects `/api/dishes` and the public detail route.
