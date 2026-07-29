package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"

	"github.com/msaeedlavasani/SabtBrooker/backend/internal/auth"
	"github.com/msaeedlavasani/SabtBrooker/backend/internal/config"
	"github.com/msaeedlavasani/SabtBrooker/backend/internal/database"
	"github.com/msaeedlavasani/SabtBrooker/backend/internal/middleware"
	"github.com/msaeedlavasani/SabtBrooker/backend/internal/workflow"
)

func main() {
	// Initialize structured logger
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	slog.Info("starting SabtBrooker backend...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	ctx := context.Background()

	// Database
	db, err := database.NewPostgres(ctx, cfg.DB)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if err := rdb.Ping(ctx).Err(); err != nil {
		slog.Error("failed to connect to Redis", "error", err)
		os.Exit(1)
	}
	defer rdb.Close()
	slog.Info("connected to Redis", "addr", cfg.Redis.Addr)

	// NATS
	nc, err := nats.Connect(cfg.NATS.URL,
		nats.MaxReconnects(cfg.NATS.MaxReconn),
		nats.ReconnectWait(2*time.Second),
		nats.DisconnectErrHandler(func(_ *nats.Conn, err error) {
			slog.Warn("NATS disconnected", "error", err)
		}),
		nats.ReconnectHandler(func(_ *nats.Conn) {
			slog.Info("NATS reconnected")
		}),
	)
	if err != nil {
		slog.Error("failed to connect to NATS", "error", err)
		os.Exit(1)
	}
	defer nc.Close()
	slog.Info("connected to NATS", "url", cfg.NATS.URL)

	// JWT Manager
	jwtManager, err := auth.NewJWTManager(cfg.JWT)
	if err != nil {
		slog.Error("failed to initialize JWT manager", "error", err)
		os.Exit(1)
	}

	// OTP Service
	otpService := auth.NewOTPService(rdb, cfg.OTP)

	// Workflow Engine
	caseSM := workflow.BuildCaseStateMachine(db.Pool, nc)

	// Create Echo server
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	// Global middleware
	e.Use(middleware.Recover())
	e.Use(middleware.RequestLogger())
	e.Use(middleware.CORS())

	// ---- Routes ----

	// Health
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "healthy"})
	})
	e.GET("/health/ready", func(c echo.Context) error {
		if err := db.HealthCheck(ctx); err != nil {
			return c.JSON(http.StatusServiceUnavailable, map[string]string{"status": "not ready"})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "ready"})
	})

	// API v1 group
	v1 := e.Group("/v1")

	// Auth routes (public)
	authGroup := v1.Group("/auth")
	authGroup.POST("/otp/send", func(c echo.Context) error {
		var req struct {
			Mobile  string `json:"mobile"`
			Purpose string `json:"purpose"`
		}
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "اطلاعات ورودی نامعتبر است"})
		}
		if req.Purpose == "" {
			req.Purpose = "auth"
		}

		otp, expiresAt, err := otpService.GenerateAndSend(c.Request().Context(), req.Mobile, req.Purpose)
		if err != nil {
			return c.JSON(http.StatusTooManyRequests, map[string]string{"error": err.Error()})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"message":    "کد تایید ارسال شد",
			"expires_in": int(time.Until(expiresAt).Seconds()),
		})
	})

	authGroup.POST("/otp/verify", func(c echo.Context) error {
		var req struct {
			Mobile  string `json:"mobile"`
			OTP     string `json:"otp"`
			Purpose string `json:"purpose"`
		}
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "اطلاعات ورودی نامعتبر است"})
		}
		if req.Purpose == "" {
			req.Purpose = "auth"
		}

		if err := otpService.Verify(c.Request().Context(), req.Mobile, req.OTP, req.Purpose); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}

		// Generate JWT tokens (simplified — in production, look up or create user)
		tokens, err := jwtManager.GenerateTokenPair(uuid.Nil, "applicant", req.Mobile)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "خطا در تولید توکن"})
		}

		return c.JSON(http.StatusOK, tokens)
	})

	// Protected routes
	protected := v1.Group("")
	protected.Use(middleware.AuthRequired(jwtManager))

	protected.GET("/auth/me", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"user_id": middleware.GetUserID(c),
			"role":    middleware.GetUserRole(c),
		})
	})

	// Cases (protected)
	cases := protected.Group("/cases")
	cases.GET("", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"message": "cases list — to be implemented"})
	})

	// Workflow status
	protected.GET("/workflow/case/:id/status", func(c echo.Context) error {
		caseID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "شناسه نامعتبر"})
		}

		_ = caseSM
		_ = caseID

		return c.JSON(http.StatusOK, map[string]string{"message": "ok"})
	})

	// Available transitions
	protected.GET("/workflow/case/:id/transitions", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"message": "transitions — to be implemented"})
	})

	// ---- Start server ----
	go func() {
		addr := cfg.Server.Host + ":" + cfg.Server.Port
		slog.Info("server started", "addr", addr)
		if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	// ---- Graceful shutdown ----
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := e.Shutdown(shutdownCtx); err != nil {
		slog.Error("server shutdown error", "error", err)
	}

	slog.Info("server stopped gracefully")
}
