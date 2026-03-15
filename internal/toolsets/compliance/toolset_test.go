package compliance

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
	assert.Equal(t, "compliance", toolset.GetName())
}

func TestToolset_IsEnabled_True(t *testing.T) {
	cfg := &config.Config{
		Tools: config.ToolsConfig{
			Compliance: config.ToolsetComplianceConfig{
				Enabled: true,
			},
		},
	}

	toolset := NewToolset(cfg, &client.Client{})
	assert.True(t, toolset.IsEnabled())

	tools := toolset.GetTools()
	require.NotEmpty(t, tools, "Should return tools when enabled")
	require.Len(t, tools, 4, "Should have 4 compliance tools")
	assert.Equal(t, "list_compliance_profiles", tools[0].GetName())
	assert.Equal(t, "get_compliance_scan_results", tools[1].GetName())
	assert.Equal(t, "list_compliance_scan_configurations", tools[2].GetName())
	assert.Equal(t, "get_compliance_check_result", tools[3].GetName())
}

func TestToolset_IsEnabled_False(t *testing.T) {
	cfg := &config.Config{
		Tools: config.ToolsConfig{
			Compliance: config.ToolsetComplianceConfig{
				Enabled: false,
			},
		},
	}

	toolset := NewToolset(cfg, &client.Client{})
	assert.False(t, toolset.IsEnabled())

	tools := toolset.GetTools()
	assert.Empty(t, tools, "Should return empty list when toolset is disabled")
}
