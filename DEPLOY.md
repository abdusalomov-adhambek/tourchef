# TourChef — VPS ga joylashtirish rejasi

**Tayyorlangan:** 2026-11-26  
**Maqsad server:** `45.138.158.39` (`vps01314.eskiz.uz`, Ubuntu 24.04, Docker 29.6)  
**Loyiha:** `~/Desktop/tourchef` → GitHub: `https://github.com/abdusalomov-adhambek/tourchef.git` (public, `main`, lokal bilan bir xil — `febc96e`)

---

## 1. Hozirgi holat tahlili

### Loyiha tomonida
| Element | Holat |
|---|---|
| `app/` | Nuxt 3.15 (Vue 3 + TS), to'liq sahifalar: public, `/client`, `/family`, `/admin`, `/masterclasses`, `/dishes`, `/regions`, `/reels` |
| `server/` | Go 1.25 + Fiber v2 + pgx, 4 ta embed migratsiya, avtomatik seed, `/health`, `/uploads` static |
| `docker-compose.yml` | Lokal uchun yaroqli, **VPS uchun mos emas** (80/443 va 5432 ni tashqariga ochadi) |
| `app/Dockerfile` | **YO'Q** — `docker-compose.yml` unda ishora qiladi, shuning uchun `docker compose up --build` hozir ishlamaydi |
| `scripts/smoke.sh` | Tayyor E2E test (admin login, chef register, dish yaratish, filter, buyurtma) |
| Demo akkauntlar | `admin@tourchef.uz/admin123`, `family123`, `client123` — ishlab chiqish uchun |

### Server tomonida
- **80/443 portlari allaqachon umumiy Caddy (`agnts-caddy-1`) ga tegishli** — uni to'xtatish yoki qayta sozlash mumkin emas. Caddy konfiguratsiyasi: `/opt/agnts/Caddyfile.cloud` (konteyner ichida `/etc/caddy/Caddyfile`).
- Caddy `agnts_default` tarmog'ida; boshqa loyihalar shunga qo'shilib, o'z nomlari bilan ko'rinadi.
- `tourchef.app.codeops.uz` DNS orqali allaqachon serverga qaraydi, lekin Caddy da unga atalgan blok **yo'q** — hozir `codeops.caddy` dagi `https://` catch-all orqali `codeops:8000` ga tushadi. Aniq blok qo'shish shart.
- Band qilingan host portlar: `13000`, `15173`, `18080`, `8090`, `8787`, `8000`, `8443`, pochta portlari.
- UFW faol: faqat `22`, `80`, `443`, `443/udp`, `8443` + `172.16.0.0/12` ochilgan.
- Disk 103 GB bo'sh, RAM 13 GB available — resurs yetarli.
- **96 ta yangilanish bor, restart talab qilinadi** — alohida texnik oynada qilish kerak.

### Xulosa
Eng yaqin o'xshash namuna serverda bor: `/opt/rustam-dev/website` (Nuxt + Go + Postgres + umumiy Caddy). Reja aynan shu pattern ga asoslanadi.

---

## 2. Qaror qilinadigan narsalar

