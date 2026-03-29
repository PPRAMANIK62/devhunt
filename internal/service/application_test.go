package service_test

import (
	"context"
	"testing"

	"github.com/PPRAMANIK62/devhunt/internal/apperr"
	"github.com/PPRAMANIK62/devhunt/internal/models"
	"github.com/PPRAMANIK62/devhunt/internal/repository/mocks"
	"github.com/PPRAMANIK62/devhunt/internal/service"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplicationService_Apply_JobClosed(t *testing.T) {
	jobRepo := &mocks.JobRepo{
		FindByIDFn: func(ctx context.Context, id string) (*models.Job, error) {
			return &models.Job{ID: id, Status: models.JobStatusClosed}, nil
		},
	}
	svc := service.NewApplicationService(&mocks.ApplicationRepo{}, jobRepo, &mocks.CompanyRepo{}, &mocks.UserRepo{}, nil)

	_, err := svc.Apply(context.Background(), "job-1", "user-1", models.ApplyRequest{})

	require.Error(t, err)
	var appErr *apperr.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperr.TypeValidation, appErr.Type)
}

func TestApplicationService_Apply_Duplicate(t *testing.T) {
	jobRepo := &mocks.JobRepo{
		FindByIDFn: func(ctx context.Context, id string) (*models.Job, error) {
			return &models.Job{ID: id, Status: models.JobStatusOpen}, nil
		},
	}
	appRepo := &mocks.ApplicationRepo{
		CreateFn: func(ctx context.Context, app *models.Application) error {
			// Simulate PostgreSQL unique constraint violation
			return &pgconn.PgError{Code: "23505"}
		},
	}
	svc := service.NewApplicationService(appRepo, jobRepo, &mocks.CompanyRepo{}, &mocks.UserRepo{}, nil)

	_, err := svc.Apply(context.Background(), "job-1", "user-1", models.ApplyRequest{})

	require.Error(t, err)
	var appErr *apperr.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperr.TypeConflict, appErr.Type)
}

func TestApplicationService_Apply_Success(t *testing.T) {
	jobRepo := &mocks.JobRepo{
		FindByIDFn: func(ctx context.Context, id string) (*models.Job, error) {
			return &models.Job{ID: id, CompanyID: "company-1", Status: models.JobStatusOpen}, nil
		},
	}
	appRepo := &mocks.ApplicationRepo{
		CreateFn: func(ctx context.Context, app *models.Application) error {
			app.ID = "app-new"
			return nil
		},
	}
	svc := service.NewApplicationService(appRepo, jobRepo, &mocks.CompanyRepo{}, &mocks.UserRepo{}, nil)

	app, err := svc.Apply(context.Background(), "job-1", "user-1", models.ApplyRequest{CoverNote: "I'm a great fit."})

	require.NoError(t, err)
	assert.Equal(t, "app-new", app.ID)
	assert.Equal(t, "job-1", app.JobID)
	assert.Equal(t, "user-1", app.UserID)
}

func TestApplicationService_ListByJobID_EnforcesOwnership(t *testing.T) {
	jobRepo := &mocks.JobRepo{
		FindByIDFn: func(ctx context.Context, id string) (*models.Job, error) {
			return &models.Job{ID: id, CompanyID: "company-A"}, nil
		},
	}
	companyRepo := &mocks.CompanyRepo{
		FindByUserIDFn: func(ctx context.Context, userID string) (*models.Company, error) {
			// Requester belongs to company-B, not company-A
			return &models.Company{ID: "company-B"}, nil
		},
	}
	svc := service.NewApplicationService(&mocks.ApplicationRepo{}, jobRepo, companyRepo, &mocks.UserRepo{}, nil)

	_, err := svc.ListByJobID(context.Background(), "job-1", "user-from-B")

	require.Error(t, err)
	var appErr *apperr.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperr.TypeForbidden, appErr.Type)
}

