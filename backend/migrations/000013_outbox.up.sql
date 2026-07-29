-- ============================================================================
-- Migration 013: Outbox Pattern — reliable messaging to organization
-- ============================================================================

CREATE TABLE outbox_messages (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    aggregate_type  VARCHAR(50) NOT NULL,        -- map_service | claim_service | cert_service
    aggregate_id    UUID NOT NULL,
    event_type      VARCHAR(100) NOT NULL,       -- submit_to_org | status_check
    payload         JSONB NOT NULL,
    status          VARCHAR(20) NOT NULL DEFAULT 'pending',  -- pending | processing | sent | failed
    retry_count     INT NOT NULL DEFAULT 0,
    last_error      TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed_at    TIMESTAMPTZ,
    correlation_id  UUID
);

-- Index for efficient polling of pending messages
CREATE INDEX idx_outbox_pending ON outbox_messages (status, created_at)
    WHERE status IN ('pending', 'failed');

-- Index for looking up messages by aggregate
CREATE INDEX idx_outbox_aggregate ON outbox_messages (aggregate_type, aggregate_id);
