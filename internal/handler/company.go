package handler

import (
	"encoding/json"
	"net/http"

	"github.com/PPRAMANIK62/devhunt/internal/apperr"
	"github.com/PPRAMANIK62/devhunt/internal/middleware"
	"github.com/PPRAMANIK62/devhunt/internal/models"
	"github.com/PPRAMANIK62/devhunt/internal/service"
	"github.com/go-chi/chi/v5"
)

type CompanyHandler struct {
	companySvc *service.CompanyService
}

func NewCompanyHandler(s *service.CompanyService) *CompanyHandler {
	return &CompanyHandler{companySvc: s}
}

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

func (h *CompanyHandler) GetMine(w http.ResponseWriter, r *http.Request) {
	company, err := h.companySvc.GetMine(r.Context(), middleware.GetUserID(r.Context()))
	if err != nil {
		writeError(w, err)
		return
	}
	writeSuccess(w, http.StatusOK, company)
}

func (h *CompanyHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	company, err := h.companySvc.GetByID(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeSuccess(w, http.StatusOK, company)
}

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

func (h *CompanyHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.companySvc.Delete(r.Context(), middleware.GetUserID(r.Context())); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
