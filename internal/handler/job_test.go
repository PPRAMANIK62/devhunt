package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/PPRAMANIK62/devhunt/internal/apperr"
	handler "github.com/PPRAMANIK62/devhunt/internal/handler"
	"github.com/PPRAMANIK62/devhunt/internal/models"
	"github.com/PPRAMANIK62/devhunt/internal/service"
	"github.com/go-chi/chi/v5"
)

type stubJobService struct {
	listFn           func(ctx context.Context, page, pageSize int, f service.ListJobsFilter) (*service.ListJobsOutput, error)
	getByIDFn        func(ctx context.Context, id string) (*models.Job, error)
	createFn         func(ctx context.Context, userID string, req models.CreateJobRequest) (*models.Job, error)
	updateFn         func(ctx context.Context, id, userID string, req models.UpdateJobRequest) (*models.Job, error)
	deleteFn         func(ctx context.Context, id, userID string) error
	listMineFn       func(ctx context.Context, userID, status string) ([]*models.Job, error)
	getFilterOptsFn  func(ctx context.Context) (*service.FilterOptions, error)
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
	if s.getFilterOptsFn != nil {
		return s.getFilterOptsFn(ctx)
	}
	return &service.FilterOptions{Locations: []string{}, Tags: []string{}}, nil
}

