package queue

import (
	"fmt"

	"github.com/PPRAMANIK62/devhunt/internal/queue/tasks"
	"github.com/hibiken/asynq"
)

func NewWorkerServer(redisURL string) (*asynq.Server, *asynq.ServeMux, error) {
	opts, err := asynq.ParseRedisURI(redisURL)
	if err != nil {
		return nil, nil, fmt.Errorf("parse redis URL: %w", err)
	}

	srv := asynq.NewServer(opts, asynq.Config{
		Concurrency: 5, // process upto 5 tasks in parallel
	})

	mux := asynq.NewServeMux()
	mux.HandleFunc(tasks.TypeApplicationConfirmation, tasks.HandleApplicationConfirmation)
	mux.HandleFunc(tasks.TypeStatusUpdate, tasks.HandleStatusUpdate)

	return srv, mux, nil
}
