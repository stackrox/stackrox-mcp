//go:build integration

package integration

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stackrox/stackrox-mcp/internal/app"
	"github.com/stackrox/stackrox-mcp/internal/config"
	"github.com/stackrox/stackrox-mcp/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupInitializedClient creates an initialized MCP client for testing.
func setupInitializedClient(t *testing.T) *testutil.MCPTestClient {
	t.Helper()

	client, err := createMCPClient(t)
	require.NoError(t, err, "Failed to create MCP client")
	t.Cleanup(func() { client.Close() })

	return client
}

// callToolAndGetResult calls a tool and verifies it succeeds.
func callToolAndGetResult(t *testing.T, client *testutil.MCPTestClient, toolName string, args map[string]any) *mcp.CallToolResult {
	t.Helper()

	ctx := context.Background()
	result, err := client.CallTool(ctx, toolName, args)
	require.NoError(t, err)
	testutil.RequireNoError(t, result)

	return result
}

// getTextContent extracts text from the first content item.
func getTextContent(t *testing.T, result *mcp.CallToolResult) string {
	t.Helper()
	require.NotEmpty(t, result.Content, "should have content in response")

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok, "expected TextContent, got %T", result.Content[0])

	return textContent.Text
}

// TestIntegration_ListTools verifies that all expected tools are registered.
func TestIntegration_ListTools(t *testing.T) {
	client := setupInitializedClient(t)

	ctx := context.Background()
	result, err := client.ListTools(ctx)
	require.NoError(t, err)

	// Verify we have tools registered
	assert.NotEmpty(t, result.Tools, "should have tools registered")

	// Check for specific tools we expect
	toolNames := make([]string, 0, len(result.Tools))
	for _, tool := range result.Tools {
		toolNames = append(toolNames, tool.Name)
	}

	assert.Contains(t, toolNames, "get_deployments_for_cve", "should have get_deployments_for_cve tool")
	assert.Contains(t, toolNames, "list_clusters", "should have list_clusters tool")
}

// TestIntegration_ToolCalls tests successful tool calls using table-driven tests.
func TestIntegration_ToolCalls(t *testing.T) {
	tests := map[string]struct {
		toolName       string
		args           map[string]any
		expectedInText []string // strings that must appear in response
	}{
		"get_deployments_for_cve with Log4Shell": {
			toolName:       "get_deployments_for_cve",
			args:           map[string]any{"cveName": Log4ShellFixture.CVEName},
			expectedInText: Log4ShellFixture.DeploymentNames,
		},
		"get_deployments_for_cve with non-existent CVE": {
			toolName:       "get_deployments_for_cve",
			args:           map[string]any{"cveName": "CVE-9999-99999"},
			expectedInText: []string{`"deployments":[]`},
		},
		"list_clusters": {
			toolName:       "list_clusters",
			args:           map[string]any{},
			expectedInText: AllClustersFixture.ClusterNames,
		},
		"get_clusters_with_orchestrator_cve": {
			toolName:       "get_clusters_with_orchestrator_cve",
			args:           map[string]any{"cveName": "CVE-2099-00001"},
			expectedInText: []string{`"clusters":`},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			client := setupInitializedClient(t)
			result := callToolAndGetResult(t, client, tt.toolName, tt.args)

			responseText := getTextContent(t, result)
			for _, expected := range tt.expectedInText {
				assert.Contains(t, responseText, expected)
			}
		})
	}
}

// TestIntegration_ToolCallErrors tests error handling using table-driven tests.
func TestIntegration_ToolCallErrors(t *testing.T) {
	tests := map[string]struct {
		toolName         string
		args             map[string]any
		expectedErrorMsg string
	}{
		"get_deployments_for_cve missing CVE name": {
			toolName:         "get_deployments_for_cve",
			args:             map[string]any{},
			expectedErrorMsg: "cveName",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			client := setupInitializedClient(t)

			ctx := context.Background()
			_, err := client.CallTool(ctx, tt.toolName, tt.args)

			// Validation errors are returned as protocol errors, not tool errors
			require.Error(t, err, "should receive protocol error for invalid params")
			assert.Contains(t, err.Error(), tt.expectedErrorMsg)
		})
	}
}

// createTestConfig creates a test configuration for the MCP server.
func createTestConfig() *config.Config {
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

// createMCPClient is a helper function that creates an MCP client with the test configuration.
func createMCPClient(t *testing.T) (*testutil.MCPTestClient, error) {
	t.Helper()

	cfg := createTestConfig()

	// Create a run function that wraps app.Run with the config
	runFunc := func(ctx context.Context, stdin io.ReadCloser, stdout io.WriteCloser) error {
		return app.Run(ctx, cfg, stdin, stdout)
	}

	return testutil.NewMCPTestClient(t, runFunc)
}