func TestJobList(t *testing.T) {
	sampleOutput := &service.ListJobsOutput{Jobs: []*models.Job{}, Total: 0, Page: 1, PageSize: 10}

	tests := []struct {
		name       string
		url        string
		listFn     func(ctx context.Context, page, pageSize int, f service.ListJobsFilter) (*service.ListJobsOutput, error)
		wantStatus int
	}{
		{
			name: "success - no filters",
			url:  "/jobs",
			listFn: func(_ context.Context, _, _ int, f service.ListJobsFilter) (*service.ListJobsOutput, error) {
				return sampleOutput, nil
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "search filter passed through",
			url:  "/jobs?q=go+engineer",
			listFn: func(_ context.Context, _, _ int, f service.ListJobsFilter) (*service.ListJobsOutput, error) {
				if f.Search != "go engineer" {
					t.Errorf("want search 'go engineer', got %q", f.Search)
				}
				return sampleOutput, nil
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "single location filter passed through",
			url:  "/jobs?location=Remote",
			listFn: func(_ context.Context, _, _ int, f service.ListJobsFilter) (*service.ListJobsOutput, error) {
				if len(f.Locations) != 1 || f.Locations[0] != "Remote" {
					t.Errorf("want locations [Remote], got %v", f.Locations)
				}
				return sampleOutput, nil
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "multiple locations passed through",
			url:  "/jobs?location=Remote&location=Austin",
			listFn: func(_ context.Context, _, _ int, f service.ListJobsFilter) (*service.ListJobsOutput, error) {
				if len(f.Locations) != 2 {
					t.Errorf("want 2 locations, got %d", len(f.Locations))
				}
				return sampleOutput, nil
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "single tag filter passed through",
			url:  "/jobs?tag=Go",
			listFn: func(_ context.Context, _, _ int, f service.ListJobsFilter) (*service.ListJobsOutput, error) {
				if len(f.Tags) != 1 || f.Tags[0] != "Go" {
					t.Errorf("want tags [Go], got %v", f.Tags)
				}
				return sampleOutput, nil
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "multiple tags passed through",
			url:  "/jobs?tag=Go&tag=Rust",
			listFn: func(_ context.Context, _, _ int, f service.ListJobsFilter) (*service.ListJobsOutput, error) {
				if len(f.Tags) != 2 {
					t.Errorf("want 2 tags, got %d", len(f.Tags))
				}
				return sampleOutput, nil
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "internal error",
			url:  "/jobs",
			listFn: func(_ context.Context, _, _ int, _ service.ListJobsFilter) (*service.ListJobsOutput, error) {
				return nil, apperr.Internal("db error", nil)
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := handler.NewJobHandlerWithService(&stubJobService{listFn: tc.listFn})

			req := httptest.NewRequest(http.MethodGet, tc.url, nil)
			rr := httptest.NewRecorder()
			h.List(rr, req)

			if rr.Code != tc.wantStatus {
				t.Errorf("want status %d, got %d", tc.wantStatus, rr.Code)
			}
		})
	}
}

func TestJobGetByID(t *testing.T) {
	sampleJob := &models.Job{ID: "j1", Title: "Backend Engineer"}

	tests := []struct {
		name       string
		id         string
		getByIDFn  func(ctx context.Context, id string) (*models.Job, error)
		wantStatus int
	}{
		{
			name: "success",
			id:   "j1",
			getByIDFn: func(_ context.Context, _ string) (*models.Job, error) {
				return sampleJob, nil
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "not found",
			id:   "missing",
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

			req := httptest.NewRequest(http.MethodGet, "/"+tc.id, nil)
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			if rr.Code != tc.wantStatus {
				t.Errorf("want status %d, got %d", tc.wantStatus, rr.Code)
			}
		})
	}
}

func TestJobCreate(t *testing.T) {
	sampleJob := &models.Job{ID: "j1", Title: "Backend Engineer"}

	tests := []struct {
		name       string
		body       any
		createFn   func(ctx context.Context, userID string, req models.CreateJobRequest) (*models.Job, error)
		wantStatus int
	}{
		{
			name: "success",
			body: models.CreateJobRequest{
				Title:       "Backend Engineer",
				Description: "We need a backend engineer with Go experience and great communication skills.",
				Location:    "Remote",
				SalaryMin:   80000,
				SalaryMax:   120000,
			},
			createFn: func(_ context.Context, _ string, _ models.CreateJobRequest) (*models.Job, error) {
				return sampleJob, nil
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "forbidden - no company profile",
			body: models.CreateJobRequest{
				Title:       "Backend Engineer",
				Description: "We need a backend engineer with Go experience and great communication skills.",
				Location:    "Remote",
				SalaryMin:   80000,
				SalaryMax:   120000,
			},
			createFn: func(_ context.Context, _ string, _ models.CreateJobRequest) (*models.Job, error) {
				return nil, apperr.Forbidden("you must have a company profile to post jobs")
			},
			wantStatus: http.StatusForbidden,
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
			h := handler.NewJobHandlerWithService(&stubJobService{createFn: tc.createFn})

			var buf bytes.Buffer
			json.NewEncoder(&buf).Encode(tc.body)
			req := httptest.NewRequest(http.MethodPost, "/jobs", &buf)
			req = withUserID(req, "u1")
			rr := httptest.NewRecorder()

			h.Create(rr, req)

			if rr.Code != tc.wantStatus {
				t.Errorf("want status %d, got %d", tc.wantStatus, rr.Code)
			}
		})
	}
}

func TestJobDelete(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		deleteFn   func(ctx context.Context, id, userID string) error
		wantStatus int
	}{
		{
			name: "success",
			id:   "j1",
			deleteFn: func(_ context.Context, _, _ string) error {
				return nil
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name: "not found",
			id:   "missing",
			deleteFn: func(_ context.Context, _, _ string) error {
				return apperr.NotFound("job")
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "forbidden - not owner",
			id:   "j2",
			deleteFn: func(_ context.Context, _, _ string) error {
				return apperr.Forbidden("you do not own this job posting")
			},
			wantStatus: http.StatusForbidden,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := handler.NewJobHandlerWithService(&stubJobService{deleteFn: tc.deleteFn})

			r := chi.NewRouter()
			r.Delete("/{id}", h.Delete)

			req := httptest.NewRequest(http.MethodDelete, "/"+tc.id, nil)
			req = withUserID(req, "u1")
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			if rr.Code != tc.wantStatus {
				t.Errorf("want status %d, got %d", tc.wantStatus, rr.Code)
			}
		})
	}
}

func TestJobListMine(t *testing.T) {
	sampleJobs := []*models.Job{
		{ID: "j1", Title: "Backend Engineer", Status: "open"},
		{ID: "j2", Title: "Frontend Engineer", Status: "draft"},
	}

	tests := []struct {
		name       string
		query      string
		listMineFn func(ctx context.Context, userID, status string) ([]*models.Job, error)
		wantStatus int
	}{
		{
			name:  "all jobs - no filter",
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
				if status != "open" {
					t.Errorf("want status filter 'open', got %q", status)
				}
				return sampleJobs[:1], nil
			},
			wantStatus: http.StatusOK,
		},
		{
			name:  "filter by draft",
			query: "?status=draft",
			listMineFn: func(_ context.Context, _, _ string) ([]*models.Job, error) {
				return sampleJobs[1:], nil
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid status param",
			query:      "?status=unknown",
			listMineFn: nil,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:  "no company profile",
			query: "",
			listMineFn: func(_ context.Context, _, _ string) ([]*models.Job, error) {
				return nil, apperr.Forbidden("you must have a company profile to list your jobs")
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

			if rr.Code != tc.wantStatus {
				t.Errorf("want status %d, got %d", tc.wantStatus, rr.Code)
			}
		})
	}
}

func TestJobGetFilterOptions(t *testing.T) {
	tests := []struct {
		name            string
		getFilterOptsFn func(ctx context.Context) (*service.FilterOptions, error)
		wantStatus      int
	}{
		{
			name: "success",
			getFilterOptsFn: func(_ context.Context) (*service.FilterOptions, error) {
				return &service.FilterOptions{Locations: []string{"Remote", "Austin"}, Tags: []string{"Go", "Rust"}}, nil
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "internal error",
			getFilterOptsFn: func(_ context.Context) (*service.FilterOptions, error) {
				return nil, apperr.Internal("db error", nil)
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := handler.NewJobHandlerWithService(&stubJobService{getFilterOptsFn: tc.getFilterOptsFn})

			req := httptest.NewRequest(http.MethodGet, "/jobs/filters", nil)
			rr := httptest.NewRecorder()
			h.GetFilterOptions(rr, req)

			if rr.Code != tc.wantStatus {
				t.Errorf("want status %d, got %d", tc.wantStatus, rr.Code)
			}
		})
	}
}
