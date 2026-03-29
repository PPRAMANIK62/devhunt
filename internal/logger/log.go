package logger

import (
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
)

func SetupLogger(env string) {
	level := slog.LevelInfo
	if env == "development" {
		level = slog.LevelDebug
	}

	var h slog.Handler
	if env == "development" {
		h = tint.NewHandler(os.Stdout, &tint.Options{
			Level:      level,
			TimeFormat: time.TimeOnly,
		})
	} else {
		// JSON in production - parsable by Datadog, Loki, CloudWatch, etc.
		h = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	}

	slog.SetDefault(slog.New(h))
}
