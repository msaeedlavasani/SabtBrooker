package handler

import (
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"

	"github.com/msaeedlavasani/SabtBrooker/backend/internal/workflow"
)

// ClaimHandler handles claim service endpoints
type ClaimHandler struct {
	db      *pgxpool.Pool
	caseSM  *workflow.StateMachine
	claimSM *workflow.StateMachine
}

// NewClaimHandler creates a new claim handler
func NewClaimHandler(db *pgxpool.Pool, caseSM, claimSM *workflow.StateMachine) *ClaimHandler {
	return &ClaimHandler{db: db, caseSM: caseSM, claimSM: claimSM}
}

// RegisterRoutes registers all claim service routes
func (h *ClaimHandler) RegisterRoutes(g *echo.Group) {
	g.GET("/:id", h.Get)
	g.POST("/:id/consent", h.RequestConsent)
	g.POST("/:id/consent/verify", h.VerifyConsent)
	g.PATCH("/:id", h.UpdateDetails)
	g.GET("/:id/documents", h.ListDocuments)
	g.POST("/:id/documents", h.AddDocument)
	g.DELETE("/:id/documents/:docId", h.DeleteDocument)
	g.POST("/:id/documents/verify", h.VerifyDocument)
	g.POST("/:id/submit", h.SubmitToOrg)
	g.POST("/:id/ai-advice", h.AiAdvice)
}

// Get returns claim service details
func (h *ClaimHandler) Get(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return BadRequest(c, "شناسه نامعتبر")
	}

	var cs struct {
		ID           uuid.UUID `json:"id"`
		CaseID       uuid.UUID `json:"case_id"`
		Status       string    `json:"status"`
		ClaimType    *string   `json:"claim_type"`
		OwnershipType *string  `json:"ownership_type"`
		MainPlate    *string   `json:"main_plate_number"`
		SubPlate     *string   `json:"sub_plate_number"`
		TrackingCode *string   `json:"tracking_code"`
		RejectReason *string   `json:"org_rejection_reason"`
		CreatedAt    string    `json:"created_at"`
	}

	err = h.db.QueryRow(c.Request().Context(), `
		SELECT id, case_id, status::text, claim_type::text, ownership_type::text,
		       main_plate_number, sub_plate_number, tracking_code,
		       org_rejection_reason, created_at::text
		FROM claim_services WHERE id = $1
	`, id).Scan(&cs.ID, &cs.CaseID, &cs.Status, &cs.ClaimType, &cs.OwnershipType,
		&cs.MainPlate, &cs.SubPlate, &cs.TrackingCode, &cs.RejectReason, &cs.CreatedAt)

	if err != nil {
		return NotFound(c, "سرویس ادعا یافت نشد")
	}

	return OK(c, cs)
}

// RequestConsent with false claim warning
func (h *ClaimHandler) RequestConsent(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return BadRequest(c, "شناسه نامعتبر")
	}

	_, err = h.db.Exec(c.Request().Context(), `
		UPDATE claim_services SET false_claim_warning_sent = true, updated_at = NOW()
		WHERE id = $1
	`, id)
	if err != nil {
		return Conflict(c, "امکان درخواست رضایت در این وضعیت وجود ندارد")
	}

	return OK(c, map[string]interface{}{
		"message": "کد تایید و هشدار قانونی ارسال شد",
		"warning": "توجه: درج ادعای واهی مطابق تبصره ۵ ماده ۱۰ قانون دارای تبعات قانونی است",
	})
}

// VerifyConsent verifies consent OTP
func (h *ClaimHandler) VerifyConsent(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return BadRequest(c, "شناسه نامعتبر")
	}

	var req struct {
		OTP string `json:"otp"`
		Ack bool   `json:"false_claim_acknowledged"`
	}
	if err := c.Bind(&req); err != nil {
		return BadRequest(c, "اطلاعات ورودی نامعتبر است")
	}
	if !req.Ack {
		return BadRequest(c, "باید صریحاً اعلام کنید که از تبعات قانونی ادعای واهی مطلع هستید")
	}

	_, err = h.db.Exec(c.Request().Context(), `
		UPDATE claim_services SET consent_granted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND consent_granted_at IS NULL
	`, id)
	if err != nil {
		return Conflict(c, "رضایت قبلاً ثبت شده است")
	}

	return OK(c, map[string]string{"message": "رضایت ثبت شد"})
}

