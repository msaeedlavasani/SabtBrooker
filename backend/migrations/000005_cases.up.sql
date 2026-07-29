-- Cases (core entity)

CREATE TABLE cases (
    id                      UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    applicant_id            UUID           NOT NULL REFERENCES users(id),

    applicant_capacity      applicant_capacity NOT NULL DEFAULT 'principal',
    legal_entity_id         VARCHAR(11),
    legal_entity_name       VARCHAR(300),
    proxy_document_type     VARCHAR(50),
    proxy_document_id       VARCHAR(100),
    proxy_document_date     DATE,
    proxy_verification_code VARCHAR(20),
    proxy_verified          BOOLEAN      DEFAULT FALSE,
    proxy_verified_by       UUID         REFERENCES experts(id),
    proxy_verification_note TEXT,

    status                  case_status  NOT NULL DEFAULT 'draft',

    claim_approved_at       TIMESTAMPTZ,
    deadline_2years         TIMESTAMPTZ,
    applicant_deceased      BOOLEAN      DEFAULT FALSE,
    deceased_national_id    VARCHAR(10),
    date_of_death           DATE,
    deadline_5months        TIMESTAMPTZ,

    province                VARCHAR(100) NOT NULL,
    city                    VARCHAR(100) NOT NULL,
    district                VARCHAR(100),
    village                 VARCHAR(100),
    postal_code             VARCHAR(10),
    address_detail          TEXT,

    map_tracking_code       VARCHAR(50),
    claim_tracking_code     VARCHAR(50),
    cert_tracking_code      VARCHAR(50),

    legal_expert_id         UUID         REFERENCES experts(id),
    survey_expert_id        UUID         REFERENCES experts(id),

    created_at              TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    completed_at            TIMESTAMPTZ
);

CREATE INDEX idx_cases_applicant ON cases (applicant_id);
CREATE INDEX idx_cases_status ON cases (status);
CREATE INDEX idx_cases_deadline ON cases (deadline_2years) WHERE status = 'claim_completed';
CREATE INDEX idx_cases_deadline_5m ON cases (deadline_5months) WHERE status = 'cert_in_progress';
CREATE INDEX idx_cases_experts ON cases (legal_expert_id, survey_expert_id);
CREATE INDEX idx_cases_tracking ON cases (map_tracking_code, claim_tracking_code, cert_tracking_code);
CREATE INDEX idx_cases_created ON cases (created_at DESC);