| Savol | Tavsiya |
|---|---|
| Domain | **`tourchef.app.codeops.uz`** (foydalanuvchi so'ragan domen). E'tibor: serverdagi ko'p foydalanuvchi ilovalari `*.apps.codeops.uz` (ko'plik) sxemasida; bu domen `*.app.codeops.uz` (yakka) ostida, lekin wildcard DNS tufayli ishlaydi. |
| DNS | **O'zgartirish shart emas** — `tourchef.app.codeops.uz` allaqachon `45.138.158.39` ga rezolyutsiya qilinadi (`dig` bilan tasdiqlangan). TLS sertifikatini Caddy o'zi oladi. |
| Seed | Birinchi ishga tushirishda `SEED=true`, demo parollarni almashtirgach `SEED=false` qilib qo'yish. |
| Portlar | Frontend `127.0.0.1:13001`, backend `127.0.0.1:18081`, Postgres — **faqat ichki tarmoqda** (host port yo'q). |

---

## 3. Maqsadli arxitektura

```
Internet ──443──▶ umumiy Caddy (TLS terminate)
                   └─ tourchef.app.codeops.uz ──▶ tourchef-frontend:3000
                         /api/**  ─▶ Nuxt Nitro proxy ─▶ tourchef-backend:8080/api/**
                         /uploads/** ─▶ Nitro proxy ─▶ tourchef-backend:8080/uploads/**

Docker Compose (project: tourchef):
  postgres  tourchef-postgres   (ichki, volume: pgdata)
  backend   tourchef-backend    127.0.0.1:18081  + agnts_default
  frontend  tourchef-frontend   127.0.0.1:13001  + agnts_default
  volume: uploads (/app/uploads)
```

**Muhim:** `/api/*` ni Caddy dan to'g'ridan-to'g'ri Go ga yubormang — frontend Nitro proxy qiladi. Aks holda `/uploads` va SSR so'rovlari buziladi. `api.*` subdomaini faqat debug/health uchun ixtiyoriy.

---

## 4. Joylashtirish oldidan tuzatishlar (kod)

### 4.1 `app/Dockerfile` — YANGI fayl (yo'q)

```dockerfile
# ---- build ----
FROM node:20-alpine AS builder
WORKDIR /app
# NUXT routeRules build vaqtida kompilyatsiya bo'ladi, shuning uchun proxy
# manzili build arg sifatida berilishi shart (runtime env yetarli emas).
ARG API_PROXY_TARGET=http://tourchef-backend:8080
ENV API_PROXY_TARGET=$API_PROXY_TARGET
COPY package.json package-lock.json ./
RUN npm ci
COPY . .
RUN npm run build

# ---- runtime ----
FROM node:20-alpine AS runtime
WORKDIR /app
ENV NODE_ENV=production \
    NITRO_HOST=0.0.0.0 \
    NITRO_PORT=3000 \
    PORT=3000
COPY --from=builder /app/.output ./.output
EXPOSE 3000
CMD ["node", ".output/server/index.mjs"]
```

### 4.2 `app/.dockerignore` — YANGI fayl

```
node_modules
.nuxt
.output
.nitro
.env
*.log
.DS_Store
```

### 4.3 `docker-compose.vps.yml` — YANGI fayl (lokal `docker-compose.yml` ga tegmaydi)

```yaml
name: tourchef

services:
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: ${POSTGRES_USER}
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
      POSTGRES_DB: ${POSTGRES_DB}
    volumes:
      - pgdata:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${POSTGRES_USER} -d ${POSTGRES_DB}"]
      interval: 5s
      timeout: 3s
      retries: 10
    networks:
      default:
        aliases: [tourchef-postgres]
    restart: unless-stopped
    # host port yo'q — faqat ichki tarmoq

  backend:
    build:
      context: ./server
    environment:
      APP_ENV: production
      PORT: 8080
      DATABASE_URL: postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@tourchef-postgres:5432/${POSTGRES_DB}?sslmode=disable
      JWT_SECRET: ${JWT_SECRET}
      JWT_EXPIRY: 168h
      SEED: ${SEED:-false}
      CORS_ORIGIN: ${SITE_URL}
      SITE_URL: ${SITE_URL}
      UPLOAD_DIR: /app/uploads
    volumes:
      - uploads:/app/uploads
    ports:
      - "127.0.0.1:18081:8080"
    depends_on:
      postgres:
        condition: service_healthy
    healthcheck:
      test: ["CMD-SHELL", "wget -q -O - http://127.0.0.1:8080/health | grep -q ok"]
      interval: 10s
      timeout: 3s
      retries: 5
    networks:
      default:
      caddy:
        aliases: [tourchef-backend]
    restart: unless-stopped

  frontend:
    build:
      context: ./app
      args:
        API_PROXY_TARGET: http://tourchef-backend:8080
    environment:
      NUXT_PUBLIC_API_BASE: /api
      NITRO_HOST: 0.0.0.0
      NITRO_PORT: 3000
    ports:
      - "127.0.0.1:13001:3000"
    depends_on:
      - backend
    networks:
      default:
      caddy:
        aliases: [tourchef-frontend]
    restart: unless-stopped

networks:
  default:
  caddy:
    name: agnts_default
    external: true

volumes:
  pgdata:
  uploads:
```

> Eslatma: `DATABASE_URL` dagi parol URL-xavfsiz bo'lishi kerak (`@`, `:`, `/`, `#` bo'lmasin). `openssl rand -hex 24` ishlatish tavsiya etiladi.

---

## 5. VPS dagi `.env` (`/opt/tourchef/.env`, gitga kirmaydi)

```env
POSTGRES_USER=tourchef
POSTGRES_PASSWORD=<openssl rand -hex 24>
POSTGRES_DB=tourchef

JWT_SECRET=<openssl rand -hex 32>
JWT_EXPIRY=168h
SEED=true

SITE_URL=https://tourchef.app.codeops.uz
```

