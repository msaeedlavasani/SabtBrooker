-- Users & Experts

CREATE TABLE users (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    national_id         VARCHAR(10)  NOT NULL UNIQUE,
    first_name          VARCHAR(100) NOT NULL,
    last_name           VARCHAR(100) NOT NULL,
    mobile              VARCHAR(11)  NOT NULL UNIQUE,
    mobile_verified     BOOLEAN      NOT NULL DEFAULT FALSE,
    birth_date          DATE,
    role                user_role    NOT NULL DEFAULT 'applicant',

    sana_status         VARCHAR(20),
    sana_checked_at     TIMESTAMPTZ,
    ncr_mobile_match    BOOLEAN,
    ncr_checked_at      TIMESTAMPTZ,
    is_alive            BOOLEAN      DEFAULT TRUE,
    alive_checked_at    TIMESTAMPTZ,

    password_hash       VARCHAR(255),
    failed_login_count  INT          NOT NULL DEFAULT 0,
    locked_until        TIMESTAMPTZ,
    last_login_at       TIMESTAMPTZ,
    last_login_ip       INET,

    is_active           BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_national_id ON users (national_id);
CREATE INDEX idx_users_mobile ON users (mobile);
CREATE INDEX idx_users_role ON users (role);
CREATE INDEX idx_users_sana_status ON users (sana_status) WHERE sana_status = 'inactive';

CREATE TABLE experts (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id             UUID         NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    expert_type         VARCHAR(20)  NOT NULL CHECK (expert_type IN ('legal', 'survey')),
    license_number      VARCHAR(50),
    license_expiry      DATE,
    organization        VARCHAR(200),
    education_level     VARCHAR(50),
    field_of_study      VARCHAR(100),
    years_of_experience INT,
    in_approved_list    BOOLEAN      NOT NULL DEFAULT FALSE,
    approved_list_ref   VARCHAR(100),
    max_active_cases    INT          NOT NULL DEFAULT 10,
    current_case_count  INT          NOT NULL DEFAULT 0,
    is_available        BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_experts_type ON experts (expert_type);
CREATE INDEX idx_experts_available ON experts (is_available, expert_type) WHERE is_available = TRUE;
CREATE INDEX idx_experts_user ON experts (user_id);
