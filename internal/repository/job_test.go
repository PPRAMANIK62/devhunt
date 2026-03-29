package repository_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/PPRAMANIK62/devhunt/internal/apperr"
	"github.com/PPRAMANIK62/devhunt/internal/database"
	"github.com/PPRAMANIK62/devhunt/internal/models"
	"github.com/PPRAMANIK62/devhunt/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testDB *pgxpool.Pool

// TestMain runs once before all tests in this package.
// It starts a Postgres container, runs migrations, runs tests, then cleans up.
func TestMain(m *testing.M) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("could not start dockertest: %v", err)
	}

	resource, err := pool.Run("postgres", "15", []string{
		"POSTGRES_USER=test",
		"POSTGRES_PASSWORD=test",
		"POSTGRES_DB=devhunt_test",
	})
	if err != nil {
		log.Fatalf("could not start postgres container: %v", err)
	}

	dsn := fmt.Sprintf("postgres://test:test@localhost:%s/devhunt_test?sslmode=disable",
		resource.GetPort("5432/tcp"))

	// Retry until Postgres is ready
	if err := pool.Retry(func() error {
		var err error
		testDB, err = database.NewPool(context.Background(), dsn)
		return err
	}); err != nil {
		log.Fatalf("could not connect to postgres: %v", err)
	}

	// Run migrations
	if _, err = testDB.Exec(context.Background(), migrationSQL); err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	code := m.Run()

	// Cleanup
	pool.Purge(resource)
	os.Exit(code)
}

func TestJobRepository_CreateAndFind(t *testing.T) {
	ctx := context.Background()
	repo := repository.NewJobRepository(testDB)

	// You'll need a company in the DB first (foreign key constraint)
	companyID := seedCompany(t, ctx)

	job := &models.Job{
		CompanyID:   companyID,
		Title:       "Go Engineer",
		Description: "Test job description for integration test.",
		Location:    "Remote",
		SalaryMin:   100000,
		SalaryMax:   150000,
		Status:      models.JobStatusOpen,
	}

	err := repo.Create(ctx, job)
	require.NoError(t, err)
	assert.NotEmpty(t, job.ID)
	assert.NotZero(t, job.CreatedAt)

	found, err := repo.FindByID(ctx, job.ID)
	require.NoError(t, err)
	assert.Equal(t, job.Title, found.Title)
	assert.Equal(t, job.CompanyID, found.CompanyID)
}

func TestJobRepository_FindByID_NotFound(t *testing.T) {
	repo := repository.NewJobRepository(testDB)
	_, err := repo.FindByID(context.Background(), "00000000-0000-0000-0000-000000000000")
	require.Error(t, err)

	var appErr *apperr.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperr.TypeNotFound, appErr.Type)
}

func TestJobRepository_List(t *testing.T) {
	ctx := context.Background()
	repo := repository.NewJobRepository(testDB)

	companyID := seedCompany(t, ctx)

	for i := 0; i < 2; i++ {
		err := repo.Create(ctx, &models.Job{
			CompanyID:   companyID,
			Title:       fmt.Sprintf("Engineer %d", i),
			Description: "Test job description for the list integration test.",
			Location:    "Remote",
			SalaryMin:   90000,
			SalaryMax:   120000,
			Status:      models.JobStatusOpen,
		})
		require.NoError(t, err)
	}

	jobs, total, err := repo.List(ctx, repository.ListFilter{
		Status:   models.JobStatusOpen,
		Page:     1,
		PageSize: 10,
	})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, total, 2)
	assert.GreaterOrEqual(t, len(jobs), 2)
}

