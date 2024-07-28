package logger

import (
	"log/slog"
	"os"
)

func NewLogger(env, version string) *slog.Logger {
	var handler slog.Handler

	switch env {
	case "dev":
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	case "prod":
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	default:
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	}

	return slog.New(handler).With("env", env, "version", version)
}
