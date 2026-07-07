package logger

import (
	"log/slog"
	"os"
)

// New creates a new structured logger.
// For "production" env, use JSON handler. For anything else, use Text handler.
// Parse logLevel string to slog.Level.
func New(env, logLevel string) *slog.Logger {
	level := parseLevel(logLevel)

	opts := &slog.HandlerOptions{
		Level: level,
	}

	var handler slog.Handler
	if env == "production" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}

func parseLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
