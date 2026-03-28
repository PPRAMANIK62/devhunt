package repository

import (
	"context"

	"github.com/PPRAMANIK62/devhunt/internal/apperr"
	"github.com/PPRAMANIK62/devhunt/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ApplicationRepository struct {
	db *pgxpool.Pool
}

func NewApplicationRepository(db *pgxpool.Pool) *ApplicationRepository {
	return &ApplicationRepository{db: db}
}

func (r *ApplicationRepository) Create(ctx context.Context, app *models.Application) error {
	query := `
		INSERT INTO applications (job_id, user_id, cover_note)
		VALUES ($1, $2, $3)
		RETURNING id, status, applied_at, updated_at
	`
	return r.db.QueryRow(ctx, query, app.JobID, app.UserID, app.CoverNote).
		Scan(&app.ID, &app.Status, &app.AppliedAt, &app.UpdatedAt)
}

// ListByUserID returns all applications for a seeker, with job title and company name.
func (r *ApplicationRepository) ListByUserID(ctx context.Context, userID string) ([]*models.Application, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			a.id, a.job_id, a.user_id, a.status, a.cover_note, a.applied_at, a.updated_at,
			j.title, j.status,
			c.id, c.name
		FROM applications a
		JOIN jobs j ON j.id = a.job_id
		JOIN companies c ON c.id = j.company_id
		WHERE a.user_id = $1
		ORDER BY a.applied_at DESC
	`, userID)
	if err != nil {
		return nil, apperr.Internal("list applications", err)
	}
	defer rows.Close()

	var apps []*models.Application
	for rows.Next() {
		a := &models.Application{}
		j := &models.Job{}
		co := &models.Company{}
		if err := rows.Scan(
			&a.ID, &a.JobID, &a.UserID, &a.Status, &a.CoverNote, &a.AppliedAt, &a.UpdatedAt,
			&j.Title, &j.Status,
			&co.ID, &co.Name,
		); err != nil {
			return nil, apperr.Internal("scan application", err)
		}
		j.Company = co
		a.Job = j
		apps = append(apps, a)
	}
	return apps, nil
}

// ListByJobID returns all applications for a job (company view), with applicant email.
func (r *ApplicationRepository) ListByJobID(ctx context.Context, jobID string) ([]*models.Application, error) {
	rows, err := r.db.Query(ctx, `
		SELECT a.id, a.job_id, a.user_id, a.status, a.cover_note, a.applied_at, a.updated_at,
		       u.id, u.email
		FROM applications a
		JOIN users u ON u.id = a.user_id
		WHERE a.job_id = $1
		ORDER BY a.applied_at DESC
	`, jobID)
	if err != nil {
		return nil, apperr.Internal("list applications", err)
	}
	defer rows.Close()

	var apps []*models.Application
	for rows.Next() {
		a := &models.Application{}
		u := &models.User{}
		if err := rows.Scan(&a.ID, &a.JobID, &a.UserID, &a.Status, &a.CoverNote, &a.AppliedAt, &a.UpdatedAt,
			&u.ID, &u.Email); err != nil {
			return nil, apperr.Internal("scan application", err)
		}
		a.User = u
		apps = append(apps, a)
	}
	return apps, nil
}

func (r *ApplicationRepository) FindByID(ctx context.Context, id string) (*models.Application, error) {
	a := &models.Application{}
	err := r.db.QueryRow(ctx, `
		SELECT id, job_id, user_id, status, cover_note, applied_at, updated_at
		FROM applications WHERE id = $1
	`, id).Scan(&a.ID, &a.JobID, &a.UserID, &a.Status, &a.CoverNote, &a.AppliedAt, &a.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, apperr.NotFound("application")
	}
	if err != nil {
		return nil, apperr.Internal("find application", err)
	}
	return a, nil
}

func (r *ApplicationRepository) UpdateStatus(ctx context.Context, id string, status models.ApplicationStatus) (*models.Application, error) {
	a := &models.Application{}
	err := r.db.QueryRow(ctx, `
		UPDATE applications SET status = $1, updated_at = NOW()
		WHERE id = $2
		RETURNING id, job_id, user_id, status, cover_note, applied_at, updated_at
	`, status, id).Scan(&a.ID, &a.JobID, &a.UserID, &a.Status, &a.CoverNote, &a.AppliedAt, &a.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, apperr.NotFound("application")
	}
	if err != nil {
		return nil, apperr.Internal("update application status", err)
	}
	return a, nil
}
