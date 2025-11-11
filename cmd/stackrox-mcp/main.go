// Package main for stackrox-mcp command.
package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/stackrox/stackrox-mcp/internal/config"
	"github.com/stackrox/stackrox-mcp/internal/logging"
	"github.com/stackrox/stackrox-mcp/internal/server"
	"github.com/stackrox/stackrox-mcp/internal/toolsets"
	toolsetConfig "github.com/stackrox/stackrox-mcp/internal/toolsets/config"
	toolsetVulnerability "github.com/stackrox/stackrox-mcp/internal/toolsets/vulnerability"
)

// getToolsets initializes and returns all available toolsets.
func getToolsets(cfg *config.Config) []toolsets.Toolset {
	return []toolsets.Toolset{
		toolsetConfig.NewToolset(cfg),
		toolsetVulnerability.NewToolset(cfg),
	}
}

func main() {
	logging.SetupLogging()

	configPath := flag.String("config", "", "Path to configuration file (optional)")

	flag.Parse()

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		logging.Fatal("Failed to load configuration", err)
	}

	slog.Info("Configuration loaded successfully", "config", cfg)

	registry := toolsets.NewRegistry(cfg, getToolsets(cfg))
	srv := server.NewServer(cfg, registry)

	// Set up context with signal handling for graceful shutdown.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		slog.Info("Received shutdown signal")
		cancel()
	}()

	slog.Info("Starting Stackrox MCP server")

	if err := srv.Start(ctx); err != nil {
		logging.Fatal("Server error", err)
	}
}
