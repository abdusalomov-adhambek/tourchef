# TourChef — API contract (backend va frontend uchun umumiy kelishuv)

Base URL: `http://localhost:8080/api` (env: `API_BASE_URL`).
JSON body. Auth: `Authorization: Bearer <JWT>` (frontend cookie `tc_token` da saqlaydi).
Rollarga ega foydalanuvchi: `client`, `chef`, `admin`. Status: `active`, `pending`, `suspended`.

## Auth

| Method | Path | Auth | Body | Response |
|---|---|---|---|---|
| POST | `/auth/register` | – | `{email, password, full_name, role: "client"|"chef", phone?}` | `{token, user}` |
| POST | `/auth/login` | – | `{email, password}` | `{token, user}` |
| POST | `/auth/forgot-password` | – | `{email}` | `{ok, message, dev}` (development rejimi: haqiqiy xat yuborilmaydi) |
| GET | `/auth/me` | yes | – | `user` |
| PATCH | `/auth/me` | yes | `{full_name?, email?, phone?}` | `user` |

`user = {id, email, full_name, role, status, phone?, created_at}`.
Chef ro'yxatdan o'tganda `chefs` jadvaliga `status='pending'` yoziladi va admin panelda ko'rinadi.

## Users / Chefs

| GET `/chefs` | public → `[{id, user_id, display_name, city, specialty, bio, avatar_url, rating, status, email}]` (faqat `approved`), `?city=` / `?region=` bo'yicha filtrlanadi |
| GET `/admin/chefs` | admin → barchasi (status bilan), `?city=` bo'yicha filtrlanadi |
| PATCH `/admin/chefs/:id` | admin → `{status: "approved"|"suspended"}` → chef |
| GET `/admin/users` | admin → `[user]` |
| PATCH `/admin/users/:id` | admin → `{status}` → user |

## Dishes (admin qo'shadi, hamma ko'radi)

| GET `/dishes` | query: `q, category, region, sort(new|name|rating), limit, offset` → `{items, total}` |
| GET `/dishes/:slug` | → dish |
| POST `/dishes` | admin → `{name, category, region, description, ingredients, image_url, status?}` → dish (`ingredients` string **yoki** `string[]` qabul qilinadi, massiv `", "` bilan birlashtiriladi) |
| PATCH `/dishes/:id` | admin → partial |
| DELETE `/dishes/:id` | admin → `{ok:true}` |
| POST `/upload` | admin/chef → `multipart/form-data`, `multiimage` (bir nechta), `file` yoki `image` maydonlari, har biri max 10 MB, jpg/png/webp/gif → `{url: birinchi, urls: [...], files: [...]}` |

Yuklangan fayllar backend tomonidan `UPLOAD_DIR` (default `./uploads`) ga saqlanadi va `/uploads/*` manzilidan ochiq ko'rsatiladi (Nuxt `/api` va `/uploads` proxy qiladi).

`dish = {id, name, slug, category, region, description, ingredients, image_url, multiimage, status, created_at, created_by}`.

`multiimage` — rasmlar URL lari massivi (`jsonb`). `image_url` har doim birinchi rasmni ko'rsatadi (public kartochkalar shu maydonni ishlatadi). Qo'lda URL kiritish olib tashlangan: barcha rasmlar `/upload` orqali yuklanadi.
`status='active'` bo'lgan taomlar landing (`/`), katalog (`/dishes`) va chef kabinetida ko'rinadi.

## Masterclasses

| GET `/masterclasses` | query: `q, city, dish, min_price, max_price, duration(1-2|2-4|4+), rating, sort(rec|price_asc|price_desc|rating|new), page, limit` → `{items, total, page, pages}` |
| GET `/masterclasses/:slug` | → `{...mc, sessions:[session], chef, restaurant, dish}` |
| GET `/masterclasses/:slug/sessions` | → `[session]` |
| POST `/masterclasses` | chef/admin → `{title, description, dish_id, city, duration_min, price, difficulty, language, image_url, restaurant_id?, status?}` → mc |
| GET `/admin/masterclasses` | admin → barchasi |
| PATCH `/admin/masterclasses/:id` | admin → `{status}` |
| GET `/chef/masterclasses` | chef → o'ziniki |

