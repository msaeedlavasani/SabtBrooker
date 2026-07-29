-- Notifications

CREATE TABLE notifications (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id         UUID                NOT NULL REFERENCES users(id),
    case_id         UUID                REFERENCES cases(id),
    channel         notification_channel NOT NULL,
    template_key    VARCHAR(50)         NOT NULL,
    content         TEXT                NOT NULL,
    status          notification_status NOT NULL DEFAULT 'pending',
    provider_ref    VARCHAR(200),
    sent_at         TIMESTAMPTZ,
    delivered_at    TIMESTAMPTZ,
    error_message   TEXT,
    retry_count     INT                 DEFAULT 0,
    max_retries     INT                 DEFAULT 3,
    read_at         TIMESTAMPTZ,
    created_at      TIMESTAMPTZ         NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notifications_user ON notifications (user_id, created_at DESC);
CREATE INDEX idx_notifications_status ON notifications (status) WHERE status = 'pending';
CREATE INDEX idx_notifications_case ON notifications (case_id);
