package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresAuditLogRepo implements AuditLogRepository
type PostgresAuditLogRepo struct {
	db *pgxpool.Pool
}

func NewPostgresAuditLogRepo(db *pgxpool.Pool) *PostgresAuditLogRepo {
	return &PostgresAuditLogRepo{db: db}
}

func (r *PostgresAuditLogRepo) Record(ctx context.Context, event *AuditEvent) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO audit_logs (event_type, actor_type, actor_id, actor_ip,
		                        resource_type, resource_id, changes, metadata, correlation_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, event.EventType, event.ActorType, event.ActorID, event.ActorIP,
		event.ResourceType, event.ResourceID, event.Changes, event.Metadata,
		event.CorrelationID,
	)
	if err != nil {
		return fmt.Errorf("failed to record audit event: %w", err)
	}
	return nil
}

// PostgresNotificationRepo implements NotificationRepository
type PostgresNotificationRepo struct {
	db *pgxpool.Pool
}

func NewPostgresNotificationRepo(db *pgxpool.Pool) *PostgresNotificationRepo {
	return &PostgresNotificationRepo{db: db}
}

func (r *PostgresNotificationRepo) Create(ctx context.Context, n *Notification) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO notifications (user_id, case_id, channel, template_key, content, status)
		VALUES ($1, $2, $3::notification_channel, $4, $5, 'pending')
	`, n.UserID, n.CaseID, n.Channel, n.TemplateKey, n.Content)
	if err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}
	return nil
}

// PostgresAIAdviceRepo implements AIAdviceRepository
type PostgresAIAdviceRepo struct {
	db *pgxpool.Pool
}

func NewPostgresAIAdviceRepo(db *pgxpool.Pool) *PostgresAIAdviceRepo {
	return &PostgresAIAdviceRepo{db: db}
}

func (r *PostgresAIAdviceRepo) Save(ctx context.Context, claimID uuid.UUID, advice *AIAdviceLog) (uuid.UUID, error) {
	var id uuid.UUID
	err := r.db.QueryRow(ctx, `
		INSERT INTO ai_advice_logs (case_id, claim_service_id, input_context,
			recommended_action, legal_references, confidence_score, model_version)
		SELECT cs.case_id, cs.id, '{}'::jsonb, $2, $3, $4, 'rule-v1.0'
		FROM claim_services cs WHERE cs.id = $1
		RETURNING id
	`, claimID, advice.Action, advice.References, advice.Confidence).Scan(&id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to save AI advice: %w", err)
	}
	return id, nil
}
