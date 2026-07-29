package outbox

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
)

// Message represents a row in the outbox_messages table
type Message struct {
	ID            uuid.UUID       `json:"id"`
	AggregateType string          `json:"aggregate_type"`
	AggregateID   uuid.UUID       `json:"aggregate_id"`
	EventType     string          `json:"event_type"`
	Payload       json.RawMessage `json:"payload"`
	Status        string          `json:"status"`
	RetryCount    int             `json:"retry_count"`
	LastError     *string         `json:"last_error,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
	ProcessedAt   *time.Time      `json:"processed_at,omitempty"`
	CorrelationID *uuid.UUID      `json:"correlation_id,omitempty"`
}

// Publisher polls the outbox table and processes pending messages
type Publisher struct {
	db       *pgxpool.Pool
	eventBus *nats.Conn
	interval time.Duration
	maxRetry int
	// sendFunc is the actual function that sends to the organization
	sendFunc func(ctx context.Context, msg Message) error
}

// NewPublisher creates a new outbox publisher
func NewPublisher(db *pgxpool.Pool, eventBus *nats.Conn, interval time.Duration) *Publisher {
	return &Publisher{
		db:       db,
		eventBus: eventBus,
		interval: interval,
		maxRetry: 3,
	}
}

// SetSender sets the function that actually sends messages to the organization
func (p *Publisher) SetSender(fn func(ctx context.Context, msg Message) error) {
	p.sendFunc = fn
}

// Start begins the poll loop — blocks until context cancelled
func (p *Publisher) Start(ctx context.Context) {
	slog.Info("outbox publisher started", "interval", p.interval)
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("outbox publisher stopped")
			return
		case <-ticker.C:
			if err := p.processBatch(ctx); err != nil {
				slog.Error("outbox batch failed", "error", err)
			}
		}
	}
}

func (p *Publisher) processBatch(ctx context.Context) error {
	tx, err := p.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	rows, err := tx.Query(ctx, `
		SELECT id, aggregate_type, aggregate_id, event_type, payload,
		       status, retry_count, last_error, created_at, processed_at,
		       correlation_id
		FROM outbox_messages
		WHERE status IN ('pending', 'failed')
		  AND retry_count < $1
		ORDER BY created_at
		LIMIT 50
		FOR UPDATE SKIP LOCKED
	`, p.maxRetry)
	if err != nil {
		return fmt.Errorf("failed to query outbox: %w", err)
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var m Message
		var lastError *string
		if err := rows.Scan(&m.ID, &m.AggregateType, &m.AggregateID, &m.EventType,
			&m.Payload, &m.Status, &m.RetryCount, &lastError,
			&m.CreatedAt, &m.ProcessedAt, &m.CorrelationID); err != nil {
			return fmt.Errorf("failed to scan outbox row: %w", err)
		}
		if lastError != nil {
			m.LastError = lastError
		}
		messages = append(messages, m)
	}

	for _, msg := range messages {
		var sendErr error
		if p.sendFunc != nil {
			sendErr = p.sendFunc(ctx, msg)
		} else {
			// Stub: publish to NATS as fallback
			sendErr = p.publishToNATS(ctx, msg)
		}

		if sendErr != nil {
			errMsg := sendErr.Error()
			tx.Exec(ctx, `
				UPDATE outbox_messages
				SET status = 'failed', retry_count = retry_count + 1, last_error = $2
				WHERE id = $1
			`, msg.ID, errMsg)

			if msg.RetryCount+1 >= p.maxRetry {
				slog.Error("outbox message exhausted retries — moving to DLQ",
					"id", msg.ID,
					"aggregate_type", msg.AggregateType,
					"retry_count", msg.RetryCount+1,
				)
				// Publish to DLQ via NATS
				if p.eventBus != nil {
					dlqData, _ := json.Marshal(map[string]interface{}{
						"original": msg,
						"error":    errMsg,
					})
					p.eventBus.Publish("integration.dlq", dlqData)
				}
			}
		} else {
			tx.Exec(ctx, `
				UPDATE outbox_messages
				SET status = 'sent', processed_at = NOW()
				WHERE id = $1
			`, msg.ID)
		}
	}

	return tx.Commit(ctx)
}

func (p *Publisher) publishToNATS(ctx context.Context, msg Message) error {
	if p.eventBus == nil {
		return fmt.Errorf("NATS not available")
	}
	subj := fmt.Sprintf("outbox.%s.%s", msg.AggregateType, msg.EventType)
	return p.eventBus.Publish(subj, msg.Payload)
}

// Recorder records a message in the outbox table within an existing transaction
type Recorder struct {
	tx pgx.Tx
}

// NewRecorder creates a recorder bound to a transaction
func NewRecorder(tx pgx.Tx) *Recorder {
	return &Recorder{tx: tx}
}

// Record inserts a message into the outbox within the current transaction
func (r *Recorder) Record(ctx context.Context, aggregateType string, aggregateID uuid.UUID, eventType string, payload interface{}) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal outbox payload: %w", err)
	}

	correlationID := uuid.New()
	_, err = r.tx.Exec(ctx, `
		INSERT INTO outbox_messages (aggregate_type, aggregate_id, event_type, payload, correlation_id)
		VALUES ($1, $2, $3, $4, $5)
	`, aggregateType, aggregateID, eventType, payloadBytes, correlationID)
	if err != nil {
		return fmt.Errorf("failed to record outbox message: %w", err)
	}

	slog.Debug("outbox message recorded",
		"aggregate_type", aggregateType,
		"aggregate_id", aggregateID,
		"event_type", eventType,
		"correlation_id", correlationID,
	)
	return nil
}
