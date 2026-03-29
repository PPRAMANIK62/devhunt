package main

import (
	"net/http"
	"time"

	"github.com/PPRAMANIK62/devhunt/internal/handler"
	appmiddleware "github.com/PPRAMANIK62/devhunt/internal/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func setupRoutes(
	authHandler *handler.AuthHandler,
	jobHandler *handler.JobHandler,
	companyHandler *handler.CompanyHandler,
	applicationHandler *handler.ApplicationHandler,
	authMW func(http.Handler) http.Handler,
	companyMW func(http.Handler) http.Handler,
) http.Handler {
	r := chi.NewRouter()

	// Global Middleware
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)
	r.Use(middleware.RequestID)
	r.Use(appmiddleware.RequestLogger) // replaces fmt.Println logging

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			// verify-email is excluded: tokens are high-entropy (can't brute-force)
			// and email clients may pre-fetch the link before the user clicks it.
			r.Get("/verify-email", authHandler.VerifyEmail)

			r.Group(func(r chi.Router) {
				r.Use(appmiddleware.RateLimit(5, time.Minute))
				r.Post("/register", authHandler.Register)
				r.Post("/login", authHandler.Login)
				r.Post("/resend-verification", authHandler.ResendVerification)
			})
		})

		r.Route("/jobs", func(r chi.Router) {
			r.Get("/", jobHandler.List)
			r.Get("/filters", jobHandler.GetFilterOptions)
			r.Get("/{id}", jobHandler.GetByID)

			r.Group(func(r chi.Router) {
				r.Use(authMW)
				r.Use(companyMW)
				r.Post("/", jobHandler.Create)
				r.Patch("/{id}", jobHandler.Update)
				r.Delete("/{id}", jobHandler.Delete)
			})
		})

		r.Route("/companies", func(r chi.Router) {
			r.Get("/{id}", companyHandler.GetByID) // public

			r.Group(func(r chi.Router) {
				r.Use(authMW)
				r.Use(companyMW)
				r.Post("/", companyHandler.Create)
				r.Get("/me", companyHandler.GetMine)
				r.Patch("/me", companyHandler.Update)
				r.Delete("/me", companyHandler.Delete)
				r.Get("/me/jobs", jobHandler.ListMine)
			})
		})

		r.Route("/jobs/{jobID}/applications", func(r chi.Router) {
			r.Use(authMW)
			r.Post("/", applicationHandler.Apply)
			r.Group(func(r chi.Router) {
				r.Use(companyMW)
				r.Get("/", applicationHandler.ListByJobID)
			})
		})

		r.Route("/applications", func(r chi.Router) {
			r.Use(authMW)
			r.Get("/", applicationHandler.ListMine)
			r.Patch("/{id}/status", applicationHandler.UpdateStatus)
		})
	})

	return r
}
