package config

import (
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewListClusterPrompt(t *testing.T) {
	t.Parallel()

	prompt := NewListClusterPrompt()

	require.NotNil(t, prompt)
	assert.Equal(t, "list-cluster", prompt.GetName())
}

func TestListClusterPrompt_GetPrompt(t *testing.T) {
	t.Parallel()

	prompt := NewListClusterPrompt()

	mcpPrompt := prompt.GetPrompt()

	require.NotNil(t, mcpPrompt)
	assert.Equal(t, "list-cluster", mcpPrompt.Name)
	assert.NotEmpty(t, mcpPrompt.Description)
	assert.Contains(t, mcpPrompt.Description, "Kubernetes")
	assert.Contains(t, mcpPrompt.Description, "StackRox")
	assert.Nil(t, mcpPrompt.Arguments)
}

func TestListClusterPrompt_GetMessages(t *testing.T) {
	t.Parallel()

	prompt := NewListClusterPrompt()

	messages, err := prompt.GetMessages(nil)

	require.NoError(t, err)
	require.Len(t, messages, 1)

	msg := messages[0]
	assert.Equal(t, mcp.Role("user"), msg.Role)

	textContent, ok := msg.Content.(*mcp.TextContent)
	require.True(t, ok, "Content should be TextContent")
	assert.NotEmpty(t, textContent.Text)
	assert.Contains(t, textContent.Text, "list_clusters")
	assert.Contains(t, textContent.Text, "Cluster ID")
	assert.Contains(t, textContent.Text, "Cluster name")
	assert.Contains(t, textContent.Text, "Cluster type")
}

func TestListClusterPrompt_RegisterWith(t *testing.T) {
	t.Parallel()

	prompt := NewListClusterPrompt()
	server := mcp.NewServer(
		&mcp.Implementation{
			Name:    "test-server",
			Version: "1.0.0",
		},
		&mcp.ServerOptions{},
	)

	assert.NotPanics(t, func() {
		prompt.RegisterWith(server)
	})
}
