package logger

import (
	"log/slog"
	"os"
)

func SetupLogger(env string) {
	level := slog.LevelInfo
	if env == "development" {
		level = slog.LevelDebug
	}

	opts := &slog.HandlerOptions{Level: level}

	var h slog.Handler
	if env == "development" {
		// Human readable text in development
		h = slog.NewTextHandler(os.Stdout, opts)
	} else {
		// JSON in production - parsable by Datadog, Loki, CloudWatch, etc.
		h = slog.NewJSONHandler(os.Stdout, opts)
	}

	slog.SetDefault(slog.New(h))
}