// UpdateDetails updates claim details
func (h *ClaimHandler) UpdateDetails(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return BadRequest(c, "شناسه نامعتبر")
	}

	var req struct {
		ClaimType            string `json:"claim_type"`
		OwnershipType        string `json:"ownership_type"`
		MainPlateNumber      string `json:"main_plate_number"`
		SubPlateNumber       string `json:"sub_plate_number"`
		PlateSection         string `json:"plate_section"`
		TotalShare           int    `json:"total_share"`
		PartialShare         int    `json:"partial_share"`
		HasGovernmentRights  bool   `json:"has_government_rights"`
		TreasuryPaymentRef   string `json:"treasury_payment_ref"`
		LegalAdviceRequested bool   `json:"legal_advice_requested"`
		LegalAdviceMethod    string `json:"legal_advice_method"`
	}
	if err := c.Bind(&req); err != nil {
		return BadRequest(c, "اطلاعات ورودی نامعتبر است")
	}

	_, err = h.db.Exec(c.Request().Context(), `
		UPDATE claim_services SET
			claim_type=$2, ownership_type=$3, main_plate_number=$4, sub_plate_number=$5,
			plate_section=$6, total_share=$7, partial_share=$8,
			has_government_rights=$9, treasury_payment_ref=$10,
			legal_advice_requested=$11, legal_advice_method=$12,
			updated_at=NOW()
		WHERE id=$1
	`, id, req.ClaimType, req.OwnershipType, req.MainPlateNumber, req.SubPlateNumber,
		req.PlateSection, req.TotalShare, req.PartialShare,
		req.HasGovernmentRights, req.TreasuryPaymentRef,
		req.LegalAdviceRequested, req.LegalAdviceMethod)

	if err != nil {
		return InternalError(c, "خطا در به‌روزرسانی اطلاعات ادعا")
	}

	return OK(c, map[string]string{"message": "اطلاعات ادعا به‌روزرسانی شد"})
}

// ListDocuments lists claim documents
func (h *ClaimHandler) ListDocuments(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return BadRequest(c, "شناسه نامعتبر")
	}

	rows, err := h.db.Query(c.Request().Context(), `
		SELECT id, doc_type::text, file_path, description,
		       COALESCE(verified_at IS NOT NULL, false) as verified,
		       COALESCE(verification_note, '') as verification_note
		FROM claim_documents WHERE claim_service_id = $1
		ORDER BY uploaded_at
	`, id)
	if err != nil {
		return InternalError(c, "خطا در دریافت مستندات")
	}
	defer rows.Close()

	type doc struct {
		ID               uuid.UUID `json:"id"`
		DocType          string    `json:"doc_type"`
		FilePath         string    `json:"file_path"`
		Description      *string   `json:"description"`
		Verified         bool      `json:"verified"`
		VerificationNote string    `json:"verification_note"`
	}

	var docs []doc
	for rows.Next() {
		var d doc
		rows.Scan(&d.ID, &d.DocType, &d.FilePath, &d.Description, &d.Verified, &d.VerificationNote)
		docs = append(docs, d)
	}

	return OK(c, map[string]interface{}{"documents": docs})
}

// AddDocument adds a supporting document
func (h *ClaimHandler) AddDocument(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return BadRequest(c, "شناسه نامعتبر")
	}

	var req struct {
		DocType     string `json:"doc_type"`
		FileID      string `json:"file_id"`
		Description string `json:"description"`
	}
	if err := c.Bind(&req); err != nil {
		return BadRequest(c, "اطلاعات ورودی نامعتبر است")
	}

	var docID uuid.UUID
	err = h.db.QueryRow(c.Request().Context(), `
		INSERT INTO claim_documents (claim_service_id, doc_type, file_path, description)
		VALUES ($1, $2, $3, $4) RETURNING id
	`, id, req.DocType, req.FileID, req.Description).Scan(&docID)

	if err != nil {
		return InternalError(c, "خطا در افزودن مستند")
	}

	return Created(c, map[string]interface{}{"id": docID, "message": "مستند اضافه شد"})
}

// DeleteDocument removes a document
func (h *ClaimHandler) DeleteDocument(c echo.Context) error {
	_, err := h.db.Exec(c.Request().Context(), `
		DELETE FROM claim_documents WHERE id = $1 AND verified_at IS NULL
	`, c.Param("docId"))
	if err != nil {
		return NotFound(c, "مستند یافت نشد یا قبلاً تایید شده است")
	}
	return NoContent(c)
}

