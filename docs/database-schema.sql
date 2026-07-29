-- ============================================================================
-- سامانه کارگزاری ماده ۱۰ — Schema پایگاه داده
-- PostgreSQL 16+
-- ============================================================================

-- ----------------------------------------------------------------------------
-- ۰. افزونه‌های مورد نیاز
-- ----------------------------------------------------------------------------
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";          -- برای pgp_sym_encrypt
CREATE EXTENSION IF NOT EXISTS "btree_gin";          -- برای composite index
CREATE EXTENSION IF NOT EXISTS "postgis";             -- برای داده‌های مکانی (GIS)

-- ----------------------------------------------------------------------------
-- ۱. ENUM Types
-- ----------------------------------------------------------------------------

-- وضعیت کلی پرونده
CREATE TYPE case_status AS ENUM (
    'draft',                -- پیش‌نویس اولیه
    'map_in_progress',      -- در حال تهیه نقشه
    'map_completed',        -- نقشه تکمیل شد
    'claim_in_progress',    -- در حال درج ادعا
    'claim_completed',      -- ادعا ثبت شد
    'cert_in_progress',     -- در حال درج گواهی اقدام
    'cert_completed',       -- گواهی اقدام ثبت شد (پایان زنجیره)
    'expired',              -- منقضی شده (گذشت مهلت ۲ ساله)
    'rejected',             -- رد شده
    'cancelled'             -- انصراف متقاضی
);

-- وضعیت سرویس نقشه
CREATE TYPE map_service_status AS ENUM (
    'pending_expert_assignment',    -- در انتظار تخصیص کارشناس
    'expert_assigned',              -- کارشناس تخصیص یافت
    'fieldwork_in_progress',        -- عملیات میدانی در حال انجام
    'fieldwork_done',               -- عملیات میدانی انجام شد
    'submitted_to_org',             -- ارسال به سازمان
    'approved',                     -- تایید توسط سازمان
    'rejected'                      -- رد توسط سازمان
);

-- وضعیت سرویس درج ادعا
CREATE TYPE claim_service_status AS ENUM (
    'pending_expert_assignment',
    'expert_assigned',
    'documents_verified',           -- مستندات توسط کارشناس تایید شد
    'submitted_to_org',
    'approved',
    'rejected'
);

-- وضعیت سرویس گواهی اقدام
CREATE TYPE cert_service_status AS ENUM (
    'pending_data',                 -- در انتظار تکمیل داده‌ها
    'submitted_to_org',
    'approved',
    'rejected'
);

-- نقش کاربری
CREATE TYPE user_role AS ENUM (
    'applicant',            -- متقاضی
    'legal_expert',         -- کارشناس امور ثبتی و حقوقی
    'survey_expert',        -- کارشناس نقشه‌بردار
    'admin',                -- ادمین کارگزاری
    'auditor'               -- حسابرس (دسترسی فقط خواندنی)
);

-- سمت متقاضی
CREATE TYPE applicant_capacity AS ENUM (
    'principal',                    -- اصیل
    'legal_rep_natural',            -- نماینده قانونی شخص حقیقی
    'legal_rep_legal'               -- نماینده قانونی شخص حقوقی
);

-- نوع ادعا
CREATE TYPE claim_type AS ENUM (
    'ownership',            -- مالکیت عین
    'easement',             -- حق ارتفاق
    'usufruct',             -- حق انتفاع
    'benefits_ownership'    -- مالکیت منافع
);

-- نوع مالکیت مورد ادعا
CREATE TYPE ownership_type AS ENUM (
    'land',                 -- عرصه
    'building',             -- اعیان
    'land_and_building'     -- عرصه و اعیان
);

-- نوع مستند ادعا
CREATE TYPE document_type AS ENUM (
    'sales_agreement',      -- مبایعه‌نامه
    'settlement',           -- صلح‌نامه
    'promissory_note',      -- قولنامه
    'partition',            -- تقسیم‌نامه
    'gift',                 -- هبه‌نامه
    'court_ruling',         -- آرای محاکم
    'cultivation_right',    -- نسق زارعانه
    'witness_testimony',    -- استشهادیه
    'inheritance_cert',     -- گواهی انحصار وراثت
    'other'                 -- سایر
);

-- مرجع اقدام (برای گواهی اقدام)
CREATE TYPE action_reference AS ENUM (
    'registration_org',     -- سازمان ثبت
    'judiciary',            -- قوه قضاییه
    'organization_board',   -- هیئت سامان‌دهی
    'determination_board'   -- هیئت تعیین تکلیف
);

-- نوع اقدام (برای گواهی اقدام)
CREATE TYPE action_type AS ENUM (
    'lawsuit',                  -- طرح دعوا در مراجع قضایی
    'initial_registration',     -- ثبت اولیه
    'organization_request',     -- طرح تقاضا در هیئت سامان‌دهی
    'determination_request'     -- طرح تقاضا در هیئت تعیین تکلیف
);

