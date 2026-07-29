package handler

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/msaeedlavasani/SabtBrooker/backend/internal/middleware"
	"github.com/msaeedlavasani/SabtBrooker/backend/internal/service"
)

// CaseHandler handles case endpoints via service layer
type CaseHandler struct {
	svc *service.CaseService
}

func NewCaseHandler(svc *service.CaseService) *CaseHandler {
	return &CaseHandler{svc: svc}
}

func (h *CaseHandler) RegisterRoutes(g *echo.Group) {
	g.GET("", h.List)
	g.POST("", h.Create)
	g.GET("/:id", h.Get)
	g.PATCH("/:id", h.Update)
	g.PUT("/:id/capacity", h.UpdateCapacity)
	g.POST("/:id/submit", h.SubmitForMap)
}

func (h *CaseHandler) List(c echo.Context) error {
	var p PageParams
	c.Bind(&p)
	p.Defaults()

	cases, err := h.svc.ListCases(
		c.Request().Context(),
		middleware.GetUserID(c),
		middleware.GetUserRole(c),
		p.Offset(),
		p.PageSize,
	)
	if err != nil {
		return BadRequest(c, err.Error())
	}
	return OK(c, cases)
}

func (h *CaseHandler) Create(c echo.Context) error {
	var input struct {
		Province      string  `json:"province"`
		City          string  `json:"city"`
		District      *string `json:"district"`
		Village       *string `json:"village"`
		AddressDetail *string `json:"address_detail"`
	}
	if err := c.Bind(&input); err != nil || input.Province == "" || input.City == "" {
		return BadRequest(c, "استان و شهر الزامی است")
	}

	created, err := h.svc.CreateCase(
		c.Request().Context(),
		middleware.GetUserID(c),
		middleware.GetUserRole(c),
		service.CreateCaseInput{
			Province:      input.Province,
			City:          input.City,
			District:      input.District,
			Village:       input.Village,
			AddressDetail: input.AddressDetail,
		},
	)
	if err != nil {
		return Unprocessable(c, err.Error(), nil)
	}
	return Created(c, created)
}

func (h *CaseHandler) Get(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return BadRequest(c, "شناسه نامعتبر")
	}
	cse, err := h.svc.GetCase(c.Request().Context(), id)
	if err != nil {
		return NotFound(c, "پرونده یافت نشد")
	}
	return OK(c, cse)
}

func (h *CaseHandler) Update(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return BadRequest(c, "شناسه نامعتبر")
	}

	var req struct {
		District      *string `json:"district"`
		Village       *string `json:"village"`
		AddressDetail *string `json:"address_detail"`
	}
	if err := c.Bind(&req); err != nil {
		return BadRequest(c, "اطلاعات ورودی نامعتبر است")
	}

	fields := make(map[string]interface{})
	if req.District != nil {
		fields["district"] = *req.District
	}
	if req.Village != nil {
		fields["village"] = *req.Village
	}
	if req.AddressDetail != nil {
		fields["address_detail"] = *req.AddressDetail
	}

	if len(fields) == 0 {
		return BadRequest(c, "حداقل یک فیلد باید مقداردهی شود")
	}

	if err := h.svc.UpdateCase(c.Request().Context(), id, fields); err != nil {
		return InternalError(c, err.Error())
	}
	return OK(c, map[string]string{"message": "پرونده به‌روزرسانی شد"})
}

func (h *CaseHandler) UpdateCapacity(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return BadRequest(c, "شناسه نامعتبر")
	}
	var req struct{ Capacity string `json:"applicant_capacity"` }
	if err := c.Bind(&req); err != nil {
		return BadRequest(c, "اطلاعات ورودی نامعتبر است")
	}
	if err := h.svc.UpdateCapacity(c.Request().Context(), id, req.Capacity); err != nil {
		return Unprocessable(c, err.Error(), nil)
	}
	return OK(c, map[string]string{"message": "سمت متقاضی به‌روزرسانی شد"})
}

func (h *CaseHandler) SubmitForMap(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return BadRequest(c, "شناسه نامعتبر")
	}
	ms, err := h.svc.SubmitForMap(c.Request().Context(), id)
	if err != nil {
		return Unprocessable(c, err.Error(), nil)
	}
	return OK(c, map[string]interface{}{
		"message":        "فرآیند تهیه نقشه آغاز شد",
		"map_service_id": ms.ID,
		"status":         ms.Status,
	})
}
