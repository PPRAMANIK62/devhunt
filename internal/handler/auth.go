package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/PPRAMANIK62/devhunt/internal/apperr"
	"github.com/PPRAMANIK62/devhunt/internal/models"
	"github.com/PPRAMANIK62/devhunt/internal/service"
)

type authServicer interface {
	Register(ctx context.Context, input service.RegisterInput) (*models.User, error)
	Login(ctx context.Context, email, password string) (*service.LoginOutput, error)
	VerifyEmail(ctx context.Context, token string) error
	ResendVerification(ctx context.Context, email string) error
}

type AuthHandler struct {
	authService authServicer
}

func NewAuthHandler(s *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: s}
}

func newAuthHandlerWithService(s authServicer) *AuthHandler {
	return &AuthHandler{authService: s}
}

type registerRequest struct {
	Email    string          `json:"email"    validate:"required,email"`
	Password string          `json:"password" validate:"required,min=8"`
	Role     models.UserRole `json:"role"     validate:"required,oneof=seeker company"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, apperr.Validation("invalid request body"))
		return
	}
	if err := validate(req); err != nil {
		writeError(w, err)
		return
	}

	user, err := h.authService.Register(r.Context(), service.RegisterInput{
		Email:    req.Email,
		Password: req.Password,
		Role:     req.Role,
	})
	if err != nil {
		writeError(w, err)
		return
	}

	writeSuccess(w, http.StatusCreated, user)
}

type loginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, apperr.Validation("invalid request body"))
		return
	}
	if err := validate(req); err != nil {
		writeError(w, err)
		return
	}

	output, err := h.authService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		writeError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, map[string]any{
		"token": output.Token,
		"user":  output.User,
	})
}

func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		writeError(w, apperr.Validation("token is required"))
		return
	}

	if err := h.authService.VerifyEmail(r.Context(), token); err != nil {
		writeError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, map[string]any{"message": "email verified successfully"})
}

type resendVerificationRequest struct {
	Email string `json:"email" validate:"required,email"`
}

func (h *AuthHandler) ResendVerification(w http.ResponseWriter, r *http.Request) {
	var req resendVerificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, apperr.Validation("invalid request body"))
		return
	}
	if err := validate(req); err != nil {
		writeError(w, err)
		return
	}

	// Always 200 — never reveal if email exists or is already verified
	_ = h.authService.ResendVerification(r.Context(), req.Email)
	writeSuccess(w, http.StatusOK, map[string]any{"message": "if that email exists and is unverified, a new link has been sent"})
}
