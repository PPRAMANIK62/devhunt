package service

import (
	"context"
	"errors"

	"github.com/PPRAMANIK62/devhunt/internal/apperr"
	"github.com/PPRAMANIK62/devhunt/internal/models"
	"github.com/PPRAMANIK62/devhunt/internal/repository"
	"github.com/jackc/pgx/v5/pgconn"
)

type ApplicationService struct {
	appRepo     *repository.ApplicationRepository
	jobRepo     *repository.JobRepository
	companyRepo *repository.CompanyRepository
}

func NewApplicationService(
	appRepo *repository.ApplicationRepository,
	jobRepo *repository.JobRepository,
	companyRepo *repository.CompanyRepository,
) *ApplicationService {
	return &ApplicationService{appRepo: appRepo, jobRepo: jobRepo, companyRepo: companyRepo}
}

func (s *ApplicationService) Apply(ctx context.Context, jobID, userID string, req models.ApplyRequest) (*models.Application, error) {
	// Make sure the job exists and is open
	job, err := s.jobRepo.FindByID(ctx, jobID)
	if err != nil {
		return nil, err
	}
	if job.Status != models.JobStatusOpen {
		return nil, apperr.Validation("this job is no longer accepting applications")
	}

	app := &models.Application{
		JobID:     jobID,
		UserID:    userID,
		CoverNote: req.CoverNote,
	}

	if err := s.appRepo.Create(ctx, app); err != nil {
		// Detect unique constraint violation (already applied)
		// pgx wraps Postgres errors as *pgconn.PgError
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, apperr.Conflict("you have already applied to this job")
		}
		return nil, apperr.Internal("create application", err)
	}

	return app, nil
}

func (s *ApplicationService) ListMine(ctx context.Context, userID string) ([]*models.Application, error) {
	return s.appRepo.ListByUserID(ctx, userID)
}

func (s *ApplicationService) UpdateStatus(ctx context.Context, id, userID string, status models.ApplicationStatus) (*models.Application, error) {
	app, err := s.appRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Only the company that owns the job can update status
	job, err := s.jobRepo.FindByID(ctx, app.JobID)
	if err != nil {
		return nil, err
	}

	company, err := s.companyRepo.FindByUserID(ctx, userID)
	if err != nil || company.ID != job.CompanyID {
		return nil, apperr.Forbidden("you do not have permission to update this application")
	}

	return s.appRepo.UpdateStatus(ctx, id, status)
}
