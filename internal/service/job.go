package service

import (
	"context"
	"errors"

	"github.com/PPRAMANIK62/devhunt/internal/apperr"
	"github.com/PPRAMANIK62/devhunt/internal/models"
	"github.com/PPRAMANIK62/devhunt/internal/repository"
)

type JobService struct {
	jobRepo     *repository.JobRepository
	companyRepo *repository.CompanyRepository
}

func NewJobService(jobRepo *repository.JobRepository, companyRepo *repository.CompanyRepository) *JobService {
	return &JobService{
		jobRepo:     jobRepo,
		companyRepo: companyRepo,
	}
}

type ListJobsOutput struct {
	Jobs     []*models.Job `json:"jobs"`
	Total    int           `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"page_size"`
}

type ListJobsFilter struct {
	Search    string
	Locations []string
	Tags      []string
	MinSalary int
}

func (s *JobService) List(ctx context.Context, page, pageSize int, f ListJobsFilter) (*ListJobsOutput, error) {
	jobs, total, err := s.jobRepo.List(ctx, repository.ListFilter{
		Status:    models.JobStatusOpen,
		Page:      page,
		PageSize:  pageSize,
		Search:    f.Search,
		Locations: f.Locations,
		Tags:      f.Tags,
		MinSalary: f.MinSalary,
	})
	if err != nil {
		return nil, err
	}
	return &ListJobsOutput{Jobs: jobs, Total: total, Page: page, PageSize: pageSize}, nil
}

func (s *JobService) GetByID(ctx context.Context, id string) (*models.Job, error) {
	return s.jobRepo.FindByID(ctx, id)
}

func (s *JobService) Create(ctx context.Context, userID string, req models.CreateJobRequest) (*models.Job, error) {
	// Business rule: you need a company profile to post jobs
	company, err := s.companyRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, apperr.Forbidden("you must have a company profile to post jobs")
	}

	job := &models.Job{
		CompanyID:   company.ID,
		Title:       req.Title,
		Description: req.Description,
		Location:    req.Location,
		SalaryMin:   req.SalaryMin,
		SalaryMax:   req.SalaryMax,
		Status:      models.JobStatusOpen,
	}

	if err := s.jobRepo.Create(ctx, job); err != nil {
		return nil, err
	}
	return job, nil
}

func (s *JobService) Update(ctx context.Context, id, userID string, req models.UpdateJobRequest) (*models.Job, error) {
	job, err := s.jobRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Ownership check — only the company that posted it can update it
	company, err := s.companyRepo.FindByUserID(ctx, userID)
	if err != nil || company.ID != job.CompanyID {
		return nil, apperr.Forbidden("you do not own this job posting")
	}

	// Only include fields that were actually provided (non-nil pointers)
	fields := map[string]any{}
	if req.Title != nil {
		fields["title"] = *req.Title
	}
	if req.Description != nil {
		fields["description"] = *req.Description
	}
	if req.Location != nil {
		fields["location"] = *req.Location
	}
	if req.SalaryMin != nil {
		fields["salary_min"] = *req.SalaryMin
	}
	if req.SalaryMax != nil {
		fields["salary_max"] = *req.SalaryMax
	}
	if req.Status != nil {
		fields["status"] = *req.Status
	}

	return s.jobRepo.Update(ctx, id, fields)
}

func (s *JobService) ListMine(ctx context.Context, userID, status string) ([]*models.Job, error) {
	company, err := s.companyRepo.FindByUserID(ctx, userID)
	if err != nil {
		var appErr *apperr.AppError
		if errors.As(err, &appErr) && appErr.Type == apperr.TypeNotFound {
			return nil, apperr.Forbidden("you must have a company profile to list your jobs")
		}
		return nil, err
	}
	return s.jobRepo.ListByCompanyID(ctx, company.ID, status)
}

type FilterOptions struct {
	Locations []string `json:"locations"`
	Tags      []string `json:"tags"`
}

func (s *JobService) GetFilterOptions(ctx context.Context) (*FilterOptions, error) {
	opts, err := s.jobRepo.GetFilterOptions(ctx)
	if err != nil {
		return nil, err
	}
	return &FilterOptions{Locations: opts.Locations, Tags: opts.Tags}, nil
}

func (s *JobService) Delete(ctx context.Context, id, userID string) error {
	job, err := s.jobRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	company, err := s.companyRepo.FindByUserID(ctx, userID)
	if err != nil || company.ID != job.CompanyID {
		return apperr.Forbidden("you do not own this job posting")
	}

	return s.jobRepo.Delete(ctx, id)
}