func TestApplicationService_ListByJobID_Success(t *testing.T) {
	expected := []*models.Application{{ID: "app-1"}, {ID: "app-2"}}

	jobRepo := &mocks.JobRepo{
		FindByIDFn: func(ctx context.Context, id string) (*models.Job, error) {
			return &models.Job{ID: id, CompanyID: "company-A"}, nil
		},
	}
	companyRepo := &mocks.CompanyRepo{
		FindByUserIDFn: func(ctx context.Context, userID string) (*models.Company, error) {
			return &models.Company{ID: "company-A"}, nil
		},
	}
	appRepo := &mocks.ApplicationRepo{
		ListByJobIDFn: func(ctx context.Context, jobID string) ([]*models.Application, error) {
			return expected, nil
		},
	}
	svc := service.NewApplicationService(appRepo, jobRepo, companyRepo, &mocks.UserRepo{}, nil)

	apps, err := svc.ListByJobID(context.Background(), "job-1", "owner-user")
	require.NoError(t, err)
	assert.Equal(t, expected, apps)
}

func TestApplicationService_UpdateStatus_EnforcesOwnership(t *testing.T) {
	appRepo := &mocks.ApplicationRepo{
		FindByIDFn: func(ctx context.Context, id string) (*models.Application, error) {
			return &models.Application{ID: id, JobID: "job-1", UserID: "applicant-1"}, nil
		},
	}
	jobRepo := &mocks.JobRepo{
		FindByIDFn: func(ctx context.Context, id string) (*models.Job, error) {
			return &models.Job{ID: id, CompanyID: "company-A"}, nil
		},
	}
	companyRepo := &mocks.CompanyRepo{
		FindByUserIDFn: func(ctx context.Context, userID string) (*models.Company, error) {
			return &models.Company{ID: "company-B"}, nil // wrong company
		},
	}
	svc := service.NewApplicationService(appRepo, jobRepo, companyRepo, &mocks.UserRepo{}, nil)

	_, err := svc.UpdateStatus(context.Background(), "app-1", "user-from-B", models.AppStatusAccepted)

	require.Error(t, err)
	var appErr *apperr.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperr.TypeForbidden, appErr.Type)
}

func TestApplicationService_UpdateStatus_Success(t *testing.T) {
	updated := false

	appRepo := &mocks.ApplicationRepo{
		FindByIDFn: func(ctx context.Context, id string) (*models.Application, error) {
			return &models.Application{ID: id, JobID: "job-1", UserID: "applicant-1"}, nil
		},
		UpdateStatusFn: func(ctx context.Context, id string, status models.ApplicationStatus) (*models.Application, error) {
			updated = true
			return &models.Application{ID: id, Status: status}, nil
		},
	}
	jobRepo := &mocks.JobRepo{
		FindByIDFn: func(ctx context.Context, id string) (*models.Job, error) {
			return &models.Job{ID: id, CompanyID: "company-A"}, nil
		},
	}
	companyRepo := &mocks.CompanyRepo{
		FindByUserIDFn: func(ctx context.Context, userID string) (*models.Company, error) {
			return &models.Company{ID: "company-A"}, nil
		},
	}
	svc := service.NewApplicationService(appRepo, jobRepo, companyRepo, &mocks.UserRepo{}, nil)

	app, err := svc.UpdateStatus(context.Background(), "app-1", "owner-user", models.AppStatusAccepted)

	require.NoError(t, err)
	assert.True(t, updated)
	assert.Equal(t, models.AppStatusAccepted, app.Status)
}

func TestApplicationService_ListMine(t *testing.T) {
	expected := []*models.Application{{ID: "app-1"}, {ID: "app-2"}}

	appRepo := &mocks.ApplicationRepo{
		ListByUserIDFn: func(ctx context.Context, userID string) ([]*models.Application, error) {
			assert.Equal(t, "user-1", userID)
			return expected, nil
		},
	}
	svc := service.NewApplicationService(appRepo, &mocks.JobRepo{}, &mocks.CompanyRepo{}, &mocks.UserRepo{}, nil)

	apps, err := svc.ListMine(context.Background(), "user-1")
	require.NoError(t, err)
	assert.Equal(t, expected, apps)
}
