package handler

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/msaeedlavasani/SabtBrooker/backend/internal/service"
)

type ClaimHandler struct {
	svc *service.ClaimService
}

func NewClaimHandler(svc *service.ClaimService) *ClaimHandler {
	return &ClaimHandler{svc: svc}
}

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
	g.POST("/:id/ai-advice", h.AIAdvice)
}

func (h *ClaimHandler) Get(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return BadRequest(c, "شناسه نامعتبر")
	}
	cs, err := h.svc.Get(c.Request().Context(), id)
	if err != nil {
		return NotFound(c, "سرویس ادعا یافت نشد")
	}
	return OK(c, cs)
}

func (h *ClaimHandler) RequestConsent(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return BadRequest(c, "شناسه نامعتبر")
	}
	if err := h.svc.RequestConsent(c.Request().Context(), id); err != nil {
		return Conflict(c, err.Error())
	}
	return OK(c, map[string]interface{}{
		"message": "کد تایید و هشدار قانونی ارسال شد",
		"warning": "توجه: درج ادعای واهی مطابق تبصره ۵ ماده ۱۰ قانون دارای تبعات قانونی است",
	})
}

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
	if err := h.svc.VerifyConsent(c.Request().Context(), id, req.Ack); err != nil {
		return Unprocessable(c, err.Error(), nil)
	}
	return OK(c, map[string]string{"message": "رضایت ثبت شد"})
}

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

	input := service.UpdateClaimInput{}
	if req.ClaimType != "" {
		input.ClaimType = &req.ClaimType
	}
	if req.OwnershipType != "" {
		input.OwnershipType = &req.OwnershipType
	}
	if req.MainPlateNumber != "" {
		input.MainPlateNumber = &req.MainPlateNumber
	}
	if req.SubPlateNumber != "" {
		input.SubPlateNumber = &req.SubPlateNumber
	}
	if req.PlateSection != "" {
		input.PlateSection = &req.PlateSection
	}
	if req.TotalShare > 0 {
		input.TotalShare = &req.TotalShare
	}
	if req.PartialShare > 0 {
		input.PartialShare = &req.PartialShare
	}
	input.HasGovernmentRights = &req.HasGovernmentRights
	if req.TreasuryPaymentRef != "" {
		input.TreasuryPaymentRef = &req.TreasuryPaymentRef
	}
	input.LegalAdviceRequested = &req.LegalAdviceRequested
	if req.LegalAdviceMethod != "" {
		input.LegalAdviceMethod = &req.LegalAdviceMethod
	}

	if err := h.svc.UpdateDetails(c.Request().Context(), id, input); err != nil {
		return InternalError(c, err.Error())
	}
	return OK(c, map[string]string{"message": "اطلاعات ادعا به‌روزرسانی شد"})
}

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
	doc, err := h.svc.AddDocument(c.Request().Context(), id, req.DocType, req.FileID, req.Description)
	if err != nil {
		return InternalError(c, err.Error())
	}
	return Created(c, map[string]interface{}{"id": doc.ID, "message": "مستند اضافه شد"})
}

func (h *ClaimHandler) DeleteDocument(c echo.Context) error {
	docID, err := uuid.Parse(c.Param("docId"))
	if err != nil {
		return BadRequest(c, "شناسه مستند نامعتبر")
	}
	if err := h.svc.DeleteDocument(c.Request().Context(), docID); err != nil {
		return NotFound(c, err.Error())
	}
	return NoContent(c)
}

func (h *ClaimHandler) ListDocuments(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return BadRequest(c, "شناسه نامعتبر")
	}
	docs, err := h.svc.ListDocuments(c.Request().Context(), id)
	if err != nil {
		return InternalError(c, err.Error())
	}
	return OK(c, map[string]interface{}{"documents": docs})
}

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
	docID, err := uuid.Parse(req.DocumentID)
	if err != nil {
		return BadRequest(c, "شناسه مستند نامعتبر")
	}
	if err := h.svc.VerifyDocument(c.Request().Context(), docID, id, req.Verified, req.Note); err != nil {
		return InternalError(c, err.Error())
	}
	return OK(c, map[string]string{"message": "نظر کارشناسی ثبت شد"})
}

func (h *ClaimHandler) SubmitToOrg(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return BadRequest(c, "شناسه نامعتبر")
	}
	result, err := h.svc.SubmitToOrg(c.Request().Context(), id)
	if err != nil {
		return Unprocessable(c, err.Error(), nil)
	}
	return OK(c, result)
}

func (h *ClaimHandler) AIAdvice(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return BadRequest(c, "شناسه نامعتبر")
	}
	advice, err := h.svc.GenerateAIAdvice(c.Request().Context(), id)
	if err != nil {
		return InternalError(c, err.Error())
	}
	return OK(c, advice)
}
