package logging

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetupLogging(t *testing.T) {
	// Clear any existing LOG_LEVEL environment variable
	require.NoError(t, os.Unsetenv("LOG_LEVEL"))

	SetupLogging()

	// Verify default log level is INFO
	handler := slog.Default().Handler()
	assert.True(t, handler.Enabled(context.Background(), slog.LevelInfo))
	assert.False(t, handler.Enabled(context.Background(), slog.LevelDebug))
}

func TestSetupLoggingWithEnvVar(t *testing.T) {
	tests := map[string]struct {
		logLevel      string
		expectedLevel slog.Level
	}{
		"DEBUG level": {
			logLevel:      "DEBUG",
			expectedLevel: slog.LevelDebug,
		},
		"INFO level": {
			logLevel:      "INFO",
			expectedLevel: slog.LevelInfo,
		},
		"WARN level": {
			logLevel:      "WARN",
			expectedLevel: slog.LevelWarn,
		},
		"ERROR level": {
			logLevel:      "ERROR",
			expectedLevel: slog.LevelError,
		},
		"Invalid level defaults to INFO": {
			logLevel:      "INVALID",
			expectedLevel: slog.LevelInfo,
		},
	}
	for testName, testCase := range tests {
		t.Run(testName, func(innerT *testing.T) {
			innerT.Setenv("LOG_LEVEL", testCase.logLevel)

			SetupLogging()

			handler := slog.Default().Handler()
			assert.True(innerT, handler.Enabled(context.Background(), testCase.expectedLevel))
		})
	}
}
