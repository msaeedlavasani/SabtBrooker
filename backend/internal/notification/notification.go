package notification

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// Service handles all notification channels
type Service struct {
	db       *pgxpool.Pool
	redisCli *redis.Client
}

// NewService creates a notification service
func NewService(db *pgxpool.Pool, redisCli *redis.Client) *Service {
	return &Service{db: db, redisCli: redisCli}
}

// SendInApp creates an in-app notification in the database
func (s *Service) SendInApp(ctx context.Context, userID uuid.UUID, caseID *uuid.UUID, templateKey, content string) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO notifications (user_id, case_id, channel, template_key, content, status)
		VALUES ($1, $2, 'in_app', $3, $4, 'pending')
	`, userID, caseID, templateKey, content)
	if err != nil {
		return fmt.Errorf("failed to create in-app notification: %w", err)
	}
	slog.Info("in-app notification created",
		"user_id", userID, "template", templateKey)
	return nil
}

// SendSMS sends an SMS through the configured provider
// فعلاً stub: فقط لاگ می‌کند — در production به SMS Gateway واقعی وصل می‌شود
func (s *Service) SendSMS(ctx context.Context, userID uuid.UUID, mobile, content string) error {
	// Rate limiting: max 5 SMS per mobile per 10 minutes
	rateKey := fmt.Sprintf("sms_rate:%s", mobile)
	count, err := s.redisCli.Incr(ctx, rateKey).Result()
	if err == nil && count == 1 {
		s.redisCli.Expire(ctx, rateKey, 600) // 10 minutes in seconds
	}
	if count > 5 {
		slog.Warn("SMS rate limited", "mobile", mobile, "count", count)
		return fmt.Errorf("تعداد پیامک‌های مجاز برای این شماره به پایان رسیده — لطفاً ۱۰ دقیقه دیگر تلاش کنید")
	}

	// Stub: log SMS instead of actually sending
	slog.Info("SMS (stub)",
		"mobile", mobile,
		"content", content,
	)

	// Record in database
	s.db.Exec(ctx, `
		INSERT INTO notifications (user_id, channel, template_key, content, status)
		VALUES ($1, 'sms', 'sms_direct', $2, 'sent')
	`, userID, content)

	return nil
}

// SendEmail sends an email — stub for now
func (s *Service) SendEmail(ctx context.Context, userID uuid.UUID, email, subject, body string) error {
	slog.Info("Email (stub)",
		"email", email,
		"subject", subject,
	)

	s.db.Exec(ctx, `
		INSERT INTO notifications (user_id, channel, template_key, content, status)
		VALUES ($1, 'email', $2, $3, 'pending')
	`, userID, subject, body)

	return nil
}

// NotifyApplicant sends both in-app and SMS notifications to an applicant
func (s *Service) NotifyApplicant(ctx context.Context, userID uuid.UUID, caseID uuid.UUID, message string) error {
	// In-app
	if err := s.SendInApp(ctx, userID, &caseID, "case_update", message); err != nil {
		slog.Error("failed to send in-app notification", "error", err)
	}

	// SMS (stub)
	var mobile string
	s.db.QueryRow(ctx, `SELECT mobile FROM users WHERE id = $1`, userID).Scan(&mobile)
	if mobile != "" {
		if err := s.SendSMS(ctx, userID, mobile, message); err != nil {
			slog.Error("failed to send SMS", "error", err)
		}
	}

	return nil
}

// NotifyExpertInApp sends in-app notification to an expert
func (s *Service) NotifyExpertInApp(ctx context.Context, expertID uuid.UUID, caseID uuid.UUID, message string) error {
	return s.SendInApp(ctx, expertID, &caseID, "expert_assignment", message)
}
