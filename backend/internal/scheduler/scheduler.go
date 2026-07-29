package scheduler

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
)

// Scheduler polls scheduled_jobs and executes due jobs
type Scheduler struct {
	db       *pgxpool.Pool
	eventBus *nats.Conn
	interval time.Duration
	jobs     map[string]JobHandler
}

// JobHandler processes a scheduled job
type JobHandler func(ctx context.Context, job ScheduledJob) error

// ScheduledJob represents a row from the scheduled_jobs table
type ScheduledJob struct {
	ID            uuid.UUID
	JobType       string
	TargetCaseID  uuid.UUID
	ScheduledAt   time.Time
	ExecutedAt    *time.Time
	Result        *string
	ErrorMessage  *string
}

// New creates a scheduler with the given poll interval
func New(db *pgxpool.Pool, eventBus *nats.Conn, interval time.Duration) *Scheduler {
	s := &Scheduler{
		db:       db,
		eventBus: eventBus,
		interval: interval,
		jobs:     make(map[string]JobHandler),
	}

	// Register built-in job handlers
	s.Register("deadline_2years", handleDeadline2Years)
	s.Register("deadline_5months", handleDeadline5Months)
	s.Register("otp_cleanup", handleOTPCleanup)

	return s
}

// Register adds a job handler
func (s *Scheduler) Register(jobType string, handler JobHandler) {
	s.jobs[jobType] = handler
}

// Start begins the poll loop — blocks until context is cancelled
func (s *Scheduler) Start(ctx context.Context) {
	slog.Info("scheduler started", "interval", s.interval)
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("scheduler stopped")
			return
		case <-ticker.C:
			if err := s.Poll(ctx); err != nil {
				slog.Error("scheduler poll failed", "error", err)
			}
		}
	}
}

// Poll fetches due jobs and executes them with FOR UPDATE SKIP LOCKED
func (s *Scheduler) Poll(ctx context.Context) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	rows, err := tx.Query(ctx, `
		SELECT id, job_type, target_case_id, scheduled_at,
		       executed_at, result, error_message
		FROM scheduled_jobs
		WHERE executed_at IS NULL AND scheduled_at <= NOW()
		ORDER BY scheduled_at
		LIMIT 100
		FOR UPDATE SKIP LOCKED
	`)
	if err != nil {
		return fmt.Errorf("failed to query scheduled jobs: %w", err)
	}
	defer rows.Close()

	var jobs []ScheduledJob
	for rows.Next() {
		var j ScheduledJob
		if err := rows.Scan(&j.ID, &j.JobType, &j.TargetCaseID, &j.ScheduledAt,
			&j.ExecutedAt, &j.Result, &j.ErrorMessage); err != nil {
			return fmt.Errorf("failed to scan job row: %w", err)
		}
		jobs = append(jobs, j)
	}

	if len(jobs) == 0 {
		return tx.Commit(ctx)
	}

	for _, job := range jobs {
		handler, ok := s.jobs[job.JobType]
		if !ok {
			slog.Warn("unknown job type", "type", job.JobType, "job_id", job.ID)
			tx.Exec(ctx, `
				UPDATE scheduled_jobs SET executed_at = NOW(), result = 'skipped',
				error_message = 'unknown job type' WHERE id = $1
			`, job.ID)
			continue
		}

		if err := handler(ctx, job); err != nil {
			slog.Error("job execution failed", "job_id", job.ID, "type", job.JobType, "error", err)
			tx.Exec(ctx, `
				UPDATE scheduled_jobs SET executed_at = NOW(), result = 'error',
				error_message = $2 WHERE id = $1
			`, job.ID, err.Error())
		} else {
			tx.Exec(ctx, `
				UPDATE scheduled_jobs SET executed_at = NOW(), result = 'success'
				WHERE id = $1
			`, job.ID)
		}
	}

	return tx.Commit(ctx)
}

