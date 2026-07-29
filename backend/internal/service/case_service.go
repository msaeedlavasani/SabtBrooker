package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/msaeedlavasani/SabtBrooker/backend/internal/repository"
	"github.com/msaeedlavasani/SabtBrooker/backend/internal/workflow"
)

// CaseService orchestrates case operations
type CaseService struct {
	caseRepo   repository.CaseRepository
	userRepo   repository.UserRepository
	mapRepo    repository.MapServiceRepository
	claimRepo  repository.ClaimServiceRepository
	certRepo   repository.CertServiceRepository
	auditRepo  repository.AuditLogRepository
	caseSM     *workflow.StateMachine
}

// NewCaseService creates a new case service
func NewCaseService(
	caseRepo repository.CaseRepository,
	userRepo repository.UserRepository,
	mapRepo repository.MapServiceRepository,
	claimRepo repository.ClaimServiceRepository,
	certRepo repository.CertServiceRepository,
	auditRepo repository.AuditLogRepository,
	caseSM *workflow.StateMachine,
) *CaseService {
	return &CaseService{
		caseRepo:  caseRepo,
		userRepo:  userRepo,
		mapRepo:   mapRepo,
		claimRepo: claimRepo,
		certRepo:  certRepo,
		auditRepo: auditRepo,
		caseSM:    caseSM,
	}
}

// ListCases returns cases for a user based on role
func (s *CaseService) ListCases(ctx context.Context, userIDStr string, role string, offset, limit int) ([]repository.Case, error) {
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, fmt.Errorf("شناسه کاربر نامعتبر است")
	}

	switch role {
	case "applicant":
		return s.caseRepo.ListByApplicant(ctx, userID, offset, limit)
	case "admin", "auditor":
		return s.caseRepo.ListAll(ctx, offset, limit)
	default:
		return nil, fmt.Errorf("نقش کاربری نامعتبر است")
	}
}

// CreateCase creates a new case for an applicant
func (s *CaseService) CreateCase(ctx context.Context, userIDStr string, role string, input CreateCaseInput) (*repository.Case, error) {
	if role != "applicant" {
		return nil, fmt.Errorf("فقط متقاضی می‌تواند پرونده ایجاد کند")
	}
	userID, _ := uuid.Parse(userIDStr)

	c := &repository.Case{
		ApplicantID:       userID,
		ApplicantCapacity: "principal",
		Province:          input.Province,
		City:              input.City,
		District:          input.District,
		Village:           input.Village,
		AddressDetail:     input.AddressDetail,
	}

	created, err := s.caseRepo.Create(ctx, c)
	if err != nil {
		return nil, fmt.Errorf("خطا در ایجاد پرونده: %w", err)
	}

	// Audit log
	s.auditRepo.Record(ctx, &repository.AuditEvent{
		EventType:    "case.created",
		ActorType:    "applicant",
		ActorID:      &userID,
		ResourceType: "case",
		ResourceID:   &created.ID,
	})

	return created, nil
}

// GetCase returns a case by ID with access control
func (s *CaseService) GetCase(ctx context.Context, id uuid.UUID, userIDStr string, role string) (*repository.Case, error) {
	c, err := s.caseRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := s.checkAccess(c, userIDStr, role); err != nil {
		return nil, err
	}

	return c, nil
}

// UpdateCase updates case fields with access control
func (s *CaseService) UpdateCase(ctx context.Context, id uuid.UUID, userIDStr string, role string, fields map[string]interface{}) error {
	c, err := s.caseRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.checkAccess(c, userIDStr, role); err != nil {
		return err
	}

	return s.caseRepo.UpdateByID(ctx, id, fields)
}

// UpdateCapacity updates the applicant capacity with access control
func (s *CaseService) UpdateCapacity(ctx context.Context, id uuid.UUID, userIDStr string, role string, capacity string) error {
	valid := map[string]bool{"principal": true, "legal_rep_natural": true, "legal_rep_legal": true}
	if !valid[capacity] {
		return fmt.Errorf("سمت متقاضی نامعتبر است — مقادیر مجاز: اصیل، نماینده قانونی شخص حقیقی، نماینده قانونی شخص حقوقی")
	}

	c, err := s.caseRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.checkAccess(c, userIDStr, role); err != nil {
		return err
	}

	return s.caseRepo.UpdateCapacity(ctx, id, capacity)
}

// SubmitForMap transitions case from draft to map_in_progress
func (s *CaseService) SubmitForMap(ctx context.Context, caseID uuid.UUID, userIDStr string, role string) (*repository.MapService, error) {
	// Check case exists and is in draft status
	c, err := s.caseRepo.GetByID(ctx, caseID)
	if err != nil {
		return nil, fmt.Errorf("پرونده یافت نشد")
	}

	if err := s.checkAccess(c, userIDStr, role); err != nil {
		return nil, err
	}

	if c.Status != "draft" {
		return nil, fmt.Errorf("پرونده در وضعیت %s است — فقط پرونده‌های پیش‌نویس قابل شروع هستند", c.Status)
	}

	// Validate applicant identity (dev mode: relaxed)
	user, err := s.userRepo.FindByID(ctx, c.ApplicantID)
	if err != nil {
		return nil, fmt.Errorf("متقاضی یافت نشد")
	}
	if !user.MobileVerified {
		return nil, fmt.Errorf("شماره موبایل متقاضی تایید نشده است")
	}

	// Transition case to map_in_progress
	if err := s.caseSM.Execute(ctx, caseID, "map_in_progress"); err != nil {
		return nil, fmt.Errorf("شروع فرآیند نقشه امکان‌پذیر نیست: %w", err)
	}

	// Create map service
	ms, err := s.mapRepo.Create(ctx, caseID)
	if err != nil {
		return nil, fmt.Errorf("خطا در ایجاد سرویس نقشه: %w", err)
	}

	// Audit log
	actorID := c.ApplicantID
	s.auditRepo.Record(ctx, &repository.AuditEvent{
		EventType:    "map.expert_assigned",
		ActorType:    "system",
		ActorID:      &actorID,
		ResourceType: "map_service",
		ResourceID:   &ms.ID,
	})

	return ms, nil
}

func (s *CaseService) checkAccess(c *repository.Case, userIDStr string, role string) error {
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return fmt.Errorf("شناسه کاربر نامعتبر است")
	}

	switch role {
	case "admin", "auditor":
		return nil
	case "applicant":
		if c.ApplicantID != userID {
			return fmt.Errorf("دسترسی غیرمجاز: شما مالک این پرونده نیستید")
		}
		return nil
	default:
		return fmt.Errorf("نقش کاربری نامعتبر")
	}
}

// CreateCaseInput represents the input for creating a case
type CreateCaseInput struct {
	Province      string
	City          string
	District      *string
	Village       *string
	AddressDetail *string
}
