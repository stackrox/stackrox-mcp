// Package main for stackrox-mcp command.
package main

import (
	"flag"
	"log/slog"
	"os"

	"github.com/stackrox/stackrox-mcp/internal/config"
	"github.com/stackrox/stackrox-mcp/internal/logging"
)

func main() {
	logging.SetupLogging()

	configPath := flag.String("config", "", "Path to configuration file (optional)")

	flag.Parse()

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		slog.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	slog.Info("Configuration loaded successfully", "config", cfg)

	slog.Info("Starting Stackrox MCP server")
}
