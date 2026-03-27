package models

import "time"

type ApplicationStatus string

const (
	AppStatusPending  ApplicationStatus = "pending"
	AppStatusReviewed ApplicationStatus = "reviewed"
	AppStatusRejected ApplicationStatus = "rejected"
	AppStatusAccepted ApplicationStatus = "accepted"
)

type Application struct {
	ID        string            `json:"id"`
	JobID     string            `json:"job_id"`
	UserID    string            `json:"user_id"`
	Status    ApplicationStatus `json:"status"`
	CoverNote string            `json:"cover_note,omitempty"`
	AppliedAt time.Time         `json:"applied_at"`
	UpdatedAt time.Time         `json:"updated_at"`
	Job       *Job              `json:"job,omitempty"`
	User      *User             `json:"user,omitempty"`
}

type ApplyRequest struct {
	CoverNote string `json:"cover_note" validate:"omitempty,max=2000"`
}

type UpdateApplicationStatusRequest struct {
	Status ApplicationStatus `json:"status" validate:"required,oneof=pending reviewed rejected accepted"`
}
