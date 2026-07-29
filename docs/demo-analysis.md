# تحلیل دموی RegBroooker
## مقایسه با طراحی تفصیلی فعلی

---

## دمو چیست؟

یک فایل HTML تکی (۷۵۳ خط) — کاملاً client-side، دیپلوی شده روی GitHub Pages.  
شبیه‌ساز frontend-only از سه سرویس (نقشه، ادعا، گواهی) با فرم‌های تعاملی.

---

## وضعیت: دمو چقدر به کارمان می‌آید؟

### نقاط قوت — قابل استفاده مجدد

| مورد | وضعیت | کاربرد در پروژه جدید |
|---|---|---|
| **Stepper/Workflow UI** | ✅ کامل | نقطه شروع عالی برای Next.js frontend |
| **فرم‌های شرطی** (نمایندگی، فوت، اشخاص ثالث) | ✅ پیاده‌سازی شده | منطق visibility عیناً قابل استفاده است |
| **OTP consent flow** با هشدار قانونی | ✅ کامل | UX همین جریان، با backend واقعی |
| **شاخه rejection کارشناس** | ✅ پیاده‌سازی شده | منطق درست — توقف فرآیند |
| **زنجیره کد رهگیری** (نقشه → ادعا → گواهی) | ✅ شبیه‌سازی شده | مفهوم درست، نیاز به backend واقعی |
| **راهنمایی حقوقی (AI/انسانی)** | ✅ شبیه‌سازی شده | Template خوب برای UI |
| **طراحی بصری** (RTL، رنگ‌ها، فونت) | ✅ قابل قبول | Design tokens قابل استخراج |
| **اصطلاحات فارسی** | ✅ دقیق و حقوقی | قابل استفاده مستقیم |

### نقاط ضعف — آنچه در دمو نیست و باید ساخته شود

| مورد | وضعیت در دمو |
|---|---|
| **Backend واقعی** | ❌ وجود ندارد — فقط simulation |
| **Authentication (JWT/OTP واقعی)** | ❌ OTP شبیه‌سازی شده (نمایش کد روی صفحه!) |
| **Database** | ❌ State در memory — با refresh از بین می‌رود |
| **API integration با سازمان** | ❌ وجود ندارد |
| **GIS / Geo-tag validation** | ❌ وجود ندارد |
| **File upload واقعی** | ❌ فقط کلیک روی دکمه — toggle visual state |
| **Audit logging** | ❌ وجود ندارد |
| **Payment integration** | ❌ وجود ندارد |
| **Deadline tracking (۲ سال، ۵ ماه)** | ❌ وجود ندارد |
| **اپ موبایل نقشه‌بردار** | ❌ وجود ندارد |
| **پنل ادمین** | ❌ وجود ندارد |
| **مدیریت کارشناسان** | ❌ فیلد کد ملی — بدون احراز واقعی |
| **احراز هویت (ثنا، شاهکار، ثبت احوال)** | ❌ checkbox دستی |
| **چند کاربر هم‌زمان** | ❌ single instance |

---

## حکم نهایی

```
┌─────────────────────────────────────────────────────────────┐
│                                                              │
│   دموی RegBroooker = Prototype / Wireframe قابل استفاده     │
│                                                              │
│   ✅ UI/logic جلو (frontend) ← قابل استخراج و استفاده مجدد  │
│   ❌ Backend، Integration، Security ← صفر — باید از نو ساخت │
│                                                              │
│   نسبت دمو به محصول نهایی: ~۱۵٪                              │
│   (فقط frontend UI logic — هیچ چیز server-side)             │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

---

## برنامه پیشنهادی برای SabtBrooker

### از دمو چه برداریم

۱. **Design tokens و layout** (رنگ‌ها، فونت Vazirmatn، اندازه‌ها، RTL)
۲. **ساختار فرم‌ها** (field types، conditional visibility، validation logic)
۳. **منطق Stepper** (زنجیره سه‌سرویسه، unlock بعدی با کد رهگیری قبلی)
۴. **OTP flow UX** (ارسال، تایمر، requireAck برای هشدار قانونی)
۵. **شاخه‌های شرطی** (نمایندگی، فوت، حقوق دولتی)

### چه چیزی از صفر بسازیم

۱. **کل backend** — Go serviceها طبق [workflow-engine.md](workflow-engine.md)
۲. **پایگاه داده** — PostgreSQL طبق [database-schema.sql](database-schema.sql)
۳. **API layer** — طبق [api-contract.yaml](api-contract.yaml)
۴. **Integration** — طبق [integration-layer.md](integration-layer.md)
۵. **Next.js frontend** — بازنویسی UI دمو با React/Next.js + اتصال به API واقعی
۶. **Flutter اپ موبایل** — برای کارشناس نقشه‌بردار

### ساختار پیشنهادی SabtBrooker

```
SabtBrooker/
├── backend/                    # Go services
│   ├── cmd/server/
│   ├── internal/
│   │   ├── auth/               # JWT + OTP
│   │   ├── case/               # Case service
│   │   ├── map/                # Map service
│   │   ├── claim/              # Claim service
│   │   ├── cert/               # Certificate service
│   │   ├── integration/        # Organization adapters
│   │   ├── geo/                # GIS + EXIF validation
│   │   ├── payment/            # Payment gateway
│   │   ├── notification/       # SMS + in-app
│   │   ├── ai/                 # Legal advice AI
│   │   └── workflow/           # State machine + Saga
│   ├── migrations/
│   └── Dockerfile
├── frontend/                   # Next.js 14+
│   ├── app/
│   ├── components/
│   └── lib/
├── mobile/                     # Flutter surveyor app
├── docs/                       # All design docs
│   ├── architecture-blueprint.md
│   ├── database-schema.sql
│   ├── api-contract.yaml
│   ├── technology-stack.md
│   ├── workflow-engine.md
│   ├── integration-layer.md
│   └── ui-ux-design.md
├── docker-compose.yml          # Local dev environment
├── .github/workflows/          # CI/CD
└── README.md
```
