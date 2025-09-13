package logger

import (
	"log/slog"
	"os"
	"strings"
)

type Logger struct {
	*slog.Logger
}

func New(level string) *Logger {
	var logLevel slog.Level

	switch strings.ToLower(level) {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn", "warning":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	logger := slog.New(handler)

	return &Logger{Logger: logger}
}

func (l *Logger) WithService(serviceName string) *Logger {
	return &Logger{
		Logger: l.Logger.With("service", serviceName),
	}
}

func (l *Logger) WithRequestID(requestID string) *Logger {
	return &Logger{
		Logger: l.Logger.With("request_id", requestID),
	}
}