---

## 6. Caddy ga qo'shiladigan blok

**Muhim:** hozir `tourchef.app.codeops.uz` hech bir aniq sayt blokiga tushmaydi va `/opt/codeops/caddy/codeops.caddy` dagi `https://` catch-all ga tushib, `codeops:8000` ga yuboriladi (noto'g'ri joy). Shu sababli `/opt/agnts/Caddyfile.cloud` ga **aniq domen blokini** qo'shamiz. Aniq hostname har doim `https://` catch-all dan ustun turadi, shuning uchun `import /etc/caddy/codeops.caddy` satridan **oldin** qo'shish tavsiya etiladi (keyin qo'shsangiz ham ishlaydi, lekin oldin qo'yish aniqroq).

```caddyfile
# TourChef — kulinar turizm marketpleysi (/opt/tourchef)
# Bu blok codeops ning catch-all (https://) ga tushib ketishining oldini oladi.
tourchef.app.codeops.uz {
	import security
	encode zstd gzip
	# /api/** va /uploads/** ni Nuxt Nitro o'zi proxy qiladi;
	# ularni to'g'ridan-to'g'ri Go ga yubormang.
	reverse_proxy tourchef-frontend:3000
}
```

> `on_demand` ishlatilmaydi — domen aniq yozilgani uchun Caddy oddiy Let's Encrypt ACME HTTP-01 orqali sertifikat oladi. `codeops` ning `on_demand_tls { ask ... }` qoidasi bu domenga ta'sir qilmaydi.
> `www.` va `api.` subdomainlari kerak bo'lsa, ularni qo'shimcha blok sifatida yozing (ixtiyoriy).

Tekshirish va qo'llash:

```bash
cp /opt/agnts/Caddyfile.cloud /opt/agnts/Caddyfile.cloud.bak-$(date +%F)
docker exec agnts-caddy-1 caddy validate --config /etc/caddy/Caddyfile --adapter caddyfile
docker exec agnts-caddy-1 caddy reload --config /etc/caddy/Caddyfile --adapter caddyfile
# reload ishlamasa (oxirgi chora, ~1s uzilish): docker restart agnts-caddy-1
```

---

## 7. Amalga oshirish bosqichlari

### Bosqich 0 — DNS va texnik xizmat (separate oyna)
1. DNS o'zgartirish kerak emas — `*.codeops.uz` wildcard yozuvi domenni `45.138.158.39` ga yo'naltiradi:
   `dig +short tourchef.app.codeops.uz` → `45.138.158.39` bo'lishi kerak.
