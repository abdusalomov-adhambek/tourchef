# TourChef — Design-2 (static HTML prototype)

`plan.txt` va `design.txt` talablari asosida yaratilgan yangi dizayn prototipi.
Eski `design/` papkasidagi "guide" sayti dizayni o'rniga **TourChef** kulinar turizm platformasi dizayni qilingan.

## Ishga tushirish

`index.html` ni brauzerda oching. Barcha sahifalar bir-biriga havola qilingan.
To'liq ro'yxat uchun `all-pages.html` ni oching.

## Fayl tuzilmasi

```
design-2/
├── index.html                  Landing page (hero, qidiruv, masterclasslar, taomlar, restoranlar, Reels preview, CTA, testimonials, footer)
├── reels.html                  To'liq ekranli vertikal Reels feed
├── masterclasses.html          Katalog + filtrlar + saralash + pagination
├── masterclass-detail.html     Masterclass detali + booking widget
├── restaurants.html            Restoranlar ro'yxati
├── restaurant.html             Restoran public profili
├── dish.html                   Taom detali (Plov)
├── login.html / register.html / forgot-password.html
├── client-*.html               Client Cabinet (9 sahifa)
├── restaurant-*.html           Chef Cabinet (10 sahifa)
├── admin-*.html                Admin Panel (12 sahifa)
└── assets/
    ├── tourchef.css            Barcha dizayn tokenlari va komponentlar
    └── app.js                  Interaksiyalar (toast, tab, modal, like/save, stepper, Reels observer)
```

## Dizayn tizimi (design.txt)

| Token | Qiymat |
|---|---|
| Background | `#F8F5EF` |
| Surface | `#FFFFFF` |
| Text | `#252923` / `#77766F` |
| Primary accent | `#C65D3A` (hover `#A94B2D`) |
| Secondary accent | `#D9A441` |
| Sage | `#7B8B68` |
| Border | `#E8E2D8` |
| Reels | `#111310` / `#FFFFFF` |

Fontlar: **DM Sans** (UI) + **Playfair Display** (editorial sarlavhalar).

## E'tiborli jihatlar

- Barcha tugmalar, tablar, filtr chiplari, like/save, modal va stepper lar ishlaydi (frontend demo darajasida, `app.js`).
- Hech qanday soxta bo'sh sahifa yo'q: har bir sahifa real ma'lumotlar bilan to'ldirilgan.
- To'lovlar **development mode** deb belgilangan (`Click (dev)`, `Payme (dev)`) — haqiqiy to'lov sifatida ko'rsatilmagan.
- Admin panel tuzilmasi `../codeops/design` dagi sidebar + KPI + jadval patternidan olingan, lekin ranglar va modullar TourChef ga moslashtirilgan.
- Bu faqat **dizayn prototipi**: ma'lumotlar statik. Keyingi bosqichda Nuxt 3 + Go/Fiber backend bilan ulanadi.
- Taom rasmlari **Wikimedia Commons** dagi haqiqiy o'zbek milliy taomlari suratlaridan olingan (Uzbek pilaf, manti, somsa, lagman, shashlik, shurpa, chuchvara, dastarkhan). Restoran interyerlari va avatarlar Unsplash dan. Production da barchasi S3-compatible storage ga yuklanadi.
- Reels dagi videolar o'rniga poster rasmlar ishlatilgan (development video reference); production da video S3/CDN dan oqadi.
