# TourChef — to'liq ishlaydigan platforma

Kulinar turizm marketplace: public sayt + Client Cabinet + Family Cabinet (Nuxt 3 / Vue 3), Admin Panel (Vue 3 sahifalari Nuxt ichida), Go/Fiber REST API va PostgreSQL ma'lumotlar bazasi.

## Tuzilma

```
tourchef/
├── app/          Nuxt 3 frontend (public, client, family, admin), Vue 3 + TypeScript
├── server/       Go 1.25 + Fiber v2 API, pgx, JWT, SQL migratsiyalar va seed
├── design/       Statik HTML prototip (dizayn manbai)
├── docs/CONTRACT.md  Backend ↔ frontend API kelishuvi
├── docker-compose.yml
└── .env.example
```

## Tez ishga tushirish (Docker, tavsiya etiladi)

```bash
cp .env.example .env
docker compose up --build
```

- Sayt: http://localhost:3000
- API: http://localhost:8080/api
- PostgreSQL: localhost:5432 (`postgres` / `postgres`, db `tourchef`)

Migratsiyalar va seed ma'lumotlari backend ishga tushganda avtomatik bajariladi (`SEED=true`).

### Demo akkauntlar

| Email | Parol | Roli |
|---|---|---|
| `admin@tourchef.uz` | `admin123` | admin |
| `family@tourchef.uz` | `family123` | family host (texnik role: `chef`) |
| `client@tourchef.uz` | `client123` | client |

## Docker'siz ishlab chiqish

```bash
# 1) Postgres
createdb tourchef   # yoki: docker compose up -d postgres

# 2) Backend
cp .env.example server/.env
cd server && go mod tidy && go run .
# http://localhost:8080/api

# 3) Frontend (yangi terminal)
cd app && npm install && npm run dev
# http://localhost:3000
```

`.env` dagi `DATABASE_URL` va `JWT_SECRET` ni o'zingizga moslang. `.env` fayli gitga kirmaydi.

## Asosiy funksiyalar

- **Landing**: API dan keladigan masterclasslar, taomlar, restoranlar va reels preview; qidiruv katalogga o'tadi.
- **Masterclass katalogi**: qidiruv, shahar, taom, narx oralig'i, davomiylik, reyting, saralash va sahifalash — barchasi backendda ishlaydi.
- **Ro'yxatdan o'tish**: `client` va `chef` rollari. Chef arizasi `pending` statusda admin panelning **Families** bo'limida ko'rinadi va admin tomonidan tasdiqlanadi.
- **Admin panel**: dashboard statistikasi, foydalanuvchilar, family tasdiqlash, masterclass yaratish va moderatsiya (hudud tanlanadi, family selecti shu hudud oilalari bilan filtrlanadi), sessionlar, buyurtmalar, reels, taomlar (to'liq CRUD, multiimage) va hisobotlar.
- **Taomlar**: admin qo'shgan taom darhol landing, `/dishes` va Family Cabinetdagi ro'yxatda ko'rinadi. Formada bir nechta rasm (`multiimage`) yuklanadi; birinchi rasm asosiy sifatida kartochkalarda ko'rsatiladi.
- **Family Cabinet** (`/family`): o'z masterclasslari va sessionlarini yaratadi, platformadagi barcha faol taomlarni ko'radi.
- **Taste of Uzbekistan**: landing dagi interaktiv xarita (14 mintaqa, haqiqiy GeoJSON chegaralari), mintaqa tanlanganda taom, madaniy ma'lumot, restoran va masterclasslar ko'rsatiladi; ma'lumotlar Admin → Taste of Uzbekistan orqali boshqariladi.
- **Client Cabinet**: dashboard, buyurtmalar, sevimlilar, sharhlar, bildirishnomalar, profil, "With Uzbekfamily" bo'limi.
- **Buyurtma**: session sig'imi tekshiriladi, to'lov development rejimida (`paid_dev`) — haqiqiy to'lov provayderi ulanmagan.

## API

To'liq endpoint va JSON shakllari: [`docs/CONTRACT.md`](docs/CONTRACT.md).

## Eslatma

`design/` papkasidagi statik HTML fayllar dizayn manbai sifatida saqlanadi; ishlab chiqarish versiyasi `app/` va `server/` da.

Xarita GeoJSON manbai: `app/public/geo/uzbekistan-regions.geojson` (MIT, © 2024 Shaxzod Uktamov — [dirixtt/GeoJSON-Uzbekistan](https://github.com/dirixtt/GeoJSON-Uzbekistan)); 14 mintaqa chegarasi haqiqiy geografik ma'lumotlardan olingan.
