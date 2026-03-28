package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/PPRAMANIK62/devhunt/internal/apperr"
	"github.com/PPRAMANIK62/devhunt/internal/middleware"
	"github.com/PPRAMANIK62/devhunt/internal/models"
	"github.com/PPRAMANIK62/devhunt/internal/service"
	"github.com/go-chi/chi/v5"
)

type jobServicer interface {
	List(ctx context.Context, page, pageSize int, f service.ListJobsFilter) (*service.ListJobsOutput, error)
	GetByID(ctx context.Context, id string) (*models.Job, error)
	Create(ctx context.Context, userID string, req models.CreateJobRequest) (*models.Job, error)
	Update(ctx context.Context, id, userID string, req models.UpdateJobRequest) (*models.Job, error)
	Delete(ctx context.Context, id, userID string) error
	ListMine(ctx context.Context, userID, status string) ([]*models.Job, error)
	GetFilterOptions(ctx context.Context) (*service.FilterOptions, error)
}

type JobHandler struct {
	jobService jobServicer
}

func NewJobHandler(s *service.JobService) *JobHandler {
	return &JobHandler{jobService: s}
}

func newJobHandlerWithService(s jobServicer) *JobHandler {
	return &JobHandler{jobService: s}
}

func (h *JobHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	pageSize, _ := strconv.Atoi(q.Get("page_size"))

	minSalary, _ := strconv.Atoi(q.Get("min_salary"))
	f := service.ListJobsFilter{
		Search:    q.Get("q"),
		Locations: q["location"],
		Tags:      q["tag"],
		MinSalary: minSalary,
	}

	output, err := h.jobService.List(r.Context(), page, pageSize, f)
	if err != nil {
		writeError(w, err)
		return
	}
	writeSuccess(w, http.StatusOK, output)
}

func (h *JobHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	job, err := h.jobService.GetByID(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeSuccess(w, http.StatusOK, job)
}

func (h *JobHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, apperr.Validation("invalid request body"))
		return
	}
	if err := validate(req); err != nil {
		writeError(w, err)
		return
	}

	job, err := h.jobService.Create(r.Context(), middleware.GetUserID(r.Context()), req)
	if err != nil {
		writeError(w, err)
		return
	}
	writeSuccess(w, http.StatusCreated, job)
}

func (h *JobHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req models.UpdateJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, apperr.Validation("invalid request body"))
		return
	}
	if err := validate(req); err != nil {
		writeError(w, err)
		return
	}

	job, err := h.jobService.Update(r.Context(), chi.URLParam(r, "id"), middleware.GetUserID(r.Context()), req)
	if err != nil {
		writeError(w, err)
		return
	}
	writeSuccess(w, http.StatusOK, job)
}

func (h *JobHandler) ListMine(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	if status != "" && status != "open" && status != "draft" && status != "closed" {
		writeError(w, apperr.Validation("status must be one of: open, draft, closed"))
		return
	}

	jobs, err := h.jobService.ListMine(r.Context(), middleware.GetUserID(r.Context()), status)
	if err != nil {
		writeError(w, err)
		return
	}
	writeSuccess(w, http.StatusOK, jobs)
}

func (h *JobHandler) GetFilterOptions(w http.ResponseWriter, r *http.Request) {
	opts, err := h.jobService.GetFilterOptions(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeSuccess(w, http.StatusOK, opts)
}

func (h *JobHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.jobService.Delete(r.Context(), chi.URLParam(r, "id"), middleware.GetUserID(r.Context())); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent) // 204 — success with no body
}
