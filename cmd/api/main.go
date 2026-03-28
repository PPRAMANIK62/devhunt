package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/PPRAMANIK62/devhunt/internal/config"
	"github.com/PPRAMANIK62/devhunt/internal/database"
	"github.com/PPRAMANIK62/devhunt/internal/handler"
	"github.com/PPRAMANIK62/devhunt/internal/middleware"
	"github.com/PPRAMANIK62/devhunt/internal/repository"
	"github.com/PPRAMANIK62/devhunt/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()

	db, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "database error: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// Repositories
	userRepo := repository.NewUserRepository(db)
	jobRepo := repository.NewJobRepository(db)
	companyRepo := repository.NewCompanyRepository(db)

	// Services
	authSvc := service.NewAuthService(userRepo, cfg.JWTSecret, cfg.JWTExpiryMinutes)
	jobSvc := service.NewJobService(jobRepo, companyRepo)
	companySvc := service.NewCompanyService(companyRepo)

	// Handlers
	authHandler := handler.NewAuthHandler(authSvc)
	jobHandler := handler.NewJobHandler(jobSvc)
	companyHandler := handler.NewCompanyHandler(companySvc)

	// Middlewares
	authMw := middleware.NewAuthMiddleware(cfg.JWTSecret)
	companyMW := middleware.NewRoleMiddleware("company")

	// Router
	router := setupRoutes(authHandler, jobHandler, companyHandler, authMw, companyMW)

	fmt.Printf("server listening on :%s\n", cfg.ServerPort)
	if err := http.ListenAndServe(":"+cfg.ServerPort, router); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}
