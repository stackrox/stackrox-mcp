package config

import (
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stackrox/stackrox-mcp/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewListClustersTool(t *testing.T) {
	tool := NewListClustersTool(&client.Client{})

	require.NotNil(t, tool)
	assert.Equal(t, "list_clusters", tool.GetName())
}

func TestListClustersTool_IsReadOnly(t *testing.T) {
	tool := NewListClustersTool(&client.Client{})

	assert.True(t, tool.IsReadOnly(), "list_clusters should be read-only")
}

func TestListClustersTool_GetTool(t *testing.T) {
	tool := NewListClustersTool(&client.Client{})

	mcpTool := tool.GetTool()

	require.NotNil(t, mcpTool)
	assert.Equal(t, "list_clusters", mcpTool.Name)
	assert.NotEmpty(t, mcpTool.Description)
}

func TestListClustersTool_RegisterWith(t *testing.T) {
	tool := NewListClustersTool(&client.Client{})
	server := mcp.NewServer(
		&mcp.Implementation{
			Name:    "test-server",
			Version: "1.0.0",
		},
		&mcp.ServerOptions{},
	)

	// Should not panic
	assert.NotPanics(t, func() {
		tool.RegisterWith(server)
	})
}
