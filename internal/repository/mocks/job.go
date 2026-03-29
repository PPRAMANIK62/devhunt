package mocks

import (
	"context"

	"github.com/PPRAMANIK62/devhunt/internal/models"
	"github.com/PPRAMANIK62/devhunt/internal/repository"
)

type JobRepo struct {
	CreateFn           func(ctx context.Context, job *models.Job) error
	FindByIDFn         func(ctx context.Context, id string) (*models.Job, error)
	ListFn             func(ctx context.Context, f repository.ListFilter) ([]*models.Job, int, error)
	UpdateFn           func(ctx context.Context, id string, fields map[string]any) (*models.Job, error)
	DeleteFn           func(ctx context.Context, id string) error
	FindByIDsFn        func(ctx context.Context, ids []string) ([]*models.Job, error)
	ListByCompanyIDFn  func(ctx context.Context, companyID, status string) ([]*models.Job, error)
	GetFilterOptionsFn func(ctx context.Context) (*repository.FilterOptions, error)
}

func (m *JobRepo) Create(ctx context.Context, job *models.Job) error {
	if m.CreateFn == nil {
		panic("mocks.JobRepo.CreateFn not set")
	}
	return m.CreateFn(ctx, job)
}

func (m *JobRepo) FindByID(ctx context.Context, id string) (*models.Job, error) {
	if m.FindByIDFn == nil {
		panic("mocks.JobRepo.FindByIDFn not set")
	}
	return m.FindByIDFn(ctx, id)
}

func (m *JobRepo) List(ctx context.Context, f repository.ListFilter) ([]*models.Job, int, error) {
	if m.ListFn == nil {
		panic("mocks.JobRepo.ListFn not set")
	}
	return m.ListFn(ctx, f)
}

func (m *JobRepo) Update(ctx context.Context, id string, fields map[string]any) (*models.Job, error) {
	if m.UpdateFn == nil {
		panic("mocks.JobRepo.UpdateFn not set")
	}
	return m.UpdateFn(ctx, id, fields)
}

func (m *JobRepo) Delete(ctx context.Context, id string) error {
	if m.DeleteFn == nil {
		panic("mocks.JobRepo.DeleteFn not set")
	}
	return m.DeleteFn(ctx, id)
}

func (m *JobRepo) FindByIDs(ctx context.Context, ids []string) ([]*models.Job, error) {
	if m.FindByIDsFn == nil {
		panic("mocks.JobRepo.FindByIDsFn not set")
	}
	return m.FindByIDsFn(ctx, ids)
}

func (m *JobRepo) ListByCompanyID(ctx context.Context, companyID, status string) ([]*models.Job, error) {
	if m.ListByCompanyIDFn == nil {
		panic("mocks.JobRepo.ListByCompanyIDFn not set")
	}
	return m.ListByCompanyIDFn(ctx, companyID, status)
}

func (m *JobRepo) GetFilterOptions(ctx context.Context) (*repository.FilterOptions, error) {
	if m.GetFilterOptionsFn == nil {
		panic("mocks.JobRepo.GetFilterOptionsFn not set")
	}
	return m.GetFilterOptionsFn(ctx)
}
