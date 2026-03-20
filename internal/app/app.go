// Package app contains the main application logic for the stackrox-mcp server.
// This is separated from the main package to allow tests to run the server in-process.
package app

import (
	"context"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/pkg/errors"
	"github.com/stackrox/stackrox-mcp/internal/client"
	"github.com/stackrox/stackrox-mcp/internal/config"
	"github.com/stackrox/stackrox-mcp/internal/server"
	"github.com/stackrox/stackrox-mcp/internal/toolsets"
	toolsetConfig "github.com/stackrox/stackrox-mcp/internal/toolsets/config"
	toolsetVulnerability "github.com/stackrox/stackrox-mcp/internal/toolsets/vulnerability"
)

// GetToolsets initializes and returns all available toolsets.
func GetToolsets(cfg *config.Config, c *client.Client) []toolsets.Toolset {
	return []toolsets.Toolset{
		toolsetConfig.NewToolset(cfg, c),
		toolsetVulnerability.NewToolset(cfg, c),
	}
}

// Run executes the MCP server with the given configuration and I/O streams.
// This function is extracted from main() to allow tests to run the server in-process.
func Run(ctx context.Context, cfg *config.Config, stdin io.ReadCloser, stdout io.WriteCloser) error {
	// Log full configuration with sensitive data redacted.
	slog.Info("Configuration loaded successfully", "config", cfg.Redacted())

	// Create a cancellable context for the entire server lifecycle
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Set up signal handling for graceful shutdown.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		slog.Info("Received shutdown signal")
		cancel()
	}()

	stackroxClient, err := client.NewClient(&cfg.Central)
	if err != nil {
		return errors.Wrap(err, "failed to create client")
	}

	registry := toolsets.NewRegistry(cfg, GetToolsets(cfg, stackroxClient))
	srv := server.NewServer(cfg, registry)

	err = stackroxClient.Connect(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to connect to central")
	}

	slog.Info("Starting StackRox MCP server")

	return errors.Wrap(srv.Start(ctx, stdin, stdout), "failed to start server")
}
