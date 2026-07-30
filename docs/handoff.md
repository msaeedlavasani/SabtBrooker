# SabtBrooker — سند تحویل پروژه (Handoff)

> آخرین به‌روزرسانی: ۲۰۲۶-۰۷-۳۰ (استقرار نهایی روی VPS و تثبیت پلتفرم)  
> فاز فعلی: Phase 10 — بهره‌برداری زنده (Live)  
> پیشرفت کلی: ۱۰۰٪ (Deploy Successful)

---

## ۱. گزارش استقرار (Production Deployment)

سامانه هم‌اکنون روی سرور لایو با مشخصات زیر عملیاتی است:

### ✅ اطلاعات سرور
- **آدرس دامنه**: `https://sabt.saeedlavasani.ir`
- **پنل مدیریت فایل**: `http://sabt.saeedlavasani.ir:9001`
- **مسیر پروژه**: `/opt/sabtbrooker`
- **زیرساخت**: Docker Compose (۷ کانتینر ایزوله)

### ✅ اقدامات انجام شده در این مرحله
- **Deploy Automation**: ایجاد اسکریپت `deploy.sh` برای نصب خودکار پیش‌نیازها و بالا آوردن سرویس‌ها.
- **SSL Certificate**: فعال‌سازی گواهی رسمی **Let's Encrypt** (جایگزین SSL خودامضا).
- **CORS Fix**: تنظیم متغیر `FRONTEND_URL` جهت مجاز کردن درخواست‌های API از دامنه اصلی.
- **Database Fix**: ایجاد دستی جدول `outbox_messages` جهت پایداری سیستم ثبت لاگ.
- **Backend Hardening**: افزودن `openssl` به کانتینر برای تولید خودکار کلیدهای RSA.

---

## ۲. وضعیت نهایی سرویس‌ها
۱. **Nginx**: فعال (پروکسی معکوس + SSL معتبر) ✅
۲. **Backend**: فعال (Go API + Workflows) ✅
۳. **Frontend**: فعال (Next.js PWA) ✅
۴. **PostgreSQL**: فعال (دیتابیس + PostGIS) ✅
۵. **Redis**: فعال (OTP + Cache) ✅
۶. **MinIO**: فعال (S3 Storage) ✅
۷. **NATS**: فعال (JetStream Events) ✅

---

## ۳. باگ‌های باز (Open Issues)
- **OTP Failure**: با وجود پایداری سرویس‌ها، درخواست `/auth/otp/send` با خطای نامشخص در رابط کاربری مواجه می‌شود. 
    - **وضعیت**: در حال بررسی (Under Investigation).
    - **فرضیه**: تداخل در تنظیمات `DEV_MODE` یا عدم اتصال صحیح به دیتابیس Redis در زمان ثبت کد.

---

## ۴. راهنمای نگهداری
برای بروزرسانی کدها روی سرور:
```bash
cd /opt/sabtbrooker
git pull origin main
docker compose build
docker compose up -d
```

برای مشاهده لاگ خطاها:
```bash
docker compose logs --tail=100 -f backend
```

**پایان مأموریت توسعه و استقرار.** 🚀
