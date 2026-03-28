package queue

import (
	"fmt"

	"github.com/PPRAMANIK62/devhunt/internal/queue/tasks"
	"github.com/hibiken/asynq"
)

func NewWorkerServer(redisURL, resendAPIKey, appBaseURL string) (*asynq.Server, *asynq.ServeMux, error) {
	opts, err := asynq.ParseRedisURI(redisURL)
	if err != nil {
		return nil, nil, fmt.Errorf("parse redis URL: %w", err)
	}

	srv := asynq.NewServer(opts, asynq.Config{
		Concurrency: 5,
	})

	mux := asynq.NewServeMux()
	mux.HandleFunc(tasks.TypeApplicationConfirmation, tasks.NewApplicationConfirmationHandler(resendAPIKey))
	mux.HandleFunc(tasks.TypeStatusUpdate, tasks.NewStatusUpdateHandler(resendAPIKey))
	mux.HandleFunc(tasks.TypeEmailVerification, tasks.NewEmailVerificationHandler(resendAPIKey, appBaseURL))

	return srv, mux, nil
}
