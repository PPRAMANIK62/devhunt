package main

import (
	"net/http"

	"github.com/PPRAMANIK62/devhunt/internal/handler"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func setupRoutes(
	authHandler *handler.AuthHandler,
	jobHandler *handler.JobHandler,
	companyHandler *handler.CompanyHandler,
	authMW func(http.Handler) http.Handler,
	companyMW func(http.Handler) http.Handler,
) http.Handler {
	r := chi.NewRouter()

	// Global Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)
	r.Use(middleware.RequestID)

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/auth/register", authHandler.Register)
		r.Post("/auth/login", authHandler.Login)

		r.Route("/jobs", func(r chi.Router) {
			r.Get("/", jobHandler.List)
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
			})
		})
	})

	return r
}
