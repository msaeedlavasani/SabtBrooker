package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/msaeedlavasani/SabtBrooker/backend/internal/payment"
	"github.com/msaeedlavasani/SabtBrooker/backend/internal/repository"
)

type PaymentService struct {
	repo     repository.PaymentRepository
	caseRepo repository.CaseRepository
	provider payment.PaymentProvider
	userRepo repository.UserRepository
}

func NewPaymentService(repo repository.PaymentRepository, caseRepo repository.CaseRepository, provider payment.PaymentProvider, userRepo repository.UserRepository) *PaymentService {
	return &PaymentService{repo: repo, caseRepo: caseRepo, provider: provider, userRepo: userRepo}
}

func (s *PaymentService) InitiatePayment(ctx context.Context, caseID uuid.UUID, userIDStr string, role string, serviceType string, callbackURL string) (string, error) {
	// 0. Auth & Access Control
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return "", fmt.Errorf("invalid user_id")
	}

	c, err := s.caseRepo.GetByID(ctx, caseID)
	if err != nil {
		return "", fmt.Errorf("case not found")
	}

	if role == "applicant" && c.ApplicantID != userID {
		return "", fmt.Errorf("unauthorized: you do not own this case")
	}

	// 1. Get active tariff
	tariff, err := s.repo.GetActiveTariff(ctx, serviceType)
	if err != nil {
		return "", err
	}

	// 2. Get real user info
	user, err := s.userRepo.FindByID(ctx, c.ApplicantID)
	if err != nil {
		return "", fmt.Errorf("applicant not found")
	}
	mobile := user.Mobile

	// 3. Request from provider
	desc := fmt.Sprintf("پرداخت تعرفه %s برای پرونده %s", serviceType, caseID.String()[:8])
	token, paymentURL, err := s.provider.Request(ctx, tariff.MaxAmount, desc, callbackURL, mobile, "")
	if err != nil {
		return "", fmt.Errorf("provider request failed: %w", err)
	}

	// 4. Record in database
	pay := &repository.Payment{
		CaseID:        caseID,
		ServiceType:   serviceType,
		Amount:        tariff.MaxAmount,
		PaymentType:   "advance",
		Status:        "pending",
		PSPToken:      token,
		PaymentURL:    paymentURL,
		TariffVersion: tariff.Version,
	}
	
	_, err = s.repo.CreatePayment(ctx, pay)
	if err != nil {
		return "", err
	}

	return paymentURL, nil
}

func (s *PaymentService) VerifyPayment(ctx context.Context, token string) error {
	// 1. Get payment from db
	pay, err := s.repo.GetPaymentByToken(ctx, token)
	if err != nil {
		return err
	}

	if pay.Status == "paid" {
		return nil // already verified
	}

	// 2. Verify with provider
	refID, err := s.provider.Verify(ctx, pay.Amount, token)
	if err != nil {
		s.repo.UpdatePaymentStatus(ctx, pay.ID, "failed", "")
		return fmt.Errorf("verification failed: %w", err)
	}

	// 3. Update status
	return s.repo.UpdatePaymentStatus(ctx, pay.ID, "paid", refID)
}