// Schedule creates a new scheduled job
func (s *Scheduler) Schedule(ctx context.Context, jobType string, caseID uuid.UUID, scheduledAt time.Time) error {
	key := fmt.Sprintf("%s_%s", jobType, caseID)
	_, err := s.db.Exec(ctx, `
		INSERT INTO scheduled_jobs (job_key, job_type, target_case_id, scheduled_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (job_key) DO NOTHING
	`, key, jobType, caseID, scheduledAt)
	return err
}

// CancelDeadline removes pending deadline jobs for a case
func (s *Scheduler) CancelDeadline(ctx context.Context, caseID uuid.UUID) error {
	_, err := s.db.Exec(ctx, `
		UPDATE scheduled_jobs SET executed_at = NOW(), result = 'skipped',
		error_message = 'case completed before deadline'
		WHERE target_case_id = $1 AND executed_at IS NULL
	`, caseID)
	return err
}

// ---------- Built-in handlers ----------

func handleDeadline2Years(ctx context.Context, job ScheduledJob) error {
	db := getDBFromCtx(ctx)
	if db == nil {
		return fmt.Errorf("no database in context")
	}

	// Check if case is already completed (race condition safety)
	var status string
	err := db.QueryRow(ctx, `SELECT status::text FROM cases WHERE id = $1`, job.TargetCaseID).Scan(&status)
	if err != nil {
		return fmt.Errorf("case %s not found: %w", job.TargetCaseID, err)
	}
	if status == "cert_completed" || status == "expired" || status == "cancelled" {
		slog.Info("deadline_2years: case already final", "case_id", job.TargetCaseID, "status", status)
		return nil
	}

	// Transition to expired
	_, err = db.Exec(ctx,
		`UPDATE cases SET status = 'expired', updated_at = NOW()
		 WHERE id = $1 AND status::text = 'claim_completed'`,
		job.TargetCaseID,
	)
	if err != nil {
		return fmt.Errorf("failed to expire case: %w", err)
	}

	slog.Info("deadline_2years: case expired", "case_id", job.TargetCaseID)

	// Publish event
	// getEventBusFromCtx(ctx) would publish case.expired event

	return nil
}

func handleDeadline5Months(ctx context.Context, job ScheduledJob) error {
	db := getDBFromCtx(ctx)
	if db == nil {
		return fmt.Errorf("no database in context")
	}

	var status string
	var applicantDeceased bool
	err := db.QueryRow(ctx,
		`SELECT status::text, COALESCE(applicant_deceased, false) FROM cases WHERE id = $1`,
		job.TargetCaseID,
	).Scan(&status, &applicantDeceased)
	if err != nil {
		return fmt.Errorf("case %s not found: %w", job.TargetCaseID, err)
	}

	if !applicantDeceased || status != "cert_in_progress" {
		slog.Info("deadline_5months: case not applicable", "case_id", job.TargetCaseID, "status", status)
		return nil
	}

	// Expire the case
	_, err = db.Exec(ctx,
		`UPDATE cases SET status = 'expired', updated_at = NOW() WHERE id = $1`,
		job.TargetCaseID,
	)
	if err != nil {
		return fmt.Errorf("failed to expire case: %w", err)
	}

	slog.Info("deadline_5months: case expired (applicant deceased)", "case_id", job.TargetCaseID)
	return nil
}

func handleOTPCleanup(ctx context.Context, job ScheduledJob) error {
	db := getDBFromCtx(ctx)
	if db == nil {
		return fmt.Errorf("no database in context")
	}

	_, err := db.Exec(ctx, `
		DELETE FROM otp_sessions WHERE expires_at < NOW() - INTERVAL '10 minutes'
	`)
	if err != nil {
		return fmt.Errorf("failed to cleanup expired OTPs: %w", err)
	}
	return nil
}

// ---- context helpers ----

type ctxKey int

const (
	ctxDBKey  ctxKey = iota
	ctxNatsKey
)

// WithDB injects database pool into context
func WithDB(ctx context.Context, db *pgxpool.Pool) context.Context {
	return context.WithValue(ctx, ctxDBKey, db)
}

func getDBFromCtx(ctx context.Context) *pgxpool.Pool {
	if db, ok := ctx.Value(ctxDBKey).(*pgxpool.Pool); ok {
		return db
	}
	return nil
}
