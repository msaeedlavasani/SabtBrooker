-- OTP Sessions & Refresh Tokens

CREATE TABLE otp_sessions (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id         UUID         REFERENCES users(id) ON DELETE CASCADE,
    mobile          VARCHAR(11)  NOT NULL,
    otp_hash        VARCHAR(255) NOT NULL,
    purpose         VARCHAR(50)  NOT NULL,
    attempts        INT          NOT NULL DEFAULT 0,
    max_attempts    INT          NOT NULL DEFAULT 3,
    expires_at      TIMESTAMPTZ  NOT NULL,
    verified_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_otp_mobile_expires ON otp_sessions (mobile, expires_at)
    WHERE verified_at IS NULL;
CREATE INDEX idx_otp_user ON otp_sessions (user_id);

CREATE TABLE refresh_tokens (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id         UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash      VARCHAR(255) NOT NULL UNIQUE,
    family          UUID         NOT NULL,
    device_info     VARCHAR(500),
    ip_address      INET,
    expires_at      TIMESTAMPTZ  NOT NULL,
    revoked_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_refresh_token_user ON refresh_tokens (user_id, revoked_at);
CREATE INDEX idx_refresh_token_family ON refresh_tokens (family);
