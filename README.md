# SabtBrooker — سامانه کارگزاری ماده ۱۰

پلتفرم کارگزاری سازمان ثبت اسناد و املاک کشور، موضوع بند (پ) ماده ۱۱۴ قانون برنامه هفتم پیشرفت.

## وضعیت پروژه

**فاز فعلی:** طراحی معماری و مستندات — آماده شروع پیاده‌سازی

## ساختار

```
SabtBrooker/
├── docs/                        # مستندات طراحی
│   ├── architecture-blueprint.md   # معماری کلی + C4 diagrams + roadmap
│   ├── database-schema.sql         # ۱۸ جدول PostgreSQL + PostGIS
│   ├── api-contract.yaml           # OpenAPI 3.1 — ۷۰+ endpoint
│   ├── technology-stack.md         # Go / Next.js / Flutter / NATS / MinIO
│   ├── workflow-engine.md          # State machine + Saga orchestrator
│   ├── integration-layer.md        # Adapter pattern + Circuit Breaker + Outbox
│   ├── ui-ux-design.md             # Wireframe سه persona
│   ├── demo-analysis.md            # تحلیل دموی قبلی (RegBroooker)
│   └── demo-reference.html         # دموی مرجع (frontend-only)
├── backend/                    # Go microservices (به‌زودی)
├── frontend/                   # Next.js 14+ (به‌زودی)
├── mobile/                     # Flutter surveyor app (به‌زودی)
└── docker-compose.yml          # محیط توسعه (به‌زودی)
```

## پشته فنی (برنامه‌ریزی شده)

| لایه | انتخاب |
|---|---|
| Backend | Go (Echo) |
| Frontend | Next.js 14+ (React) |
| Mobile | Flutter |
| Database | PostgreSQL 16 + PostGIS |
| Cache | Redis Sentinel |
| Message Broker | NATS JetStream |
| Storage | MinIO (S3-compatible) |
| API Gateway | Traefik |
| Monitoring | Grafana LGTM |
| CI/CD | GitLab CI + ArgoCD |

## پیش‌نیازهای شروع پیاده‌سازی

سه مدرک از سازمان ثبت باید اخذ شود:

1. **مستندات فنی مرکز تبادل اطلاعات ملی** — پروتکل، فرمت پیام، احراز هویت
2. **مستندات API/فرمت سامانه مانا** — schema نقشه و جدول توصیفی
3. **متن کامل دستورالعمل بند (پ) ماده ۱۱۴** — فهرست کارشناسان، تعرفه‌ها

## دموی مرجع

دموی [RegBroooker](https://github.com/msaeedlavasani/RegBroooker) یک prototype frontend-only است که گردش‌کار سه‌سرویسه را شبیه‌سازی می‌کند.
فایل [docs/demo-reference.html](docs/demo-reference.html) برای مرجع UI نگه‌داری شده است.

## مجوز

Private — کلیه حقوق محفوظ
