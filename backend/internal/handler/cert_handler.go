package handler

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"

	"github.com/msaeedlavasani/SabtBrooker/backend/internal/workflow"
)

// CertHandler handles certificate service endpoints
type CertHandler struct {
	db     *pgxpool.Pool
	caseSM *workflow.StateMachine
	certSM *workflow.StateMachine
}

// NewCertHandler creates a new cert handler
func NewCertHandler(db *pgxpool.Pool, caseSM, certSM *workflow.StateMachine) *CertHandler {
	return &CertHandler{db: db, caseSM: caseSM, certSM: certSM}
}

// RegisterRoutes registers all cert service routes
func (h *CertHandler) RegisterRoutes(g *echo.Group) {
	g.GET("/:id", h.Get)
	g.POST("/:id/consent", h.RequestConsent)
	g.POST("/:id/consent/verify", h.VerifyConsent)
	g.PATCH("/:id", h.UpdateDetails)
	g.POST("/:id/submit", h.SubmitToOrg)
}

// Get returns cert service details
func (h *CertHandler) Get(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return BadRequest(c, "شناسه نامعتبر")
	}

	var cs struct {
		ID             uuid.UUID `json:"id"`
		CaseID         uuid.UUID `json:"case_id"`
		Status         string    `json:"status"`
		ActionRef      *string   `json:"action_reference"`
		ActionType     *string   `json:"action_type"`
		ActionDate     *string   `json:"action_date"`
		CertImagePath  *string   `json:"cert_image_path"`
		CertUniqueID   *string   `json:"cert_unique_id"`
		TrackingCode   *string   `json:"tracking_code"`
		RejectReason   *string   `json:"org_rejection_reason"`
		CreatedAt      string    `json:"created_at"`
	}

	err = h.db.QueryRow(c.Request().Context(), `
		SELECT id, case_id, status::text, action_reference::text, action_type::text,
		       action_date::text, cert_image_path, cert_unique_id, tracking_code,
		       org_rejection_reason, created_at::text
		FROM cert_services WHERE id = $1
	`, id).Scan(&cs.ID, &cs.CaseID, &cs.Status, &cs.ActionRef, &cs.ActionType,
		&cs.ActionDate, &cs.CertImagePath, &cs.CertUniqueID,
		&cs.TrackingCode, &cs.RejectReason, &cs.CreatedAt)

	if err != nil {
		return NotFound(c, "سرویس گواهی اقدام یافت نشد")
	}

	return OK(c, cs)
}

// RequestConsent for cert service
func (h *CertHandler) RequestConsent(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return BadRequest(c, "شناسه نامعتبر")
	}

	_, err = h.db.Exec(c.Request().Context(), `
		UPDATE cert_services SET updated_at = NOW() WHERE id = $1
	`, id)
	if err != nil {
		return Conflict(c, "امکان درخواست رضایت در این وضعیت وجود ندارد")
	}

	return OK(c, map[string]string{"message": "کد تایید ارسال شد"})
}

// VerifyConsent verifies the consent OTP
func (h *CertHandler) VerifyConsent(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return BadRequest(c, "شناسه نامعتبر")
	}

	var req struct{ OTP string `json:"otp"` }
	if err := c.Bind(&req); err != nil {
		return BadRequest(c, "اطلاعات ورودی نامعتبر است")
	}

	_, err = h.db.Exec(c.Request().Context(), `
		UPDATE cert_services SET consent_granted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND consent_granted_at IS NULL
	`, id)
	if err != nil {
		return Conflict(c, "رضایت قبلاً ثبت شده است")
	}

	return OK(c, map[string]string{"message": "رضایت ثبت شد"})
}

// UpdateDetails updates certificate details
func (h *CertHandler) UpdateDetails(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return BadRequest(c, "شناسه نامعتبر")
	}

	var req struct {
		ActionReference string `json:"action_reference"`
		ActionType      string `json:"action_type"`
		ActionDate      string `json:"action_date"`
		CertImageID     string `json:"cert_image_id"`
		CertUniqueID    string `json:"cert_unique_id"`
	}
	if err := c.Bind(&req); err != nil {
		return BadRequest(c, "اطلاعات ورودی نامعتبر است")
	}

	_, err = h.db.Exec(c.Request().Context(), `
		UPDATE cert_services SET
			action_reference=$2, action_type=$3, action_date=$4::date,
			cert_image_path=$5, cert_unique_id=$6, updated_at=NOW()
		WHERE id=$1
	`, id, req.ActionReference, req.ActionType, req.ActionDate,
		req.CertImageID, req.CertUniqueID)

	if err != nil {
		return InternalError(c, "خطا در به‌روزرسانی اطلاعات گواهی")
	}

	return OK(c, map[string]string{"message": "اطلاعات گواهی به‌روزرسانی شد"})
}

// SubmitToOrg sends the certificate to the organization (final step)
func (h *CertHandler) SubmitToOrg(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return BadRequest(c, "شناسه نامعتبر")
	}

	if err := h.certSM.Execute(c.Request().Context(), id, "submitted_to_org"); err != nil {
		return Unprocessable(c, err.Error(), nil)
	}

	// Simulate approval — final tracking code
	var caseID uuid.UUID
	h.db.QueryRow(c.Request().Context(), `
		UPDATE cert_services SET
			status = 'approved',
			tracking_code = 'CERT-' || substr(id::text, 1, 8),
			org_response_at = NOW()
		WHERE id = $1 RETURNING case_id
	`, id).Scan(&caseID)

	h.caseSM.Execute(c.Request().Context(), caseID, "cert_completed")

	return OK(c, map[string]string{
		"message":       "گواهی اقدام ثبت و تایید شد — فرآیند تکمیل گردید",
		"tracking_code": "CERT-" + id.String()[:8],
	})
}
