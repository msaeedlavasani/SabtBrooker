package repository

import (
	"time"

	"github.com/google/uuid"
)

// مدل‌های دامین — مستقل از دیتابیس

type User struct {
	ID             uuid.UUID  `json:"id"`
	NationalID     string     `json:"national_id"`
	FirstName      string     `json:"first_name"`
	LastName       string     `json:"last_name"`
	Mobile         string     `json:"mobile"`
	MobileVerified bool       `json:"mobile_verified"`
	BirthDate      *time.Time `json:"birth_date,omitempty"`
	Role           string     `json:"role"`
	SanaStatus     *string    `json:"sana_status,omitempty"`
	IsAlive        bool       `json:"is_alive"`
	IsActive       bool       `json:"is_active"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type Case struct {
	ID                 uuid.UUID  `json:"id"`
	ApplicantID        uuid.UUID  `json:"applicant_id"`
	ApplicantCapacity  string     `json:"applicant_capacity"`
	Status             string     `json:"status"`
	Province           string     `json:"province"`
	City               string     `json:"city"`
	District           *string    `json:"district,omitempty"`
	Village            *string    `json:"village,omitempty"`
	AddressDetail      *string    `json:"address_detail,omitempty"`
	LegalExpertID      *uuid.UUID `json:"legal_expert_id,omitempty"`
	SurveyExpertID     *uuid.UUID `json:"survey_expert_id,omitempty"`
	ProxyVerified      bool       `json:"proxy_verified"`
	MapTrackingCode    *string    `json:"map_tracking_code,omitempty"`
	ClaimTrackingCode  *string    `json:"claim_tracking_code,omitempty"`
	CertTrackingCode   *string    `json:"cert_tracking_code,omitempty"`
	Deadline2Years     *time.Time `json:"deadline_2years,omitempty"`
	ApplicantDeceased  bool       `json:"applicant_deceased"`
	Deadline5Months    *time.Time `json:"deadline_5months,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
	CompletedAt        *time.Time `json:"completed_at,omitempty"`
}

type MapService struct {
	ID                   uuid.UUID  `json:"id"`
	CaseID               uuid.UUID  `json:"case_id"`
	Status               string     `json:"status"`
	PropertyType         *string    `json:"property_type,omitempty"`
	ApproxAreaSqm        *float64   `json:"approx_area_sqm,omitempty"`
	LandUse              *string    `json:"land_use,omitempty"`
	OwnershipType        *string    `json:"ownership_type,omitempty"`
	HasBuilding          bool       `json:"has_building"`
	AnnexCount           int        `json:"annex_count"`
	GeoLatitude          *float64   `json:"geo_latitude,omitempty"`
	GeoLongitude         *float64   `json:"geo_longitude,omitempty"`
	GrantAccessToOthers  bool       `json:"grant_access_to_others"`
	ConsentGrantedAt     *time.Time `json:"consent_granted_at,omitempty"`
	FieldworkStartedAt   *time.Time `json:"fieldwork_started_at,omitempty"`
	FieldworkCompletedAt *time.Time `json:"fieldwork_completed_at,omitempty"`
	MapFilePath          *string    `json:"map_file_path,omitempty"`
	MapFormat            *string    `json:"map_format,omitempty"`
	DescriptiveTable     []byte     `json:"descriptive_table,omitempty"`
	SubmittedToOrgAt     *time.Time `json:"submitted_to_org_at,omitempty"`
	TrackingCode         *string    `json:"tracking_code,omitempty"`
	OrgResponseAt        *time.Time `json:"org_response_at,omitempty"`
	OrgRejectionReason   *string    `json:"org_rejection_reason,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

type ClaimService struct {
	ID                    uuid.UUID  `json:"id"`
	CaseID                uuid.UUID  `json:"case_id"`
	Status                string     `json:"status"`
	MapTrackingCode       *string    `json:"map_tracking_code,omitempty"`
	MapTrackingValid      bool       `json:"map_tracking_valid"`
	FalseClaimWarningSent bool       `json:"false_claim_warning_sent"`
	ConsentGrantedAt      *time.Time `json:"consent_granted_at,omitempty"`
	ClaimType             *string    `json:"claim_type,omitempty"`
	OwnershipType         *string    `json:"ownership_type,omitempty"`
	MainPlateNumber       *string    `json:"main_plate_number,omitempty"`
	SubPlateNumber        *string    `json:"sub_plate_number,omitempty"`
	PlateSection          *string    `json:"plate_section,omitempty"`
	TotalShare            *int       `json:"total_share,omitempty"`
	PartialShare          *int       `json:"partial_share,omitempty"`
	SubmittedToOrgAt      *time.Time `json:"submitted_to_org_at,omitempty"`
	TrackingCode          *string    `json:"tracking_code,omitempty"`
	OrgResponseAt         *time.Time `json:"org_response_at,omitempty"`
	OrgRejectionReason    *string    `json:"org_rejection_reason,omitempty"`
	HasGovernmentRights   bool       `json:"has_government_rights"`
	TreasuryPaymentRef    *string    `json:"treasury_payment_ref,omitempty"`
	LegalAdviceRequested  bool       `json:"legal_advice_requested"`
	LegalAdviceMethod     *string    `json:"legal_advice_method,omitempty"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}

