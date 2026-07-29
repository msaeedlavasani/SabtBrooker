package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/msaeedlavasani/SabtBrooker/backend/internal/repository"
	"github.com/msaeedlavasani/SabtBrooker/backend/internal/workflow"
)

// MapService orchestrates map service operations
type MapService struct {
	mapRepo  repository.MapServiceRepository
	caseSM   *workflow.StateMachine
	mapSM    *workflow.StateMachine
	auditRepo repository.AuditLogRepository
}

// NewMapService creates a new map service
func NewMapService(
	mapRepo repository.MapServiceRepository,
	caseSM *workflow.StateMachine,
	mapSM *workflow.StateMachine,
	auditRepo repository.AuditLogRepository,
) *MapService {
	return &MapService{
		mapRepo:   mapRepo,
		caseSM:    caseSM,
		mapSM:     mapSM,
		auditRepo: auditRepo,
	}
}

// Get returns a map service by ID
func (s *MapService) Get(ctx context.Context, id uuid.UUID) (*repository.MapService, error) {
	return s.mapRepo.GetByID(ctx, id)
}

// RequestConsent triggers OTP consent for the applicant
func (s *MapService) RequestConsent(ctx context.Context, id uuid.UUID) error {
	ms, err := s.mapRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if ms.Status != "expert_assigned" && ms.Status != "pending_expert_assignment" {
		return fmt.Errorf("درخواست رضایت فقط در وضعیت‌های اولیه سرویس نقشه امکان‌پذیر است")
	}

	return s.mapRepo.RequestConsent(ctx, id)
}

// VerifyConsent verifies the consent OTP
func (s *MapService) VerifyConsent(ctx context.Context, id uuid.UUID) error {
	return s.mapRepo.VerifyConsent(ctx, id)
}

// AssignExpert assigns an expert to the map service
func (s *MapService) AssignExpert(ctx context.Context, id uuid.UUID, expertID uuid.UUID) error {
	return s.mapRepo.AssignExpert(ctx, id, expertID)
}

// StartFieldwork transitions map to fieldwork_in_progress
func (s *MapService) StartFieldwork(ctx context.Context, id uuid.UUID, role string) error {
	if role != "survey_expert" {
		return fmt.Errorf("فقط کارشناس نقشه‌بردار می‌تواند عملیات میدانی را شروع کند")
	}

	return s.mapSM.Execute(ctx, id, "fieldwork_in_progress")
}

// SubmitFieldwork submits fieldwork data (map file + descriptive table)
func (s *MapService) SubmitFieldwork(ctx context.Context, id uuid.UUID, mapFileID string, descriptiveTable map[string]interface{}) error {
	return s.mapRepo.SubmitFieldwork(ctx, id, mapFileID, descriptiveTable)
}

// SubmitToOrg submits the map to the organization and transitions the case
func (s *MapService) SubmitToOrg(ctx context.Context, id uuid.UUID) (map[string]string, error) {
	// Validate: fieldwork must be done
	ms, err := s.mapRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if ms.Status != "fieldwork_done" {
		return nil, fmt.Errorf("ابتدا باید عملیات میدانی تکمیل و ثبت شود")
	}

	// Transition map to submitted
	if err := s.mapSM.Execute(ctx, id, "submitted_to_org"); err != nil {
		return nil, fmt.Errorf("امکان ارسال به سازمان وجود ندارد: %w", err)
	}

	// Simulated: submit and auto-approve (در نسخه نهایی: Outbox Pattern)
	caseID, trackingCode, err := s.mapRepo.SubmitToOrg(ctx, id, "submitted_to_org")
	if err != nil {
		return nil, err
	}

	// Transition case to map_completed
	if err := s.caseSM.Execute(ctx, caseID, "map_completed"); err != nil {
		return nil, fmt.Errorf("خطا در به‌روزرسانی وضعیت پرونده: %w", err)
	}

	// Audit log
	s.auditRepo.Record(ctx, &repository.AuditEvent{
		EventType:    "map.approved",
		ActorType:    "system",
		ResourceType: "map_service",
		ResourceID:   &id,
	})

	return map[string]string{
		"message":       "نقشه به سازمان ارسال و تایید شد",
		"tracking_code": trackingCode,
	}, nil
}
