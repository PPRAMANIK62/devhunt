package main

import (
	"net/http"

	"github.com/PPRAMANIK62/devhunt/internal/handler"
	appmiddleware "github.com/PPRAMANIK62/devhunt/internal/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func setupRoutes(authHandler *handler.AuthHandler, authMW func(http.Handler) http.Handler) http.Handler {
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
		r.Post("/auth/register", authHandler.Register)
		r.Post("/auth/login", authHandler.Login)

		// Example protected route (expand later)
		r.Group(func(r chi.Router) {
			r.Use(authMW)
			r.Get("/me", func(w http.ResponseWriter, r *http.Request) {
				userID := appmiddleware.GetUserID(r.Context())
				w.Write([]byte(`{"user_id":"` + userID + `"}`))
			})
		})
	})

	return r
}