2. Caddy dagi hozirgi holatni tasdiqlang (hech qanday aniq blok yo'q, catch-all ga tushadi):
   `grep -n 'app.codeops.uz' /opt/agnts/Caddyfile.cloud /opt/codeops/caddy/codeops.caddy`.
3. (Tavsiya) 96 ta yangilanishni o'rnating va reboot qiling — **deploy bilan aralashtirmang**:
   `apt update && apt upgrade -y && reboot`.

### Bosqich 1 — Kodni tayyorlash (lokaldan)
1. Yuqoridagi 3 faylni (`app/Dockerfile`, `app/.dockerignore`, `docker-compose.vps.yml`) qo'shing.
2. `git add ... && git commit && git push origin main`.
3. Tasdiqlang: `git ls-remote origin refs/heads/main` lokal `HEAD` bilan bir xil bo'lishi kerak.

### Bosqich 2 — Serverga olish
```bash
ssh root@45.138.158.39
git clone https://github.com/abdusalomov-adhambek/tourchef.git /opt/tourchef
cd /opt/tourchef
```

### Bosqich 3 — Sirlarni yozish
```bash
cat > /opt/tourchef/.env <<'EOF'
POSTGRES_USER=tourchef
POSTGRES_PASSWORD=...hex...
POSTGRES_DB=tourchef
JWT_SECRET=...hex...
JWT_EXPIRY=168h
SEED=true
SITE_URL=https://tourchef.app.codeops.uz
EOF
chmod 600 /opt/tourchef/.env
```

### Bosqich 4 — Stack ni ishga tushirish
```bash
cd /opt/tourchef
docker compose -f docker-compose.vps.yml up -d --build
docker compose -f docker-compose.vps.yml logs -f
```
Migratsiyalar va seed backend startida avtomatik bajariladi.

### Bosqich 5 — Ichki tekshiruv (Caddy siz)
```bash
curl -s http://127.0.0.1:18081/health          # {"status":"ok"}
curl -s -o /dev/null -w '%{http_code}\n' http://127.0.0.1:13001   # 200
curl -s http://127.0.0.1:13001/api/masterclasses | head -c 200
docker compose -f docker-compose.vps.yml ps
```

### Bosqich 6 — Caddy + TLS
6-bo'limdagi blokni qo'shing, `validate` + `reload`. So'ngra:
```bash
curl -sI https://tourchef.app.codeops.uz | head -5
```

### Bosqich 7 — Smoke test
```bash
cd /opt/tourchef
API=https://tourchef.app.codeops.uz/api bash scripts/smoke.sh
```
> Test ma'lumot qoldiradi (1 chef, 1 taom, 1 buyurtma). Keyin tozalang yoki demo ma'lumot deb qabul qiling.

### Bosqich 8 — Production xavfsizligi (majburiy)
1. Admin panel orqali `admin@tourchef.uz` parolini `admin123` dan kuchlisiga almashtiring; `family`, `client` akkauntlarini ham.
2. `.env` da `SEED=false` qiling:
   ```bash
   sed -i 's/^SEED=true/SEED=false/' /opt/tourchef/.env
   docker compose -f docker-compose.vps.yml up -d --force-recreate backend
   ```

---

## 8. Backup va kuzatuv

```bash
# DB zaxira (kunlik cron uchun)
mkdir -p /backups/tourchef
docker compose -f /opt/tourchef/docker-compose.vps.yml exec -T postgres \
  pg_dump -U tourchef tourchef | gzip > /backups/tourchef/db-$(date +%F).sql.gz

# Rasmlar
tar czf /backups/tourchef/uploads-$(date +%F).tar.gz \
  /var/lib/docker/volumes/tourchef_uploads/_data
```

- Loglar: `docker compose -f docker-compose.vps.yml logs --tail=200 backend|frontend`.
- Yangilanish: `git pull && docker compose -f docker-compose.vps.yml up -d --build`.
- Rollback: Caddy blokini olib tashlang + `.bak` fayldan tiklang, `docker compose ... down` ( volumes saqlanadi; `-v` qo'ymang!).

---

## 9. Xavflar va nazorat ro'yxati

| # | Xavf | Chora |
|---|---|---|
| 1 | `app/Dockerfile` yo'q | 4.1 bo'yicha qo'shish, shartsiz |
| 2 | Nitro proxy build vaqtida yoziladi | `API_PROXY_TARGET` ni **build arg** sifatida bering |
| 3 | 80/443 ni egallash | Faqat `127.0.0.1` portlar; Caddy ga faqat blok qo'shish |
| 4 | Tarmoq nomi to'qnashuvi | `tourchef-frontend` / `tourchef-backend` / `tourchef-postgres` aliaslari |
| 5 | Postgres tashqariga ochilishi | Compose da postgres uchun `ports` yo'q |
| 6 | Demo parollar | Birinchi kiritishdayoq almashtirish, keyin `SEED=false` |
| 7 | Paroldagi maxsus belgilar DSN ni buzadi | Faqat hex parollar |
| 8 | Caddy xatosi boshqa saytlarni buzadi | Zaxira + `caddy validate` + reload |
| 9 | Server yangilanish talab qiladi | Deploy dan oldin alohida reboot oynasi |
| 10 | Upload hajmi 10 MB ( kodda qattiq ) | Kerak bo'lsa `handlers_upload.go` dagi `maxUploadBytes` o'zgartirish |
| 11 | Domen `codeops` catch-all ga tushib ketadi | `tourchef.app.codeops.uz` uchun aniq Caddy bloki; `caddy validate` bilan tasdiqlash |
| 12 | `*.app` va `*.apps` nomlash farqi | Siz so'ragan `tourchef.app.codeops.uz` ishlatiladi; agar platforma konvensiyasiga rioya qilish kerak bo'lsa `tourchef.apps.codeops.uz` ga o'zgartiring |

---

## 10. Taxminiy vaqt va natija

- Tayyorgarlik + push: ~10 daqiqa
- Clone + build + start: ~10–15 daqiqa (Go va npm birinchi build sekinroq)
- Caddy + DNS + TLS: ~5 daqiqa
- Test + xavfsizlik: ~10 daqiqa

**Natija:** `https://tourchef.app.codeops.uz` — to'liq ishlaydigan platforma (public sayt, client/family kabinet, admin panel), API va rasmlar bir origin orqali, Postgres va uploads doimiy volume larda.
