# SabtBrooker — سند تحویل پروژه

> آخرین به‌روزرسانی: ۲۰۲۶-۰۷-۲۹ (شروع فاز ۲)  
> فاز فعلی: Phase 2 — پیاده‌سازی Frontend  
> پیشرفت کلی: ~۶۵٪

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
آخرین commit: `fcc2cde` — `docs: finalize Handoff status for Phase 1 completion`

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

## ۴. کدهای نوشته شده

### Backend (Go)
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
├── migrations/                     # ۱۳ جفت up/down migration
├── Dockerfile                      # Multi-stage build (golang:1.25 → alpine:3.20)
├── go.mod                          # Go 1.25 dependencies
└── go.sum
```

### Frontend (Next.js)
```
frontend/
├── src/
│   ├── app/                        # App Router (Login, Dashboard)
│   ├── lib/                        # API client (Axios)
│   ├── components/                 # UI Components
│   └── globals.css                 # Theme (Navy, Brass, Paper)
├── tailwind.config.ts
└── next.config.ts
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

# Backend
cd backend && make run

# Frontend
cd frontend && npm run dev
```

---

## ۸. تسک‌های باقی‌مانده (فاز ۲ و بعد)

| اولویت | کار | وضعیت |
|---|---|---|
| 🔴 ۱ | **Frontend: ایجاد پرونده** | در حال پیاده‌سازی |
| 🔴 ۲ | **Frontend: نقشه‌برداری** | — |
| 🔴 ۳ | **تکمیل workflow handlers** | — |
| 🟡 ۴ | **اپ Flutter نقشه‌بردار** | — |
| 🟢 ۵ | **پرداخت** | — |
| 🟢 ۶ | **SMS Gateway** | — |
| 🟢 ۷ | **Integration واقعی با سازمان** | — |

---

## ۹. نکات مهم فنی

1. **Go نسخه 1.26.5** روی مک در مسیر `/usr/local/go/bin/go` نصب شده است.
2. **CI و Dockerfile** برای سازگاری با `go.mod` روی نسخه **1.25** تنظیم شده‌اند.
3. **Postgres روی پورت 5433** (نه 5432 — conflict با Local)
4. **توکن JWT توی context به صورت string** ذخیره میشه (نه uuid.UUID)
5. **Frontend** با Next.js 15 و Tailwind CSS پیاده‌سازی شده است.

---

## ۱۰. workflow تست کامل API

```bash
# ۱. OTP
curl -s -X POST localhost:8080/v1/auth/otp/send -H "Content-Type: application/json" -d '{"mobile":"09121112233"}'
# → dev_otp: "XXXXX"

# ۲. Verify + توکن
curl -s -X POST localhost:8080/v1/auth/otp/verify -H "Content-Type: application/json" -d '{"mobile":"09121112233","otp":"XXXXX"}'
# → access_token
```
