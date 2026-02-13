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

	"github.com/stackrox/stackrox-mcp/internal/client"
	"github.com/stackrox/stackrox-mcp/internal/config"
	"github.com/stackrox/stackrox-mcp/internal/server"
	"github.com/stackrox/stackrox-mcp/internal/toolsets"
	toolsetConfig "github.com/stackrox/stackrox-mcp/internal/toolsets/config"
	toolsetVulnerability "github.com/stackrox/stackrox-mcp/internal/toolsets/vulnerability"
)

// getToolsets initializes and returns all available toolsets.
func getToolsets(cfg *config.Config, c *client.Client) []toolsets.Toolset {
	return []toolsets.Toolset{
		toolsetConfig.NewToolset(cfg, c),
		toolsetVulnerability.NewToolset(cfg, c),
	}
}

// Run executes the MCP server with the given configuration and I/O streams.
// If stdin/stdout are nil, os.Stdin/os.Stdout will be used.
// This function is extracted from main() to allow tests to run the server in-process.
func Run(ctx context.Context, cfg *config.Config, stdin io.ReadCloser, stdout io.WriteCloser) error {
	// Log full configuration with sensitive data redacted.
	slog.Info("Configuration loaded successfully", "config", cfg.Redacted())

	stackroxClient, err := client.NewClient(&cfg.Central)
	if err != nil {
		return err
	}

	registry := toolsets.NewRegistry(cfg, getToolsets(cfg, stackroxClient))
	srv := server.NewServer(cfg, registry)

	err = stackroxClient.Connect(ctx)
	if err != nil {
		return err
	}

	// Set up signal handling for graceful shutdown.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Create a cancellable context from the input context
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		<-sigChan
		slog.Info("Received shutdown signal")
		cancel()
	}()

	slog.Info("Starting StackRox MCP server")

	return srv.Start(ctx, stdin, stdout)
}
