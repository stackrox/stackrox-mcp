package app

import (
	"testing"

	"github.com/stackrox/stackrox-mcp/internal/client"
	"github.com/stackrox/stackrox-mcp/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestGetToolsets(t *testing.T) {
	allToolsets := GetToolsets(&config.Config{}, &client.Client{})

	toolsetNames := []string{}
	for _, toolset := range allToolsets {
		toolsetNames = append(toolsetNames, toolset.GetName())
	}

	assert.Contains(t, toolsetNames, "config_manager")
	assert.Contains(t, toolsetNames, "vulnerability")
}
