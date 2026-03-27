package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func setupRoutes() http.Handler {
	r := chi.NewRouter()

	// Global Middleware
	r.Use(middleware.RequestID) // adds X-Request-ID to every request
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer) // catches panics, returns 500 instead of crashing

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{status:"ok"}`))
	})

	r.Route("/v1", func(r chi.Router) {
		// Auth, jobs, applications will be added here
	})

	return r
}
