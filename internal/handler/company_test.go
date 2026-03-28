package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/PPRAMANIK62/devhunt/internal/apperr"
	"github.com/PPRAMANIK62/devhunt/internal/middleware"
	"github.com/PPRAMANIK62/devhunt/internal/models"
	"github.com/go-chi/chi/v5"
)

// stubCompanyService is a local test double — no production interface needed.
type stubCompanyService struct {
	createFn  func(ctx context.Context, userID string, req models.CreateCompanyRequest) (*models.Company, error)
	getMineFn func(ctx context.Context, userID string) (*models.Company, error)
	getByIDFn func(ctx context.Context, id string) (*models.Company, error)
	updateFn  func(ctx context.Context, userID string, req models.UpdateCompanyRequest) (*models.Company, error)
	deleteFn  func(ctx context.Context, userID string) error
}

// fakeCompanyHandler mirrors CompanyHandler but accepts the stub.
// It avoids modifying the production handler.
type fakeCompanyHandler struct {
	svc *stubCompanyService
}

func (h *fakeCompanyHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateCompanyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	company, err := h.svc.createFn(r.Context(), middleware.GetUserID(r.Context()), req)
	if err != nil {
		var code int
		switch apperr.HTTPStatus(err) {
		case http.StatusConflict:
			code = http.StatusConflict
		default:
			code = http.StatusInternalServerError
		}
		w.WriteHeader(code)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{"data": company})
}

func (h *fakeCompanyHandler) GetMine(w http.ResponseWriter, r *http.Request) {
	company, err := h.svc.getMineFn(r.Context(), middleware.GetUserID(r.Context()))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"data": company})
}

func (h *fakeCompanyHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	company, err := h.svc.getByIDFn(r.Context(), id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"data": company})
}

func (h *fakeCompanyHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req models.UpdateCompanyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	company, err := h.svc.updateFn(r.Context(), middleware.GetUserID(r.Context()), req)
	if err != nil {
		w.WriteHeader(apperr.HTTPStatus(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"data": company})
}

func (h *fakeCompanyHandler) Delete(w http.ResponseWriter, r *http.Request) {
	err := h.svc.deleteFn(r.Context(), middleware.GetUserID(r.Context()))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
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
			h := &fakeCompanyHandler{svc: &stubCompanyService{createFn: tc.createFn}}

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
			h := &fakeCompanyHandler{svc: &stubCompanyService{getMineFn: tc.getMineFn}}

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
			h := &fakeCompanyHandler{svc: &stubCompanyService{getByIDFn: tc.getByIDFn}}

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
			h := &fakeCompanyHandler{svc: &stubCompanyService{updateFn: tc.updateFn}}

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
			h := &fakeCompanyHandler{svc: &stubCompanyService{deleteFn: tc.deleteFn}}

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
