package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/PPRAMANIK62/devhunt/internal/apperr"
	"github.com/PPRAMANIK62/devhunt/internal/middleware"
	"github.com/PPRAMANIK62/devhunt/internal/models"
	"github.com/PPRAMANIK62/devhunt/internal/service"
	"github.com/go-chi/chi/v5"
)

type applicationServicer interface {
	Apply(ctx context.Context, jobID, userID string, req models.ApplyRequest) (*models.Application, error)
	ListMine(ctx context.Context, userID string) ([]*models.Application, error)
	ListByJobID(ctx context.Context, jobID, userID string) ([]*models.Application, error)
	UpdateStatus(ctx context.Context, id, userID string, status models.ApplicationStatus) (*models.Application, error)
}

type ApplicationHandler struct {
	appService applicationServicer
}

func NewApplicationHandler(s *service.ApplicationService) *ApplicationHandler {
	return &ApplicationHandler{appService: s}
}

func newApplicationHandlerWithService(s applicationServicer) *ApplicationHandler {
	return &ApplicationHandler{appService: s}
}

// Apply godoc
// @Summary      Apply to a job
// @Tags         applications
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        jobID    path  string              true  "Job ID"
// @Param        request  body  models.ApplyRequest true  "Application details"
// @Success      201  {object}  map[string]any
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      409  {object}  map[string]string
// @Router       /jobs/{jobID}/applications [post]
func (h *ApplicationHandler) Apply(w http.ResponseWriter, r *http.Request) {
	var req models.ApplyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, apperr.Validation("invalid request body"))
		return
	}
	if err := validate(req); err != nil {
		writeError(w, err)
		return
	}

	jobID := chi.URLParam(r, "jobID")
	userID := middleware.GetUserID(r.Context())

	app, err := h.appService.Apply(r.Context(), jobID, userID, req)
	if err != nil {
		writeError(w, err)
		return
	}
	writeSuccess(w, http.StatusCreated, app)
}

// ListByJobID godoc
// @Summary      List applications for a job (company only)
// @Tags         applications
// @Produce      json
// @Security     BearerAuth
// @Param        jobID  path  string  true  "Job ID"
// @Success      200  {object}  map[string]any
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /jobs/{jobID}/applications [get]
func (h *ApplicationHandler) ListByJobID(w http.ResponseWriter, r *http.Request) {
	apps, err := h.appService.ListByJobID(r.Context(), chi.URLParam(r, "jobID"), middleware.GetUserID(r.Context()))
	if err != nil {
		writeError(w, err)
		return
	}
	writeSuccess(w, http.StatusOK, apps)
}

// ListMine godoc
// @Summary      List the authenticated user's applications
// @Tags         applications
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]any
// @Failure      401  {object}  map[string]string
// @Router       /applications [get]
func (h *ApplicationHandler) ListMine(w http.ResponseWriter, r *http.Request) {
	apps, err := h.appService.ListMine(r.Context(), middleware.GetUserID(r.Context()))
	if err != nil {
		writeError(w, err)
		return
	}
	writeSuccess(w, http.StatusOK, apps)
}

// UpdateStatus godoc
// @Summary      Update the status of an application (company only)
// @Tags         applications
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path  string                                  true  "Application ID"
// @Param        request  body  models.UpdateApplicationStatusRequest  true  "New status"
// @Success      200  {object}  map[string]any
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /applications/{id}/status [patch]
func (h *ApplicationHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	var req models.UpdateApplicationStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, apperr.Validation("invalid request body"))
		return
	}
	if err := validate(req); err != nil {
		writeError(w, err)
		return
	}

	app, err := h.appService.UpdateStatus(
		r.Context(),
		chi.URLParam(r, "id"),
		middleware.GetUserID(r.Context()),
		req.Status,
	)
	if err != nil {
		writeError(w, err)
		return
	}
	writeSuccess(w, http.StatusOK, app)
}
