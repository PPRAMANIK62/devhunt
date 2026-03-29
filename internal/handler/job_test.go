package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/PPRAMANIK62/devhunt/internal/apperr"
	"github.com/PPRAMANIK62/devhunt/internal/handler"
	"github.com/PPRAMANIK62/devhunt/internal/middleware"
	"github.com/PPRAMANIK62/devhunt/internal/models"
	"github.com/PPRAMANIK62/devhunt/internal/repository"
	"github.com/PPRAMANIK62/devhunt/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

// ── stub service for unit tests ──────────────────────────────────────────────

type stubJobService struct {
	listFn             func(ctx context.Context, page, pageSize int, f service.ListJobsFilter) (*service.ListJobsOutput, error)
	getByIDFn          func(ctx context.Context, id string) (*models.Job, error)
	createFn           func(ctx context.Context, userID string, req models.CreateJobRequest) (*models.Job, error)
	updateFn           func(ctx context.Context, id, userID string, req models.UpdateJobRequest) (*models.Job, error)
	deleteFn           func(ctx context.Context, id, userID string) error
	listMineFn         func(ctx context.Context, userID, status string) ([]*models.Job, error)
	getFilterOptionsFn func(ctx context.Context) (*service.FilterOptions, error)
}

func (s *stubJobService) List(ctx context.Context, page, pageSize int, f service.ListJobsFilter) (*service.ListJobsOutput, error) {
	return s.listFn(ctx, page, pageSize, f)
}
func (s *stubJobService) GetByID(ctx context.Context, id string) (*models.Job, error) {
	return s.getByIDFn(ctx, id)
}
func (s *stubJobService) Create(ctx context.Context, userID string, req models.CreateJobRequest) (*models.Job, error) {
	return s.createFn(ctx, userID, req)
}
func (s *stubJobService) Update(ctx context.Context, id, userID string, req models.UpdateJobRequest) (*models.Job, error) {
	return s.updateFn(ctx, id, userID, req)
}
func (s *stubJobService) Delete(ctx context.Context, id, userID string) error {
	return s.deleteFn(ctx, id, userID)
}
func (s *stubJobService) ListMine(ctx context.Context, userID, status string) ([]*models.Job, error) {
	return s.listMineFn(ctx, userID, status)
}
func (s *stubJobService) GetFilterOptions(ctx context.Context) (*service.FilterOptions, error) {
	if s.getFilterOptionsFn != nil {
		return s.getFilterOptionsFn(ctx)
	}
	return &service.FilterOptions{Locations: []string{}, Tags: []string{}}, nil
}

// ── stub-based unit tests ────────────────────────────────────────────────────