// VerifyDocument marks a document as verified/rejected by legal expert
func (h *ClaimHandler) VerifyDocument(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return BadRequest(c, "شناسه نامعتبر")
	}

	var req struct {
		DocumentID string `json:"document_id"`
		Verified   bool   `json:"verified"`
		Note       string `json:"note"`
	}
	if err := c.Bind(&req); err != nil {
		return BadRequest(c, "اطلاعات ورودی نامعتبر است")
	}

	_, err = h.db.Exec(c.Request().Context(), `
		UPDATE claim_documents SET
			verified_by = (SELECT legal_expert_id FROM cases c
			               JOIN claim_services cs ON c.id = cs.case_id WHERE cs.id = $2),
			verified_at = CASE WHEN $3 THEN NOW() ELSE NULL END,
			verification_note = $4
		WHERE id = $1
	`, req.DocumentID, id, req.Verified, req.Note)

	if err != nil {
		return InternalError(c, "خطا در ثبت تایید مستند")
	}

	return OK(c, map[string]string{"message": "نظر کارشناسی ثبت شد"})
}

// SubmitToOrg sends the claim to the organization
func (h *ClaimHandler) SubmitToOrg(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return BadRequest(c, "شناسه نامعتبر")
	}

	// Check all documents are verified
	var unverified int
	h.db.QueryRow(c.Request().Context(), `
		SELECT COUNT(*) FROM claim_documents
		WHERE claim_service_id = $1 AND verified_at IS NULL
	`, id).Scan(&unverified)

	if unverified > 0 {
		return Unprocessable(c, "تمام مستندات باید توسط کارشناس تایید شوند", map[string]int{"unverified_count": unverified})
	}

	if err := h.claimSM.Execute(c.Request().Context(), id, "submitted_to_org"); err != nil {
		return Unprocessable(c, err.Error(), nil)
	}

	// Simulate approval
	var caseID uuid.UUID
	h.db.QueryRow(c.Request().Context(), `
		UPDATE claim_services SET
			status = 'approved',
			tracking_code = 'CLAIM-' || substr(id::text, 1, 8),
			org_response_at = NOW()
		WHERE id = $1 RETURNING case_id
	`, id).Scan(&caseID)

	h.caseSM.Execute(c.Request().Context(), caseID, "claim_completed")

	return OK(c, map[string]string{
		"message":       "ادعا ثبت و تایید شد (شبیه‌سازی)",
		"tracking_code": "CLAIM-" + id.String()[:8],
	})
}

// AiAdvice provides AI-powered legal guidance
func (h *ClaimHandler) AiAdvice(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return BadRequest(c, "شناسه نامعتبر")
	}

	// Get claim context
	var claimType, ownershipType string
	h.db.QueryRow(c.Request().Context(), `
		SELECT COALESCE(claim_type::text, ''), COALESCE(ownership_type::text, '')
		FROM claim_services WHERE id = $1
	`, id).Scan(&claimType, &ownershipType)

	// Simple rule-based advice (in production: call AI/RAG service)
	advice := generateLegalAdvice(claimType, ownershipType)

	var adviceID uuid.UUID
	h.db.QueryRow(c.Request().Context(), `
		INSERT INTO ai_advice_logs (case_id, claim_service_id, input_context,
			recommended_action, legal_references, confidence_score, model_version)
		SELECT cs.case_id, cs.id, '{}'::jsonb, $2, $3, $4, 'rule-v1.0'
		FROM claim_services cs WHERE cs.id = $1
		RETURNING id
	`, id, advice.Action, advice.References, advice.Confidence).Scan(&adviceID)

	slog.Info("AI advice generated", "claim_id", id, "advice_id", adviceID)

	return OK(c, map[string]interface{}{
		"id":                 adviceID,
		"recommended_action": advice.Action,
		"legal_references":   advice.References,
		"confidence_score":   advice.Confidence,
		"disclaimer":         "این یک نظر رسمی حقوقی نیست و صرفاً راهنمایی برای انتخاب نوع اقدام قانونی است",
	})
}

type legalAdvice struct {
	Action     string   `json:"action"`
	References []string `json:"references"`
	Confidence float64  `json:"confidence"`
}

func generateLegalAdvice(claimType, ownershipType string) legalAdvice {
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
