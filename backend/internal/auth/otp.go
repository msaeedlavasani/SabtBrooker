package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"github.com/msaeedlavasani/SabtBrooker/backend/internal/config"
	"github.com/msaeedlavasani/SabtBrooker/backend/internal/notification"
	"github.com/redis/go-redis/v9"
)

// OTPService handles one-time password operations
type OTPService struct {
	redis  *redis.Client
	cfg    config.OTPConfig
	notify *notification.Service
}

// NewOTPService creates a new OTP service
func NewOTPService(rdb *redis.Client, cfg config.OTPConfig, notify *notification.Service) *OTPService {
	return &OTPService{redis: rdb, cfg: cfg, notify: notify}
}

// GenerateAndSend creates an OTP, stores it in Redis, and sends via SMS
func (s *OTPService) GenerateAndSend(ctx context.Context, mobile, purpose string) (string, time.Time, error) {
	// Rate limiting check
	if err := s.checkRateLimit(ctx, mobile); err != nil {
		return "", time.Time{}, err
	}

	// Generate random OTP
	otp, err := generateOTP(s.cfg.Length)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to generate OTP: %w", err)
	}

	expiresAt := time.Now().Add(s.cfg.TTL)

	// Store in Redis: key = otp:{mobile}:{purpose}
	key := fmt.Sprintf("otp:%s:%s", mobile, purpose)

	// Store OTP hash + metadata
	hash := hashOTP(otp, mobile)
	pipe := s.redis.Pipeline()
	pipe.HSet(ctx, key,
		"hash", hash,
		"attempts", 0,
		"purpose", purpose,
		"created_at", time.Now().Unix(),
	)
	pipe.Expire(ctx, key, s.cfg.TTL)
	if _, err := pipe.Exec(ctx); err != nil {
		return "", time.Time{}, fmt.Errorf("failed to store OTP: %w", err)
	}

	// 3. Send SMS
	if s.notify != nil {
		// Run in background to avoid blocking response
		go func() {
			s.notify.SendOTP(context.Background(), mobile, otp)
		}()
	}

	return otp, expiresAt, nil
}

// Verify checks if the provided OTP is valid
func (s *OTPService) Verify(ctx context.Context, mobile, otpCode, purpose string) error {
	key := fmt.Sprintf("otp:%s:%s", mobile, purpose)

	// Get current attempts
	attempts, err := s.redis.HGet(ctx, key, "attempts").Int()
	if err == redis.Nil {
		return ErrOTPNotFound
	}
	if err != nil {
		return fmt.Errorf("failed to read OTP state: %w", err)
	}

	if attempts >= s.cfg.MaxAttempts {
		return ErrOTPExceeded
	}

	// Increment attempts
	s.redis.HIncrBy(ctx, key, "attempts", 1)

	// Get stored hash
	storedHash, err := s.redis.HGet(ctx, key, "hash").Result()
	if err != nil {
		return ErrOTPNotFound
	}

	expectedHash := hashOTP(otpCode, mobile)
	if storedHash != expectedHash {
		return ErrOTPInvalid
	}

	// Delete OTP on success (one-time use)
	s.redis.Del(ctx, key)

	return nil
}

// checkRateLimit prevents OTP spam
func (s *OTPService) checkRateLimit(ctx context.Context, mobile string) error {
	rateKey := fmt.Sprintf("otp_rate:%s", mobile)

	count, err := s.redis.Incr(ctx, rateKey).Result()
	if err != nil {
		return err
	}

	// Set expiry on first request
	if count == 1 {
		s.redis.Expire(ctx, rateKey, s.cfg.RateWindow)
	}

	if count > int64(s.cfg.RateLimit) {
		return ErrOTPRateLimited
	}

	return nil
}

func generateOTP(length int) (string, error) {
	const digits = "0123456789"
	result := make([]byte, length)
	for i := range result {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", err
		}
		result[i] = digits[n.Int64()]
	}
	// Ensure no all-zeros
	if string(result) == "00000" || string(result) == "000000" {
		result[0] = '1'
	}
	return string(result), nil
}

func hashOTP(otp, salt string) string {
	h := sha256.New()
	h.Write([]byte(otp + ":" + salt))
	return hex.EncodeToString(h.Sum(nil))
}

// Custom errors
var (
	ErrOTPNotFound    = fmt.Errorf("کد تایید یافت نشد یا منقضی شده است")
	ErrOTPInvalid     = fmt.Errorf("کد تایید اشتباه است")
	ErrOTPExceeded    = fmt.Errorf("تعداد دفعات مجاز به پایان رسید. لطفاً کد جدید درخواست کنید")
	ErrOTPRateLimited = fmt.Errorf("تعداد درخواست بیش از حد مجاز است. لطفاً دقایقی دیگر تلاش کنید")
)
