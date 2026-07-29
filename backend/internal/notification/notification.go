package notification

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// SMSProvider defines the interface for different SMS gateways
type SMSProvider interface {
	Send(ctx context.Context, mobile, content string) error
	SendPattern(ctx context.Context, mobile, pattern string, tokens map[string]string) error
}

// ConsoleSMSProvider logs SMS to console (for dev)
type ConsoleSMSProvider struct{}

func (p *ConsoleSMSProvider) Send(ctx context.Context, mobile, content string) error {
	slog.Info("SMS (Console)", "mobile", mobile, "content", content)
	return nil
}
func (p *ConsoleSMSProvider) SendPattern(ctx context.Context, mobile, pattern string, tokens map[string]string) error {
	slog.Info("SMS Pattern (Console)", "mobile", mobile, "pattern", pattern, "tokens", tokens)
	return nil
}

// KavenegarProvider implements SMSProvider for Kavenegar service
type KavenegarProvider struct {
	apiKey string
	client *http.Client
}

func NewKavenegarProvider(apiKey string) *KavenegarProvider {
	return &KavenegarProvider{
		apiKey: apiKey,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (p *KavenegarProvider) Send(ctx context.Context, mobile, content string) error {
	endpoint := fmt.Sprintf("https://api.kavenegar.com/v1/%s/sms/send.json", p.apiKey)
	form := url.Values{}
	form.Add("receptor", mobile)
	form.Add("message", content)

	resp, err := p.client.PostForm(endpoint, form)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("kavenegar error: status %d", resp.StatusCode)
	}
	return nil
}

func (p *KavenegarProvider) SendPattern(ctx context.Context, mobile, pattern string, tokens map[string]string) error {
	endpoint := fmt.Sprintf("https://api.kavenegar.com/v1/%s/verify/lookup.json", p.apiKey)
	form := url.Values{}
	form.Add("receptor", mobile)
	form.Add("template", pattern)
	for k, v := range tokens {
		form.Add(k, v)
	}

	resp, err := p.client.PostForm(endpoint, form)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("kavenegar error: status %d", resp.StatusCode)
	}
	return nil
}

// Service handles all notification channels
type Service struct {
	db          *pgxpool.Pool
	redisCli    *redis.Client
	smsProvider SMSProvider
}

// NewService creates a notification service
func NewService(db *pgxpool.Pool, redisCli *redis.Client, sms SMSProvider) *Service {
	if sms == nil {
		sms = &ConsoleSMSProvider{}
	}
	return &Service{db: db, redisCli: redisCli, smsProvider: sms}
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
func (s *Service) SendSMS(ctx context.Context, userID uuid.UUID, mobile, content string) error {
	// Rate limiting: max 5 SMS per mobile per 10 minutes
	rateKey := fmt.Sprintf("sms_rate:%s", mobile)
	count, err := s.redisCli.Incr(ctx, rateKey).Result()
	if err == nil && count == 1 {
		s.redisCli.Expire(ctx, rateKey, 600)
	}
	if count > 5 {
		return fmt.Errorf("تعداد پیامک‌های مجاز به پایان رسیده")
	}

	// Send via provider
	if err := s.smsProvider.Send(ctx, mobile, content); err != nil {
		slog.Error("SMS provider error", "error", err)
		return err
	}

	// Record in database
	s.db.Exec(ctx, `
		INSERT INTO notifications (user_id, channel, template_key, content, status)
		VALUES ($1, 'sms', 'sms_direct', $2, 'sent')
	`, userID, content)

	return nil
}

// SendOTP sends an OTP via pattern lookup (fast & reliable)
func (s *Service) SendOTP(ctx context.Context, mobile, code string) error {
	return s.smsProvider.SendPattern(ctx, mobile, "otp-template", map[string]string{"token": code})
}

// SendEmail sends an email — stub for now
func (s *Service) SendEmail(ctx context.Context, userID uuid.UUID, email, subject, body string) error {
	slog.Info("Email (stub)", "email", email, "subject", subject)
	s.db.Exec(ctx, `
		INSERT INTO notifications (user_id, channel, template_key, content, status)
		VALUES ($1, 'email', $2, $3, 'pending')
	`, userID, subject, body)
	return nil
}

// NotifyApplicant sends both in-app and SMS notifications to an applicant
func (s *Service) NotifyApplicant(ctx context.Context, userID uuid.UUID, caseID uuid.UUID, message string) error {
	s.SendInApp(ctx, userID, &caseID, "case_update", message)
	var mobile string
	s.db.QueryRow(ctx, `SELECT mobile FROM users WHERE id = $1`, userID).Scan(&mobile)
	if mobile != "" {
		s.SendSMS(ctx, userID, mobile, message)
	}
	return nil
}

// NotifyExpertInApp sends in-app notification to an expert
func (s *Service) NotifyExpertInApp(ctx context.Context, expertID uuid.UUID, caseID uuid.UUID, message string) error {
	return s.SendInApp(ctx, expertID, &caseID, "expert_assignment", message)
}
