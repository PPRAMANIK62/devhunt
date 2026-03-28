package queue

import (
	"fmt"

	"github.com/PPRAMANIK62/devhunt/internal/queue/tasks"
	"github.com/hibiken/asynq"
)

type Client struct {
	client *asynq.Client
}

func NewClient(redisURL string) (*Client, error) {
	opts, err := asynq.ParseRedisURI(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parse redis URL: %w", err)
	}
	return &Client{client: asynq.NewClient(opts)}, nil
}

func (c *Client) Close() error {
	return c.client.Close()
}

func (c *Client) EnqueueApplicationConfirmation(p tasks.ApplicationConfirmationPayload) error {
	task, err := tasks.NewApplicationConfirmationTask(p)
	if err != nil {
		return err
	}
	_, err = c.client.Enqueue(task, asynq.MaxRetry(3))
	return err
}

func (c *Client) EnqueueStatusUpdate(p tasks.StatusUpdatePayload) error {
	task, err := tasks.NewStatusUpdateTask(p)
	if err != nil {
		return err
	}
	_, err = c.client.Enqueue(task, asynq.MaxRetry(3))
	return err
}
