package service

import (
	"context"
	"errors"
	"log/slog"

	"github.com/PPRAMANIK62/devhunt/internal/apperr"
	"github.com/PPRAMANIK62/devhunt/internal/models"
	"github.com/PPRAMANIK62/devhunt/internal/queue"
	"github.com/PPRAMANIK62/devhunt/internal/queue/tasks"
	"github.com/PPRAMANIK62/devhunt/internal/repository"
	"github.com/jackc/pgx/v5/pgconn"
)

type ApplicationService struct {
	appRepo     repository.ApplicationRepo
	jobRepo     repository.JobRepo
	companyRepo repository.CompanyRepo
	userRepo    repository.UserRepo
	queue       *queue.Client // nil - no background jobs
}

func NewApplicationService(
	appRepo repository.ApplicationRepo,
	jobRepo repository.JobRepo,
	companyRepo repository.CompanyRepo,
	userRepo repository.UserRepo,
	q *queue.Client,
) *ApplicationService {
	return &ApplicationService{
		appRepo:     appRepo,
		jobRepo:     jobRepo,
		companyRepo: companyRepo,
		userRepo:    userRepo,
		queue:       q,
	}
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

	// Enqueue the email - don't fail the request if this doesn't work
	if s.queue != nil {
		user, err := s.userRepo.FindByID(ctx, userID)
		if err != nil {
			slog.Error("failed to fetch user for confirmation email", "error", err, "user_id", userID)
		} else {
			company, err := s.companyRepo.FindByID(ctx, job.CompanyID)
			if err != nil {
				slog.Error("failed to fetch company for confirmation email", "error", err, "company_id", job.CompanyID)
			} else {
				if err := s.queue.EnqueueApplicationConfirmation(tasks.ApplicationConfirmationPayload{
					ApplicantEmail: user.Email,
					JobTitle:       job.Title,
					CompanyName:    company.Name,
					ApplicationID:  app.ID,
				}); err != nil {
					// Log and continue - the application was saved, the email can be retried manually
					slog.Error("failed to enqueue confirmation email",
						"error", err,
						"application_id", app.ID,
					)
				}
			}
		}
	}

	return app, nil
}

func (s *ApplicationService) ListMine(ctx context.Context, userID string) ([]*models.Application, error) {
	return s.appRepo.ListByUserID(ctx, userID)
}

func (s *ApplicationService) ListByJobID(ctx context.Context, jobID, userID string) ([]*models.Application, error) {
	job, err := s.jobRepo.FindByID(ctx, jobID)
	if err != nil {
		return nil, err
	}
	company, err := s.companyRepo.FindByUserID(ctx, userID)
	if err != nil || company.ID != job.CompanyID {
		return nil, apperr.Forbidden("you do not own this job posting")
	}
	return s.appRepo.ListByJobID(ctx, jobID)
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

	updated, err := s.appRepo.UpdateStatus(ctx, id, status)
	if err != nil {
		return nil, err
	}

	// Enqueue the status update email - don't fail the request if this doesn't work
	if s.queue != nil {
		user, err := s.userRepo.FindByID(ctx, app.UserID)
		if err != nil {
			slog.Error("failed to fetch user for status update email", "error", err, "user_id", app.UserID)
		} else {
			if err := s.queue.EnqueueStatusUpdate(tasks.StatusUpdatePayload{
				ApplicantEmail: user.Email,
				JobTitle:       job.Title,
				NewStatus:      string(status),
			}); err != nil {
				slog.Error("failed to enqueue status update email",
					"error", err,
					"application_id", id,
				)
			}
		}
	}

	return updated, nil
}
