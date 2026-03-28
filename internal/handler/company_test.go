package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/PPRAMANIK62/devhunt/internal/apperr"
	"github.com/PPRAMANIK62/devhunt/internal/handler"
	"github.com/PPRAMANIK62/devhunt/internal/middleware"
	"github.com/PPRAMANIK62/devhunt/internal/models"
	"github.com/go-chi/chi/v5"
)

// stubCompanyService is a test double that satisfies the companyServicer interface.
type stubCompanyService struct {
	createFn  func(ctx context.Context, userID string, req models.CreateCompanyRequest) (*models.Company, error)
	getMineFn func(ctx context.Context, userID string) (*models.Company, error)
	getByIDFn func(ctx context.Context, id string) (*models.Company, error)
	updateFn  func(ctx context.Context, userID string, req models.UpdateCompanyRequest) (*models.Company, error)
	deleteFn  func(ctx context.Context, userID string) error
}

func (s *stubCompanyService) Create(ctx context.Context, userID string, req models.CreateCompanyRequest) (*models.Company, error) {
	return s.createFn(ctx, userID, req)
}

func (s *stubCompanyService) GetMine(ctx context.Context, userID string) (*models.Company, error) {
	return s.getMineFn(ctx, userID)
}

func (s *stubCompanyService) GetByID(ctx context.Context, id string) (*models.Company, error) {
	return s.getByIDFn(ctx, id)
}

func (s *stubCompanyService) Update(ctx context.Context, userID string, req models.UpdateCompanyRequest) (*models.Company, error) {
	return s.updateFn(ctx, userID, req)
}

func (s *stubCompanyService) Delete(ctx context.Context, userID string) error {
	return s.deleteFn(ctx, userID)
}

// withUserID injects a user_id into the request context (mimics authMW).
func withUserID(r *http.Request, userID string) *http.Request {
	ctx := context.WithValue(r.Context(), middleware.ContextKeyUserID, userID)
	return r.WithContext(ctx)
}

func TestCompanyCreate(t *testing.T) {
	sampleCompany := &models.Company{ID: "c1", UserID: "u1", Name: "Acme", Slug: "acme"}

	tests := []struct {
		name       string
		body       any
		createFn   func(ctx context.Context, userID string, req models.CreateCompanyRequest) (*models.Company, error)
		wantStatus int
	}{
		{
			name: "success",
			body: models.CreateCompanyRequest{Name: "Acme", Slug: "acme"},
			createFn: func(_ context.Context, _ string, _ models.CreateCompanyRequest) (*models.Company, error) {
				return sampleCompany, nil
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "conflict when company already exists",
			body: models.CreateCompanyRequest{Name: "Acme", Slug: "acme"},
			createFn: func(_ context.Context, _ string, _ models.CreateCompanyRequest) (*models.Company, error) {
				return nil, apperr.Conflict("company profile already exists")
			},
			wantStatus: http.StatusConflict,
		},
		{
			name:       "bad json body",
			body:       "not-json{{",
			createFn:   nil,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := handler.NewCompanyHandlerWithService(&stubCompanyService{createFn: tc.createFn})

			var buf bytes.Buffer
			json.NewEncoder(&buf).Encode(tc.body)
			req := httptest.NewRequest(http.MethodPost, "/companies", &buf)
			req = withUserID(req, "u1")
			rr := httptest.NewRecorder()

			h.Create(rr, req)

			if rr.Code != tc.wantStatus {
				t.Errorf("want status %d, got %d", tc.wantStatus, rr.Code)
			}
		})
	}
}

