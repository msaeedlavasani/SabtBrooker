# SabtBrooker — سند تحویل پروژه (Handoff)

> آخرین به‌روزرسانی: ۲۰۲۶-۰۸-۰۲ (ایجاد زیرساخت استقرار Docker و اسکریپت deploy.sh)  
> فاز فعلی: Phase 9 — آماده استقرار Production  
> پیشرفت کلی: ۱۰۰٪ (Ready for Deploy)

---

## ۱. خلاصه اقدامات انجام شده

در این مرحله، تمام زیرساخت‌های لازم برای استقرار روی VPS تکمیل شد:

### ✅ زیرساخت Docker (Deployment Ready)
- **Dockerfile فرانت‌اِند**: ساخته شد (multi-stage build برای Next.js).
- **Dockerfile بک‌اِند**: اصلاح شد (هماهنگ‌سازی پورت‌ها + تولید خودکار کلید JWT).
- **Nginx Reverse Proxy**: فایل `nginx/nginx.conf` با Rate Limiting و Security Headers.
- **docker-compose.yml**: تکمیل شده با ۷ سرویس کامل.
- **entrypoint.sh**: اسکریپت خودکار تولید کلیدهای RSA برای JWT.

### ✅ اسکریپت استقرار یک‌خطی
- **deploy.sh**: نصب Docker، clone کد، ساخت SSL، بالا آوردن همه سرویس‌ها.
- **.env.production.example**: قالب فایل تنظیمات Production با توضیحات فارسی.

---

## ۲. دستور استقرار یک‌خطی (One-liner)
```bash
curl -fsSL https://raw.githubusercontent.com/msaeedlavasani/SabtBrooker/main/deploy.sh | bash -s -- YOUR_DOMAIN.com
```

---

## ۳. وضعیت نهایی
کلیه بخش‌های پروژه برای استقرار روی سرور Production آماده است. تنها پیش‌نیاز: تهیه VPS و دامنه.

### فایل‌های کلیدی استقرار
| فایل | کاربرد |
|---|---|
| `deploy.sh` | اسکریپت اصلی استقرار |
| `docker-compose.yml` | تعریف ۷ سرویس Docker |
| `.env.production.example` | قالب تنظیمات Production |
| `nginx/nginx.conf` | تنظیمات پروکسی معکوس |
| `backend/entrypoint.sh` | تولید خودکار کلید JWT |
| `backend/Dockerfile` | بیلد ایمیج Go |
| `frontend/Dockerfile` | بیلد ایمیج Next.js |
