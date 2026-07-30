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

## ۳. اصلاحات موقت و تست
- **SMS Integration (MeliPayamak)**: اتصال واقعی به پنل ملی‌پیامک (وب‌سرویس BaseServiceNumber) پیاده‌سازی شد. سیستم اکنون قادر به ارسال پیامک واقعی با استفاده از پترن (BodyID) است.
- **Fixed OTP (Demo Mode)**: قابلیت ورود با کد ثابت **`1234`** در حالت `DEV_MODE=true` برای موارد اضطراری حفظ شده است.
- **Network Resiliency**: اسکریپت استقرار با قابلیت تنظیم خودکار DNS (شکن) و GOPROXY به‌روزرسانی شد.
- **Dependency Fix**: وابستگی کانتینر بک‌اِند به `openssl` اصلاح شد.

### ⚠️ نکات مهم پیکربندی
- **Docker Env Escaping**: در فایل `.env.production` برای پسوردهایی که دارای کاراکتر `$` هستند، باید از `$$` استفاده شود تا داکر آن را به عنوان متغیر سیستم تفسیر نکند (مثلاً `B96$$1`).
- **SMS Pattern**: کد پترن فعلی فعال روی سرور: `123456`.

---

## ۴. راهنمای نگهداری
برای بروزرسانی کدها و بیلد مجدد روی سرور:
```bash
cd /opt/sabtbrooker
git pull origin main
docker compose build --build-arg GOPROXY=https://goproxy.io,direct backend
docker compose up -d
```

برای مشاهده لاگ خطاها:
```bash
docker compose logs --tail=100 -f backend
```

**پایان مأموریت توسعه و استقرار.** 🚀
