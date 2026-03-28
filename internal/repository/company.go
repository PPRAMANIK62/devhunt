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

type CompanyRepository struct {
	db *pgxpool.Pool
}

func NewCompanyRepository(db *pgxpool.Pool) *CompanyRepository {
	return &CompanyRepository{db: db}
}

func (r *CompanyRepository) FindByUserID(ctx context.Context, userID string) (*models.Company, error) {
	query := `SELECT id, user_id, name, slug, description, website, created_at, updated_at FROM companies WHERE user_id = $1`

	c := &models.Company{}
	err := r.db.QueryRow(ctx, query, userID).
		Scan(
			&c.ID,
			&c.UserID,
			&c.Name,
			&c.Slug,
			&c.Description,
			&c.Website,
			&c.CreatedAt,
			&c.UpdatedAt,
		)
	if err == pgx.ErrNoRows {
		return nil, apperr.NotFound("company")
	}
	if err != nil {
		return nil, apperr.Internal("find company", err)
	}

	return c, nil
}

func (r *CompanyRepository) Create(ctx context.Context, company *models.Company) error {
	query := `
		INSERT INTO companies (user_id, name, slug, description, website)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`
	return r.db.
		QueryRow(
			ctx,
			query,
			company.UserID,
			company.Name,
			company.Slug,
			company.Description,
			company.Website,
		).Scan(&company.ID, &company.CreatedAt, &company.UpdatedAt)
}

func (r *CompanyRepository) FindByID(ctx context.Context, id string) (*models.Company, error) {
	query := `SELECT id, user_id, name, slug, description, website, created_at, updated_at FROM companies WHERE id = $1`

	c := &models.Company{}
	err := r.db.QueryRow(ctx, query, id).
		Scan(&c.ID, &c.UserID, &c.Name, &c.Slug, &c.Description, &c.Website, &c.CreatedAt, &c.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, apperr.NotFound("company")
	}
	if err != nil {
		return nil, apperr.Internal("find company by id", err)
	}
	return c, nil
}

func (r *CompanyRepository) Update(ctx context.Context, userID string, fields map[string]any) (*models.Company, error) {
	if len(fields) == 0 {
		return r.FindByUserID(ctx, userID)
	}

	setClauses := make([]string, 0, len(fields))
	args := make([]any, 0, len(fields)+1)
	i := 1
	for col, val := range fields {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", col, i))
		args = append(args, val)
		i++
	}
	args = append(args, userID)

	query := fmt.Sprintf(
		`UPDATE companies SET %s, updated_at = now() WHERE user_id = $%d
		 RETURNING id, user_id, name, slug, description, website, created_at, updated_at`,
		strings.Join(setClauses, ", "), i,
	)

	c := &models.Company{}
	err := r.db.QueryRow(ctx, query, args...).
		Scan(&c.ID, &c.UserID, &c.Name, &c.Slug, &c.Description, &c.Website, &c.CreatedAt, &c.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, apperr.NotFound("company")
	}
	if err != nil {
		return nil, apperr.Internal("update company", err)
	}
	return c, nil
}

func (r *CompanyRepository) Delete(ctx context.Context, userID string) error {
	query := `DELETE FROM companies WHERE user_id = $1`
	result, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return apperr.Internal("delete company", err)
	}
	if result.RowsAffected() == 0 {
		return apperr.NotFound("company")
	}
	return nil
}
