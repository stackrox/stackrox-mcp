package testutil

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

// SetupInitializedClient creates an initialized MCP client with automatic cleanup.
func SetupInitializedClient(t *testing.T, createClient func(*testing.T) (*MCPTestClient, error)) *MCPTestClient {
	t.Helper()

	client, err := createClient(t)
	require.NoError(t, err, "Failed to create MCP client")
	t.Cleanup(func() { _ = client.Close() })

	return client
}

// CallToolAndGetResult calls a tool and verifies it succeeds.
func CallToolAndGetResult(
	t *testing.T,
	client *MCPTestClient,
	toolName string,
	args map[string]any,
) *mcp.CallToolResult {
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
