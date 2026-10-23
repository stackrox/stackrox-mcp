package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetupLogging(t *testing.T) {
	// Clear any existing LOG_LEVEL environment variable.
	require.NoError(t, os.Unsetenv("LOG_LEVEL"))

	SetupLogging()

	// Verify default log level is INFO.
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

func TestSetupLogging_InvalidLevelRemainsWithinOneRecord(t *testing.T) {
	tests := map[string]struct {
		logLevel string
		json     bool
	}{
		"text invalid level": {logLevel: "INVALID"},
		"text CR/LF input":   {logLevel: "INVALID\r\nforged log entry"},
		"JSON invalid level": {logLevel: "INVALID", json: true},
		"JSON CR/LF input":   {logLevel: "INVALID\r\nforged log entry", json: true},
	}
	for testName, testCase := range tests {
		t.Run(testName, func(innerT *testing.T) {
			originalLogger := slog.Default()

			innerT.Cleanup(func() { slog.SetDefault(originalLogger) })

			var output bytes.Buffer

			var handler slog.Handler = slog.NewTextHandler(&output, nil)

			if testCase.json {
				handler = slog.NewJSONHandler(&output, nil)
			}

			slog.SetDefault(slog.New(handler))
			innerT.Setenv("LOG_LEVEL", testCase.logLevel)

			SetupLogging()

			assert.Contains(innerT, output.String(), "Invalid LOG_LEVEL, defaulting to INFO")
			assert.Equal(innerT, 1, strings.Count(output.String(), "\n"))
			assert.NotContains(innerT, output.String(), "\r")

			if testCase.json {
				var record map[string]any
				require.NoError(innerT, json.Unmarshal(output.Bytes(), &record))
				assert.Equal(innerT, testCase.logLevel, record["provided"])
			} else {
				expectedValue := testCase.logLevel
				if strings.ContainsAny(expectedValue, "\r\n") {
					expectedValue = strconv.Quote(expectedValue)
				}

				assert.Contains(innerT, output.String(), "provided="+expectedValue)
			}

			configuredHandler := slog.Default().Handler()
			assert.True(innerT, configuredHandler.Enabled(context.Background(), slog.LevelInfo))
			assert.False(innerT, configuredHandler.Enabled(context.Background(), slog.LevelDebug))
		})
	}
}
