# SabtBrooker — سند تحویل پروژه

> این فایل رو به جلسه جدید بده تا دقیقاً از همینجا ادامه بده.

---

## ۱. پروژه چیه؟

**سامانه کارگزاری ماده ۱۰ قانون الزام به ثبت رسمی معاملات اموال غیرمنقول**

ما به عنوان **کارگزار سازمان ثبت اسناد و املاک کشور** باید سامانه‌ای طراحی و اجرا کنیم که سه خدمت زنجیره‌ای ارائه بده:

1. **تهیه نقشه ثبتی** → کد رهگیری نقشه
2. **درج ادعا** (ماده ۱۰) → کد رهگیری ادعا  
3. **درج گواهی اقدام** (ماده ۱۰) → کد رهگیری نهایی

خروجی هر مرحله، ورودی مرحله بعد است. کل زنجیره باید state machine و deadline tracking (۲ سال / ۵ ماه) داشته باشد.

---

## ۲. ریپوی گیت

**https://github.com/msaeedlavasani/SabtBrooker**

Branch: `main`  
آخرین commit: `06e05b4` — `fix: relax identity guard for dev mode`

CI سبز است (Lint → Test → Build → Security Scan).

---

## ۳. فایل‌های کلیدی — همینا رو اول بخون

| اولویت | فایل | محتوا |
|---|---|---|
| ۱ | `docs/architecture-blueprint.md` | معماری کلی، ۱۰ لایه زیرساخت، مدل تهدید، نقشه راه ۳۰ هفته |
| ۲ | `docs/database-schema.sql` | کاملترین مرجع — تمام Enumها و ۱۸ جدول |
| ۳ | `docs/workflow-engine.md` | State machine، ۱۲ transition، Saga orchestrator |
| ۴ | `docs/api-contract.yaml` | OpenAPI 3.1 — ۷۰+ endpoint |
| ۵ | `docs/technology-stack.md` | Go/Next.js/Flutter/NATS/MinIO — با دلیل |
| ۶ | `docs/ui-ux-design.md` | Wireframe سه persona |

بقیه اسناد: `integration-layer.md`, `demo-analysis.md`, `demo-reference.html`

---

## ۴. کدهای نوشته شده (backend/)

```
backend/
├── cmd/server/main.go              # ورودی اصلی — همه routeها اینجان
├── internal/
│   ├── config/config.go            # ۱۲-factor config از env
│   ├── database/postgres.go        # Connection pool (pgx v5)
│   ├── auth/
│   │   ├── jwt.go                  # RS256 JWT — auto key-gen
│   │   └── otp.go                  # Redis OTP + rate limiting
│   ├── middleware/middleware.go     # Auth, CORS, panic recovery
│   ├── workflow/
│   │   ├── statemachine.go         # موتور State Machine عمومی
│   │   └── case_transitions.go     # ۱۲ transition case با guard/effect
│   └── handler/
│       ├── response.go             # Helperهای JSON response
│       ├── case_handler.go         # Case CRUD + capacity + submit
│       ├── map_handler.go          # Map service (۵ مرحله)
│       ├── claim_handler.go        # Claim service + docs + AI advice
│       └── cert_handler.go         # Cert service (۴ مرحله)
├── migrations/                     # ۱۲ جفت up/down migration
├── Dockerfile                      # Multi-stage build (golang:1.23 → alpine:3.20)
├── go.mod                          # Go 1.23 dependencies
└── go.sum
```

---

## ۵. زیرساخت (docker-compose.yml)

| سرویس | پورت | وضعیت |
|---|---|---|
| PostgreSQL 16 + PostGIS | `5433` (host) | ✅ |
| Redis 7 | `6379` | ✅ |
| NATS JetStream | `4222`, `8222` (mon) | ✅ |
| MinIO (S3) | `9000`, `9001` (console) | ✅ |
| Backend Go | `8080` | ✅ |

---

## ۶. وضعیت فعلی API — چیزایی که کار می‌کنه