-- وضعیت پرداخت
CREATE TYPE payment_status AS ENUM (
    'pending',              -- در انتظار پرداخت
    'paid',                 -- پرداخت شده
    'partially_refunded',   -- بخشی مسترد شد
    'fully_refunded',       -- کاملاً مسترد شد
    'failed'                -- ناموفق
);

-- نوع پرداخت
CREATE TYPE payment_type AS ENUM (
    'advance',              -- علی‌الحساب
    'final_settlement',     -- تسویه نهایی
    'refund'                -- استرداد
);

-- کانال اطلاع‌رسانی
CREATE TYPE notification_channel AS ENUM (
    'sms',
    'in_app',
    'email'
);

-- وضعیت اطلاع‌رسانی
CREATE TYPE notification_status AS ENUM (
    'pending',
    'sent',
    'delivered',
    'failed'
);

-- نوع رویداد Audit Log
CREATE TYPE audit_event_type AS ENUM (
    'case.created',
    'case.status_changed',
    'case.expired',
    'case.cancelled',
    'map.expert_assigned',
    'map.fieldwork_started',
    'map.fieldwork_completed',
    'map.submitted_to_org',
    'map.approved',
    'map.rejected',
    'claim.expert_assigned',
    'claim.documents_verified',
    'claim.submitted_to_org',
    'claim.approved',
    'claim.rejected',
    'cert.submitted_to_org',
    'cert.approved',
    'cert.rejected',
    'payment.created',
    'payment.paid',
    'payment.refunded',
    'expert.created',
    'expert.updated',
    'expert.deactivated',
    'user.login',
    'user.login_failed',
    'otp.sent',
    'otp.verified',
    'otp.failed',
    'ai_advice.requested',
    'document.uploaded',
    'document.verified'
);


-- ----------------------------------------------------------------------------
-- ۲. جداول اصلی
-- ----------------------------------------------------------------------------

