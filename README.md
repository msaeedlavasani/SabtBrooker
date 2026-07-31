# SabtBrooker

سامانه هوشمند مدیریت پرونده‌های ثبتی مبتنی بر معماری Microservice

---

## وضعیت پروژه

**Status:** Production (Demo Ready)

سامانه روی VPS مستقر شده و هم‌اکنون از طریق دامنه زیر در دسترس است:

**Frontend**

https://sabt.saeedlavasani.ir

**API**

https://sabt.saeedlavasani.ir/api

---

# ویژگی‌ها

- معماری Microservice
- Backend با Go
- Frontend با Next.js
- PostgreSQL + PostGIS
- Redis
- MinIO (S3 Storage)
- NATS JetStream
- JWT Authentication
- Docker Compose
- Reverse Proxy با Nginx
- SSL (Let's Encrypt)
- PWA Ready

---

# معماری سیستم

```
                Internet
                     │
                     ▼
               Nginx (SSL)
                     │
         ┌───────────┴───────────┐
         ▼                       ▼
   Next.js Frontend        Go Backend API
                                   │
          ┌─────────────┬──────────────┬─────────────┐
          ▼             ▼              ▼             ▼
      PostgreSQL      Redis         MinIO         NATS
```

---

# تکنولوژی‌ها

## Backend

- Go
- Echo Framework
- JWT
- PostgreSQL
- Redis
- MinIO
- NATS

---

## Frontend

- Next.js
- React
- TypeScript
- Axios
- PWA

---

## Infrastructure

- Docker
- Docker Compose
- Nginx
- Let's Encrypt

---

# ساختار پروژه

```
backend/
frontend/

docker-compose.yml
deploy.sh
.env.production

README.md
HANDOFF.md
```

---

# سرویس‌های Production

| Service | Status |
|----------|--------|
| Frontend | ✅ |
| Backend | ✅ |
| PostgreSQL | ✅ |
| Redis | ✅ |
| MinIO | ✅ |
| NATS | ✅ |
| Nginx | ✅ |

---

# اجرای پروژه

## Clone

```bash
git clone <repository-url>

cd SabtBrooker
```

---

## Build

```bash
docker compose build
```

---

## Run

```bash
docker compose up -d
```

---

## مشاهده وضعیت

```bash
docker compose ps
```

---

# متغیرهای محیطی

نمونه فایل:

```
.env.production
```

نمونه تنظیمات:

```env
NEXT_PUBLIC_API_URL=https://your-domain/api

DB_HOST=postgres
DB_PORT=5432
DB_NAME=sabtbrooker
DB_USER=sabtbrooker
DB_PASSWORD=********

REDIS_ADDR=redis:6379

MINIO_ENDPOINT=minio:9000

NATS_URL=nats://nats:4222
```

---

# Deploy روی سرور

```bash
git pull

docker compose down

docker compose build --no-cache

docker compose up -d
```

---

# دستورات مفید

مشاهده سرویس‌ها

```bash
docker compose ps
```

لاگ Backend

```bash
docker compose logs -f backend
```

لاگ Frontend

```bash
docker compose logs -f frontend
```

ورود به Backend

```bash
docker exec -it sabtbrooker-backend sh
```

ورود به Frontend

```bash
docker exec -it sabtbrooker-frontend sh
```

ورود به Database

```bash
docker exec -it sabtbrooker-db psql -U sabtbrooker -d sabtbrooker
```

---

# احراز هویت

نسخه فعلی جهت ارائه (Demo) آماده شده است.

در این نسخه:

- ارسال OTP همیشه موفق است.
- کد 1234 معتبر در نظر گرفته می‌شود.
- Backend بدون بررسی واقعی OTP توکن JWT صادر می‌کند.
- در صورت نبود کاربر، به صورت خودکار ساخته می‌شود.

> این رفتار فقط برای نسخه Demo فعال است و باید قبل از انتشار نهایی به حالت استاندارد بازگردد.

---

# نکته مهم درباره Next.js

متغیر

```
NEXT_PUBLIC_API_URL
```

در زمان Build داخل فایل‌های `.next` کامپایل می‌شود.

بک‌اِند نیز آدرس را در فایل `.env.production` می‌خواند. 

---

# وضعیت فعلی پروژه

- ✅ Production Deployment
- ✅ HTTPS فعال
- ✅ Reverse Proxy
- ✅ Dockerized
- ✅ JWT Authentication
- ✅ Redis Connected
- ✅ PostgreSQL Connected
- ✅ MinIO Connected
- ✅ NATS Connected
- ✅ PWA
- ✅ Demo Login

---

# Roadmap

نسخه‌های بعدی شامل موارد زیر خواهند بود:

- پنل کارشناسان
- پنل مدیر سیستم
- مدیریت کاربران
- احراز هویت واقعی با OTP
- اتصال کامل به ملی پیامک
- داشبورد مدیریتی
- Notification Center
- Audit Log
- CI/CD Pipeline
- Backup خودکار

---

# License

Private Repository

All Rights Reserved © Mohammad Saeed Lavasani
