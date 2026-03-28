package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	resend "github.com/resend/resend-go/v2"

	"github.com/hibiken/asynq"
)

const (
	TypeApplicationConfirmation = "email:application_confirmation"
	TypeStatusUpdate            = "email:status_update"
	TypeEmailVerification       = "email:email_verification"

	fromAddress = "DevHunt <noreply@purbayan.me>"
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

func NewApplicationConfirmationHandler(apiKey string) asynq.HandlerFunc {
	return func(ctx context.Context, t *asynq.Task) error {
		var p ApplicationConfirmationPayload
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			slog.Error("invalid task payload", "type", t.Type(), "error", err)
			return nil // malformed payload — don't retry
		}

		if apiKey == "" {
			slog.Warn("RESEND_API_KEY not set, skipping confirmation email", "to", p.ApplicantEmail)
			return nil
		}

		client := resend.NewClient(apiKey)
		body := fmt.Sprintf(
			"Hi,\n\nWe've received your application for %s at %s.\nApplication ID: %s\n\nYou can track your status at DevHunt.\n\nGood luck!",
			p.JobTitle, p.CompanyName, p.ApplicationID,
		)

		_, err := client.Emails.Send(&resend.SendEmailRequest{
			From:    fromAddress,
			To:      []string{p.ApplicantEmail},
			Subject: fmt.Sprintf("Application received – %s at %s", p.JobTitle, p.CompanyName),
			Text:    body,
		})
		if err != nil {
			// Return the error so Asynq retries
			return fmt.Errorf("send confirmation email: %w", err)
		}

		slog.Info("confirmation email sent", "to", p.ApplicantEmail, "job", p.JobTitle)
		return nil
	}
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

func NewStatusUpdateHandler(apiKey string) asynq.HandlerFunc {
	return func(ctx context.Context, t *asynq.Task) error {
		var p StatusUpdatePayload
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			slog.Error("invalid task payload", "type", t.Type(), "error", err)
			return nil
		}

		if apiKey == "" {
			slog.Warn("RESEND_API_KEY not set, skipping status update email", "to", p.ApplicantEmail)
			return nil
		}

		client := resend.NewClient(apiKey)
		body := fmt.Sprintf(
			"Hi,\n\nYour application for %s has been updated.\nNew status: %s\n\nLog in to DevHunt to view details.",
			p.JobTitle, p.NewStatus,
		)

		_, err := client.Emails.Send(&resend.SendEmailRequest{
			From:    fromAddress,
			To:      []string{p.ApplicantEmail},
			Subject: "Your application status has been updated",
			Text:    body,
		})
		if err != nil {
			return fmt.Errorf("send status update email: %w", err)
		}

		slog.Info("status update email sent", "to", p.ApplicantEmail, "status", p.NewStatus)
		return nil
	}
}

type EmailVerificationPayload struct {
	Email string `json:"email"`
	Token string `json:"token"`
}

func NewEmailVerificationTask(p EmailVerificationPayload) (*asynq.Task, error) {
	data, err := json.Marshal(p)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}
	return asynq.NewTask(TypeEmailVerification, data), nil
}

func NewEmailVerificationHandler(apiKey, appBaseURL string) asynq.HandlerFunc {
	return func(ctx context.Context, t *asynq.Task) error {
		var p EmailVerificationPayload
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			slog.Error("invalid task payload", "type", t.Type(), "error", err)
			return nil
		}

		if apiKey == "" {
			slog.Warn("RESEND_API_KEY not set, skipping verification email", "to", p.Email)
			return nil
		}

		link := fmt.Sprintf("%s/verify-email?token=%s", appBaseURL, p.Token)
		body := fmt.Sprintf(
			"Hi,\n\nPlease verify your email address by clicking the link below:\n\n%s\n\nThis link expires in 24 hours.\n\nIf you didn't create an account, you can ignore this email.",
			link,
		)

		client := resend.NewClient(apiKey)
		_, err := client.Emails.Send(&resend.SendEmailRequest{
			From:    fromAddress,
			To:      []string{p.Email},
			Subject: "Verify your DevHunt email address",
			Text:    body,
		})
		if err != nil {
			return fmt.Errorf("send verification email: %w", err)
		}

		slog.Info("verification email sent", "to", p.Email)
		return nil
	}
}
