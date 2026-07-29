package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/msaeedlavasani/SabtBrooker/backend/internal/repository"
	"github.com/msaeedlavasani/SabtBrooker/backend/internal/workflow"
)

// CertService orchestrates certificate service operations
type CertService struct {
	certRepo  repository.CertServiceRepository
	caseSM    *workflow.StateMachine
	certSM    *workflow.StateMachine
	auditRepo repository.AuditLogRepository
}

// NewCertService creates a new cert service
func NewCertService(
	certRepo repository.CertServiceRepository,
	caseSM *workflow.StateMachine,
	certSM *workflow.StateMachine,
	auditRepo repository.AuditLogRepository,
) *CertService {
	return &CertService{
		certRepo:  certRepo,
		caseSM:    caseSM,
		certSM:    certSM,
		auditRepo: auditRepo,
	}
}

// Get returns a cert service by ID
func (s *CertService) Get(ctx context.Context, id uuid.UUID) (*repository.CertService, error) {
	return s.certRepo.GetByID(ctx, id)
}

// RequestConsent triggers OTP consent
func (s *CertService) RequestConsent(ctx context.Context, id uuid.UUID) error {
	return s.certRepo.RequestConsent(ctx, id)
}

// VerifyConsent verifies consent OTP
func (s *CertService) VerifyConsent(ctx context.Context, id uuid.UUID) error {
	return s.certRepo.VerifyConsent(ctx, id)
}

// UpdateDetails updates certificate metadata
func (s *CertService) UpdateDetails(ctx context.Context, id uuid.UUID, input UpdateCertInput) error {
	fields := make(map[string]interface{})
	if input.ActionReference != nil {
		fields["action_reference"] = *input.ActionReference
	}
	if input.ActionType != nil {
		fields["action_type"] = *input.ActionType
	}
	if input.ActionDate != nil {
		t, err := time.Parse("2006-01-02", *input.ActionDate)
		if err != nil {
			return fmt.Errorf("فرمت تاریخ نامعتبر است — باید YYYY-MM-DD باشد")
		}
		fields["action_date"] = t
	}
	if input.CertImageID != nil {
		fields["cert_image_path"] = *input.CertImageID
	}
	if input.CertUniqueID != nil {
		fields["cert_unique_id"] = *input.CertUniqueID
	}
	return s.certRepo.UpdateDetails(ctx, id, fields)
}

// SubmitToOrg submits the certificate to the organization (final step)
func (s *CertService) SubmitToOrg(ctx context.Context, id uuid.UUID) (map[string]string, error) {
	// Validate: must be in pending_data status
	cs, err := s.certRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if cs.Status != "pending_data" {
		return nil, fmt.Errorf("ابتدا باید اطلاعات گواهی تکمیل شود")
	}

	// Transition cert to submitted
	if err := s.certSM.Execute(ctx, id, "submitted_to_org"); err != nil {
		return nil, fmt.Errorf("امکان ارسال گواهی وجود ندارد: %w", err)
	}

	// Simulated: submit and auto-approve
	caseID, trackingCode, err := s.certRepo.SubmitToOrg(ctx, id)
	if err != nil {
		return nil, err
	}

	// Transition case to cert_completed (end of chain)
	if err := s.caseSM.Execute(ctx, caseID, "cert_completed"); err != nil {
		return nil, fmt.Errorf("خطا در تکمیل پرونده: %w", err)
	}

	// Audit log
	s.auditRepo.Record(ctx, &repository.AuditEvent{
		EventType:    "cert.approved",
		ActorType:    "system",
		ResourceType: "cert_service",
		ResourceID:   &id,
	})

	return map[string]string{
		"message":       "گواهی اقدام ثبت و تایید شد — فرآیند تکمیل گردید",
		"tracking_code": trackingCode,
	}, nil
}

// UpdateCertInput for updating certificate details
type UpdateCertInput struct {
	ActionReference *string
	ActionType      *string
	ActionDate      *string
	CertImageID     *string
	CertUniqueID    *string
}
