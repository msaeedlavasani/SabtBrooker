package handler

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/msaeedlavasani/SabtBrooker/backend/internal/service"
)

type CertHandler struct {
	svc *service.CertService
}

func NewCertHandler(svc *service.CertService) *CertHandler {
	return &CertHandler{svc: svc}
}

func (h *CertHandler) RegisterRoutes(g *echo.Group) {
	g.GET("/:id", h.Get)
	g.POST("/:id/consent", h.RequestConsent)
	g.POST("/:id/consent/verify", h.VerifyConsent)
	g.PATCH("/:id", h.UpdateDetails)
	g.POST("/:id/submit", h.SubmitToOrg)
}

func (h *CertHandler) Get(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return BadRequest(c, "شناسه نامعتبر")
	}
	cs, err := h.svc.Get(c.Request().Context(), id)
	if err != nil {
		return NotFound(c, "سرویس گواهی اقدام یافت نشد")
	}
	return OK(c, cs)
}

func (h *CertHandler) RequestConsent(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return BadRequest(c, "شناسه نامعتبر")
	}
	if err := h.svc.RequestConsent(c.Request().Context(), id); err != nil {
		return Conflict(c, err.Error())
	}
	return OK(c, map[string]string{"message": "کد تایید ارسال شد"})
}

func (h *CertHandler) VerifyConsent(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return BadRequest(c, "شناسه نامعتبر")
	}
	if err := h.svc.VerifyConsent(c.Request().Context(), id); err != nil {
		return Conflict(c, err.Error())
	}
	return OK(c, map[string]string{"message": "رضایت ثبت شد"})
}

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

	input := service.UpdateCertInput{}
	if req.ActionReference != "" {
		input.ActionReference = &req.ActionReference
	}
	if req.ActionType != "" {
		input.ActionType = &req.ActionType
	}
	if req.ActionDate != "" {
		input.ActionDate = &req.ActionDate
	}
	if req.CertImageID != "" {
		input.CertImageID = &req.CertImageID
	}
	if req.CertUniqueID != "" {
		input.CertUniqueID = &req.CertUniqueID
	}

	if err := h.svc.UpdateDetails(c.Request().Context(), id, input); err != nil {
		return Unprocessable(c, err.Error(), nil)
	}
	return OK(c, map[string]string{"message": "اطلاعات گواهی به‌روزرسانی شد"})
}

func (h *CertHandler) SubmitToOrg(c echo.Context) error {
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
