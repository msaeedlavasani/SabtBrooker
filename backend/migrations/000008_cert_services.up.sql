-- Certificate Service

CREATE TABLE cert_services (
    id                      UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    case_id                 UUID                NOT NULL REFERENCES cases(id) ON DELETE CASCADE,
    status                  cert_service_status NOT NULL DEFAULT 'pending_data',

    claim_tracking_code     VARCHAR(50),
    claim_tracking_valid    BOOLEAN       DEFAULT FALSE,

    consent_otp_id          UUID          REFERENCES otp_sessions(id),
    consent_granted_at      TIMESTAMPTZ,

    action_reference        action_reference,
    action_type             action_type,
    action_date             DATE,
    cert_image_path         VARCHAR(500),
    cert_unique_id          VARCHAR(100),

    submitted_to_org_at     TIMESTAMPTZ,
    tracking_code           VARCHAR(50),
    org_response_at         TIMESTAMPTZ,
    org_rejection_reason    TEXT,
    org_response_raw        JSONB,

    created_at              TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_cert_services_case ON cert_services (case_id);
CREATE INDEX idx_cert_services_status ON cert_services (status);
CREATE INDEX idx_cert_services_tracking ON cert_services (tracking_code);
