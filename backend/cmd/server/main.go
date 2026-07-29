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
	"github.com/msaeedlavasani/SabtBrooker/backend/internal/handler"
	"github.com/msaeedlavasani/SabtBrooker/backend/internal/middleware"
	"github.com/msaeedlavasani/SabtBrooker/backend/internal/workflow"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))
	slog.Info("starting SabtBrooker backend...")

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
	rdb := redis.NewClient(&redis.Options{Addr: cfg.Redis.Addr, Password: cfg.Redis.Password, DB: cfg.Redis.DB})
	if err := rdb.Ping(ctx).Err(); err != nil {
		slog.Error("failed to connect to Redis", "error", err)
		os.Exit(1)
	}
	defer rdb.Close()

	// NATS (non-fatal if unavailable — backend works without it)
	nc, _ := nats.Connect(cfg.NATS.URL,
		nats.MaxReconnects(3),
		nats.ReconnectWait(2*time.Second),
	)
	if nc != nil {
		defer nc.Close()
		slog.Info("connected to NATS", "url", cfg.NATS.URL)
	}

	// JWT Manager
	jwtManager, err := auth.NewJWTManager(cfg.JWT)
	if err != nil {
		slog.Error("failed to initialize JWT manager", "error", err)
		os.Exit(1)
	}

	// OTP Service
	otpService := auth.NewOTPService(rdb, cfg.OTP)

	// Workflow Engines
	caseSM := workflow.BuildCaseStateMachine(db.Pool, nc)
	mapSM := workflow.NewStateMachine("map_service", db.Pool, nc)
	claimSM := workflow.NewStateMachine("claim_service", db.Pool, nc)
	certSM := workflow.NewStateMachine("cert_service", db.Pool, nc)

	// Map service transitions
	mapSM.AddTransition(workflow.Transition{From: "pending_expert_assignment", To: "expert_assigned"})
	mapSM.AddTransition(workflow.Transition{From: "expert_assigned", To: "fieldwork_in_progress"})
	mapSM.AddTransition(workflow.Transition{From: "fieldwork_in_progress", To: "fieldwork_done"})
	mapSM.AddTransition(workflow.Transition{From: "fieldwork_done", To: "submitted_to_org"})
	mapSM.AddTransition(workflow.Transition{From: "submitted_to_org", To: "approved"})
	mapSM.AddTransition(workflow.Transition{From: "submitted_to_org", To: "rejected"})

	// Claim service transitions
	claimSM.AddTransition(workflow.Transition{From: "pending_expert_assignment", To: "expert_assigned"})
	claimSM.AddTransition(workflow.Transition{From: "expert_assigned", To: "documents_verified"})
	claimSM.AddTransition(workflow.Transition{From: "documents_verified", To: "submitted_to_org"})
	claimSM.AddTransition(workflow.Transition{From: "submitted_to_org", To: "approved"})
	claimSM.AddTransition(workflow.Transition{From: "submitted_to_org", To: "rejected"})

	// Cert service transitions
	certSM.AddTransition(workflow.Transition{From: "pending_data", To: "submitted_to_org"})
	certSM.AddTransition(workflow.Transition{From: "submitted_to_org", To: "approved"})
	certSM.AddTransition(workflow.Transition{From: "submitted_to_org", To: "rejected"})

	// Handlers
	caseHandler := handler.NewCaseHandler(db.Pool, caseSM)
	mapHandler := handler.NewMapHandler(db.Pool, caseSM, mapSM)
	claimHandler := handler.NewClaimHandler(db.Pool, caseSM, claimSM)
	certHandler := handler.NewCertHandler(db.Pool, caseSM, certSM)

	// Echo server
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Use(middleware.Recover())
	e.Use(middleware.RequestLogger())
	e.Use(middleware.CORS())

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

	// Public auth routes
	v1 := e.Group("/v1")
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
		code, expiresAt, err := otpService.GenerateAndSend(c.Request().Context(), req.Mobile, req.Purpose)
		if err != nil {
			return c.JSON(http.StatusTooManyRequests, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusOK, map[string]interface{}{
			"message":    "کد تایید ارسال شد",
			"expires_in": int(time.Until(expiresAt).Seconds()),
			"dev_otp":    code, // فقط در محیط توسعه — در production حذف شود
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

	// Business logic routes
	caseHandler.RegisterRoutes(protected.Group("/cases"))
	mapHandler.RegisterRoutes(protected.Group("/map-services"))
	claimHandler.RegisterRoutes(protected.Group("/claim-services"))
	certHandler.RegisterRoutes(protected.Group("/cert-services"))

	// Start server
	go func() {
		addr := cfg.Server.Host + ":" + cfg.Server.Port
		slog.Info("server started", "addr", addr)
		if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("shutting down...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()
	if err := e.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown error", "error", err)
	}
	slog.Info("server stopped")
}
