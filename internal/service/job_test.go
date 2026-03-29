package service_test

import (
	"context"
	"testing"

	"github.com/PPRAMANIK62/devhunt/internal/apperr"
	"github.com/PPRAMANIK62/devhunt/internal/models"
	"github.com/PPRAMANIK62/devhunt/internal/repository/mocks"
	"github.com/PPRAMANIK62/devhunt/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJobService_Create_RequiresCompanyProfile(t *testing.T) {
	companyRepo := &mocks.CompanyRepo{
		FindByUserIDFn: func(ctx context.Context, userID string) (*models.Company, error) {
			return nil, apperr.NotFound("company")
		},
	}
	jobRepo := &mocks.JobRepo{}

	svc := service.NewJobService(jobRepo, companyRepo, nil)

	_, err := svc.Create(context.Background(), "user-123", models.CreateJobRequest{
		Title:       "Go Engineer",
		Description: "We need a Go engineer with experience in distributed systems.",
		Location:    "Remote",
		SalaryMin:   100000,
		SalaryMax:   150000,
	})

	require.Error(t, err)
	var appErr *apperr.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperr.TypeForbidden, appErr.Type)
}

func TestJobService_Create_Success(t *testing.T) {
	companyRepo := &mocks.CompanyRepo{
		FindByUserIDFn: func(ctx context.Context, userID string) (*models.Company, error) {
			return &models.Company{ID: "company-1"}, nil
		},
	}
	jobRepo := &mocks.JobRepo{
		CreateFn: func(ctx context.Context, job *models.Job) error {
			job.ID = "job-new"
			return nil
		},
	}

	svc := service.NewJobService(jobRepo, companyRepo, nil)

	job, err := svc.Create(context.Background(), "user-1", models.CreateJobRequest{
		Title:       "Go Engineer",
		Description: "We need a Go engineer with experience in distributed systems.",
		Location:    "Remote",
		SalaryMin:   100000,
		SalaryMax:   150000,
	})

	require.NoError(t, err)
	assert.Equal(t, "job-new", job.ID)
	assert.Equal(t, "company-1", job.CompanyID)
}

func TestJobService_Delete_EnforcesOwnership(t *testing.T) {
	jobRepo := &mocks.JobRepo{
		FindByIDFn: func(ctx context.Context, id string) (*models.Job, error) {
			return &models.Job{ID: id, CompanyID: "company-A"}, nil
		},
		DeleteFn: func(ctx context.Context, id string) error {
			return nil
		},
	}
	companyRepo := &mocks.CompanyRepo{
		FindByUserIDFn: func(ctx context.Context, userID string) (*models.Company, error) {
			// Returns company-B, not company-A — different owner
			return &models.Company{ID: "company-B"}, nil
		},
	}

	svc := service.NewJobService(jobRepo, companyRepo, nil)

	err := svc.Delete(context.Background(), "job-1", "user-from-company-B")

	require.Error(t, err)
	var appErr *apperr.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperr.TypeForbidden, appErr.Type)
}

func TestJobService_Delete_Success(t *testing.T) {
	deleted := false

	jobRepo := &mocks.JobRepo{
		FindByIDFn: func(ctx context.Context, id string) (*models.Job, error) {
			return &models.Job{ID: id, CompanyID: "company-A"}, nil
		},
		DeleteFn: func(ctx context.Context, id string) error {
			deleted = true
			return nil
		},
	}
	companyRepo := &mocks.CompanyRepo{
		FindByUserIDFn: func(ctx context.Context, userID string) (*models.Company, error) {
			return &models.Company{ID: "company-A"}, nil
		},
	}

	svc := service.NewJobService(jobRepo, companyRepo, nil)
	err := svc.Delete(context.Background(), "job-1", "owner-user")

	require.NoError(t, err)
	assert.True(t, deleted)
}

func TestJobService_Update_EnforcesOwnership(t *testing.T) {
	jobRepo := &mocks.JobRepo{
		FindByIDFn: func(ctx context.Context, id string) (*models.Job, error) {
			return &models.Job{ID: id, CompanyID: "company-A"}, nil
		},
	}
	companyRepo := &mocks.CompanyRepo{
		FindByUserIDFn: func(ctx context.Context, userID string) (*models.Company, error) {
			return &models.Company{ID: "company-B"}, nil
		},
	}

	svc := service.NewJobService(jobRepo, companyRepo, nil)

	title := "Updated Title"
	_, err := svc.Update(context.Background(), "job-1", "user-from-company-B", models.UpdateJobRequest{
		Title: &title,
	})

	require.Error(t, err)
	var appErr *apperr.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperr.TypeForbidden, appErr.Type)
}

func TestJobService_ListMine_RequiresCompanyProfile(t *testing.T) {
	companyRepo := &mocks.CompanyRepo{
		FindByUserIDFn: func(ctx context.Context, userID string) (*models.Company, error) {
			return nil, apperr.NotFound("company")
		},
	}
	jobRepo := &mocks.JobRepo{}

	svc := service.NewJobService(jobRepo, companyRepo, nil)

	_, err := svc.ListMine(context.Background(), "user-no-company", "")

	require.Error(t, err)
	var appErr *apperr.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperr.TypeForbidden, appErr.Type)
}

func TestJobService_ListMine_Success(t *testing.T) {
	expected := []*models.Job{
		{ID: "job-1", CompanyID: "company-A"},
		{ID: "job-2", CompanyID: "company-A"},
	}

	companyRepo := &mocks.CompanyRepo{
		FindByUserIDFn: func(ctx context.Context, userID string) (*models.Company, error) {
			return &models.Company{ID: "company-A"}, nil
		},
	}
	jobRepo := &mocks.JobRepo{
		ListByCompanyIDFn: func(ctx context.Context, companyID, status string) ([]*models.Job, error) {
			assert.Equal(t, "company-A", companyID)
			return expected, nil
		},
	}

	svc := service.NewJobService(jobRepo, companyRepo, nil)

	jobs, err := svc.ListMine(context.Background(), "owner-user", "")
	require.NoError(t, err)
	assert.Equal(t, expected, jobs)
}

func TestJobService_GetByID_NotFound(t *testing.T) {
	jobRepo := &mocks.JobRepo{
		FindByIDFn: func(ctx context.Context, id string) (*models.Job, error) {
			return nil, apperr.NotFound("job")
		},
	}
	companyRepo := &mocks.CompanyRepo{}

	svc := service.NewJobService(jobRepo, companyRepo, nil)

	_, err := svc.GetByID(context.Background(), "missing-id")

	require.Error(t, err)
	var appErr *apperr.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperr.TypeNotFound, appErr.Type)
}