func TestJobHandler_List(t *testing.T) {
	sampleOutput := &service.ListJobsOutput{
		Jobs:     []*models.Job{{ID: "j1", Title: "Go Engineer"}},
		Total:    1,
		Page:     1,
		PageSize: 20,
	}

	tests := []struct {
		name       string
		listFn     func(context.Context, int, int, service.ListJobsFilter) (*service.ListJobsOutput, error)
		wantStatus int
	}{
		{
			name: "success",
			listFn: func(_ context.Context, _, _ int, _ service.ListJobsFilter) (*service.ListJobsOutput, error) {
				return sampleOutput, nil
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "service error returns 500",
			listFn: func(_ context.Context, _, _ int, _ service.ListJobsFilter) (*service.ListJobsOutput, error) {
				return nil, apperr.Internal("db", nil)
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := handler.NewJobHandlerWithService(&stubJobService{listFn: tc.listFn})
			req := httptest.NewRequest(http.MethodGet, "/jobs", nil)
			rr := httptest.NewRecorder()
			h.List(rr, req)
			assert.Equal(t, tc.wantStatus, rr.Code)
		})
	}
}

func TestJobHandler_GetByID(t *testing.T) {
	tests := []struct {
		name       string
		getByIDFn  func(context.Context, string) (*models.Job, error)
		wantStatus int
	}{
		{
			name: "success",
			getByIDFn: func(_ context.Context, id string) (*models.Job, error) {
				return &models.Job{ID: id, Title: "Go Engineer"}, nil
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "not found",
			getByIDFn: func(_ context.Context, _ string) (*models.Job, error) {
				return nil, apperr.NotFound("job")
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := handler.NewJobHandlerWithService(&stubJobService{getByIDFn: tc.getByIDFn})

			r := chi.NewRouter()
			r.Get("/{id}", h.GetByID)

			req := httptest.NewRequest(http.MethodGet, "/job-1", nil)
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)
			assert.Equal(t, tc.wantStatus, rr.Code)
		})
	}
}

func TestJobHandler_Create_Success(t *testing.T) {
	h := handler.NewJobHandlerWithService(&stubJobService{
		createFn: func(_ context.Context, _ string, _ models.CreateJobRequest) (*models.Job, error) {
			return &models.Job{ID: "j-new", Title: "Go Engineer"}, nil
		},
	})

	body, _ := json.Marshal(models.CreateJobRequest{
		Title:       "Go Engineer",
		Description: "We need a Go engineer with experience in distributed systems.",
		Location:    "Remote",
		SalaryMin:   100000,
		SalaryMax:   150000,
	})
	req := httptest.NewRequest(http.MethodPost, "/jobs", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withUserID(req, "user-1")
	rr := httptest.NewRecorder()

	h.Create(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
}

func TestJobHandler_Create_Forbidden(t *testing.T) {
	h := handler.NewJobHandlerWithService(&stubJobService{
		createFn: func(_ context.Context, _ string, _ models.CreateJobRequest) (*models.Job, error) {
			return nil, apperr.Forbidden("you must have a company profile to post jobs")
		},
	})

	body, _ := json.Marshal(models.CreateJobRequest{
		Title:       "Go Engineer",
		Description: "We need a Go engineer with experience in distributed systems.",
		Location:    "Remote",
		SalaryMin:   100000,
		SalaryMax:   150000,
	})
	req := httptest.NewRequest(http.MethodPost, "/jobs", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withUserID(req, "user-no-company")
	rr := httptest.NewRecorder()

	h.Create(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestJobHandler_Delete_Success(t *testing.T) {
	h := handler.NewJobHandlerWithService(&stubJobService{
		deleteFn: func(_ context.Context, _, _ string) error { return nil },
	})

	r := chi.NewRouter()
	r.Delete("/{id}", h.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/job-1", nil)
	req = withUserID(req, "owner-user")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
}

func TestJobHandler_Delete_Forbidden(t *testing.T) {
	h := handler.NewJobHandlerWithService(&stubJobService{
		deleteFn: func(_ context.Context, _, _ string) error {
			return apperr.Forbidden("you do not own this job posting")
		},
	})

	r := chi.NewRouter()
	r.Delete("/{id}", h.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/job-1", nil)
	req = withUserID(req, "other-user")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestJobHandler_ListMine(t *testing.T) {
	sampleJobs := []*models.Job{
		{ID: "j1", Title: "Backend Engineer", Status: "open"},
		{ID: "j2", Title: "Frontend Engineer", Status: "draft"},
	}

	tests := []struct {
		name       string
		query      string
		listMineFn func(context.Context, string, string) ([]*models.Job, error)
		wantStatus int
	}{
		{
			name:  "success - no filter",
			query: "",
			listMineFn: func(_ context.Context, _, _ string) ([]*models.Job, error) {
				return sampleJobs, nil
			},
			wantStatus: http.StatusOK,
		},
		{
			name:  "filter by open",
			query: "?status=open",
			listMineFn: func(_ context.Context, _, status string) ([]*models.Job, error) {
				assert.Equal(t, "open", status)
				return sampleJobs[:1], nil
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid status param returns 400",
			query:      "?status=unknown",
			listMineFn: nil,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:  "service forbidden returns 403",
			query: "",
			listMineFn: func(_ context.Context, _, _ string) ([]*models.Job, error) {
				return nil, apperr.Forbidden("you must have a company profile")
			},
			wantStatus: http.StatusForbidden,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := handler.NewJobHandlerWithService(&stubJobService{listMineFn: tc.listMineFn})
			req := httptest.NewRequest(http.MethodGet, "/companies/me/jobs"+tc.query, nil)
			req = withUserID(req, "u1")
			rr := httptest.NewRecorder()
			h.ListMine(rr, req)
			assert.Equal(t, tc.wantStatus, rr.Code)
		})
	}
}

func TestJobHandler_GetFilterOptions(t *testing.T) {
	tests := []struct {
		name               string
		getFilterOptionsFn func(context.Context) (*service.FilterOptions, error)
		wantStatus         int
	}{
		{
			name: "success",
			getFilterOptionsFn: func(_ context.Context) (*service.FilterOptions, error) {
				return &service.FilterOptions{Locations: []string{"Remote", "Austin"}, Tags: []string{"Go", "Rust"}}, nil
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "service error returns 500",
			getFilterOptionsFn: func(_ context.Context) (*service.FilterOptions, error) {
				return nil, apperr.Internal("db error", nil)
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := handler.NewJobHandlerWithService(&stubJobService{getFilterOptionsFn: tc.getFilterOptionsFn})
			req := httptest.NewRequest(http.MethodGet, "/jobs/filters", nil)
			rr := httptest.NewRecorder()
			h.GetFilterOptions(rr, req)
			assert.Equal(t, tc.wantStatus, rr.Code)
		})
	}
}

// ── integration tests (require Docker via TestMain) ──────────────────────────

const testJWTSecret = "test-jwt-secret"

func TestJobHandler_Create_Unauthorized(t *testing.T) {
	// Build the full server (same wiring as main.go, using test DB)
	srv := buildTestServer(t)

	body, _ := json.Marshal(map[string]any{
		"title":       "Go Engineer",
		"description": "A sufficiently long description for testing purposes here.",
		"location":    "Remote",
		"salary_min":  100000,
		"salary_max":  150000,
	})

	req := httptest.NewRequest(http.MethodPost, "/v1/jobs", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	// No Authorization header

	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestJobHandler_Create_ValidationError(t *testing.T) {
	srv := buildTestServer(t)
	token := loginAsCompany(t, srv)

	// Missing required fields
	body, _ := json.Marshal(map[string]any{"title": "x"})

	req := httptest.NewRequest(http.MethodPost, "/v1/jobs", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]any
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Equal(t, "VALIDATION_ERROR", resp["code"])
}

func TestJobHandler_FullLifecycle(t *testing.T) {
	srv := buildTestServer(t)
	token := loginAsCompany(t, srv)

	// Create
	createBody, _ := json.Marshal(map[string]any{
		"title":       "Senior Go Engineer",
		"description": "We need an experienced Go engineer for our infrastructure team.",
		"location":    "Remote",
		"salary_min":  120000,
		"salary_max":  160000,
	})
	req := httptest.NewRequest(http.MethodPost, "/v1/jobs", bytes.NewReader(createBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var createResp struct{ Data struct{ ID string } }
	json.NewDecoder(w.Body).Decode(&createResp)
	jobID := createResp.Data.ID

	// Get
	req = httptest.NewRequest(http.MethodGet, "/v1/jobs/"+jobID, nil)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Delete
	req = httptest.NewRequest(http.MethodDelete, "/v1/jobs/"+jobID, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNoContent, w.Code)

	// Get after delete — should 404
	req = httptest.NewRequest(http.MethodGet, "/v1/jobs/"+jobID, nil)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// buildTestServer wires up the full server using the test database.
func buildTestServer(t *testing.T) http.Handler {
	t.Helper()

	userRepo := repository.NewUserRepository(testDB)
	jobRepo := repository.NewJobRepository(testDB)
	companyRepo := repository.NewCompanyRepository(testDB)
	appRepo := repository.NewApplicationRepository(testDB)

	authSvc := service.NewAuthService(userRepo, testJWTSecret, 60, nil)
	jobSvc := service.NewJobService(jobRepo, companyRepo, nil)
	companySvc := service.NewCompanyService(companyRepo)
	appSvc := service.NewApplicationService(appRepo, jobRepo, companyRepo, userRepo, nil)

	authH := handler.NewAuthHandler(authSvc)
	jobH := handler.NewJobHandler(jobSvc)
	companyH := handler.NewCompanyHandler(companySvc)
	_ = handler.NewApplicationHandler(appSvc) // wired but not routed in this test server

	authMW := middleware.NewAuthMiddleware(testJWTSecret)
	companyMW := middleware.NewRoleMiddleware("company")

	r := chi.NewRouter()
	r.Route("/v1", func(r chi.Router) {
		r.Post("/auth/login", authH.Login)
		r.Route("/jobs", func(r chi.Router) {
			r.Get("/{id}", jobH.GetByID)
			r.Group(func(r chi.Router) {
				r.Use(authMW)
				r.Use(companyMW)
				r.Post("/", jobH.Create)
				r.Delete("/{id}", jobH.Delete)
			})
		})
		r.Group(func(r chi.Router) {
			r.Use(authMW)
			r.Use(companyMW)
			r.Post("/companies", companyH.Create)
		})
	})

	return r
}

// loginAsCompany inserts a company-role user directly, logs in, creates a company
// profile, and returns the JWT token.
func loginAsCompany(t *testing.T, srv http.Handler) string {
	t.Helper()
	ctx := context.Background()

	n := time.Now().UnixNano()
	email := fmt.Sprintf("co+%d@test.com", n)
	hash, err := bcrypt.GenerateFromPassword([]byte("pass123"), bcrypt.MinCost)
	require.NoError(t, err)

	_, err = testDB.Exec(ctx, `
		INSERT INTO users (email, password_hash, role, email_verified)
		VALUES ($1, $2, 'company', true)
	`, email, string(hash))
	require.NoError(t, err)

	// Login to get JWT
	loginBody, _ := json.Marshal(map[string]string{"email": email, "password": "pass123"})
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewReader(loginBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var loginResp struct {
		Data struct{ Token string }
	}
	json.NewDecoder(w.Body).Decode(&loginResp)
	token := loginResp.Data.Token
	require.NotEmpty(t, token)

	// Create company profile
	slug := fmt.Sprintf("co-%d", n)
	compBody, _ := json.Marshal(map[string]string{"name": "Test Co", "slug": slug})
	req = httptest.NewRequest(http.MethodPost, "/v1/companies", bytes.NewReader(compBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	return token
}
