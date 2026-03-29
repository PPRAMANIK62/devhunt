package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

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

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}

	logger.SetupLogger(cfg.Env)
	slog.Info("devhunt starting", "port", cfg.ServerPort, "env", cfg.Env)

	ctx := context.Background()

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
		appCache, err = cache.New(cfg.RedisURL)
		if err != nil {
			// Not fatal - app works without caching
			slog.Warn("redis unavailable, caching disabled", "error", err)
		} else {
			defer appCache.Close()
		}

		queueClient, err = queue.NewClient(cfg.RedisURL)
		if err != nil {
			slog.Warn("queue client unavailable, email notifications disabled", "error", err)
		} else {
			defer queueClient.Close()
		}

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

	slog.Info("server listening", "port", cfg.ServerPort)
	if err := http.ListenAndServe(":"+cfg.ServerPort, router); err != nil {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}