type CertService struct {
	ID               uuid.UUID  `json:"id"`
	CaseID           uuid.UUID  `json:"case_id"`
	Status           string     `json:"status"`
	ClaimTrackingCode *string   `json:"claim_tracking_code,omitempty"`
	ClaimTrackingValid bool     `json:"claim_tracking_valid"`
	ConsentGrantedAt *time.Time `json:"consent_granted_at,omitempty"`
	ActionReference  *string    `json:"action_reference,omitempty"`
	ActionType       *string    `json:"action_type,omitempty"`
	ActionDate       *time.Time `json:"action_date,omitempty"`
	CertImagePath    *string    `json:"cert_image_path,omitempty"`
	CertUniqueID     *string    `json:"cert_unique_id,omitempty"`
	SubmittedToOrgAt *time.Time `json:"submitted_to_org_at,omitempty"`
	TrackingCode     *string    `json:"tracking_code,omitempty"`
	OrgResponseAt    *time.Time `json:"org_response_at,omitempty"`
	OrgRejectionReason *string  `json:"org_rejection_reason,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type ClaimDocument struct {
	ID              uuid.UUID  `json:"id"`
	ClaimServiceID  uuid.UUID  `json:"claim_service_id"`
	DocType         string     `json:"doc_type"`
	FilePath        string     `json:"file_path"`
	Description     *string    `json:"description,omitempty"`
	Verified        bool       `json:"verified"`
	VerificationNote *string   `json:"verification_note,omitempty"`
	UploadedAt      time.Time  `json:"uploaded_at"`
}

type AuditEvent struct {
	ID            uuid.UUID  `json:"id"`
	EventType     string     `json:"event_type"`
	ActorType     string     `json:"actor_type"`
	ActorID       *uuid.UUID `json:"actor_id,omitempty"`
	ActorIP       *string    `json:"actor_ip,omitempty"`
	ResourceType  string     `json:"resource_type"`
	ResourceID    *uuid.UUID `json:"resource_id,omitempty"`
	Changes       []byte     `json:"changes,omitempty"`
	Metadata      []byte     `json:"metadata,omitempty"`
	CorrelationID *uuid.UUID `json:"correlation_id,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

type Notification struct {
	ID           uuid.UUID `json:"id"`
	UserID       uuid.UUID `json:"user_id"`
	CaseID       *uuid.UUID `json:"case_id,omitempty"`
	Channel      string    `json:"channel"`
	TemplateKey  string    `json:"template_key"`
	Content      string    `json:"content"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

type AIAdviceLog struct {
	ID         uuid.UUID `json:"id"`
	CaseID     uuid.UUID `json:"case_id"`
	Action     string    `json:"recommended_action"`
	References []string  `json:"legal_references"`
	Confidence float64   `json:"confidence_score"`
}

// Input structs for mutations

type SubmitFieldworkInput struct {
	PropertyType        string                 `json:"property_type"`
	ApproxAreaSqm       float64                `json:"approx_area_sqm"`
	LandUse             string                 `json:"land_use"`
	OwnershipType       string                 `json:"ownership_type"`
	HasBuilding         bool                   `json:"has_building"`
	AnnexCount          int                    `json:"annex_count"`
	GeoLatitude         float64                `json:"geo_latitude"`
	GeoLongitude        float64                `json:"geo_longitude"`
	MapFilePath         string                 `json:"map_file_path"`
	MapFormat           string                 `json:"map_format"`
	DescriptiveTable    map[string]interface{} `json:"descriptive_table"`
	Photos              []MapPhotoInput        `json:"photos"`
	GrantAccessToOthers bool                   `json:"grant_access_to_others"`
}

type MapPhotoInput struct {
	FilePath  string  `json:"file_path"`
	Side      string  `json:"side"` // north, south, east, west
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type UpdateClaimInput struct {
	ClaimType            string                 `json:"claim_type"`
	OwnershipType        string                 `json:"ownership_type"`
	MainPlateNumber      string                 `json:"main_plate_number"`
	SubPlateNumber       string                 `json:"sub_plate_number"`
	PlateSection         string                 `json:"plate_section"`
	TotalShare           int                    `json:"total_share"`
	PartialShare         int                    `json:"partial_share"`
	HasGovernmentRights  bool                   `json:"has_government_rights"`
	TreasuryPaymentRef   string                 `json:"treasury_payment_ref"`
	LegalAdviceRequested bool                   `json:"legal_advice_requested"`
	LegalAdviceMethod    string                 `json:"legal_advice_method"`
	Documents            []ClaimDocumentInput   `json:"documents"`
}

type ClaimDocumentInput struct {
	DocType     string `json:"doc_type"`
	FilePath    string `json:"file_path"`
	Description string `json:"description"`
}

type UpdateCertInput struct {
	ActionReference string    `json:"action_reference"`
	ActionType      string    `json:"action_type"`
	ActionDate      time.Time `json:"action_date"`
	CertImagePath   string    `json:"cert_image_path"`
	CertUniqueID    string    `json:"cert_unique_id"`
}
