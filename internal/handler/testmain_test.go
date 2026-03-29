package handler_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/PPRAMANIK62/devhunt/internal/database"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ory/dockertest/v3"
)

var testDB *pgxpool.Pool

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

	if err := pool.Retry(func() error {
		var err error
		testDB, err = database.NewPool(context.Background(), dsn)
		return err
	}); err != nil {
		log.Fatalf("could not connect to postgres: %v", err)
	}

	if _, err = testDB.Exec(context.Background(), handlerTestMigrationSQL); err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	code := m.Run()
	pool.Purge(resource)
	os.Exit(code)
}

// handlerTestMigrationSQL contains both migration files for handler integration tests.
var handlerTestMigrationSQL = `
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
