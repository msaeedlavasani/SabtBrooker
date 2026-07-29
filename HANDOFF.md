# HANDOFF — وضعیت پروژه و تسک‌های باقی‌مانده

> آخرین به‌روزرسانی: ۲۰۲۶-۰۷-۲۹ (جلسه دوم پیاده‌سازی)  
> فاز فعلی: Phase 1 — تکمیل هسته (۹۰٪)  
> پیشرفت کلی: ~۵۰٪

---

## خلاصه وضعیت

### ✅ انجام شده — فاز ۱: هسته (تکمیل شد)

| بخش | وضعیت | توضیح |
|---|---|---|
| مستندات طراحی | کامل | ۹ فایل در docs/ |
| Database Schema | کامل | ۱۸ جدول + ۱۳ migration (شامل outbox) |
| CI Pipeline | کامل | GitHub Actions: lint, test, build, security scan |
| Docker Compose | کامل | PostgreSQL + Redis + NATS + MinIO + Backend |
| Config | کامل | بارگذاری از env |
| Auth Service | کامل | JWT (RS256) + OTP + Refresh Token rotation |
| State Machine | کامل | Case/Map/Claim/Cert با Transition + Guard + Effect |
| Repository Layer | کامل | ۷ repository با interface و پیاده‌سازی PostgreSQL |
| Service Layer | کامل | CaseService, MapService, ClaimService, CertService |
| Audit Logger | کامل | PostgresAuditLogRepo — ثبت خودکار در تمام transitionها |
| File Storage | کامل | MinIO: Upload, Download, Delete, Presigned URL (5min) |
| Notification Service | کامل | In-App (DB) + SMS (stub با rate limiting) + Email (stub) |
| Scheduler | کامل | Poll-based با FOR UPDATE SKIP LOCKED — deadline_2years, 5months, otp_cleanup |
| Outbox Pattern | کامل | جدول outbox_messages + Publisher (poll) + Recorder (tx-aware) |
| Handlers | کامل | بازنویسی شده با service layer + AuthHandler مجزا |
| Middleware | کامل | Auth (JWT)، CORS، Logger، Recovery، RequireRole |
| AI Advisor | پایه | Rule-based advice با disclaimer قانونی |
| تست‌ها | پایه | State machine unit tests + service business logic tests |

### 🔴 مسدود (وابسته به سازمان ثبت)

| مورد | تأثیر |
|---|---|
| مستندات فنی مرکز تبادل اطلاعات ملی | مسدودیت Integration Gateway |
| مستندات API سامانه مانا | مسدودیت GIS Service |
| متن دستورالعمل بند (پ) ماده ۱۱۴ | مسدودیت مدیریت کارشناسان و پرداخت |

---

## تسک‌های باقی‌مانده به ترتیب اولویت

### فاز ۱: تکمیل هسته ✅ (انجام شد)

| # | تسک | وضعیت |
|---|---|---|
| 1 | Repository Layer | ✅ |
| 2 | Service Layer | ✅ |
| 3 | Audit Logger | ✅ |
| 4 | File Storage (MinIO) | ✅ |
| 5 | Notification Service | ✅ |
| 6 | Scheduler | ✅ |
| 7 | Outbox Pattern | ✅ |
| 8 | بازنویسی handlerها | ✅ |
| 9 | تست‌های unit | ✅ (پایه) |

### فاز ۲: سرویس نقشه (وابسته به مانا)

| # | تسک | وابستگی |
|---|---|---|
| 10 | GIS Service (PostGIS queries, Geo-tag) | مستندات مانا 🔴 |
| 11 | Web Panel کارشناس نقشه‌بردار | #10 |
| 12 | Mobile App (Flutter) | #10 |

### فاز ۳: ادعا و گواهی

| # | تسک | وابستگی |
|---|---|---|
| 13 | پورتال متقاضی (Next.js) | — |
| 14 | AI Advisor (RAG واقعی) | — |
| 15 | Integration Gateway (اتصال به سازمان) | مستندات سازمان 🔴 |

