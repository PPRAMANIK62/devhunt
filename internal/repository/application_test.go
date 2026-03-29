package repository_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/PPRAMANIK62/devhunt/internal/apperr"
	"github.com/PPRAMANIK62/devhunt/internal/models"
	"github.com/PPRAMANIK62/devhunt/internal/repository"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplicationRepository_CreateAndFind(t *testing.T) {
	ctx := context.Background()
	appRepo := repository.NewApplicationRepository(testDB)

	companyID := seedCompany(t, ctx)
	userID := seedUser(t, ctx)
	jobID := seedJob(t, ctx, companyID)

	app := &models.Application{
		JobID:     jobID,
		UserID:    userID,
		CoverNote: "I am a great candidate.",
	}

	err := appRepo.Create(ctx, app)
	require.NoError(t, err)
	assert.NotEmpty(t, app.ID)
	assert.NotZero(t, app.AppliedAt)
	assert.Equal(t, models.AppStatusPending, app.Status)

	found, err := appRepo.FindByID(ctx, app.ID)
	require.NoError(t, err)
	assert.Equal(t, app.JobID, found.JobID)
	assert.Equal(t, app.UserID, found.UserID)
	assert.Equal(t, "I am a great candidate.", found.CoverNote)
}

func TestApplicationRepository_Create_Duplicate(t *testing.T) {
	ctx := context.Background()
	appRepo := repository.NewApplicationRepository(testDB)

	companyID := seedCompany(t, ctx)
	userID := seedUser(t, ctx)
	jobID := seedJob(t, ctx, companyID)

	// First application
	err := appRepo.Create(ctx, &models.Application{JobID: jobID, UserID: userID})
	require.NoError(t, err)

	// Second application to same job — should violate unique constraint
	err = appRepo.Create(ctx, &models.Application{JobID: jobID, UserID: userID})
	require.Error(t, err)

	var pgErr *pgconn.PgError
	require.True(t, errors.As(err, &pgErr), "expected pgconn.PgError, got %T", err)
	assert.Equal(t, "23505", pgErr.Code)
}

func TestApplicationRepository_FindByID_NotFound(t *testing.T) {
	appRepo := repository.NewApplicationRepository(testDB)
	_, err := appRepo.FindByID(context.Background(), "00000000-0000-0000-0000-000000000000")
	require.Error(t, err)

	var appErr *apperr.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperr.TypeNotFound, appErr.Type)
}

func TestApplicationRepository_UpdateStatus(t *testing.T) {
	ctx := context.Background()
	appRepo := repository.NewApplicationRepository(testDB)

	companyID := seedCompany(t, ctx)
	userID := seedUser(t, ctx)
	jobID := seedJob(t, ctx, companyID)

	app := &models.Application{JobID: jobID, UserID: userID}
	require.NoError(t, appRepo.Create(ctx, app))
	assert.Equal(t, models.AppStatusPending, app.Status)

	updated, err := appRepo.UpdateStatus(ctx, app.ID, models.AppStatusAccepted)
	require.NoError(t, err)
	assert.Equal(t, models.AppStatusAccepted, updated.Status)

	found, err := appRepo.FindByID(ctx, app.ID)
	require.NoError(t, err)
	assert.Equal(t, models.AppStatusAccepted, found.Status)
}

func TestApplicationRepository_ListByUserID(t *testing.T) {
	ctx := context.Background()
	appRepo := repository.NewApplicationRepository(testDB)

	companyID := seedCompany(t, ctx)
	userID := seedUser(t, ctx)

	// Apply to two jobs
	for i := 0; i < 2; i++ {
		jobID := seedJob(t, ctx, companyID)
		err := appRepo.Create(ctx, &models.Application{JobID: jobID, UserID: userID})
		require.NoError(t, err)
	}

	apps, err := appRepo.ListByUserID(ctx, userID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(apps), 2)
	for _, a := range apps {
		assert.Equal(t, userID, a.UserID)
		assert.NotNil(t, a.Job, "expected Job to be populated")
	}
}

// seedUser inserts a seeker user and returns its ID.
func seedUser(t *testing.T, ctx context.Context) string {
	t.Helper()
	var userID string
	err := testDB.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, role, email_verified)
		VALUES ($1, 'hash', 'seeker', true)
		RETURNING id
	`, fmt.Sprintf("seeker+%d@test.com", time.Now().UnixNano())).Scan(&userID)
	require.NoError(t, err)
	return userID
}

// seedJob inserts a job for the given company and returns its ID.
func seedJob(t *testing.T, ctx context.Context, companyID string) string {
	t.Helper()
	var jobID string
	err := testDB.QueryRow(ctx, `
		INSERT INTO jobs (company_id, title, description, location, salary_min, salary_max, status)
		VALUES ($1, 'Test Job', 'Description for the seed job.', 'Remote', 80000, 120000, 'open')
		RETURNING id
	`, companyID).Scan(&jobID)
	require.NoError(t, err)
	return jobID
}
