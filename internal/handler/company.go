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

type companyServicer interface {
	Create(ctx context.Context, userID string, req models.CreateCompanyRequest) (*models.Company, error)
	GetMine(ctx context.Context, userID string) (*models.Company, error)
	GetByID(ctx context.Context, id string) (*models.Company, error)
	Update(ctx context.Context, userID string, req models.UpdateCompanyRequest) (*models.Company, error)
	Delete(ctx context.Context, userID string) error
}

type CompanyHandler struct {
	companySvc companyServicer
}

func NewCompanyHandler(s *service.CompanyService) *CompanyHandler {
	return &CompanyHandler{companySvc: s}
}

func newCompanyHandlerWithService(s companyServicer) *CompanyHandler {
	return &CompanyHandler{companySvc: s}
}

// Create godoc
// @Summary      Create a company profile
// @Tags         companies
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body  models.CreateCompanyRequest  true  "Company details"
// @Success      201  {object}  map[string]any
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      409  {object}  map[string]string
// @Router       /companies [post]
func (h *CompanyHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateCompanyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, apperr.Validation("invalid request body"))
		return
	}
	if err := validate(req); err != nil {
		writeError(w, err)
		return
	}

	company, err := h.companySvc.Create(r.Context(), middleware.GetUserID(r.Context()), req)
	if err != nil {
		writeError(w, err)
		return
	}
	writeSuccess(w, http.StatusCreated, company)
}

// GetMine godoc
// @Summary      Get the authenticated company's profile
// @Tags         companies
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]any
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /companies/me [get]
func (h *CompanyHandler) GetMine(w http.ResponseWriter, r *http.Request) {
	company, err := h.companySvc.GetMine(r.Context(), middleware.GetUserID(r.Context()))
	if err != nil {
		writeError(w, err)
		return
	}
	writeSuccess(w, http.StatusOK, company)
}

// GetByID godoc
// @Summary      Get a company by ID
// @Tags         companies
// @Produce      json
// @Param        id  path  string  true  "Company ID"
// @Success      200  {object}  map[string]any
// @Failure      404  {object}  map[string]string
// @Router       /companies/{id} [get]
func (h *CompanyHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	company, err := h.companySvc.GetByID(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeSuccess(w, http.StatusOK, company)
}

// Update godoc
// @Summary      Update the authenticated company's profile
// @Tags         companies
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body  models.UpdateCompanyRequest  true  "Fields to update"
// @Success      200  {object}  map[string]any
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /companies/me [patch]
func (h *CompanyHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req models.UpdateCompanyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, apperr.Validation("invalid request body"))
		return
	}
	if err := validate(req); err != nil {
		writeError(w, err)
		return
	}

	company, err := h.companySvc.Update(r.Context(), middleware.GetUserID(r.Context()), req)
	if err != nil {
		writeError(w, err)
		return
	}
	writeSuccess(w, http.StatusOK, company)
}

// Delete godoc
// @Summary      Delete the authenticated company's profile
// @Tags         companies
// @Security     BearerAuth
// @Success      204
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /companies/me [delete]
func (h *CompanyHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.companySvc.Delete(r.Context(), middleware.GetUserID(r.Context())); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
