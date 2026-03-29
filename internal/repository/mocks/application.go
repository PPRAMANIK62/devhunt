package mocks

import (
	"context"

	"github.com/PPRAMANIK62/devhunt/internal/models"
)

type ApplicationRepo struct {
	CreateFn         func(ctx context.Context, app *models.Application) error
	FindByIDFn       func(ctx context.Context, id string) (*models.Application, error)
	ListByUserIDFn   func(ctx context.Context, userID string) ([]*models.Application, error)
	ListByJobIDFn    func(ctx context.Context, jobID string) ([]*models.Application, error)
	UpdateStatusFn   func(ctx context.Context, id string, status models.ApplicationStatus) (*models.Application, error)
}

func (m *ApplicationRepo) Create(ctx context.Context, app *models.Application) error {
	if m.CreateFn == nil {
		panic("mocks.ApplicationRepo.CreateFn not set")
	}
	return m.CreateFn(ctx, app)
}

func (m *ApplicationRepo) FindByID(ctx context.Context, id string) (*models.Application, error) {
	if m.FindByIDFn == nil {
		panic("mocks.ApplicationRepo.FindByIDFn not set")
	}
	return m.FindByIDFn(ctx, id)
}

func (m *ApplicationRepo) ListByUserID(ctx context.Context, userID string) ([]*models.Application, error) {
	if m.ListByUserIDFn == nil {
		panic("mocks.ApplicationRepo.ListByUserIDFn not set")
	}
	return m.ListByUserIDFn(ctx, userID)
}

func (m *ApplicationRepo) ListByJobID(ctx context.Context, jobID string) ([]*models.Application, error) {
	if m.ListByJobIDFn == nil {
		panic("mocks.ApplicationRepo.ListByJobIDFn not set")
	}
	return m.ListByJobIDFn(ctx, jobID)
}

func (m *ApplicationRepo) UpdateStatus(ctx context.Context, id string, status models.ApplicationStatus) (*models.Application, error) {
	if m.UpdateStatusFn == nil {
		panic("mocks.ApplicationRepo.UpdateStatusFn not set")
	}
	return m.UpdateStatusFn(ctx, id, status)
}
