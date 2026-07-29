package repository

import (
	"context"

	"github.com/google/uuid"
)

// UserRepository — مدیریت کاربران
type UserRepository interface {
	FindByMobile(ctx context.Context, mobile string) (*User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*User, error)
	FindOrCreateByMobile(ctx context.Context, mobile string) (*User, error)
	UpdateMobileVerified(ctx context.Context, id uuid.UUID, verified bool) error
}

// CaseRepository — مدیریت پرونده‌ها
type CaseRepository interface {
	ListByApplicant(ctx context.Context, applicantID uuid.UUID, offset, limit int) ([]Case, error)
	ListAll(ctx context.Context, offset, limit int) ([]Case, error)
	Create(ctx context.Context, c *Case) (*Case, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Case, error)
	UpdateByID(ctx context.Context, id uuid.UUID, fields map[string]interface{}) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
	UpdateCapacity(ctx context.Context, id uuid.UUID, capacity string) error
}

// MapServiceRepository — مدیریت سرویس نقشه
type MapServiceRepository interface {
	Create(ctx context.Context, caseID uuid.UUID) (*MapService, error)
	GetByID(ctx context.Context, id uuid.UUID) (*MapService, error)
	RequestConsent(ctx context.Context, id uuid.UUID) error
	VerifyConsent(ctx context.Context, id uuid.UUID) error
	AssignExpert(ctx context.Context, id uuid.UUID, expertID uuid.UUID) error
	SubmitFieldwork(ctx context.Context, id uuid.UUID, input SubmitFieldworkInput) error
	SubmitToOrg(ctx context.Context, id uuid.UUID, status string) (caseID uuid.UUID, trackingCode string, err error)
}

// ClaimServiceRepository — مدیریت سرویس ادعا
type ClaimServiceRepository interface {
	Create(ctx context.Context, caseID uuid.UUID) (*ClaimService, error)
	GetByID(ctx context.Context, id uuid.UUID) (*ClaimService, error)
	RequestConsent(ctx context.Context, id uuid.UUID) error
	VerifyConsent(ctx context.Context, id uuid.UUID) error
	UpdateDetails(ctx context.Context, id uuid.UUID, input UpdateClaimInput) error
	AddDocument(ctx context.Context, claimID uuid.UUID, docType, fileID, description string) (*ClaimDocument, error)
	DeleteDocument(ctx context.Context, docID uuid.UUID) error
	ListDocuments(ctx context.Context, claimID uuid.UUID) ([]ClaimDocument, error)
	VerifyDocument(ctx context.Context, docID uuid.UUID, claimID uuid.UUID, verified bool, note string) error
	CountUnverifiedDocuments(ctx context.Context, claimID uuid.UUID) (int, error)
	SubmitToOrg(ctx context.Context, id uuid.UUID) (caseID uuid.UUID, trackingCode string, err error)
}

// CertServiceRepository — مدیریت سرویس گواهی اقدام
type CertServiceRepository interface {
	Create(ctx context.Context, caseID uuid.UUID) (*CertService, error)
	GetByID(ctx context.Context, id uuid.UUID) (*CertService, error)
	RequestConsent(ctx context.Context, id uuid.UUID) error
	VerifyConsent(ctx context.Context, id uuid.UUID) error
	UpdateDetails(ctx context.Context, id uuid.UUID, input UpdateCertInput) error
	SubmitToOrg(ctx context.Context, id uuid.UUID) (caseID uuid.UUID, trackingCode string, err error)
}

// AuditLogRepository — لاگ ممیزی (append-only)
type AuditLogRepository interface {
	Record(ctx context.Context, event *AuditEvent) error
}

// NotificationRepository — مدیریت اطلاع‌رسانی
type NotificationRepository interface {
	Create(ctx context.Context, n *Notification) error
}

// AIAdviceRepository — لاگ راهنمایی هوش مصنوعی
type AIAdviceRepository interface {
	Save(ctx context.Context, claimID uuid.UUID, advice *AIAdviceLog) (uuid.UUID, error)
}
