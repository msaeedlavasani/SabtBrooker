package handler

import (
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"

	"github.com/msaeedlavasani/SabtBrooker/backend/internal/middleware"
	"github.com/msaeedlavasani/SabtBrooker/backend/internal/workflow"
)

// CaseHandler handles case-related endpoints
type CaseHandler struct {
	db    *pgxpool.Pool
	caseSM *workflow.StateMachine
}

// NewCaseHandler creates a new case handler
func NewCaseHandler(db *pgxpool.Pool, sm *workflow.StateMachine) *CaseHandler {
	return &CaseHandler{db: db, caseSM: sm}
}

// RegisterRoutes registers all case routes
func (h *CaseHandler) RegisterRoutes(g *echo.Group) {
	g.GET("", h.List)
	g.POST("", h.Create)
	g.GET("/:id", h.Get)
	g.PATCH("/:id", h.Update)
	g.PUT("/:id/capacity", h.UpdateCapacity)
	g.POST("/:id/submit", h.SubmitForMap)
}

// List returns all cases for the authenticated user
func (h *CaseHandler) List(c echo.Context) error {
	userID := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)

	var p PageParams
	c.Bind(&p)
	p.Defaults()

	var rows pgx.Rows
	var err error

	switch role {
	case "applicant":
		rows, err = h.db.Query(c.Request().Context(), `
			SELECT id, status, province, city, address_detail,
			       map_tracking_code, claim_tracking_code, cert_tracking_code,
			       deadline_2years, created_at, updated_at
			FROM cases WHERE applicant_id = $1
			ORDER BY created_at DESC LIMIT $2 OFFSET $3
		`, userID, p.PageSize, p.Offset())
	case "admin", "auditor":
		rows, err = h.db.Query(c.Request().Context(), `
			SELECT id, status, province, city, address_detail,
			       map_tracking_code, claim_tracking_code, cert_tracking_code,
			       deadline_2years, created_at, updated_at
			FROM cases ORDER BY created_at DESC LIMIT $1 OFFSET $2
		`, p.PageSize, p.Offset())
	default:
		// expert — only assigned cases
		rows, err = h.db.Query(c.Request().Context(), `
			SELECT id, status, province, city, address_detail,
			       map_tracking_code, claim_tracking_code, cert_tracking_code,
			       deadline_2years, created_at, updated_at
			FROM cases
			WHERE legal_expert_id = $1 OR survey_expert_id = $1
			ORDER BY created_at DESC LIMIT $2 OFFSET $3
		`, userID, p.PageSize, p.Offset())
	}

	if err != nil {
		slog.Error("failed to list cases", "error", err)
		return InternalError(c, "خطا در دریافت لیست پرونده‌ها")
	}
	defer rows.Close()

	type item struct {
		ID                uuid.UUID `json:"id"`
		Status            string    `json:"status"`
		Province          string    `json:"province"`
		City              string    `json:"city"`
		AddressDetail     *string   `json:"address_detail"`
		MapTrackingCode   *string   `json:"map_tracking_code"`
		ClaimTrackingCode *string   `json:"claim_tracking_code"`
		CertTrackingCode  *string   `json:"cert_tracking_code"`
		Deadline2Years    *string   `json:"deadline_2years"`
		CreatedAt         string    `json:"created_at"`
		UpdatedAt         string    `json:"updated_at"`
	}

	var items []item
	for rows.Next() {
		var i item
		rows.Scan(&i.ID, &i.Status, &i.Province, &i.City, &i.AddressDetail,
			&i.MapTrackingCode, &i.ClaimTrackingCode, &i.CertTrackingCode,
			&i.Deadline2Years, &i.CreatedAt, &i.UpdatedAt)
		items = append(items, i)
	}

	return OK(c, map[string]interface{}{
		"items":     items,
		"total":     len(items),
		"page":      p.Page,
		"page_size": p.PageSize,
	})
}

// Create creates a new case
func (h *CaseHandler) Create(c echo.Context) error {
	userID := middleware.GetUserID(c)

	var req struct {
		Province      string `json:"province"`
		City          string `json:"city"`
		District      string `json:"district"`
		Village       string `json:"village"`
		PostalCode    string `json:"postal_code"`
		AddressDetail string `json:"address_detail"`
	}
	if err := c.Bind(&req); err != nil {
		return BadRequest(c, "اطلاعات ورودی نامعتبر است")
	}
	if req.Province == "" || req.City == "" {
		return BadRequest(c, "استان و شهر الزامی است")
	}

	var id uuid.UUID
	err := h.db.QueryRow(c.Request().Context(), `
		INSERT INTO cases (applicant_id, province, city, district, village, postal_code, address_detail)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`, userID, req.Province, req.City, req.District, req.Village, req.PostalCode, req.AddressDetail).Scan(&id)

	if err != nil {
		slog.Error("failed to create case", "error", err)
		return InternalError(c, "خطا در ایجاد پرونده")
	}

	return Created(c, map[string]interface{}{
		"id":      id,
		"message": "پرونده با موفقیت ایجاد شد",
	})
}