func TestJobRepository_Update(t *testing.T) {
	ctx := context.Background()
	repo := repository.NewJobRepository(testDB)

	companyID := seedCompany(t, ctx)
	job := &models.Job{
		CompanyID:   companyID,
		Title:       "Before Update",
		Description: "Original description for the update integration test.",
		Location:    "Remote",
		SalaryMin:   80000,
		SalaryMax:   100000,
		Status:      models.JobStatusOpen,
	}
	require.NoError(t, repo.Create(ctx, job))

	updated, err := repo.Update(ctx, job.ID, map[string]any{"title": "After Update"})
	require.NoError(t, err)
	assert.Equal(t, "After Update", updated.Title)

	found, err := repo.FindByID(ctx, job.ID)
	require.NoError(t, err)
	assert.Equal(t, "After Update", found.Title)
}

func TestJobRepository_Delete(t *testing.T) {
	ctx := context.Background()
	repo := repository.NewJobRepository(testDB)

	companyID := seedCompany(t, ctx)
	job := &models.Job{
		CompanyID:   companyID,
		Title:       "To Be Deleted",
		Description: "This job will be deleted in the integration test.",
		Location:    "Remote",
		SalaryMin:   70000,
		SalaryMax:   90000,
		Status:      models.JobStatusOpen,
	}
	require.NoError(t, repo.Create(ctx, job))

	err := repo.Delete(ctx, job.ID)
	require.NoError(t, err)

	_, err = repo.FindByID(ctx, job.ID)
	require.Error(t, err)
	var appErr *apperr.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperr.TypeNotFound, appErr.Type)
}

// migrationSQL contains both migration files to set up the test schema.
var migrationSQL = `
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TYPE user_role AS ENUM ('seeker', 'company', 'admin');
CREATE TYPE job_status AS ENUM ('open', 'closed', 'draft');
CREATE TYPE application_status AS ENUM ('pending', 'reviewed', 'rejected', 'accepted');

CREATE TABLE users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email         TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role          user_role NOT NULL DEFAULT 'seeker',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE companies (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    slug        TEXT NOT NULL UNIQUE,
    description TEXT,
    website     TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE jobs (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id  UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    title       TEXT NOT NULL,
    description TEXT NOT NULL,
    location    TEXT NOT NULL,
    salary_min  INTEGER NOT NULL CHECK (salary_min >= 0),
    salary_max  INTEGER NOT NULL CHECK (salary_max >= salary_min),
    status      job_status NOT NULL DEFAULT 'open',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE tags (
    id   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE job_tags (
    job_id UUID REFERENCES jobs(id) ON DELETE CASCADE,
    tag_id UUID REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (job_id, tag_id)
);

CREATE TABLE applications (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_id     UUID NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status     application_status NOT NULL DEFAULT 'pending',
    cover_note TEXT,
    applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (job_id, user_id)
);

CREATE INDEX idx_jobs_company_id ON jobs(company_id);
CREATE INDEX idx_jobs_status     ON jobs(status);
CREATE INDEX idx_jobs_created_at ON jobs(created_at DESC);
CREATE INDEX idx_applications_job_id  ON applications(job_id);
CREATE INDEX idx_applications_user_id ON applications(user_id);

ALTER TABLE users
  ADD COLUMN email_verified                BOOLEAN     NOT NULL DEFAULT false,
  ADD COLUMN verification_token            TEXT        UNIQUE,
  ADD COLUMN verification_token_expires_at TIMESTAMPTZ;
`

func seedCompany(t *testing.T, ctx context.Context) string {
	t.Helper()

	var userID string
	err := testDB.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, role, email_verified)
		VALUES ($1, 'hash', 'company', true)
		RETURNING id
	`, fmt.Sprintf("seed+%d@test.com", time.Now().UnixNano())).Scan(&userID)
	require.NoError(t, err)

	var companyID string
	err = testDB.QueryRow(ctx, `
		INSERT INTO companies (user_id, name, slug)
		VALUES ($1, 'Test Co', $2)
		RETURNING id
	`, userID, fmt.Sprintf("test-co-%d", time.Now().UnixNano())).Scan(&companyID)
	require.NoError(t, err)

	return companyID
}
