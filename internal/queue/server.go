package queue

import (
	"fmt"
	"log/slog"

	"github.com/PPRAMANIK62/devhunt/internal/queue/tasks"
	"github.com/hibiken/asynq"
)

// slogAdapter bridges asynq's Logger interface to slog.
type slogAdapter struct{}

func (slogAdapter) Debug(args ...any) { slog.Debug(fmt.Sprint(args...)) }
func (slogAdapter) Info(args ...any)  { slog.Info(fmt.Sprint(args...)) }
func (slogAdapter) Warn(args ...any)  { slog.Warn(fmt.Sprint(args...)) }
func (slogAdapter) Error(args ...any) { slog.Error(fmt.Sprint(args...)) }
func (slogAdapter) Fatal(args ...any) { slog.Error(fmt.Sprint(args...)) }

func NewWorkerServer(redisURL, resendAPIKey, appBaseURL string) (*asynq.Server, *asynq.ServeMux, error) {
	opts, err := asynq.ParseRedisURI(redisURL)
	if err != nil {
		return nil, nil, fmt.Errorf("parse redis URL: %w", err)
	}

	srv := asynq.NewServer(opts, asynq.Config{
		Concurrency: 5,
		Logger:      slogAdapter{},
		LogLevel:    asynq.WarnLevel,
	})

	mux := asynq.NewServeMux()
	mux.HandleFunc(tasks.TypeApplicationConfirmation, tasks.NewApplicationConfirmationHandler(resendAPIKey))
	mux.HandleFunc(tasks.TypeStatusUpdate, tasks.NewStatusUpdateHandler(resendAPIKey))
	mux.HandleFunc(tasks.TypeEmailVerification, tasks.NewEmailVerificationHandler(resendAPIKey, appBaseURL))

	return srv, mux, nil
}