// Get returns case details with all linked services
func (h *CaseHandler) Get(c echo.Context) error {
	caseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return BadRequest(c, "شناسه پرونده نامعتبر است")
	}

	var cs struct {
		ID                uuid.UUID
		Status            string
		Province          string
		City              string
		District          *string
		Village           *string
		PostalCode        *string
		AddressDetail     *string
		MapTrackingCode   *string
		ClaimTrackingCode *string
		CertTrackingCode  *string
		Deadline2Years    *string
		Deadline5Months   *string
		CreatedAt         string
		UpdatedAt         string
		CompletedAt       *string
	}

	err = h.db.QueryRow(c.Request().Context(), `
		SELECT id, status::text, province, city, district, village, postal_code, address_detail,
		       map_tracking_code, claim_tracking_code, cert_tracking_code,
		       deadline_2years::text, deadline_5months::text,
		       created_at::text, updated_at::text, completed_at::text
		FROM cases WHERE id = $1
	`, caseID).Scan(&cs.ID, &cs.Status, &cs.Province, &cs.City,
		&cs.District, &cs.Village, &cs.PostalCode, &cs.AddressDetail,
		&cs.MapTrackingCode, &cs.ClaimTrackingCode, &cs.CertTrackingCode,
		&cs.Deadline2Years, &cs.Deadline5Months,
		&cs.CreatedAt, &cs.UpdatedAt, &cs.CompletedAt)

	if err != nil {
		return NotFound(c, "پرونده یافت نشد")
	}

	return OK(c, cs)
}

// Update updates case address info (only in draft status)
func (h *CaseHandler) Update(c echo.Context) error {
	caseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return BadRequest(c, "شناسه پرونده نامعتبر است")
	}

	var req struct {
		Province      string `json:"province"`
		City          string `json:"city"`
		District      string `json:"district"`
		Village       string `json:"village"`
		PostalCode    string `json:"postal_code"`
		AddressDetail string `json:"address_detail"`
	}
	if err := c.Bind(&req); err != nil {
		return BadRequest(c, "اطلاعات ورودی نامعتبر است")
	}

	tag, err := h.db.Exec(c.Request().Context(), `
		UPDATE cases SET province=$2, city=$3, district=$4, village=$5,
		       postal_code=$6, address_detail=$7, updated_at=NOW()
		WHERE id=$1 AND status='draft'
	`, caseID, req.Province, req.City, req.District, req.Village, req.PostalCode, req.AddressDetail)

	if err != nil || tag.RowsAffected() == 0 {
		return Conflict(c, "پرونده قابل ویرایش نیست (فقط در وضعیت پیش‌نویس)")
	}

	return OK(c, map[string]string{"message": "به‌روزرسانی شد"})
}

// UpdateCapacity updates applicant capacity and representation info
func (h *CaseHandler) UpdateCapacity(c echo.Context) error {
	caseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return BadRequest(c, "شناسه پرونده نامعتبر است")
	}

	var req struct {
		ApplicantCapacity      string `json:"applicant_capacity"`
		LegalEntityID          string `json:"legal_entity_id"`
		LegalEntityName        string `json:"legal_entity_name"`
		ProxyDocumentType      string `json:"proxy_document_type"`
		ProxyDocumentID        string `json:"proxy_document_id"`
		ProxyDocumentDate      string `json:"proxy_document_date"`
		ProxyVerificationCode  string `json:"proxy_verification_code"`
	}
	if err := c.Bind(&req); err != nil {
		return BadRequest(c, "اطلاعات ورودی نامعتبر است")
	}

	_, err = h.db.Exec(c.Request().Context(), `
		UPDATE cases SET
			applicant_capacity=$2, legal_entity_id=$3, legal_entity_name=$4,
			proxy_document_type=$5, proxy_document_id=$6, proxy_document_date=$7::date,
			proxy_verification_code=$8, updated_at=NOW()
		WHERE id=$1
	`, caseID, req.ApplicantCapacity, req.LegalEntityID, req.LegalEntityName,
		req.ProxyDocumentType, req.ProxyDocumentID, req.ProxyDocumentDate,
		req.ProxyVerificationCode)

	if err != nil {
		return InternalError(c, "خطا در به‌روزرسانی اطلاعات نمایندگی")
	}

	return OK(c, map[string]string{"message": "اطلاعات نمایندگی ثبت شد"})
}

// SubmitForMap transitions the case from draft to map_in_progress
func (h *CaseHandler) SubmitForMap(c echo.Context) error {
	caseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return BadRequest(c, "شناسه پرونده نامعتبر است")
	}

	if err := h.caseSM.Execute(c.Request().Context(), caseID, "map_in_progress"); err != nil {
		return Unprocessable(c, err.Error(), nil)
	}

	return OK(c, map[string]string{
		"message": "پرونده وارد مرحله تهیه نقشه ثبتی شد",
		"status":  "map_in_progress",
	})
}
