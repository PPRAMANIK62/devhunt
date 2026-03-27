package models

import "time"

type JobStatus string

const (
	JobStatusOpen   JobStatus = "open"
	JobStatusClosed JobStatus = "closed"
	JobStatusDraft  JobStatus = "draft"
)

type Job struct {
	ID          string    `json:"id"`
	CompanyID   string    `json:"company_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Location    string    `json:"location"`
	SalaryMin   int       `json:"salary_min"`
	SalaryMax   int       `json:"salary_max"`
	Status      JobStatus `json:"status"`
	Tags        []string  `json:"tags,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Company     *Company  `json:"company,omitempty"`
}

type CreateJobRequest struct {
	Title       string   `json:"title"       validate:"required,min=3,max=200"`
	Description string   `json:"description" validate:"required,min=50"`
	Location    string   `json:"location"    validate:"required"`
	SalaryMin   int      `json:"salary_min"  validate:"required,min=0"`
	SalaryMax   int      `json:"salary_max"  validate:"required,gtefield=SalaryMin"`
	Tags        []string `json:"tags"        validate:"omitempty,max=10"`
}

// `UpdateJobRequest` uses pointers (`*string`, `*int`). This lets you
// distinguish "field was not sent" (nil) from "field was sent as empty string".
// Crucial for PATCH endpoints.
type UpdateJobRequest struct {
	Title       *string    `json:"title"       validate:"omitempty,min=3,max=200"`
	Description *string    `json:"description" validate:"omitempty,min=50"`
	Location    *string    `json:"location"`
	SalaryMin   *int       `json:"salary_min"  validate:"omitempty,min=0"`
	SalaryMax   *int       `json:"salary_max"`
	Status      *JobStatus `json:"status"      validate:"omitempty,oneof=open closed draft"`
}