### فاز ۴: تکمیل

| # | تسک | وابستگی |
|---|---|---|
| 16 | Payment Gateway | PSP credentials 🔴 |
| 17 | SMS Gateway واقعی | SMS provider credentials 🔴 |
| 18 | Monitoring & Alerting (Grafana) | — |
| 19 | Backup & DR | — |
| 20 | Kubernetes manifests | — |

---

## ساختار پروژه

```
SabtBrooker/
├── HANDOFF.md                      
├── README.md
├── .env.example
├── docker-compose.yml
├── Makefile
├── docs/                           ← ۹ فایل مستندات کامل
├── backend/                        ← Go microservice
│   ├── cmd/server/main.go          ← ✅ ساده‌سازی شده با service layer
│   ├── migrations/                 ← ۱۳ migration
│   └── internal/
│       ├── config/                 ← ✅
│       ├── database/               ← ✅
│       ├── auth/                   ← ✅ JWT + OTP
│       ├── middleware/             ← ✅ Auth, CORS, Logger, RequireRole
│       ├── workflow/               ← ✅ State Machine + tests
│       ├── handler/                ← ✅ Auth + Case + Map + Claim + Cert
│       ├── repository/             ← ✅ ۷ repo (user, case, map, claim, cert, audit, ai)
│       ├── service/                ← ✅ ۴ service (case, map, claim, cert) + tests
│       ├── storage/                ← ✅ MinIO
│       ├── notification/           ← ✅ SMS/In-App
│       ├── scheduler/              ← ✅ Deadline tracking + tests
│       └── outbox/                 ← ✅ Outbox Pattern
├── frontend/                       ← خالی — Next.js
└── mobile/                         ← خالی — Flutter
```

---

## تغییرات این جلسه

### فایل‌های جدید (۱۹ عدد)
- `HANDOFF.md`
- `internal/repository/`: models.go, interfaces.go, user_repo.go, case_repo.go, map_repo.go, claim_repo.go, cert_repo.go, aux_repo.go
- `internal/service/`: case_service.go, map_service.go, claim_service.go, cert_service.go
- `internal/storage/storage.go`
- `internal/notification/notification.go`
- `internal/scheduler/scheduler.go`
- `internal/outbox/outbox.go`
- `internal/handler/auth_handler.go`
- `backend/migrations/000013_outbox.up.sql` + `.down.sql`

### فایل‌های بازنویسی شده (۴ عدد)
- `cmd/server/main.go` — از direct DB به service layer مهاجرت کرد
- `internal/handler/case_handler.go`
- `internal/handler/map_handler.go`
- `internal/handler/claim_handler.go`
- `internal/handler/cert_handler.go`

### تست‌ها (۳ عدد)
- `internal/workflow/statemachine_test.go` — ۱۰ تست
- `internal/service/claim_service_test.go` — ۶ تست
- `internal/scheduler/scheduler_test.go` — ۶ تست (۴ integration skipped)

---

## راهنمای توسعه

### دستورات پرکاربرد
```bash
make docker-up     # راه‌اندازی زیرساخت
make migrate-up    # اجرای migrationها (شامل 013_outbox)
make run           # اجرای backend
make test          # اجرای تست‌ها
make build         # build binary
```

### الگوی معماری
```
Handler → Service → Repository → PostgreSQL
              ↓
    Notification / Scheduler / Audit / Outbox
```

---

## نقاط اتصال خارجی (نیاز به credentials)

| سرویس | وضعیت | توضیح |
|---|---|---|
| SMS Gateway | stub | log-based — ارائه‌دهنده پیامک ایرانی |
| Payment Gateway | نیازمند PSP | درگاه بانکی |
| شاهکار | نیازمند API | تطابق کد ملی-موبایل |
| ثنا | نیازمند API | استعلام وضعیت ثبت‌نام |
| ثبت احوال | نیازمند API | استعلام حیات/فوت |
