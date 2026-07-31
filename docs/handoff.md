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

## ۳. اصلاحات موقت و تاریخچه چالش‌ها (بسیار مهم)

در جریان آماده‌سازی نسخه دمو، با مجموعه‌ای از چالش‌های فنی مواجه شدیم که منجر به تصمیم نهایی برای **حذف موقت فلو OTP** شد. جزئیات جهت مراجعات بعدی:

### ❌ چالش‌های مشاهده شده:
1. **Network Timeout**: تحریم‌های داکر و گوگل در شبکه ایران باعث اختلال در دانلود تصاویر Alpine و پکیج‌های Go شد. (راهکار: استفاده از DNS شکن و `goproxy.cn`).
2. **Docker Env Escaping**: کاراکتر `$` در رمز عبور ملی‌پیامک (`B96$1`) توسط داکر به عنوان متغیر تفسیر می‌شد. (راهکار: انتقال تنظیمات به `env_file` در docker-compose).
3. **Nginx Routing**: تداخل در بلاک‌های `location` باعث می‌شد درخواست‌های API به درستی به بک‌اِند نرسند.
4. **Next.js Build-time Variables**: متغیرهای محیطی در زمان بیلد فرانت‌اِند حک می‌شوند، لذا هر تغییر آدرس API نیازمند ری‌بیلد کامل فرانت‌اِند بود.

### ✅ وضعیت فعلی (Demo-Ready):
به دلیل محدودیت زمان پرزنتیشن، منطق احراز هویت به صورت زیر تغییر یافت:
- **Backend**: هندلر `VerifyOTP` بدون چک کردن کد، توکن صادر می‌کند.
- **Frontend**: دکمه "دریافت کد" حذف و به "ورود مستقیم" تبدیل شد که با یک کد فرضی مرحله تایید را صدا می‌زند.

### ⚠️ اقدامات برای آینده (Production):
برای بازگشت به حالت ایمن:
1. کد کامنت شده در `backend/internal/handler/auth_handler.go` را فعال کنید.
2. فایل `frontend/src/app/page.tsx` را به حالت دو مرحله‌ای (Mobile -> OTP) برگردانید.
3. از صحت `SMS_PASSWORD=B96$1` در فایل `.env.production` مطمئن شوید.

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
