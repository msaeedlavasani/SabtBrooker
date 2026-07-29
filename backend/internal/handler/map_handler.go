package handler

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/msaeedlavasani/SabtBrooker/backend/internal/middleware"
	"github.com/msaeedlavasani/SabtBrooker/backend/internal/service"
)

type MapHandler struct {
	svc *service.MapService
}

func NewMapHandler(svc *service.MapService) *MapHandler {
	return &MapHandler{svc: svc}
}

func (h *MapHandler) RegisterRoutes(g *echo.Group) {
	g.GET("/:id", h.Get)
	g.POST("/:id/consent", h.RequestConsent)
	g.POST("/:id/consent/verify", h.VerifyConsent)
	g.POST("/:id/accept", h.Accept)
	g.POST("/:id/fieldwork/start", h.StartFieldwork)
	g.POST("/:id/fieldwork/submit", h.SubmitFieldwork)
	g.POST("/:id/submit", h.SubmitToOrg)
}

func (h *MapHandler) Get(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return BadRequest(c, "شناسه نامعتبر")
	}
	ms, err := h.svc.Get(c.Request().Context(), id)
	if err != nil {
		return NotFound(c, "سرویس نقشه یافت نشد")
	}
	return OK(c, ms)
}

func (h *MapHandler) RequestConsent(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return BadRequest(c, "شناسه نامعتبر")
	}
	if err := h.svc.RequestConsent(c.Request().Context(), id); err != nil {
		return Conflict(c, err.Error())
	}
	return OK(c, map[string]string{"message": "کد تایید ارسال شد — لطفاً از طریق /consent/verify تایید کنید"})
}

func (h *MapHandler) VerifyConsent(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return BadRequest(c, "شناسه نامعتبر")
	}
	if err := h.svc.VerifyConsent(c.Request().Context(), id); err != nil {
		return Conflict(c, err.Error())
	}
	return OK(c, map[string]string{"message": "رضایت ثبت شد — کارشناس می‌تواند عملیات میدانی را آغاز کند"})
}

func (h *MapHandler) Accept(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return BadRequest(c, "شناسه نامعتبر")
	}
	
	// Get expert ID from context (JWT)
	expertID := middleware.GetUserID(c)
	
	if err := h.svc.AssignExpert(c.Request().Context(), id, expertID); err != nil {
		return Unprocessable(c, err.Error(), nil)
	}
	
	return OK(c, map[string]string{"message": "پرونده توسط شما پذیرفته شد"})
}

func (h *MapHandler) StartFieldwork(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return BadRequest(c, "شناسه نامعتبر")
	}
	if err := h.svc.StartFieldwork(c.Request().Context(), id, middleware.GetUserRole(c)); err != nil {
		return Unprocessable(c, err.Error(), nil)
	}
	return OK(c, map[string]string{"message": "عملیات میدانی آغاز شد"})
}

func (h *MapHandler) SubmitFieldwork(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return BadRequest(c, "شناسه نامعتبر")
	}

	var req struct {
		MapFileID        string                 `json:"map_file_id"`
		DescriptiveTable map[string]interface{} `json:"descriptive_table"`
	}
	if err := c.Bind(&req); err != nil {
		return BadRequest(c, "اطلاعات ورودی نامعتبر است")
	}

	if err := h.svc.SubmitFieldwork(c.Request().Context(), id, req.MapFileID, req.DescriptiveTable); err != nil {
		return InternalError(c, err.Error())
	}
	return OK(c, map[string]string{"message": "اطلاعات میدانی ثبت شد — آماده ارسال به سازمان"})
}

func (h *MapHandler) SubmitToOrg(c echo.Context) error {
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
