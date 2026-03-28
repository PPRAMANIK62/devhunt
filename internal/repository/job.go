package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/PPRAMANIK62/devhunt/internal/apperr"
	"github.com/PPRAMANIK62/devhunt/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type JobRepository struct {
	db *pgxpool.Pool
}

func NewJobRepository(db *pgxpool.Pool) *JobRepository {
	return &JobRepository{db: db}
}

func (r *JobRepository) Create(ctx context.Context, job *models.Job) error {
	query := `
			INSERT INTO jobs (company_id, title, description, location, salary_min, salary_max, status)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING id, created_at, updated_at
		`
	return r.db.QueryRow(ctx, query,
		job.CompanyID, job.Title, job.Description,
		job.Location, job.SalaryMin, job.SalaryMax, job.Status,
	).Scan(&job.ID, &job.CreatedAt, &job.UpdatedAt)
}

type ListFilter struct {
	Status   models.JobStatus
	Page     int
	PageSize int
}

func (r *JobRepository) List(ctx context.Context, f ListFilter) ([]*models.Job, int, error) {
	if f.PageSize == 0 {
		f.PageSize = 20
	}
	if f.Page < 1 {
		f.Page = 1
	}
	offset := (f.Page - 1) * f.PageSize

	var total int
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM jobs WHERE status = $1`, f.Status).Scan(&total); err != nil {
		return nil, 0, apperr.Internal("count jobs", err)
	}

	rows, err := r.db.Query(ctx, `
		SELECT id, company_id, title, description, location, salary_min, salary_max, status, created_at, updated_at
		FROM jobs WHERE status = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, f.Status, f.PageSize, offset)
	if err != nil {
		return nil, 0, apperr.Internal("list jobs", err)
	}
	defer rows.Close()

	var jobs []*models.Job
	for rows.Next() {
		j := &models.Job{}
		if err := rows.Scan(&j.ID, &j.CompanyID, &j.Title, &j.Description,
			&j.Location, &j.SalaryMin, &j.SalaryMax, &j.Status, &j.CreatedAt, &j.UpdatedAt); err != nil {
			return nil, 0, apperr.Internal("scan job", err)
		}
		jobs = append(jobs, j)
	}

	return jobs, total, nil
}

func (r *JobRepository) FindByID(ctx context.Context, id string) (*models.Job, error) {
	j := &models.Job{}
	err := r.db.
		QueryRow(ctx, `
			SELECT id, company_id, title, description, location, salary_min, salary_max, status, created_at, updated_at
			FROM jobs WHERE id = $1
		`, id).Scan(&j.ID, &j.CompanyID, &j.Title, &j.Description,
		&j.Location, &j.SalaryMin, &j.SalaryMax, &j.Status, &j.CreatedAt, &j.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, apperr.NotFound("job")
	}
	if err != nil {
		return nil, apperr.Internal("find job", err)
	}

	return j, nil
}

func (r *JobRepository) Update(ctx context.Context, id string, fields map[string]any) (*models.Job, error) {
	if len(fields) == 0 {
		return r.FindByID(ctx, id)
	}

	// Build SET claude dynamically from only the fields provided
	clauses := []string{}
	args := []any{}
	i := 1
	for col, val := range fields {
		clauses = append(clauses, fmt.Sprintf("%s = $%d", col, i))
		args = append(args, val)
		i++
	}
	clauses = append(clauses, "updated_at = NOW()")
	args = append(args, id)

	query := fmt.Sprintf(`
		UPDATE jobs SET %s WHERE id = $%d
		RETURNING id, company_id, title, description, location, salary_min, salary_max, status, created_at, updated_at
	`, strings.Join(clauses, ", "), i)

	j := &models.Job{}
	err := r.db.
		QueryRow(ctx, query, args...).
		Scan(&j.ID, &j.CompanyID, &j.Title, &j.Description,
			&j.Location, &j.SalaryMin, &j.SalaryMax, &j.Status, &j.CreatedAt, &j.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, apperr.NotFound("job")
	}
	if err != nil {
		return nil, apperr.Internal("update job", err)
	}

	return j, nil
}

func (r *JobRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.Exec(ctx, `DELETE FROM jobs WHERE id = $1`, id)
	if err != nil {
		return apperr.Internal("delete job", err)
	}
	if result.RowsAffected() == 0 {
		return apperr.NotFound("job")
	}

	return nil
}

func (r *JobRepository) FindByIDs(ctx context.Context, ids []string) ([]*models.Job, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, company_id, title, description, location, salary_min, salary_max, status, created_at, updated_at
		FROM jobs WHERE id = ANY($1)
	`, ids)
	if err != nil {
		return nil, apperr.Internal("find jobs by ids", err)
	}
	defer rows.Close()

	var jobs []*models.Job
	for rows.Next() {
		j := &models.Job{}
		if err := rows.Scan(&j.ID, &j.CompanyID, &j.Title, &j.Description,
			&j.Location, &j.SalaryMin, &j.SalaryMax, &j.Status, &j.CreatedAt, &j.UpdatedAt); err != nil {
			return nil, apperr.Internal("scan job", err)
		}
		jobs = append(jobs, j)
	}
	return jobs, nil
}