```
✅ POST /v1/auth/otp/send       — ارسال OTP (dev_otp توی response)
✅ POST /v1/auth/otp/verify      — تایید OTP + دریافت JWT
✅ GET  /v1/auth/me              — پروفایل کاربر
✅ POST /v1/cases                — ایجاد پرونده
✅ GET  /v1/cases                — لیست پرونده‌ها
✅ GET  /v1/cases/:id            — جزئیات پرونده
✅ PATCH /v1/cases/:id           — ویرایش (فقط draft)
✅ PUT  /v1/cases/:id/capacity   — ثبت سمت و نمایندگی
✅ POST /v1/cases/:id/submit     — شروع سرویس نقشه (draft→map_in_progress)
✅ GET  /v1/map-services/:id     — جزئیات سرویس نقشه
✅ POST /v1/map-services/:id/consent        — درخواست رضایت
✅ POST /v1/map-services/:id/consent/verify — تایید رضایت
✅ POST /v1/map-services/:id/fieldwork/start — شروع عملیات میدانی
✅ POST /v1/map-services/:id/fieldwork/submit — ثبت نتیجه میدانی
✅ POST /v1/map-services/:id/submit          — ارسال به سازمان
✅ GET/POST/PATCH claim-services و cert-services (همه endpointها)
```

**چیزایی که شبیه‌سازی شدن (نیاز به integration واقعی):**
- ارسال به سازمان → وضعیت رو مستقیم approved می‌کنه
- OTP → کد توی response برمی‌گرده (dev mode)
- احراز شاهکار/ثنا → skip شده

---

## ۷. دستورات پرکاربرد

```bash
cd /Users/msl/Documents/GitHub/SabtBrooker
export PATH="/usr/local/go/bin:$HOME/go/bin:$PATH"

make docker-up          # بالا آوردن همه سرویس‌ها
make docker-down        # خاموش کردن
make migrate-up         # اجرای migrationها
make run                # اجرای backend بدون Docker
make test               # اجرای تست‌ها
make docker-up -d --build backend  # rebuild فقط backend

# تست API
curl -s http://localhost:8080/health
```

---

## ۸. کارهای انجام نشده (اولویت‌بندی شده)

| اولویت | کار | توضیح |
|---|---|---|
| 🔴 ۱ | **Frontend Next.js** | تبدیل `docs/demo-reference.html` به اپ React واقعی |
| 🔴 ۲ | **تکمیل workflow handlers** | Map service هنوز چند stage دستی نیاز داره (عکس، نقشه، جدول توصیفی) |
| 🟡 ۳ | **اپ Flutter نقشه‌بردار** | عکس‌برداری Geo-tagged + آفلاین |
| 🟡 ۴ | **Scheduler** | deadlineهای ۲ سال و ۵ ماه |
| 🟡 ۵ | **File upload واقعی** | اتصال به MinIO با Pre-signed URL |
| 🟢 ۶ | **پرداخت** | اتصال به PSP |
| 🟢 ۷ | **SMS Gateway** | ارسال واقعی پیامک |
| 🟢 ۸ | **Integration واقعی با سازمان** | منتظر مستندات از سازمان ثبت |
| 🟢 ۹ | **تست نفوذ** | OWASP ASVS Level 2 |

---

## ۹. نکات مهم فنی

1. **Go روی مسیر `/usr/local/go/bin/go`** نسخه 1.26.5 نصب شده
2. **golang-migrate CLI** با `make migrate-install` نصب میشه
3. **Postgres روی پورت 5433** (نه 5432 — conflict با Local)
4. **توکن JWT توی context به صورت string** ذخیره میشه (نه uuid.UUID)
5. **NATS non-fatal** — backend بدون NATS هم کار می‌کنه
6. **dev_otp** توی response برمی‌گرده — فقط برای محیط توسعه
7. **کاربر auto-create** میشه موقع OTP verify (با national_id موقت)
8. **guardهای identity** در dev mode سختگیرانه نیستن (شاهکار و ثنا skip)

---

## ۱۰. workflow تست کامل API

```bash
# ۱. OTP
curl -s -X POST localhost:8080/v1/auth/otp/send -H "Content-Type: application/json" -d '{"mobile":"09121112233"}'
# → dev_otp: "XXXXX"

# ۲. Verify + توکن
curl -s -X POST localhost:8080/v1/auth/otp/verify -H "Content-Type: application/json" -d '{"mobile":"09121112233","otp":"XXXXX"}'
# → access_token

# ۳. پرونده
TOKEN="..."
curl -s -X POST localhost:8080/v1/cases -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" -d '{"province":"تهران","city":"تهران","address_detail":"..."}'
# → case id

# ۴. شروع نقشه
curl -s -X POST localhost:8080/v1/cases/{id}/submit -H "Authorization: Bearer $TOKEN"
# → map_in_progress

# ۵. ادامه زنجیره با map-services و claim-services و cert-services...
```
