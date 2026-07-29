-- Payments & Tariffs

CREATE TABLE tariffs (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    service_type    VARCHAR(20)  NOT NULL,
    max_amount      BIGINT       NOT NULL,
    non_refundable  BIGINT       DEFAULT 0,
    effective_from  DATE         NOT NULL,
    effective_to    DATE,
    version         VARCHAR(20)  NOT NULL,
    description     TEXT,
    created_by      UUID         REFERENCES users(id),
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tariffs_service ON tariffs (service_type, effective_from DESC);
CREATE UNIQUE INDEX idx_tariffs_active ON tariffs (service_type) WHERE effective_to IS NULL;

CREATE TABLE payments (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    case_id             UUID            NOT NULL REFERENCES cases(id),
    service_type        VARCHAR(20)     NOT NULL CHECK (service_type IN ('map', 'claim', 'cert', 'legal_advice')),
    amount              BIGINT          NOT NULL,
    payment_type        payment_type    NOT NULL,
    status              payment_status  NOT NULL DEFAULT 'pending',

    psp_reference       VARCHAR(100),
    psp_token           VARCHAR(500),
    payment_url         TEXT,
    paid_at             TIMESTAMPTZ,

    refund_amount       BIGINT,
    refund_reference    VARCHAR(100),
    refunded_at         TIMESTAMPTZ,
    refund_reason       TEXT,

    tariff_version      VARCHAR(20),
    meta                JSONB,

    created_at          TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_payments_case ON payments (case_id);
CREATE INDEX idx_payments_status ON payments (status);
CREATE INDEX idx_payments_psp ON payments (psp_reference);
