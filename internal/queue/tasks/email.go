package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/hibiken/asynq"
)

const (
	TypeApplicationConfirmation = "email:application_confirmation"
	TypeStatusUpdate            = "email:status_update"
)

// ApplicationConfirmationPayload is stored in Redis and passed to the worker.
// Must be JSON-serializable - no unexported fields, no channels, no funcs
type ApplicationConfirmationPayload struct {
	ApplicantEmail string `json:"applicant_email"`
	JobTitle       string `json:"job_title"`
	CompanyName    string `json:"company_name"`
	ApplicationID  string `json:"application_id"`
}

func NewApplicationConfirmationTask(p ApplicationConfirmationPayload) (*asynq.Task, error) {
	data, err := json.Marshal(p)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}
	return asynq.NewTask(TypeApplicationConfirmation, data), nil
}

// HandleApplicationConfirmation runs in the worker process
func HandleApplicationConfirmation(ctx context.Context, t *asynq.Task) error {
	var p ApplicationConfirmationPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		// Malperformed payload - don't retry, it'll never succeed
		slog.Error("invalid task payload", "type", t.Type(), "error", err)
		return nil
	}

	// Replace this with a real email provider call (Resend, Mailgun, etc.)
	slog.Info("sending confirmation email",
		"to", p.ApplicantEmail,
		"job", p.JobTitle,
		"company", p.CompanyName,
	)

	// Return an error here to trigger a retry
	return nil
}

type StatusUpdatePayload struct {
	ApplicantEmail string `json:"applicant_email"`
	JobTitle       string `json:"job_title"`
	NewStatus      string `json:"new_status"`
}

func NewStatusUpdateTask(p StatusUpdatePayload) (*asynq.Task, error) {
	data, err := json.Marshal(p)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}
	return asynq.NewTask(TypeStatusUpdate, data), nil
}

func HandleStatusUpdate(ctx context.Context, t *asynq.Task) error {
	var p StatusUpdatePayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		slog.Error("invalid task payload", "type", t.Type(), "error", err)
		return nil
	}

	slog.Info("sending status update email",
		"to", p.ApplicantEmail,
		"status", p.NewStatus,
	)
	return nil
}
