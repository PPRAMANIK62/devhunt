package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/PPRAMANIK62/devhunt/docs"
	"github.com/PPRAMANIK62/devhunt/internal/cache"
	"github.com/PPRAMANIK62/devhunt/internal/config"
	"github.com/PPRAMANIK62/devhunt/internal/database"
	"github.com/PPRAMANIK62/devhunt/internal/handler"
	"github.com/PPRAMANIK62/devhunt/internal/logger"
	"github.com/PPRAMANIK62/devhunt/internal/middleware"
	"github.com/PPRAMANIK62/devhunt/internal/queue"
	"github.com/PPRAMANIK62/devhunt/internal/repository"
	"github.com/PPRAMANIK62/devhunt/internal/service"
)

// @title           DevHunt API
// @version         1.0
// @description     A job board API for developers.

// @host            localhost:8080
// @BasePath        /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and the JWT token.

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}

	// Logger
	logger.SetupLogger(cfg.Env)
	slog.Info("devhunt starting", "port", cfg.ServerPort, "env", cfg.Env)

	// Run migrations
	if err := database.RunMigrations(cfg.DatabaseURL); err != nil {
		slog.Error("migrations failed", "error", err)
		os.Exit(1)
	}
	slog.Info("migrations applied")

	ctx := context.Background()

	// Postgres (required)
	db, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("database error", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	var appCache *cache.Cache
	var queueClient *queue.Client
	if cfg.RedisURL != "" {
		var err error
		// Redis (optional)
		appCache, err = cache.New(cfg.RedisURL)
		if err != nil {
			// Not fatal - app works without caching
			slog.Warn("redis unavailable, caching and email notifications disabled", "error", err)
		} else {
			defer appCache.Close()

			// Queue (optional)
			queueClient, err = queue.NewClient(cfg.RedisURL)
			if err != nil {
				slog.Warn("queue client unavailable, email notifications disabled", "error", err)
			} else {
				defer queueClient.Close()
			}

			// email worker
			workerSrv, workerMux, err := queue.NewWorkerServer(cfg.RedisURL, cfg.ResendAPIKey, cfg.AppBaseURL)
			if err != nil {
				slog.Warn("queue worker unavailable", "error", err)
			} else {
				go func() {
					slog.Info("queue worker started")
					if err := workerSrv.Run(workerMux); err != nil {
						slog.Error("queue worker stopped", "error", err)
					}
				}()
				defer workerSrv.Shutdown()
			}
		}
	}

	// Repositories
	userRepo := repository.NewUserRepository(db)
	jobRepo := repository.NewJobRepository(db)
	companyRepo := repository.NewCompanyRepository(db)
	applicationRepo := repository.NewApplicationRepository(db)

	// Services
	authSvc := service.NewAuthService(userRepo, cfg.JWTSecret, cfg.JWTExpiryMinutes, queueClient)
	jobSvc := service.NewJobService(jobRepo, companyRepo, appCache)
	companySvc := service.NewCompanyService(companyRepo)
	applicationSvc := service.NewApplicationService(applicationRepo, jobRepo, companyRepo, userRepo, queueClient)

	// Handlers
	authHandler := handler.NewAuthHandler(authSvc)
	jobHandler := handler.NewJobHandler(jobSvc)
	companyHandler := handler.NewCompanyHandler(companySvc)
	applicationHandler := handler.NewApplicationHandler(applicationSvc)

	// Middlewares
	authMw := middleware.NewAuthMiddleware(cfg.JWTSecret)
	companyMW := middleware.NewRoleMiddleware("company")

	// Router
	router := setupRoutes(
		authHandler,
		jobHandler,
		companyHandler,
		applicationHandler,
		authMw,
		companyMW,
	)

	// HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Run server in background so signal listener isn't blocked
	go func() {
		slog.Info("server listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown
	// Block until SIGTERM or Ctrl+C
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	sig := <-quit

	slog.Info("shutdown signal received", "signal", sig.String())

	// Give in-flight requests 30 seconds to finish
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("forced shutdown", "error", err)
	}

	slog.Info("server stopped cleanly")
	// defer db.Close() and other deferred cleanup runs here
}
