package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"

	"github.com/msaeedlavasani/SabtBrooker/backend/internal/auth"
	"github.com/msaeedlavasani/SabtBrooker/backend/internal/config"
	"github.com/msaeedlavasani/SabtBrooker/backend/internal/database"
	"github.com/msaeedlavasani/SabtBrooker/backend/internal/handler"
	"github.com/msaeedlavasani/SabtBrooker/backend/internal/middleware"
	"github.com/msaeedlavasani/SabtBrooker/backend/internal/notification"
	"github.com/msaeedlavasani/SabtBrooker/backend/internal/outbox"
	"github.com/msaeedlavasani/SabtBrooker/backend/internal/repository"
	"github.com/msaeedlavasani/SabtBrooker/backend/internal/scheduler"
	"github.com/msaeedlavasani/SabtBrooker/backend/internal/service"
	"github.com/msaeedlavasani/SabtBrooker/backend/internal/storage"
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

	// ── Infrastructure ──────────────────────────────────────────

	db, err := database.NewPostgres(ctx, cfg.DB)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	store, err := storage.NewFileStorage(ctx, cfg.MinIO)
	if err != nil {
		slog.Error("failed to initialize storage", "error", err)
		os.Exit(1)
	}

	rdb := redis.NewClient(&redis.Options{Addr: cfg.Redis.Addr, Password: cfg.Redis.Password, DB: cfg.Redis.DB})
	if err := rdb.Ping(ctx).Err(); err != nil {
		slog.Error("failed to connect to Redis", "error", err)
		os.Exit(1)
	}
	defer rdb.Close()

	nc, _ := nats.Connect(cfg.NATS.URL,
		nats.MaxReconnects(3),
		nats.ReconnectWait(2*time.Second),
	)
	if nc != nil {
		defer nc.Close()
		slog.Info("connected to NATS", "url", cfg.NATS.URL)
	}

	// ── Repositories ────────────────────────────────────────────

	userRepo := repository.NewPostgresUserRepo(db.Pool)
	caseRepo := repository.NewPostgresCaseRepo(db.Pool)
	mapRepo := repository.NewPostgresMapServiceRepo(db.Pool)
	claimRepo := repository.NewPostgresClaimServiceRepo(db.Pool)
	certRepo := repository.NewPostgresCertServiceRepo(db.Pool)
	auditRepo := repository.NewPostgresAuditLogRepo(db.Pool)
	aiAdviceRepo := repository.NewPostgresAIAdviceRepo(db.Pool)

	// ── Workflow Engines ────────────────────────────────────────

	caseSM := workflow.BuildCaseStateMachine(db.Pool, nc)
	mapSM := workflow.NewStateMachine("map_service", db.Pool, nc)
	claimSM := workflow.NewStateMachine("claim_service", db.Pool, nc)
	certSM := workflow.NewStateMachine("cert_service", db.Pool, nc)

	// Register transitions
	mapSM.AddTransition(workflow.Transition{From: "pending_expert_assignment", To: "expert_assigned"})
	mapSM.AddTransition(workflow.Transition{From: "expert_assigned", To: "fieldwork_in_progress"})
	mapSM.AddTransition(workflow.Transition{From: "fieldwork_in_progress", To: "fieldwork_done"})
	mapSM.AddTransition(workflow.Transition{From: "fieldwork_done", To: "submitted_to_org"})
	mapSM.AddTransition(workflow.Transition{From: "submitted_to_org", To: "approved"})
	mapSM.AddTransition(workflow.Transition{From: "submitted_to_org", To: "rejected"})

	claimSM.AddTransition(workflow.Transition{From: "pending_expert_assignment", To: "expert_assigned"})
	claimSM.AddTransition(workflow.Transition{From: "expert_assigned", To: "documents_verified"})
	claimSM.AddTransition(workflow.Transition{From: "documents_verified", To: "submitted_to_org"})
	claimSM.AddTransition(workflow.Transition{From: "submitted_to_org", To: "approved"})
	claimSM.AddTransition(workflow.Transition{From: "submitted_to_org", To: "rejected"})

	certSM.AddTransition(workflow.Transition{From: "pending_data", To: "submitted_to_org"})
	certSM.AddTransition(workflow.Transition{From: "submitted_to_org", To: "approved"})
	certSM.AddTransition(workflow.Transition{From: "submitted_to_org", To: "rejected"})

	// ── Services ────────────────────────────────────────────────

	notifySvc := notification.NewService(db.Pool, rdb, nil)
	caseSvc := service.NewCaseService(caseRepo, userRepo, mapRepo, claimRepo, certRepo, auditRepo, caseSM)
	mapSvc := service.NewMapService(mapRepo, caseSM, mapSM, auditRepo)
	claimSvc := service.NewClaimService(claimRepo, caseSM, claimSM, auditRepo, aiAdviceRepo)
	certSvc := service.NewCertService(certRepo, caseSM, certSM, auditRepo)

	// ── Auth ────────────────────────────────────────────────────

	otpService := auth.NewOTPService(rdb, cfg.OTP, notifySvc)
	jwtManager, err := auth.NewJWTManager(cfg.JWT)
	if err != nil {
		slog.Error("failed to initialize JWT manager", "error", err)
		os.Exit(1)
	}

	// ── Handlers ────────────────────────────────────────────────

	authHandler := handler.NewAuthHandler(jwtManager, otpService, userRepo)
	caseHandler := handler.NewCaseHandler(caseSvc)
	mapHandler := handler.NewMapHandler(mapSvc)
	claimHandler := handler.NewClaimHandler(claimSvc)
	certHandler := handler.NewCertHandler(certSvc)
	storageHandler := handler.NewStorageHandler(store)

	// ── Background Services ─────────────────────────────────────

	// Outbox publisher (submits queued messages to org)
	outboxPublisher := outbox.NewPublisher(db.Pool, nc, 1*time.Second)
	go outboxPublisher.Start(ctx)

	// Scheduler (deadline tracking)
	sched := scheduler.New(db.Pool, nc, 1*time.Hour)
	schedCtx := scheduler.WithDB(ctx, db.Pool)
	go sched.Start(schedCtx)

	// ── HTTP Server ─────────────────────────────────────────────

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

	// Public routes
	v1 := e.Group("/v1")
	authHandler.RegisterRoutes(v1.Group("/auth"))

	// Protected routes
	v1.Group("")
	protected := v1.Group("")
	protected.Use(middleware.AuthRequired(jwtManager))

	caseHandler.RegisterRoutes(protected.Group("/cases"))
	mapHandler.RegisterRoutes(protected.Group("/map-services"))
	claimHandler.RegisterRoutes(protected.Group("/claim-services"))
	certHandler.RegisterRoutes(protected.Group("/cert-services"))
	storageHandler.RegisterRoutes(protected.Group("/storage"))

	// ── Graceful Shutdown ───────────────────────────────────────

	go func() {
		addr := cfg.Server.Host + ":" + cfg.Server.Port
		slog.Info("server started", "addr", addr)
		if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

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
