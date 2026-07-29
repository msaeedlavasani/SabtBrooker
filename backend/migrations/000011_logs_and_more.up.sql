-- Audit Logs, Integration Logs, AI Advice, Scheduled Jobs, Outbox, System Configs

CREATE TABLE audit_logs (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    event_type      audit_event_type NOT NULL,
    event_timestamp TIMESTAMPTZ      NOT NULL DEFAULT NOW(),

    actor_type      VARCHAR(20)      NOT NULL,
    actor_id        UUID             REFERENCES users(id),
    actor_ip        INET,
    actor_user_agent TEXT,

    resource_type   VARCHAR(50)      NOT NULL,
    resource_id     UUID,

    changes         JSONB,
    metadata        JSONB,

    correlation_id  UUID,
    trace_id        VARCHAR(64)
);

CREATE INDEX idx_audit_timestamp ON audit_logs (event_timestamp DESC);
CREATE INDEX idx_audit_event ON audit_logs (event_type, event_timestamp DESC);
CREATE INDEX idx_audit_actor ON audit_logs (actor_id, event_timestamp DESC);
CREATE INDEX idx_audit_resource ON audit_logs (resource_type, resource_id);
CREATE INDEX idx_audit_correlation ON audit_logs (correlation_id);

-- Prevent deletes/updates on audit_logs (append-only)
REVOKE DELETE, UPDATE ON audit_logs FROM PUBLIC;

CREATE TABLE integration_logs (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    case_id             UUID         REFERENCES cases(id),
    service_type        VARCHAR(20)  NOT NULL,
    direction           VARCHAR(10)  NOT NULL,
    endpoint            VARCHAR(200) NOT NULL,
    request_payload     JSONB,
    response_payload    JSONB,
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

CREATE TABLE ai_advice_logs (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    case_id             UUID         NOT NULL REFERENCES cases(id),
    claim_service_id    UUID         NOT NULL REFERENCES claim_services(id),

    input_context       JSONB        NOT NULL,
    input_documents     TEXT[],

    recommended_action  VARCHAR(100),
    legal_references    TEXT[],
    confidence_score    DECIMAL(3,2),
    raw_ai_response     TEXT,

    was_helpful         BOOLEAN,
    escalated_to_human  BOOLEAN      DEFAULT FALSE,

    model_version       VARCHAR(50),
    prompt_version      VARCHAR(50),
    latency_ms          INT,

    created_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ai_advice_case ON ai_advice_logs (case_id);

CREATE TABLE outbox_messages (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    aggregate_type  VARCHAR(50) NOT NULL,
    aggregate_id    UUID NOT NULL,
    event_type      VARCHAR(100) NOT NULL,
    payload         JSONB NOT NULL,
    status          VARCHAR(20) NOT NULL DEFAULT 'pending',
    retry_count     INT NOT NULL DEFAULT 0,
    last_error      TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed_at    TIMESTAMPTZ,
    correlation_id  UUID
);

CREATE INDEX idx_outbox_pending ON outbox_messages (status, created_at)
    WHERE status IN ('pending', 'failed');
CREATE INDEX idx_outbox_aggregate ON outbox_messages (aggregate_type, aggregate_id);

CREATE TABLE scheduled_jobs (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    job_key         VARCHAR(100) NOT NULL UNIQUE,
    job_type        VARCHAR(50)  NOT NULL,
    target_case_id  UUID         REFERENCES cases(id),
    scheduled_at    TIMESTAMPTZ  NOT NULL,
    executed_at     TIMESTAMPTZ,
    result          VARCHAR(20),
    error_message   TEXT,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_jobs_scheduled ON scheduled_jobs (scheduled_at) WHERE executed_at IS NULL;
CREATE INDEX idx_jobs_case ON scheduled_jobs (target_case_id);

CREATE TABLE system_configs (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    config_key      VARCHAR(100) NOT NULL UNIQUE,
    config_value    JSONB        NOT NULL,
    description     TEXT,
    updated_by      UUID         REFERENCES users(id),
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
