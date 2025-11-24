package config

import (
	"testing"

	"github.com/stackrox/stackrox-mcp/internal/client"
	"github.com/stackrox/stackrox-mcp/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewToolset(t *testing.T) {
	toolset := NewToolset(&config.Config{}, &client.Client{})
	require.NotNil(t, toolset)
	assert.Equal(t, "config_manager", toolset.GetName())
}

func TestToolset_IsEnabled_True(t *testing.T) {
	cfg := &config.Config{
		Tools: config.ToolsConfig{
			ConfigManager: config.ToolConfigManagerConfig{
				Enabled: true,
			},
		},
	}

	toolset := NewToolset(cfg, &client.Client{})
	assert.True(t, toolset.IsEnabled())

	tools := toolset.GetTools()
	require.NotEmpty(t, tools, "Should return tools when enabled")
	require.Len(t, tools, 1, "Should have list_clusters tool")
	assert.Equal(t, "list_clusters", tools[0].GetName())
}

func TestToolset_IsEnabled_False(t *testing.T) {
	cfg := &config.Config{
		Tools: config.ToolsConfig{
			ConfigManager: config.ToolConfigManagerConfig{
				Enabled: false,
			},
		},
	}

	toolset := NewToolset(cfg, &client.Client{})
	assert.False(t, toolset.IsEnabled())

	tools := toolset.GetTools()
	assert.Empty(t, tools, "Should return empty list when toolset is disabled")
}
