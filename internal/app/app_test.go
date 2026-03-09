package app

import (
	"testing"

	"github.com/stackrox/stackrox-mcp/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetToolsets(t *testing.T) {
	cfg := &config.Config{
		Central: config.CentralConfig{
			URL: "https://example.com",
		},
		Tools: config.ToolsConfig{
			ConfigManager: config.ToolConfigManagerConfig{
				Enabled: false,
			},
			Vulnerability: config.ToolsetVulnerabilityConfig{
				Enabled: true,
			},
		},
	}

	// We can't create a real client without a valid connection,
	// so we pass nil and just test that getToolsets returns something
	toolsets := getToolsets(cfg, nil)

	require.NotNil(t, toolsets, "getToolsets should return a non-nil slice")
	assert.NotEmpty(t, toolsets, "getToolsets should return at least one toolset")
	assert.Len(t, toolsets, 2, "getToolsets should return 2 toolsets")
}
