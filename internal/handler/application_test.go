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
	"github.com/PPRAMANIK62/devhunt/internal/models"
	"github.com/go-chi/chi/v5"
)

type stubApplicationService struct {
	applyFn        func(ctx context.Context, jobID, userID string, req models.ApplyRequest) (*models.Application, error)
	listMineFn     func(ctx context.Context, userID string) ([]*models.Application, error)
	updateStatusFn func(ctx context.Context, id, userID string, status models.ApplicationStatus) (*models.Application, error)
}

func (s *stubApplicationService) Apply(ctx context.Context, jobID, userID string, req models.ApplyRequest) (*models.Application, error) {
	return s.applyFn(ctx, jobID, userID, req)
}
func (s *stubApplicationService) ListMine(ctx context.Context, userID string) ([]*models.Application, error) {
	return s.listMineFn(ctx, userID)
}
func (s *stubApplicationService) UpdateStatus(ctx context.Context, id, userID string, status models.ApplicationStatus) (*models.Application, error) {
	return s.updateStatusFn(ctx, id, userID, status)
}

func TestApplicationApply(t *testing.T) {
	sampleApp := &models.Application{ID: "a1", JobID: "j1", UserID: "u1", Status: models.AppStatusPending}

	tests := []struct {
		name       string
		jobID      string
		body       any
		applyFn    func(ctx context.Context, jobID, userID string, req models.ApplyRequest) (*models.Application, error)
		wantStatus int
	}{
		{
			name:  "success",
			jobID: "j1",
			body:  models.ApplyRequest{CoverNote: "I am a great fit."},
			applyFn: func(_ context.Context, _, _ string, _ models.ApplyRequest) (*models.Application, error) {
				return sampleApp, nil
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:  "job not open",
			jobID: "j2",
			body:  models.ApplyRequest{},
			applyFn: func(_ context.Context, _, _ string, _ models.ApplyRequest) (*models.Application, error) {
				return nil, apperr.Validation("this job is no longer accepting applications")
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:  "duplicate application",
			jobID: "j1",
			body:  models.ApplyRequest{},
			applyFn: func(_ context.Context, _, _ string, _ models.ApplyRequest) (*models.Application, error) {
				return nil, apperr.Conflict("you have already applied to this job")
			},
			wantStatus: http.StatusConflict,
		},
		{
			name:       "bad json body",
			jobID:      "j1",
			body:       "not-json{{",
			applyFn:    nil,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := handler.NewApplicationHandlerWithService(&stubApplicationService{applyFn: tc.applyFn})

			r := chi.NewRouter()
			r.Post("/jobs/{jobID}/applications", h.Apply)

			var buf bytes.Buffer
			json.NewEncoder(&buf).Encode(tc.body)
			req := httptest.NewRequest(http.MethodPost, "/jobs/"+tc.jobID+"/applications", &buf)
			req = withUserID(req, "u1")
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			if rr.Code != tc.wantStatus {
				t.Errorf("want status %d, got %d", tc.wantStatus, rr.Code)
			}
		})
	}
}

func TestApplicationListMine(t *testing.T) {
	tests := []struct {
		name       string
		listMineFn func(ctx context.Context, userID string) ([]*models.Application, error)
		wantStatus int
	}{
		{
			name: "success",
			listMineFn: func(_ context.Context, _ string) ([]*models.Application, error) {
				return []*models.Application{{ID: "a1", JobID: "j1", UserID: "u1", Status: models.AppStatusPending}}, nil
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "internal error",
			listMineFn: func(_ context.Context, _ string) ([]*models.Application, error) {
				return nil, apperr.Internal("db error", nil)
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := handler.NewApplicationHandlerWithService(&stubApplicationService{listMineFn: tc.listMineFn})

			req := httptest.NewRequest(http.MethodGet, "/applications", nil)
			req = withUserID(req, "u1")
			rr := httptest.NewRecorder()

			h.ListMine(rr, req)

			if rr.Code != tc.wantStatus {
				t.Errorf("want status %d, got %d", tc.wantStatus, rr.Code)
			}
		})
	}
}

func TestApplicationUpdateStatus(t *testing.T) {
	sampleApp := &models.Application{ID: "a1", JobID: "j1", UserID: "u1", Status: models.AppStatusReviewed}

	tests := []struct {
		name           string
		id             string
		body           any
		updateStatusFn func(ctx context.Context, id, userID string, status models.ApplicationStatus) (*models.Application, error)
		wantStatus     int
	}{
		{
			name: "success",
			id:   "a1",
			body: models.UpdateApplicationStatusRequest{Status: models.AppStatusReviewed},
			updateStatusFn: func(_ context.Context, _, _ string, _ models.ApplicationStatus) (*models.Application, error) {
				return sampleApp, nil
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "forbidden - not company owner",
			id:   "a1",
			body: models.UpdateApplicationStatusRequest{Status: models.AppStatusRejected},
			updateStatusFn: func(_ context.Context, _, _ string, _ models.ApplicationStatus) (*models.Application, error) {
				return nil, apperr.Forbidden("you do not have permission to update this application")
			},
			wantStatus: http.StatusForbidden,
		},
		{
			name: "not found",
			id:   "missing",
			body: models.UpdateApplicationStatusRequest{Status: models.AppStatusAccepted},
			updateStatusFn: func(_ context.Context, _, _ string, _ models.ApplicationStatus) (*models.Application, error) {
				return nil, apperr.NotFound("application")
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:           "bad json body",
			id:             "a1",
			body:           "not-json{{",
			updateStatusFn: nil,
			wantStatus:     http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := handler.NewApplicationHandlerWithService(&stubApplicationService{updateStatusFn: tc.updateStatusFn})

			r := chi.NewRouter()
			r.Patch("/{id}/status", h.UpdateStatus)

			var buf bytes.Buffer
			json.NewEncoder(&buf).Encode(tc.body)
			req := httptest.NewRequest(http.MethodPatch, "/"+tc.id+"/status", &buf)
			req = withUserID(req, "u1")
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			if rr.Code != tc.wantStatus {
				t.Errorf("want status %d, got %d", tc.wantStatus, rr.Code)
			}
		})
	}
}
