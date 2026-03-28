package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/PPRAMANIK62/devhunt/internal/apperr"
	"github.com/PPRAMANIK62/devhunt/internal/models"
	"github.com/PPRAMANIK62/devhunt/internal/service"
)

type stubAuthService struct {
	registerFn func(ctx context.Context, input service.RegisterInput) (*models.User, error)
	loginFn    func(ctx context.Context, email, password string) (*service.LoginOutput, error)
}

type fakeAuthHandler struct {
	svc *stubAuthService
}

func (h *fakeAuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string          `json:"email"`
		Password string          `json:"password"`
		Role     models.UserRole `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	user, err := h.svc.registerFn(r.Context(), service.RegisterInput{
		Email:    req.Email,
		Password: req.Password,
		Role:     req.Role,
	})
	if err != nil {
		w.WriteHeader(apperr.HTTPStatus(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{"data": user})
}

func (h *fakeAuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	output, err := h.svc.loginFn(r.Context(), req.Email, req.Password)
	if err != nil {
		w.WriteHeader(apperr.HTTPStatus(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"token": output.Token, "user": output.User}})
}

func TestAuthRegister(t *testing.T) {
	sampleUser := &models.User{ID: "u1", Email: "a@b.com", Role: "company"}

	tests := []struct {
		name       string
		body       any
		registerFn func(ctx context.Context, input service.RegisterInput) (*models.User, error)
		wantStatus int
	}{
		{
			name: "success",
			body: map[string]string{"email": "a@b.com", "password": "secret123", "role": "company"},
			registerFn: func(_ context.Context, _ service.RegisterInput) (*models.User, error) {
				return sampleUser, nil
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "conflict when user already exists",
			body: map[string]string{"email": "a@b.com", "password": "secret123", "role": "company"},
			registerFn: func(_ context.Context, _ service.RegisterInput) (*models.User, error) {
				return nil, apperr.Conflict("email already in use")
			},
			wantStatus: http.StatusConflict,
		},
		{
			name:       "bad json body",
			body:       "not-json{{",
			registerFn: nil,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := &fakeAuthHandler{svc: &stubAuthService{registerFn: tc.registerFn}}

			var buf bytes.Buffer
			json.NewEncoder(&buf).Encode(tc.body)
			req := httptest.NewRequest(http.MethodPost, "/auth/register", &buf)
			rr := httptest.NewRecorder()

			h.Register(rr, req)

			if rr.Code != tc.wantStatus {
				t.Errorf("want status %d, got %d", tc.wantStatus, rr.Code)
			}
		})
	}
}

func TestAuthLogin(t *testing.T) {
	sampleOutput := &service.LoginOutput{
		Token: "jwt.token.here",
		User:  &models.User{ID: "u1", Email: "a@b.com", Role: "company"},
	}

	tests := []struct {
		name       string
		body       any
		loginFn    func(ctx context.Context, email, password string) (*service.LoginOutput, error)
		wantStatus int
	}{
		{
			name: "success",
			body: map[string]string{"email": "a@b.com", "password": "secret123"},
			loginFn: func(_ context.Context, _, _ string) (*service.LoginOutput, error) {
				return sampleOutput, nil
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "invalid credentials",
			body: map[string]string{"email": "a@b.com", "password": "wrong"},
			loginFn: func(_ context.Context, _, _ string) (*service.LoginOutput, error) {
				return nil, apperr.Unauthorized("invalid credentials")
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:    "bad json body",
			body:    "not-json{{",
			loginFn: nil,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := &fakeAuthHandler{svc: &stubAuthService{loginFn: tc.loginFn}}

			var buf bytes.Buffer
			json.NewEncoder(&buf).Encode(tc.body)
			req := httptest.NewRequest(http.MethodPost, "/auth/login", &buf)
			rr := httptest.NewRecorder()

			h.Login(rr, req)

			if rr.Code != tc.wantStatus {
				t.Errorf("want status %d, got %d", tc.wantStatus, rr.Code)
			}
		})
	}
}
