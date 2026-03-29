package mocks

import (
	"context"

	"github.com/PPRAMANIK62/devhunt/internal/models"
)

type CompanyRepo struct {
	FindByUserIDFn func(ctx context.Context, userID string) (*models.Company, error)
	CreateFn       func(ctx context.Context, company *models.Company) error
	FindByIDFn     func(ctx context.Context, id string) (*models.Company, error)
	UpdateFn       func(ctx context.Context, userID string, fields map[string]any) (*models.Company, error)
	DeleteFn       func(ctx context.Context, userID string) error
}

func (m *CompanyRepo) FindByUserID(ctx context.Context, userID string) (*models.Company, error) {
	if m.FindByUserIDFn == nil {
		panic("mocks.CompanyRepo.FindByUserIDFn not set")
	}
	return m.FindByUserIDFn(ctx, userID)
}

func (m *CompanyRepo) Create(ctx context.Context, company *models.Company) error {
	if m.CreateFn == nil {
		panic("mocks.CompanyRepo.CreateFn not set")
	}
	return m.CreateFn(ctx, company)
}

func (m *CompanyRepo) FindByID(ctx context.Context, id string) (*models.Company, error) {
	if m.FindByIDFn == nil {
		panic("mocks.CompanyRepo.FindByIDFn not set")
	}
	return m.FindByIDFn(ctx, id)
}

func (m *CompanyRepo) Update(ctx context.Context, userID string, fields map[string]any) (*models.Company, error) {
	if m.UpdateFn == nil {
		panic("mocks.CompanyRepo.UpdateFn not set")
	}
	return m.UpdateFn(ctx, userID, fields)
}

func (m *CompanyRepo) Delete(ctx context.Context, userID string) error {
	if m.DeleteFn == nil {
		panic("mocks.CompanyRepo.DeleteFn not set")
	}
	return m.DeleteFn(ctx, userID)
}
