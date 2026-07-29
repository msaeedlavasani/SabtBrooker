package handler

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/msaeedlavasani/SabtBrooker/backend/internal/middleware"
	"github.com/msaeedlavasani/SabtBrooker/backend/internal/service"
)

type PaymentHandler struct {
	svc *service.PaymentService
}

func NewPaymentHandler(svc *service.PaymentService) *PaymentHandler {
	return &PaymentHandler{svc: svc}
}

func (h *PaymentHandler) RegisterRoutes(g *echo.Group) {
	g.POST("/initiate", h.Initiate)
	g.GET("/callback", h.Callback)
}

func (h *PaymentHandler) Initiate(c echo.Context) error {
	var req struct {
		CaseID      string `json:"case_id"`
		ServiceType string `json:"service_type"`
		CallbackURL string `json:"callback_url"`
	}
	if err := c.Bind(&req); err != nil {
		return BadRequest(c, "invalid request")
	}

	caseID, err := uuid.Parse(req.CaseID)
	if err != nil {
		return BadRequest(c, "invalid case_id")
	}

	url, err := h.svc.InitiatePayment(
		c.Request().Context(),
		caseID,
		middleware.GetUserID(c),
		middleware.GetUserRole(c),
		req.ServiceType,
		req.CallbackURL,
	)
	if err != nil {
		return InternalError(c, err.Error())
	}

	return OK(c, map[string]string{"payment_url": url})
}

func (h *PaymentHandler) Callback(c echo.Context) error {
	authority := c.QueryParam("Authority")
	status := c.QueryParam("Status")

	if status != "OK" {
		return c.Redirect(http.StatusFound, "/payment/failed")
	}

	err := h.svc.VerifyPayment(c.Request().Context(), authority)
	if err != nil {
		return c.Redirect(http.StatusFound, fmt.Sprintf("/payment/error?msg=%s", err.Error()))
	}

	return c.Redirect(http.StatusFound, "/payment/success")
}
