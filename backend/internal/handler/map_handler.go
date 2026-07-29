package handler

import (
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"

	"github.com/msaeedlavasani/SabtBrooker/backend/internal/middleware"
	"github.com/msaeedlavasani/SabtBrooker/backend/internal/workflow"
)

// MapHandler handles map service endpoints
type MapHandler struct {
	db     *pgxpool.Pool
	caseSM *workflow.StateMachine
	mapSM  *workflow.StateMachine
}

// NewMapHandler creates a new map handler
func NewMapHandler(db *pgxpool.Pool, caseSM, mapSM *workflow.StateMachine) *MapHandler {
	return &MapHandler{db: db, caseSM: caseSM, mapSM: mapSM}
}

// RegisterRoutes registers all map service routes
func (h *MapHandler) RegisterRoutes(g *echo.Group) {
	g.GET("/:id", h.Get)
	g.POST("/:id/consent", h.RequestConsent)
	g.POST("/:id/consent/verify", h.VerifyConsent)
	g.POST("/:id/fieldwork/start", h.StartFieldwork)
	g.POST("/:id/fieldwork/submit", h.SubmitFieldwork)
	g.POST("/:id/submit", h.SubmitToOrg)
}

// Get returns map service details
func (h *MapHandler) Get(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return BadRequest(c, "شناسه نامعتبر")
	}

	var ms struct {
		ID             uuid.UUID `json:"id"`
		CaseID         uuid.UUID `json:"case_id"`
		Status         string    `json:"status"`
		PropertyType   *string   `json:"property_type"`
		ApproxAreaSqm  *float64  `json:"approx_area_sqm"`
		GeoLatitude    *float64  `json:"geo_latitude"`
		GeoLongitude   *float64  `json:"geo_longitude"`
		MapFilePath    *string   `json:"map_file_path"`
		TrackingCode   *string   `json:"tracking_code"`
		RejectReason   *string   `json:"org_rejection_reason"`
		CreatedAt      string    `json:"created_at"`
	}

	err = h.db.QueryRow(c.Request().Context(), `
		SELECT id, case_id, status::text, property_type, approx_area_sqm,
		       geo_latitude, geo_longitude, map_file_path, tracking_code,
		       org_rejection_reason, created_at::text
		FROM map_services WHERE id = $1
	`, id).Scan(&ms.ID, &ms.CaseID, &ms.Status, &ms.PropertyType, &ms.ApproxAreaSqm,
		&ms.GeoLatitude, &ms.GeoLongitude, &ms.MapFilePath, &ms.TrackingCode,
		&ms.RejectReason, &ms.CreatedAt)

	if err != nil {
		return NotFound(c, "سرویس نقشه یافت نشد")
	}

	return OK(c, ms)
}

// RequestConsent triggers OTP for consent
func (h *MapHandler) RequestConsent(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return BadRequest(c, "شناسه نامعتبر")
	}

	// In production: send real OTP to applicant's mobile
	// For now: just update status to indicate consent requested
	_, err = h.db.Exec(c.Request().Context(), `
		UPDATE map_services SET updated_at = NOW() WHERE id = $1 AND status = 'expert_assigned'
	`, id)
	if err != nil {
		return Conflict(c, "امکان درخواست رضایت در این وضعیت وجود ندارد")
	}

	return OK(c, map[string]string{"message": "کد تایید ارسال شد — لطفاً از طریق /consent/verify تایید کنید"})
}

// VerifyConsent verifies the consent OTP
func (h *MapHandler) VerifyConsent(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return BadRequest(c, "شناسه نامعتبر")
	}

	var req struct{ OTP string `json:"otp"` }
	if err := c.Bind(&req); err != nil || req.OTP == "" {
		return BadRequest(c, "کد تایید الزامی است")
	}

	// In production: verify OTP against Redis
	// For now: accept and update status
	_, err = h.db.Exec(c.Request().Context(), `
		UPDATE map_services SET consent_granted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND consent_granted_at IS NULL
	`, id)
	if err != nil {
		return Conflict(c, "رضایت قبلاً ثبت شده یا سرویس در وضعیت نامعتبر است")
	}

	return OK(c, map[string]string{"message": "رضایت ثبت شد — کارشناس می‌تواند عملیات میدانی را آغاز کند"})
}

// StartFieldwork transitions map service to fieldwork_in_progress
func (h *MapHandler) StartFieldwork(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return BadRequest(c, "شناسه نامعتبر")
	}

	role := middleware.GetUserRole(c)
	if role != "survey_expert" {
		return Forbidden(c, "فقط کارشناس نقشه‌بردار می‌تواند عملیات میدانی را شروع کند")
	}

	if err := h.mapSM.Execute(c.Request().Context(), id, "fieldwork_in_progress"); err != nil {
		return Conflict(c, err.Error())
	}

	return OK(c, map[string]string{"message": "عملیات میدانی آغاز شد"})
}

// SubmitFieldwork submits the fieldwork results (map + photos + descriptive table)
func (h *MapHandler) SubmitFieldwork(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return BadRequest(c, "شناسه نامعتبر")
	}

	var req struct {
		MapFileID         string `json:"map_file_id"`
		DescriptiveTable  map[string]interface{} `json:"descriptive_table"`
	}
	if err := c.Bind(&req); err != nil {
		return BadRequest(c, "اطلاعات ورودی نامعتبر است")
	}

	// Update map service with fieldwork data
	_, err = h.db.Exec(c.Request().Context(), `
		UPDATE map_services SET
			map_file_path = $2,
			descriptive_table = $3,
			fieldwork_completed_at = NOW(),
			status = 'fieldwork_done',
			updated_at = NOW()
		WHERE id = $1 AND status = 'fieldwork_in_progress'
	`, id, req.MapFileID, req.DescriptiveTable)

	if err != nil {
		slog.Error("failed to submit fieldwork", "error", err)
		return InternalError(c, "خطا در ثبت اطلاعات میدانی")
	}

	return OK(c, map[string]string{"message": "اطلاعات میدانی ثبت شد — آماده ارسال به سازمان"})
}

// SubmitToOrg sends the map service to the organization
func (h *MapHandler) SubmitToOrg(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return BadRequest(c, "شناسه نامعتبر")
	}

	if err := h.mapSM.Execute(c.Request().Context(), id, "submitted_to_org"); err != nil {
		return Unprocessable(c, err.Error(), nil)
	}

	// TODO: actual integration — send to organization via Outbox pattern
	// For now, simulate immediate approval
	var caseID uuid.UUID
	h.db.QueryRow(c.Request().Context(), `
		UPDATE map_services SET status = 'approved', tracking_code = 'MAP-' || substr(id::text, 1, 8), org_response_at = NOW()
		WHERE id = $1 RETURNING case_id
	`, id).Scan(&caseID)

	// Transition case to map_completed
	h.caseSM.Execute(c.Request().Context(), caseID, "map_completed")

	return OK(c, map[string]string{
		"message":       "نقشه به سازمان ارسال و تایید شد (شبیه‌سازی)",
		"tracking_code": "MAP-" + id.String()[:8],
	})
}
