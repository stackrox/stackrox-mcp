//go:build integration

package testutil

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/stackrox/stackrox-mcp/internal/app"
	"github.com/stackrox/stackrox-mcp/internal/config"
)

// CreateIntegrationTestConfig creates a test configuration for integration tests.
// This connects to a local WireMock instance at localhost:8081.
func CreateIntegrationTestConfig() *config.Config {
	return &config.Config{
		Central: config.CentralConfig{
			URL:                   "localhost:8081",
			AuthType:              "static",
			APIToken:              "test-token-admin",
			InsecureSkipTLSVerify: true,
			RequestTimeout:        30 * time.Second,
			MaxRetries:            3,
			InitialBackoff:        time.Second,
			MaxBackoff:            10 * time.Second,
		},
		Server: config.ServerConfig{
			Type: config.ServerTypeStdio,
		},
		Tools: config.ToolsConfig{
			Vulnerability: config.ToolsetVulnerabilityConfig{
				Enabled: true,
			},
			ConfigManager: config.ToolConfigManagerConfig{
				Enabled: true,
			},
		},
	}
}

// CreateIntegrationMCPClient creates an MCP client with integration test configuration.
func CreateIntegrationMCPClient(t *testing.T) (*MCPTestClient, error) {
	t.Helper()

	cfg := CreateIntegrationTestConfig()

	// Create a run function that wraps app.Run with the config
	runFunc := func(ctx context.Context, stdin io.ReadCloser, stdout io.WriteCloser) error {
		return app.Run(ctx, cfg, stdin, stdout)
	}

	return NewMCPTestClient(t, runFunc)
}
