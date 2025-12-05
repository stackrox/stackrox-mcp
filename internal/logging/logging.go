// Package logging provides setting up log level for structured logging.
package logging

import (
	"log/slog"
	"os"
	"strings"
)

// SetupLogging configures the global slog logger with JSON output.
// Log level can be configured via the LOG_LEVEL environment variable.
// Supported values: DEBUG, INFO, WARN, ERROR (case-insensitive).
// Default: INFO.
func SetupLogging() {
	// Parse log level from environment variable, default to INFO
	logLevel := slog.LevelInfo

	if levelStr := os.Getenv("LOG_LEVEL"); levelStr != "" {
		switch strings.ToUpper(levelStr) {
		case "DEBUG":
			logLevel = slog.LevelDebug
		case "INFO":
			logLevel = slog.LevelInfo
		case "WARN":
			logLevel = slog.LevelWarn
		case "ERROR":
			logLevel = slog.LevelError
		default:
			slog.Warn("Invalid LOG_LEVEL, defaulting to INFO", "provided", levelStr)
		}
	}

	// Initialize slog with JSON handler.
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: logLevel,
	}))

	slog.SetDefault(logger)
}

// Fatal logs and error and exits with exit code 1.
func Fatal(msg string, err error) {
	slog.Error(msg, "error", err)
	os.Exit(1)
}
