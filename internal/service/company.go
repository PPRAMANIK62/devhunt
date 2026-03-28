package service

import (
	"context"

	"github.com/PPRAMANIK62/devhunt/internal/apperr"
	"github.com/PPRAMANIK62/devhunt/internal/models"
	"github.com/PPRAMANIK62/devhunt/internal/repository"
)

type CompanyService struct {
	repo *repository.CompanyRepository
}

func NewCompanyService(repo *repository.CompanyRepository) *CompanyService {
	return &CompanyService{repo: repo}
}

func (s *CompanyService) Create(ctx context.Context, userID string, req models.CreateCompanyRequest) (*models.Company, error) {
	// Enforce one-company-per-user
	existing, _ := s.repo.FindByUserID(ctx, userID)
	if existing != nil {
		return nil, apperr.Conflict("company profile already exists")
	}

	company := &models.Company{
		UserID:      userID,
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		Website:     req.Website,
	}
	if err := s.repo.Create(ctx, company); err != nil {
		return nil, err
	}
	return company, nil
}

func (s *CompanyService) GetMine(ctx context.Context, userID string) (*models.Company, error) {
	return s.repo.FindByUserID(ctx, userID)
}

func (s *CompanyService) GetByID(ctx context.Context, id string) (*models.Company, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *CompanyService) Update(ctx context.Context, userID string, req models.UpdateCompanyRequest) (*models.Company, error) {
	fields := map[string]any{}
	if req.Name != nil {
		fields["name"] = *req.Name
	}
	if req.Slug != nil {
		fields["slug"] = *req.Slug
	}
	if req.Description != nil {
		fields["description"] = *req.Description
	}
	if req.Website != nil {
		fields["website"] = *req.Website
	}
	return s.repo.Update(ctx, userID, fields)
}

func (s *CompanyService) Delete(ctx context.Context, userID string) error {
	return s.repo.Delete(ctx, userID)
}