func TestCompanyGetMine(t *testing.T) {
	sampleCompany := &models.Company{ID: "c1", UserID: "u1", Name: "Acme", Slug: "acme"}

	tests := []struct {
		name       string
		getMineFn  func(ctx context.Context, userID string) (*models.Company, error)
		wantStatus int
	}{
		{
			name: "success",
			getMineFn: func(_ context.Context, _ string) (*models.Company, error) {
				return sampleCompany, nil
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "not found",
			getMineFn: func(_ context.Context, _ string) (*models.Company, error) {
				return nil, apperr.NotFound("company")
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := handler.NewCompanyHandlerWithService(&stubCompanyService{getMineFn: tc.getMineFn})

			req := httptest.NewRequest(http.MethodGet, "/companies/me", nil)
			req = withUserID(req, "u1")
			rr := httptest.NewRecorder()

			h.GetMine(rr, req)

			if rr.Code != tc.wantStatus {
				t.Errorf("want status %d, got %d", tc.wantStatus, rr.Code)
			}
		})
	}
}

func TestCompanyGetByID(t *testing.T) {
	sampleCompany := &models.Company{ID: "c1", UserID: "u1", Name: "Acme", Slug: "acme"}

	tests := []struct {
		name       string
		id         string
		getByIDFn  func(ctx context.Context, id string) (*models.Company, error)
		wantStatus int
	}{
		{
			name: "success",
			id:   "c1",
			getByIDFn: func(_ context.Context, id string) (*models.Company, error) {
				return sampleCompany, nil
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "not found",
			id:   "missing",
			getByIDFn: func(_ context.Context, _ string) (*models.Company, error) {
				return nil, apperr.NotFound("company")
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := handler.NewCompanyHandlerWithService(&stubCompanyService{getByIDFn: tc.getByIDFn})

			r := chi.NewRouter()
			r.Get("/{id}", h.GetByID)

			req := httptest.NewRequest(http.MethodGet, "/"+tc.id, nil)
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			if rr.Code != tc.wantStatus {
				t.Errorf("want status %d, got %d", tc.wantStatus, rr.Code)
			}
		})
	}
}

func TestCompanyUpdate(t *testing.T) {
	sampleCompany := &models.Company{ID: "c1", UserID: "u1", Name: "Updated", Slug: "updated"}
	name := "Updated"
	slug := "updated"

	tests := []struct {
		name       string
		body       any
		updateFn   func(ctx context.Context, userID string, req models.UpdateCompanyRequest) (*models.Company, error)
		wantStatus int
	}{
		{
			name: "success partial update",
			body: models.UpdateCompanyRequest{Name: &name, Slug: &slug},
			updateFn: func(_ context.Context, _ string, _ models.UpdateCompanyRequest) (*models.Company, error) {
				return sampleCompany, nil
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "not found",
			body: models.UpdateCompanyRequest{Name: &name},
			updateFn: func(_ context.Context, _ string, _ models.UpdateCompanyRequest) (*models.Company, error) {
				return nil, apperr.NotFound("company")
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := handler.NewCompanyHandlerWithService(&stubCompanyService{updateFn: tc.updateFn})

			var buf bytes.Buffer
			json.NewEncoder(&buf).Encode(tc.body)
			req := httptest.NewRequest(http.MethodPatch, "/companies/me", &buf)
			req = withUserID(req, "u1")
			rr := httptest.NewRecorder()

			h.Update(rr, req)

			if rr.Code != tc.wantStatus {
				t.Errorf("want status %d, got %d", tc.wantStatus, rr.Code)
			}
		})
	}
}

func TestCompanyDelete(t *testing.T) {
	tests := []struct {
		name       string
		deleteFn   func(ctx context.Context, userID string) error
		wantStatus int
	}{
		{
			name: "success",
			deleteFn: func(_ context.Context, _ string) error {
				return nil
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name: "not found",
			deleteFn: func(_ context.Context, _ string) error {
				return apperr.NotFound("company")
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := handler.NewCompanyHandlerWithService(&stubCompanyService{deleteFn: tc.deleteFn})

			req := httptest.NewRequest(http.MethodDelete, "/companies/me", nil)
			req = withUserID(req, "u1")
			rr := httptest.NewRecorder()

			h.Delete(rr, req)

			if rr.Code != tc.wantStatus {
				t.Errorf("want status %d, got %d", tc.wantStatus, rr.Code)
			}
		})
	}
}