-- ============================================================
-- ۲.۱ کاربران (متقاضیان + کارشناسان + ادمین)
-- ============================================================
CREATE TABLE users (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    national_id         VARCHAR(10)  NOT NULL UNIQUE,       -- کد ملی
    first_name          VARCHAR(100) NOT NULL,
    last_name           VARCHAR(100) NOT NULL,
    mobile              VARCHAR(11)  NOT NULL UNIQUE,       -- 09xxxxxxxxx
    mobile_verified     BOOLEAN      NOT NULL DEFAULT FALSE,
    birth_date          DATE,                               -- تاریخ تولد (برای محاسبه سن)
    role                user_role    NOT NULL DEFAULT 'applicant',

    -- وضعیت‌های استعلامی (cache شده)
    sana_status         VARCHAR(20),                        -- ثنا: active / inactive / unknown
    sana_checked_at     TIMESTAMPTZ,
    ncr_mobile_match    BOOLEAN,                            -- تطابق کد ملی و شماره از شاهکار
    ncr_checked_at      TIMESTAMPTZ,
    is_alive            BOOLEAN      DEFAULT TRUE,          -- احراز حیات از ثبت احوال
    alive_checked_at    TIMESTAMPTZ,

    -- امنیتی
    password_hash       VARCHAR(255),                       -- bcrypt hash (nullable برای OTP-only)
    failed_login_count  INT          NOT NULL DEFAULT 0,
    locked_until        TIMESTAMPTZ,
    last_login_at       TIMESTAMPTZ,
    last_login_ip       INET,

    -- وضعیت
    is_active           BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Indexها
CREATE INDEX idx_users_national_id ON users (national_id);
CREATE INDEX idx_users_mobile ON users (mobile);
CREATE INDEX idx_users_role ON users (role);
CREATE INDEX idx_users_sana_status ON users (sana_status) WHERE sana_status = 'inactive';

COMMENT ON TABLE users IS 'تمامی کاربران سامانه: متقاضیان، کارشناسان، ادمین';
COMMENT ON COLUMN users.sana_status IS 'کش یک‌ساعته — وضعیت ثبت‌نام در سامانه ثنا';
COMMENT ON COLUMN users.ncr_mobile_match IS 'کش یک‌روزه — تطابق کد ملی و شماره موبایل از شاهکار';


-- ============================================================
-- ۲.۲ کارشناسان (جزئیات تخصصی)
-- ============================================================
CREATE TABLE experts (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id             UUID         NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    expert_type         VARCHAR(20)  NOT NULL CHECK (expert_type IN ('legal', 'survey')),

    -- اطلاعات تخصصی
    license_number      VARCHAR(50),                        -- شماره پروانه/مجوز
    license_expiry      DATE,
    organization        VARCHAR(200),                       -- سازمان/شرکت محل اشتغال
    education_level     VARCHAR(50),
    field_of_study      VARCHAR(100),
    years_of_experience INT,

    -- وضعیت در فهرست کارشناسان مجاز (دستورالعمل ماده ۴/۵)
    in_approved_list    BOOLEAN      NOT NULL DEFAULT FALSE,
    approved_list_ref   VARCHAR(100),                       -- شماره ثبت در فهرست مجاز

    -- مدیریتی
    max_active_cases    INT          NOT NULL DEFAULT 10,   -- حداکثر پرونده هم‌زمان
    current_case_count  INT          NOT NULL DEFAULT 0,
    is_available        BOOLEAN      NOT NULL DEFAULT TRUE,

    created_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_experts_type ON experts (expert_type);
CREATE INDEX idx_experts_available ON experts (is_available, expert_type) WHERE is_available = TRUE;
CREATE INDEX idx_experts_user ON experts (user_id);

COMMENT ON TABLE experts IS 'کارشناسان: ثبتی-حقوقی (legal) و نقشه‌بردار (survey)';


-- ============================================================
-- ۲.۳ OTP Sessions
-- ============================================================
CREATE TABLE otp_sessions (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id         UUID         REFERENCES users(id) ON DELETE CASCADE,
    mobile          VARCHAR(11)  NOT NULL,
    otp_hash        VARCHAR(255) NOT NULL,                  -- SHA-256(otp + salt)
    purpose         VARCHAR(50)  NOT NULL,                  -- 'auth', 'consent', 'warning'
    attempts        INT          NOT NULL DEFAULT 0,
    max_attempts    INT          NOT NULL DEFAULT 3,
    expires_at      TIMESTAMPTZ  NOT NULL,                  -- معمولاً ۲ دقیقه
    verified_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_otp_mobile_expires ON otp_sessions (mobile, expires_at) WHERE verified_at IS NULL;
CREATE INDEX idx_otp_user ON otp_sessions (user_id);

COMMENT ON TABLE otp_sessions IS 'جلسات رمز یکبارمصرف — auth, consent, warning';


-- ============================================================
-- ۲.۴ Refresh Tokens
-- ============================================================
CREATE TABLE refresh_tokens (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id         UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash      VARCHAR(255) NOT NULL UNIQUE,           -- SHA-256(token)
    family          UUID         NOT NULL,                  -- rotation family
    device_info     VARCHAR(500),
    ip_address      INET,
    expires_at      TIMESTAMPTZ  NOT NULL,                  -- معمولاً ۷ روز
    revoked_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_refresh_token_user ON refresh_tokens (user_id, revoked_at);
CREATE INDEX idx_refresh_token_family ON refresh_tokens (family);

COMMENT ON TABLE refresh_tokens IS 'Refresh Token با قابلیت rotation و تشخیص reuse';


-- ============================================================
-- ۲.۵ پرونده (Case)
-- ============================================================
CREATE TABLE cases (
    id                      UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    applicant_id            UUID           NOT NULL REFERENCES users(id),

    -- احراز نمایندگی
    applicant_capacity      applicant_capacity NOT NULL DEFAULT 'principal',
    -- فیلدهای نماینده قانونی شخص حقوقی
    legal_entity_id         VARCHAR(11),                    -- شناسه ملی شخص حقوقی
    legal_entity_name       VARCHAR(300),
    -- فیلدهای مشترک نماینده
    proxy_document_type     VARCHAR(50),                    -- نوع سند رسمی نمایندگی
    proxy_document_id       VARCHAR(100),                   -- شناسه سند رسمی
    proxy_document_date     DATE,                           -- تاریخ تنظیم سند
    proxy_verification_code VARCHAR(20),                    -- رمز تصدیق
    proxy_verified          BOOLEAN      DEFAULT FALSE,     -- تایید نهایی توسط کارشناس
    proxy_verified_by       UUID         REFERENCES experts(id),
    proxy_verification_note TEXT,

    -- وضعیت
    status                  case_status  NOT NULL DEFAULT 'draft',

    -- مهلت‌های قانونی
    claim_approved_at       TIMESTAMPTZ,                    -- تاریخ ثبت ادعا → مبنای deadline_2years
    deadline_2years         TIMESTAMPTZ,                    -- claim_approved_at + 2 years
    applicant_deceased      BOOLEAN      DEFAULT FALSE,     -- فوت متقاضی
    deceased_national_id    VARCHAR(10),                    -- کد ملی متوفی
    date_of_death           DATE,                           -- تاریخ فوت
    deadline_5months        TIMESTAMPTZ,                    -- max(date_of_death + 5m, original_deadline)

    -- آدرس ملک
    province                VARCHAR(100) NOT NULL,
    city                    VARCHAR(100) NOT NULL,
    district                VARCHAR(100),
    village                 VARCHAR(100),
    postal_code             VARCHAR(10),
    address_detail          TEXT,

    -- ردیابی کدها
    map_tracking_code       VARCHAR(50),
    claim_tracking_code     VARCHAR(50),
    cert_tracking_code      VARCHAR(50),

    -- تخصیص کارشناس
    legal_expert_id         UUID         REFERENCES experts(id),
    survey_expert_id        UUID         REFERENCES experts(id),

    -- ابرداده
    created_at              TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    completed_at            TIMESTAMPTZ
);

-- Indexها
CREATE INDEX idx_cases_applicant ON cases (applicant_id);
CREATE INDEX idx_cases_status ON cases (status);
CREATE INDEX idx_cases_deadline ON cases (deadline_2years) WHERE status = 'claim_completed';
CREATE INDEX idx_cases_deadline_5m ON cases (deadline_5months) WHERE status = 'cert_in_progress';
CREATE INDEX idx_cases_experts ON cases (legal_expert_id, survey_expert_id);
CREATE INDEX idx_cases_tracking ON cases (map_tracking_code, claim_tracking_code, cert_tracking_code);
CREATE INDEX idx_cases_created ON cases (created_at DESC);

COMMENT ON TABLE cases IS 'پرونده — موجودیت مرکزی که کل گردش‌کار سه‌مرحله‌ای را ردیابی می‌کند';


-- ============================================================
-- ۲.۶ سرویس تهیه نقشه ثبتی
-- ============================================================
CREATE TABLE map_services (
    id                      UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    case_id                 UUID              NOT NULL REFERENCES cases(id) ON DELETE CASCADE,

    status                  map_service_status NOT NULL DEFAULT 'pending_expert_assignment',

    -- اطلاعات ملک (مرحله ۱)
    property_type           VARCHAR(50),                    -- مسکونی/تجاری/زراعی/...
    approx_area_sqm         DECIMAL(10,2),                  -- مساحت تقریبی (مترمربع)
    land_use                VARCHAR(50),                    -- کاربری
    ownership_type          VARCHAR(50),                    -- نوع مالکیت
    has_building            BOOLEAN       DEFAULT FALSE,    -- وجود اعیانی
    annex_count             INT           DEFAULT 0,        -- تعداد منضمات

    -- مختصات جغرافیایی
    geo_latitude            DECIMAL(10,7),
    geo_longitude           DECIMAL(10,7),
    geo_location            GEOGRAPHY(POINT, 4326),         -- PostGIS point

    -- دسترسی به اشخاص ثالث
    grant_access_to_others  BOOLEAN       DEFAULT FALSE,
    access_granted_to       JSONB,                          -- [{national_id, name, relation}]

    -- رضایت (مرحله ۲)
    consent_otp_id          UUID          REFERENCES otp_sessions(id),
    consent_granted_at      TIMESTAMPTZ,

    -- عملیات میدانی (مرحله ۳)
    fieldwork_started_at    TIMESTAMPTZ,
    fieldwork_completed_at  TIMESTAMPTZ,

    -- خروجی
    map_file_path           VARCHAR(500),                   -- مسیر فایل نقشه در Object Storage
    map_format              VARCHAR(20),                    -- dxf / dwg / geojson
    descriptive_table       JSONB,                          -- جدول توصیفی طبق فرمت مانا
    submitted_to_org_at     TIMESTAMPTZ,

    -- پاسخ سازمان (مرحله ۴)
    tracking_code           VARCHAR(50),                    -- کد رهگیری نقشه (صادره از سازمان)
    org_response_at         TIMESTAMPTZ,
    org_rejection_reason    TEXT,
    org_response_raw        JSONB,                          -- پاسخ خام سازمان برای دیباگ

    created_at              TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_map_services_case ON map_services (case_id);
CREATE INDEX idx_map_services_status ON map_services (status);
CREATE INDEX idx_map_services_tracking ON map_services (tracking_code);
CREATE INDEX idx_map_services_geo ON map_services USING GIST (geo_location);

COMMENT ON TABLE map_services IS 'سرویس تهیه نقشه ثبتی — ۵ مرحله';


-- ============================================================
-- ۲.۷ عکس‌های میدانی نقشه
-- ============================================================
CREATE TABLE map_photos (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    map_service_id  UUID         NOT NULL REFERENCES map_services(id) ON DELETE CASCADE,
    file_path       VARCHAR(500) NOT NULL,                  -- مسیر در Object Storage
    side            VARCHAR(20)  NOT NULL CHECK (side IN ('north', 'south', 'east', 'west')),
    photo_latitude  DECIMAL(10,7),
    photo_longitude DECIMAL(10,7),
    photo_location  GEOGRAPHY(POINT, 4326),
    photo_taken_at  TIMESTAMPTZ,                            -- زمان واقعی عکس (EXIF)
    exif_valid      BOOLEAN      DEFAULT FALSE,             -- اعتبارسنجی Geo-tag
    exif_validation_note TEXT,
    file_size       BIGINT,                                 -- بایت
    checksum_sha256 VARCHAR(64),
    uploaded_by     UUID         REFERENCES users(id),

    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_map_photos_service ON map_photos (map_service_id);
CREATE INDEX idx_map_photos_geo ON map_photos USING GIST (photo_location);

COMMENT ON TABLE map_photos IS '۴ عکس از ۴ ضلع ملک با Geo-tag — اعتبارسنجی مختصات در برابر مختصات اعلامی';


-- ============================================================
-- ۲.۸ سرویس درج ادعا
-- ============================================================
CREATE TABLE claim_services (
    id                      UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    case_id                 UUID                NOT NULL REFERENCES cases(id) ON DELETE CASCADE,

    status                  claim_service_status NOT NULL DEFAULT 'pending_expert_assignment',

    -- داده‌های اولیه (مرحله ۱)
    map_tracking_code       VARCHAR(50),                    -- کد رهگیری نقشه (اعتبارسنجی می‌شود)
    map_tracking_valid      BOOLEAN       DEFAULT FALSE,

    -- رضایت + هشدار (مرحله ۲)
    consent_otp_id          UUID          REFERENCES otp_sessions(id),
    false_claim_warning_sent BOOLEAN      DEFAULT FALSE,    -- پیامک هشدار ادعای واهی
    consent_granted_at      TIMESTAMPTZ,

    -- اطلاعات ادعا (مرحله ۳)
    claim_type              claim_type,                     -- نوع ادعا
    ownership_type          ownership_type,                 -- عرصه/اعیان/عرصه و اعیان
    -- اطلاعات پلاک ثبتی
    main_plate_number       VARCHAR(50),                    -- پلاک ثبتی اصلی
    sub_plate_number        VARCHAR(50),                    -- پلاک ثبتی فرعی
    plate_section           VARCHAR(50),                    -- بخش
    -- کسر سهم (برای مالکیت عین)
    total_share             INT,                            -- سهم کل (مخرج)
    partial_share           INT,                            -- سهم جزء (صورت)

    -- response سازمان (مرحله ۴)
    submitted_to_org_at     TIMESTAMPTZ,
    tracking_code           VARCHAR(50),                    -- کد رهگیری ادعا
    org_response_at         TIMESTAMPTZ,
    org_rejection_reason    TEXT,
    org_response_raw        JSONB,

    -- حقوق دولتی
    has_government_rights   BOOLEAN      DEFAULT FALSE,
    treasury_payment_ref    VARCHAR(100),                   -- شماره واریز به خزانه

    -- راهنمایی ثبتی (مرحله ۵)
    legal_advice_requested  BOOLEAN      DEFAULT FALSE,
    legal_advice_method     VARCHAR(20)  DEFAULT 'human',   -- 'human' / 'ai' / 'both'

    created_at              TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_claim_services_case ON claim_services (case_id);
CREATE INDEX idx_claim_services_status ON claim_services (status);
CREATE INDEX idx_claim_services_tracking ON claim_services (tracking_code);
CREATE INDEX idx_claim_services_map_tracking ON claim_services (map_tracking_code);

COMMENT ON TABLE claim_services IS 'سرویس درج ادعا — ۵ مرحله';


-- ============================================================
-- ۲.۹ مستندات ادعا
-- ============================================================
CREATE TABLE claim_documents (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    claim_service_id UUID        NOT NULL REFERENCES claim_services(id) ON DELETE CASCADE,
    doc_type        document_type NOT NULL,
    file_path       VARCHAR(500) NOT NULL,
    file_size       BIGINT,
    checksum_sha256 VARCHAR(64),
    description     TEXT,                                   -- توضیحات متقاضی درباره سند
    verified_by     UUID         REFERENCES experts(id),    -- کارشناس تاییدکننده
    verified_at     TIMESTAMPTZ,
    verification_note TEXT,

    uploaded_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_claim_docs_service ON claim_documents (claim_service_id);
CREATE INDEX idx_claim_docs_type ON claim_documents (doc_type);

COMMENT ON TABLE claim_documents IS 'مستندات پشتیبان ادعا — مبایعه‌نامه، قولنامه، آرا محاکم و...';


-- ============================================================
-- ۲.۱۰ سرویس درج گواهی اقدام
-- ============================================================
CREATE TABLE cert_services (
    id                      UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    case_id                 UUID                NOT NULL REFERENCES cases(id) ON DELETE CASCADE,

    status                  cert_service_status NOT NULL DEFAULT 'pending_data',

    -- داده‌های اولیه (مرحله ۱)
    claim_tracking_code     VARCHAR(50),                    -- کد رهگیری ادعای مرتبط
    claim_tracking_valid    BOOLEAN       DEFAULT FALSE,    -- اعتبارسنجی (کمتر از ۲ سال)

    -- رضایت (مرحله ۲)
    consent_otp_id          UUID          REFERENCES otp_sessions(id),
    consent_granted_at      TIMESTAMPTZ,

    -- گواهی اقدام (مرحله ۳)
    action_reference        action_reference,               -- مرجع اقدام
    action_type             action_type,                    -- نوع اقدام
    action_date             DATE,                           -- تاریخ صدور گواهی اقدام
    cert_image_path         VARCHAR(500),                   -- تصویر گواهی اقدام (اجباری)
    cert_unique_id          VARCHAR(100),                   -- شناسه یکتای گواهی (اختیاری)

    -- ارسال به سازمان (مرحله ۴)
    submitted_to_org_at     TIMESTAMPTZ,
    tracking_code           VARCHAR(50),                    -- کد رهگیری گواهی اقدام (خروجی نهایی)
    org_response_at         TIMESTAMPTZ,
    org_rejection_reason    TEXT,
    org_response_raw        JSONB,

    created_at              TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_cert_services_case ON cert_services (case_id);
CREATE INDEX idx_cert_services_status ON cert_services (status);
CREATE INDEX idx_cert_services_tracking ON cert_services (tracking_code);

COMMENT ON TABLE cert_services IS 'سرویس درج گواهی اقدام — ۴ مرحله';


-- ============================================================
-- ۲.۱۱ پرداخت‌ها
-- ============================================================
CREATE TABLE payments (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    case_id             UUID            NOT NULL REFERENCES cases(id),
    service_type        VARCHAR(20)     NOT NULL CHECK (service_type IN ('map', 'claim', 'cert', 'legal_advice')),

    amount              BIGINT          NOT NULL,           -- مبلغ به ریال
    payment_type        payment_type    NOT NULL,
    status              payment_status  NOT NULL DEFAULT 'pending',

    -- درگاه پرداخت
    psp_reference       VARCHAR(100),                       -- کد پیگیری PSP
    psp_token           VARCHAR(500),                       -- توکن پرداخت
    payment_url         TEXT,                               -- URL درگاه
    paid_at             TIMESTAMPTZ,

    -- استرداد
    refund_amount       BIGINT,
    refund_reference    VARCHAR(100),
    refunded_at         TIMESTAMPTZ,
    refund_reason       TEXT,

    -- تعرفه
    tariff_version      VARCHAR(20),                        -- نسخه تعرفه مبنای محاسبه

    -- ابرداده
    meta                JSONB,
    created_at          TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_payments_case ON payments (case_id);
CREATE INDEX idx_payments_status ON payments (status);
CREATE INDEX idx_payments_psp ON payments (psp_reference);

COMMENT ON TABLE payments IS 'تمامی تراکنش‌های مالی: علی‌الحساب، تسویه نهایی، استرداد';


-- ============================================================
-- ۲.۱۲ تعرفه‌ها
-- ============================================================
CREATE TABLE tariffs (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    service_type    VARCHAR(20)  NOT NULL,                  -- 'map', 'claim', 'cert', 'legal_advice'
    max_amount      BIGINT       NOT NULL,                  -- سقف تعرفه به ریال
    non_refundable  BIGINT       DEFAULT 0,                 -- بخش غیرقابل استرداد
    effective_from  DATE         NOT NULL,
    effective_to    DATE,                                   -- NULL = تا اطلاع ثانوی
    version         VARCHAR(20)  NOT NULL,
    description     TEXT,
    created_by      UUID         REFERENCES users(id),
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tariffs_service ON tariffs (service_type, effective_from DESC);
CREATE UNIQUE INDEX idx_tariffs_active ON tariffs (service_type) WHERE effective_to IS NULL;

COMMENT ON TABLE tariffs IS 'تعرفه‌های اعلامی سازمان — فقط آخرین نسخه فعال معتبر است';


-- ============================================================
-- ۲.۱۳ اطلاع‌رسانی‌ها
-- ============================================================
CREATE TABLE notifications (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id         UUID                NOT NULL REFERENCES users(id),
    case_id         UUID                REFERENCES cases(id),
    channel         notification_channel NOT NULL,
    template_key    VARCHAR(50)         NOT NULL,           -- کلید قالب پیام
    content         TEXT                NOT NULL,           -- متن نهایی ارسالی
    status          notification_status NOT NULL DEFAULT 'pending',
    provider_ref    VARCHAR(200),                           -- reference ID از ارائه‌دهنده پیامک
    sent_at         TIMESTAMPTZ,
    delivered_at    TIMESTAMPTZ,
    error_message   TEXT,
    retry_count     INT                 DEFAULT 0,
    max_retries     INT                 DEFAULT 3,
    created_at      TIMESTAMPTZ         NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notifications_user ON notifications (user_id, created_at DESC);
CREATE INDEX idx_notifications_status ON notifications (status) WHERE status = 'pending';
CREATE INDEX idx_notifications_case ON notifications (case_id);

COMMENT ON TABLE notifications IS 'تمامی پیامک‌ها، نوتیفیکیشن‌ها و ایمیل‌های ارسالی';


-- ============================================================
-- ۲.۱۴ راهنمایی حقوقی مبتنی بر AI
-- ============================================================
CREATE TABLE ai_advice_logs (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    case_id             UUID         NOT NULL REFERENCES cases(id),
    claim_service_id    UUID         NOT NULL REFERENCES claim_services(id),

    -- ورودی
    input_context       JSONB        NOT NULL,              -- context ارسالی به AI (claim_type, docs, etc.)
    input_documents     TEXT[],                              -- خلاصه مستندات

    -- خروجی
    recommended_action  VARCHAR(100),                       -- اقدام پیشنهادی
    legal_references    TEXT[],                              -- مواد قانونی مرتبط
    confidence_score    DECIMAL(3,2),                       -- 0.00 تا 1.00
    raw_ai_response     TEXT,                               -- پاسخ خام برای audit

    -- ارزیابی
    was_helpful         BOOLEAN,                            -- بازخورد کاربر
    escalated_to_human  BOOLEAN      DEFAULT FALSE,         -- ارجاع به کارشناس

    -- مدل
    model_version       VARCHAR(50),
    prompt_version      VARCHAR(50),
    latency_ms          INT,

    created_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ai_advice_case ON ai_advice_logs (case_id);

COMMENT ON TABLE ai_advice_logs IS 'لاگ راهنمایی هوش مصنوعی — صرفاً مشاوره، نه نظر رسمی حقوقی';


-- ============================================================
-- ۲.۱۵ لاگ ممیزی (Audit Log) — Append-Only
-- ============================================================
CREATE TABLE audit_logs (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    event_type      audit_event_type NOT NULL,
    event_timestamp TIMESTAMPTZ      NOT NULL DEFAULT NOW(),

    -- بازیگر
    actor_type      VARCHAR(20)      NOT NULL,              -- 'applicant' / 'expert' / 'admin' / 'system'
    actor_id        UUID             REFERENCES users(id),
    actor_ip        INET,
    actor_user_agent TEXT,

    -- منبع
    resource_type   VARCHAR(50)      NOT NULL,              -- 'case' / 'map_service' / 'claim_service' / 'cert_service' / 'payment'
    resource_id     UUID,

    -- تغییرات
    changes         JSONB,                                  -- {field, from, to}[]
    metadata        JSONB,                                  -- extra context

    -- همبستگی
    correlation_id  UUID,                                   -- برای ردیابی یک تراکنش در کل سیستم
    trace_id        VARCHAR(64)                             -- OpenTelemetry trace ID
);

-- Indexها
CREATE INDEX idx_audit_timestamp ON audit_logs (event_timestamp DESC);
CREATE INDEX idx_audit_event ON audit_logs (event_type, event_timestamp DESC);
CREATE INDEX idx_audit_actor ON audit_logs (actor_id, event_timestamp DESC);
CREATE INDEX idx_audit_resource ON audit_logs (resource_type, resource_id);
CREATE INDEX idx_audit_correlation ON audit_logs (correlation_id);

-- سیاست: جدول append-only — حتی ادمین هم نمی‌تواند رکورد را حذف یا ویرایش کند
-- (با REVOKE DELETE, UPDATE ON audit_logs FROM all roles اعمال می‌شود)

COMMENT ON TABLE audit_logs IS 'لاگ ممیزی — Immutable, Append-Only — قابل export برای سازمان';


-- ============================================================
-- ۲.۱۶ یکپارچه‌سازی با سازمان (Integration Log)
-- ============================================================
CREATE TABLE integration_logs (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    case_id             UUID         REFERENCES cases(id),
    service_type        VARCHAR(20)  NOT NULL,              -- 'map' / 'claim' / 'cert'
    direction           VARCHAR(10)  NOT NULL,              -- 'outbound' / 'inbound'
    endpoint            VARCHAR(200) NOT NULL,              -- URL/endpoint مقصد
    request_payload     JSONB,                              -- payload ارسالی (حذف داده حساس)
    response_payload    JSONB,                              -- payload دریافتی
    response_http_code  INT,
    success             BOOLEAN      NOT NULL,
    error_message       TEXT,
    latency_ms          INT,
    retry_count         INT          DEFAULT 0,
    correlation_id      UUID,

    created_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_intlog_case ON integration_logs (case_id, created_at DESC);
CREATE INDEX idx_intlog_success ON integration_logs (success, created_at DESC) WHERE success = FALSE;
CREATE INDEX idx_intlog_service ON integration_logs (service_type, created_at DESC);

COMMENT ON TABLE integration_logs IS 'لاگ تمام ارتباطات با سازمان — برای دیباگ و مانیتورینگ';


-- ============================================================
-- ۲.۱۷ Scheduled Jobs (ردیابی مهلت‌ها)
-- ============================================================
CREATE TABLE scheduled_jobs (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    job_key         VARCHAR(100) NOT NULL UNIQUE,           -- 'deadline_2years_case_xxx'
    job_type        VARCHAR(50)  NOT NULL,                  -- 'deadline_2years' / 'deadline_5months' / 'otp_cleanup'
    target_case_id  UUID         REFERENCES cases(id),
    scheduled_at    TIMESTAMPTZ  NOT NULL,
    executed_at     TIMESTAMPTZ,
    result          VARCHAR(20),                            -- 'success' / 'skipped' / 'error'
    error_message   TEXT,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_jobs_scheduled ON scheduled_jobs (scheduled_at) WHERE executed_at IS NULL;
CREATE INDEX idx_jobs_case ON scheduled_jobs (target_case_id);

COMMENT ON TABLE scheduled_jobs IS 'کارهای زمان‌بندی شده — مهلت‌های قانونی و پاکسازی';


-- ============================================================
-- ۲.۱۸ تنظیمات سیستم
-- ============================================================
CREATE TABLE system_configs (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    config_key      VARCHAR(100) NOT NULL UNIQUE,
    config_value    JSONB        NOT NULL,
    description     TEXT,
    updated_by      UUID         REFERENCES users(id),
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- مقادیر پیش‌فرض
INSERT INTO system_configs (config_key, config_value, description) VALUES
    ('otp.expiry_seconds', '120', 'مدت اعتبار OTP به ثانیه'),
    ('otp.max_attempts', '3', 'حداکثر تلاش ناموفق OTP'),
    ('otp.rate_limit.per_10min', '5', 'حداکثر درخواست OTP در ۱۰ دقیقه به ازای هر شماره'),
    ('security.max_failed_login', '5', 'حداکثر تلاش ناموفق ورود پیش از قفل'),
    ('security.lock_duration_minutes', '30', 'مدت قفل حساب پس از تلاش ناموفق'),
    ('integration.retry.max_attempts', '3', 'حداکثر تلاش مجدد ارسال به سازمان'),
    ('integration.retry.backoff_seconds', '[1,2,4,8,16]', 'فواصل تلاش مجدد به ثانیه'),
    ('deadline.check_interval_hours', '24', 'فاصله زمانی بررسی مهلت‌ها'),
    ('storage.presigned_url_expiry_seconds', '300', 'مدت اعتبار Pre-signed URL فایل‌ها');


-- ----------------------------------------------------------------------------
-- ۳. استراتژی رمزنگاری داده حساس
-- ----------------------------------------------------------------------------

-- ۳.۱ رمزنگاری فیلدهای حساس در سطح دیتابیس (Encryption at Rest)
-- فیلدهای زیر در لایه application قبل از ذخیره‌سازی رمزنگاری می‌شوند:
--   - users.national_id (با salt ثابت برای جستجوی exact match)
--   - cases.legal_entity_id
--   - cases.deceased_national_id
--   - map_services.geo_latitude, geo_longitude (مختصات دقیق)
--   - claim_services.main_plate_number, sub_plate_number
--   - payments.psp_token
--
-- روش: AES-256-GCM با کلید مستر ذخیره‌شده در Vault/KMS
-- (نه در کد application و نه در خود دیتابیس)


-- ----------------------------------------------------------------------------
-- ۴. Row-Level Security (RLS) — ایزوله‌سازی داده
-- ----------------------------------------------------------------------------
ALTER TABLE cases ENABLE ROW LEVEL SECURITY;
ALTER TABLE map_services ENABLE ROW LEVEL SECURITY;
ALTER TABLE claim_services ENABLE ROW LEVEL SECURITY;
ALTER TABLE cert_services ENABLE ROW LEVEL SECURITY;
ALTER TABLE map_photos ENABLE ROW LEVEL SECURITY;
ALTER TABLE claim_documents ENABLE ROW LEVEL SECURITY;

-- سیاست: متقاضی فقط پرونده‌های خودش را می‌بیند
CREATE POLICY applicant_access ON cases
    FOR SELECT
    TO applicant_role
    USING (applicant_id = current_setting('app.current_user_id')::UUID);

-- سیاست: کارشناس فقط پرونده‌های تخصیص‌یافته به خودش را می‌بیند
CREATE POLICY expert_access ON cases
    FOR SELECT
    TO expert_role
    USING (
        legal_expert_id = current_setting('app.current_expert_id')::UUID
        OR survey_expert_id = current_setting('app.current_expert_id')::UUID
    );

-- سیاست: ادمین همه را می‌بیند
CREATE POLICY admin_access ON cases
    FOR ALL
    TO admin_role
    USING (TRUE);

-- سیاست: حسابرس فقط خواندنی دارد
CREATE POLICY auditor_access ON cases
    FOR SELECT
    TO auditor_role
    USING (TRUE);


-- ----------------------------------------------------------------------------
-- ۵. استراتژی Migration و Versioning
-- ----------------------------------------------------------------------------
-- migrations/
--   001_extensions.sql          ← افزونه‌ها
--   002_enum_types.sql          ← تمام ENUMها
--   003_users.sql               ← جدول users
--   004_experts.sql             ← جدول experts
--   005_cases.sql               ← جدول cases
--   006_map_services.sql        ← سرویس نقشه
--   007_claim_services.sql      ← سرویس ادعا
--   008_cert_services.sql       ← سرویس گواهی
--   009_payments.sql            ← پرداخت‌ها + تعرفه‌ها
--   010_notifications.sql       ← اطلاع‌رسانی
--   011_audit_logs.sql          ← لاگ‌ها
--   012_scheduled_jobs.sql      ← کارهای زمان‌بندی
--   013_rls_policies.sql        ← سیاست‌های RLS
--   014_seed_data.sql           ← داده‌های اولیه (system_configs, ...)
--
-- ابزار پیشنهادی: golang-migrate / Flyway / Alembic
-- Migration naming: {YYYYMMDDHHMMSS}_{description}.sql
-- هر migration باید reversible باشد (down migration)
