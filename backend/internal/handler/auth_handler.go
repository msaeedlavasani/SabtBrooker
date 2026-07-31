package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/msaeedlavasani/SabtBrooker/backend/internal/auth"
	"github.com/msaeedlavasani/SabtBrooker/backend/internal/middleware"
	"github.com/msaeedlavasani/SabtBrooker/backend/internal/repository"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	jwtManager *auth.JWTManager
	otpService *auth.OTPService
	userRepo   repository.UserRepository
}

func NewAuthHandler(jwtManager *auth.JWTManager, otpService *auth.OTPService, userRepo repository.UserRepository) *AuthHandler {
	return &AuthHandler{
		jwtManager: jwtManager,
		otpService: otpService,
		userRepo:   userRepo,
	}
}

func (h *AuthHandler) RegisterRoutes(g *echo.Group) {
	g.POST("/otp/send", h.SendOTP)
	g.POST("/otp/verify", h.VerifyOTP)
	g.GET("/me", h.Me, middleware.AuthRequired(h.jwtManager))
}

// SendOTP generates and sends an OTP code
func (h *AuthHandler) SendOTP(c echo.Context) error {
	var req struct {
		Mobile  string `json:"mobile"`
		Purpose string `json:"purpose"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "اطلاعات ورودی نامعتبر است"})
	}

	// BYPASSED FOR DEMO - Always return success
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":    "کد تایید ارسال شد (حالت دمو)",
		"expires_in": 120,
		"dev_otp":    "1234",
	})
}

// VerifyOTP validates OTP and returns JWT tokens
func (h *AuthHandler) VerifyOTP(c echo.Context) error {
	var req struct {
		Mobile  string `json:"mobile"`
		OTP     string `json:"otp"`
		Purpose string `json:"purpose"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "اطلاعات ورودی نامعتبر است"})
	}
	
	// DEMO MODE: OTP verification is bypassed.
	// In production, uncomment the following block:
	/*
	if err := h.otpService.Verify(c.Request().Context(), req.Mobile, req.OTP, req.Purpose); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	*/

	// Find or create user
	user, err := h.userRepo.FindOrCreateByMobile(c.Request().Context(), req.Mobile)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "خطا در ایجاد کاربر"})
	}

	// Generate JWT tokens
	tokens, err := h.jwtManager.GenerateTokenPair(user.ID, user.Role, user.Mobile)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "خطا در تولید توکن"})
	}

	return c.JSON(http.StatusOK, tokens)
}

// Me returns the current authenticated user info
func (h *AuthHandler) Me(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"user_id": middleware.GetUserID(c),
		"role":    middleware.GetUserRole(c),
	})
}
