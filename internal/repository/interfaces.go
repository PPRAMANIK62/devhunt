package repository

import (
	"context"
	"time"

	"github.com/PPRAMANIK62/devhunt/internal/models"
)

type JobRepo interface {
	Create(ctx context.Context, job *models.Job) error
	FindByID(ctx context.Context, id string) (*models.Job, error)
	List(ctx context.Context, f ListFilter) ([]*models.Job, int, error)
	Update(ctx context.Context, id string, fields map[string]any) (*models.Job, error)
	Delete(ctx context.Context, id string) error
	FindByIDs(ctx context.Context, ids []string) ([]*models.Job, error)
	ListByCompanyID(ctx context.Context, companyID, status string) ([]*models.Job, error)
	GetFilterOptions(ctx context.Context) (*FilterOptions, error)
}

type CompanyRepo interface {
	FindByUserID(ctx context.Context, userID string) (*models.Company, error)
	Create(ctx context.Context, company *models.Company) error
	FindByID(ctx context.Context, id string) (*models.Company, error)
	Update(ctx context.Context, userID string, fields map[string]any) (*models.Company, error)
	Delete(ctx context.Context, userID string) error
}

type UserRepo interface {
	Create(ctx context.Context, user *models.User) error
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	FindByID(ctx context.Context, id string) (*models.User, error)
	FindByVerificationToken(ctx context.Context, token string) (*models.User, error)
	SetVerificationToken(ctx context.Context, userID, token string, expiresAt time.Time) error
	SetVerified(ctx context.Context, userID string) error
}

type ApplicationRepo interface {
	Create(ctx context.Context, app *models.Application) error
	FindByID(ctx context.Context, id string) (*models.Application, error)
	ListByUserID(ctx context.Context, userID string) ([]*models.Application, error)
	ListByJobID(ctx context.Context, jobID string) ([]*models.Application, error)
	UpdateStatus(ctx context.Context, id string, status models.ApplicationStatus) (*models.Application, error)
}
