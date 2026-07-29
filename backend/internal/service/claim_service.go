package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/msaeedlavasani/SabtBrooker/backend/internal/repository"
	"github.com/msaeedlavasani/SabtBrooker/backend/internal/workflow"
)

// ClaimService orchestrates claim service operations
type ClaimService struct {
	claimRepo    repository.ClaimServiceRepository
	caseSM       *workflow.StateMachine
	claimSM      *workflow.StateMachine
	auditRepo    repository.AuditLogRepository
	aiAdviceRepo repository.AIAdviceRepository
}

// NewClaimService creates a new claim service
func NewClaimService(
	claimRepo repository.ClaimServiceRepository,
	caseSM *workflow.StateMachine,
	claimSM *workflow.StateMachine,
	auditRepo repository.AuditLogRepository,
	aiAdviceRepo repository.AIAdviceRepository,
) *ClaimService {
	return &ClaimService{
		claimRepo:    claimRepo,
		caseSM:       caseSM,
		claimSM:      claimSM,
		auditRepo:    auditRepo,
		aiAdviceRepo: aiAdviceRepo,
	}
}

// Get returns a claim service by ID
func (s *ClaimService) Get(ctx context.Context, id uuid.UUID) (*repository.ClaimService, error) {
	return s.claimRepo.GetByID(ctx, id)
}

// RequestConsent sends OTP + false claim warning
func (s *ClaimService) RequestConsent(ctx context.Context, id uuid.UUID) error {
	return s.claimRepo.RequestConsent(ctx, id)
}

// VerifyConsent verifies consent OTP after false claim warning acknowledged
func (s *ClaimService) VerifyConsent(ctx context.Context, id uuid.UUID, ack bool) error {
	if !ack {
		return fmt.Errorf("باید صریحاً اعلام کنید که از تبعات قانونی ادعای واهی مطلع هستید")
	}
	return s.claimRepo.VerifyConsent(ctx, id)
}

// UpdateDetails updates claim metadata and documents
func (s *ClaimService) UpdateDetails(ctx context.Context, id uuid.UUID, input repository.UpdateClaimInput) error {
	return s.claimRepo.UpdateDetails(ctx, id, input)
}

// AddDocument adds a supporting document to the claim
func (s *ClaimService) AddDocument(ctx context.Context, claimID uuid.UUID, docType, fileID, description string) (*repository.ClaimDocument, error) {
	return s.claimRepo.AddDocument(ctx, claimID, docType, fileID, description)
}

// DeleteDocument removes a document (only if not verified)
func (s *ClaimService) DeleteDocument(ctx context.Context, docID uuid.UUID) error {
	return s.claimRepo.DeleteDocument(ctx, docID)
}

// ListDocuments returns all documents for a claim
func (s *ClaimService) ListDocuments(ctx context.Context, claimID uuid.UUID) ([]repository.ClaimDocument, error) {
	return s.claimRepo.ListDocuments(ctx, claimID)
}

// VerifyDocument marks a document as verified/rejected
func (s *ClaimService) VerifyDocument(ctx context.Context, docID uuid.UUID, claimID uuid.UUID, verified bool, note string) error {
	return s.claimRepo.VerifyDocument(ctx, docID, claimID, verified, note)
}

// SubmitToOrg submits the claim to the organization after verification
func (s *ClaimService) SubmitToOrg(ctx context.Context, id uuid.UUID) (map[string]string, error) {
	// Pre-condition: all documents must be verified
	unverifiedCount, err := s.claimRepo.CountUnverifiedDocuments(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("خطا در بررسی مستندات: %w", err)
	}
	if unverifiedCount > 0 {
		return nil, fmt.Errorf("تمام مستندات باید توسط کارشناس تایید شوند — %d مستند تایید نشده", unverifiedCount)
	}

	// Transition claim to submitted
	if err := s.claimSM.Execute(ctx, id, "submitted_to_org"); err != nil {
		return nil, fmt.Errorf("امکان ارسال ادعا وجود ندارد: %w", err)
	}

	// Simulated: submit and auto-approve
	caseID, trackingCode, err := s.claimRepo.SubmitToOrg(ctx, id)
	if err != nil {
		return nil, err
	}

	// Transition case to claim_completed
	if err := s.caseSM.Execute(ctx, caseID, "claim_completed"); err != nil {
		return nil, fmt.Errorf("خطا در به‌روزرسانی وضعیت پرونده: %w", err)
	}

	// Audit log
	s.auditRepo.Record(ctx, &repository.AuditEvent{
		EventType:    "claim.approved",
		ActorType:    "system",
		ResourceType: "claim_service",
		ResourceID:   &id,
	})

	return map[string]string{
		"message":       "ادعا ثبت و تایید شد",
		"tracking_code": trackingCode,
	}, nil
}

// GenerateAIAdvice generates legal advice based on claim data
func (s *ClaimService) GenerateAIAdvice(ctx context.Context, id uuid.UUID) (map[string]interface{}, error) {
	cs, err := s.claimRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	claimType := ""
	ownershipType := ""
	if cs.ClaimType != nil {
		claimType = *cs.ClaimType
	}
	if cs.OwnershipType != nil {
		ownershipType = *cs.OwnershipType
	}

	advice := generateLegalAdvice(claimType, ownershipType)

	adviceID, err := s.aiAdviceRepo.Save(ctx, id, &repository.AIAdviceLog{
		CaseID:     cs.CaseID,
		Action:     advice.Action,
		References: advice.References,
		Confidence: advice.Confidence,
	})
	if err != nil {
		return nil, fmt.Errorf("خطا در ذخیره راهنمایی: %w", err)
	}

	return map[string]interface{}{
		"id":                 adviceID,
		"recommended_action": advice.Action,
		"legal_references":   advice.References,
		"confidence_score":   advice.Confidence,
		"disclaimer":         "این یک نظر رسمی حقوقی نیست و صرفاً راهنمایی برای انتخاب نوع اقدام قانونی است",
	}, nil
}

// Legal advice struct
type legalAdvice struct {
	Action     string
	References []string
	Confidence float64
}

func generateLegalAdvice(claimType, ownershipType string) legalAdvice {
	if strings.TrimSpace(ownershipType) != "" {
		_ = ownershipType // برای استفاده در نسخه‌های بعدی با AI واقعی
	}

	switch claimType {
	case "ownership":
		return legalAdvice{
			Action:     "طرح دعوای اثبات مالکیت در مراجع قضایی",
			References: []string{"ماده ۲۲ قانون ثبت", "ماده ۱۰ قانون الزام به ثبت رسمی معاملات"},
			Confidence: 0.85,
		}
	case "easement":
		return legalAdvice{
			Action:     "طرح دعوای اثبات حق ارتفاق در مراجع قضایی",
			References: []string{"ماده ۹۳ قانون مدنی", "ماده ۱۰ قانون الزام"},
			Confidence: 0.80,
		}
	default:
		return legalAdvice{
			Action:     "مراجعه به کارشناس ثبتی برای تعیین مسیر حقوقی مناسب",
			References: []string{"ماده ۱۰ قانون الزام به ثبت رسمی معاملات"},
			Confidence: 0.70,
		}
	}
}
