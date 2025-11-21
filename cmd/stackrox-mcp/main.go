// Package main for stackrox-mcp command.
package main

import (
	"log/slog"

	"github.com/stackrox/stackrox-mcp/internal/logging"
)

func main() {
	logging.SetupLogging()

	slog.Info("Starting Stackrox MCP server")
}
