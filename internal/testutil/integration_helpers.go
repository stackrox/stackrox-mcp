//go:build integration

package testutil

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stackrox/stackrox-mcp/internal/app"
	"github.com/stackrox/stackrox-mcp/internal/config"
	"github.com/stretchr/testify/require"
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

// SetupInitializedClient creates an initialized MCP client for testing with automatic cleanup.
func SetupInitializedClient(t *testing.T, createClient func(*testing.T) (*MCPTestClient, error)) *MCPTestClient {
	t.Helper()

	client, err := createClient(t)
	require.NoError(t, err, "Failed to create MCP client")
	t.Cleanup(func() { client.Close() })

	return client
}

// CallToolAndGetResult calls a tool and verifies it succeeds.
func CallToolAndGetResult(t *testing.T, client *MCPTestClient, toolName string, args map[string]any) *mcp.CallToolResult {
	t.Helper()

	ctx := context.Background()
	result, err := client.CallTool(ctx, toolName, args)
	require.NoError(t, err)
	RequireNoError(t, result)

	return result
}

// GetTextContent extracts text from the first content item.
func GetTextContent(t *testing.T, result *mcp.CallToolResult) string {
	t.Helper()
	require.NotEmpty(t, result.Content, "should have content in response")

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok, "expected TextContent, got %T", result.Content[0])

	return textContent.Text
}
