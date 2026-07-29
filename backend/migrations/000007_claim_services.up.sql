-- Claim Service + Documents

CREATE TABLE claim_services (
    id                      UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    case_id                 UUID                NOT NULL REFERENCES cases(id) ON DELETE CASCADE,
    status                  claim_service_status NOT NULL DEFAULT 'pending_expert_assignment',

    map_tracking_code       VARCHAR(50),
    map_tracking_valid      BOOLEAN       DEFAULT FALSE,

    consent_otp_id          UUID          REFERENCES otp_sessions(id),
    false_claim_warning_sent BOOLEAN      DEFAULT FALSE,
    consent_granted_at      TIMESTAMPTZ,

    claim_type              claim_type,
    ownership_type          ownership_type,
    main_plate_number       VARCHAR(50),
    sub_plate_number        VARCHAR(50),
    plate_section           VARCHAR(50),
    total_share             INT,
    partial_share           INT,

    submitted_to_org_at     TIMESTAMPTZ,
    tracking_code           VARCHAR(50),
    org_response_at         TIMESTAMPTZ,
    org_rejection_reason    TEXT,
    org_response_raw        JSONB,

    has_government_rights   BOOLEAN      DEFAULT FALSE,
    treasury_payment_ref    VARCHAR(100),

    legal_advice_requested  BOOLEAN      DEFAULT FALSE,
    legal_advice_method     VARCHAR(20)  DEFAULT 'human',

    created_at              TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_claim_services_case ON claim_services (case_id);
CREATE INDEX idx_claim_services_status ON claim_services (status);
CREATE INDEX idx_claim_services_tracking ON claim_services (tracking_code);

CREATE TABLE claim_documents (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    claim_service_id UUID        NOT NULL REFERENCES claim_services(id) ON DELETE CASCADE,
    doc_type        document_type NOT NULL,
    file_path       VARCHAR(500) NOT NULL,
    file_size       BIGINT,
    checksum_sha256 VARCHAR(64),
    description     TEXT,
    verified_by     UUID         REFERENCES experts(id),
    verified_at     TIMESTAMPTZ,
    verification_note TEXT,
    uploaded_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_claim_docs_service ON claim_documents (claim_service_id);
CREATE INDEX idx_claim_docs_type ON claim_documents (doc_type);