`mc = {id, title, slug, description, dish_id, dish, chef_id, chef, restaurant_id, restaurant, city, duration_min, price, difficulty, language, rating, image_url, status, created_at}`.

## Sessions

| POST `/chef/sessions` | chef → `{masterclass_id, starts_at, capacity, price, status?}` → session |
| GET `/chef/sessions` | chef → o'ziniki |
| GET `/admin/sessions` | admin → barchasi |

`session = {id, masterclass_id, starts_at, capacity, booked, price, status}`.

## Restaurants & Reels

| GET `/restaurants` | → `[restaurant]` |
| GET `/restaurants/:slug` | → restaurant (masterclasslari bilan) |
| GET `/reels` | → `[reel]` |

`restaurant = {id, name, slug, city, cuisine, description, image_url, rating, status}`.
`reel = {id, title, caption, image_url, video_url, views, likes, chef, restaurant, status}`.

## Bookings / Favorites / Reviews / Notifications

| POST `/bookings` | client → `{session_id, guests}` → booking (payment `dev` rejim: `payment_status='paid_dev'`). Har bir bron uchun QR avtomatik yaratiladi: `qr_code_url` maydonida `/uploads/qr-<id>.png` qaytadi |
| GET `/client/bookings/:id/qr` | client → `{url, booking_number}` — eski bronlar uchun QR yaratib beradi; `?refresh=1` bilan qayta yaratadi. QR rangi: `#C65D3A` (oq fon) |
| GET `/client/bookings` | client → o'ziniki |
| PATCH `/client/bookings/:id/cancel` | client → booking |
| GET `/admin/bookings` | admin → barchasi |
| PATCH `/admin/bookings/:id` | admin → `{status, payment_status?}` |
| GET `/client/favorites` / POST `/client/favorites` `{masterclass_id}` / DELETE `/client/favorites/:id` |
| GET `/reviews?masterclass_id=` / POST `/client/reviews` `{masterclass_id, rating, comment}` |
| GET `/client/notifications` → `[notification]` |

## Taste of Uzbekistan (interaktiv xarita)

| GET `/v1/regions` | public → barcha `published` mintaqa-taom yozuvlari (marker ma'lumotlari bilan) |
| GET `/v1/regions/:slug` | mintaqa detali |
| GET `/v1/regions/:slug/featured-dish` | taom ma'lumoti |
| GET `/v1/regions/:slug/masterclasses` | shu mintaqa shahridagi masterclasslar → `{items, total}` |
| GET `/v1/regions/:slug/restaurants` | shu mintaqadagi restoranlar → `[restaurant]` |
| GET `/admin/regions` / POST `/admin/regions` / PATCH `/admin/regions/:id` / DELETE `/admin/regions/:id` | admin CRUD |

`region = {id, slug, geo_id, name, name_uz, city, dish_name, dish_image_url, description, ingredients[], cultural_info, latitude, longitude, status, created_at}`.
`geo_id` — `app/public/geo/uzbekistan-regions.geojson` dagi mintaqa `id` si bilan mos keladi.

## Admin

| GET `/admin/stats` | → `{users, chefs, restaurants, dishes, masterclasses, bookings, revenue, pending_chefs, pending_mc}` |
| GET `/admin/reels` | admin → barcha reels (moderatsiya navbati bilan) |
| PATCH `/admin/reels/:id` | admin → `{status: "published"|"rejected"|"pending"}` → reel |

## Xatolar

`{error: "message"}` + tegishli HTTP status (400/401/403/404/409/500).
Frontend barcha xatolarni toast xabarida ko'rsatadi.

## Seed ma'lumotlari (dev)

- `admin@tourchef.uz / admin123` (role admin)
- `family@tourchef.uz / family123` (role chef, approved)
- `client@tourchef.uz / client123` (role client)
- 3 ta restoran, 8 ta taom, 5 ta masterclass, sessionlar, reels, 2 ta pending chef arizasi.

## Muhim

- Admin taom qo'shganda (POST `/dishes`) u darhol `/dishes`, landing va `/chef` sahifalarida ko'rinishi kerak.
- Chef ro'yxatdan o'tganda (POST `/auth/register` role=chef) u `/admin/chefs` da `pending` statusda ko'rinishi kerak.
- Masterclass katalogidagi barcha filterlar backendda ishlashi shart (frontend faqat query yuboradi).
