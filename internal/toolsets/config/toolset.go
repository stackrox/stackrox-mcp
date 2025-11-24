// Package config provides functionality for config manager toolset.
package config

import (
	"github.com/stackrox/stackrox-mcp/internal/config"
	"github.com/stackrox/stackrox-mcp/internal/toolsets"
)

// Toolset implements the config management toolset.
type Toolset struct {
	cfg   *config.Config
	tools []toolsets.Tool
}

// NewToolset creates a new config management toolset.
func NewToolset(cfg *config.Config) *Toolset {
	return &Toolset{
		cfg: cfg,
		tools: []toolsets.Tool{
			NewListClustersTool(),
		},
	}
}

// GetName returns the toolset name.
func (t *Toolset) GetName() string {
	return "config_manager"
}

// IsEnabled checks if this toolset is enabled in configuration.
func (t *Toolset) IsEnabled() bool {
	return t.cfg.Tools.ConfigManager.Enabled
}

// GetTools returns all tools.
func (t *Toolset) GetTools() []toolsets.Tool {
	if !t.IsEnabled() {
		return []toolsets.Tool{}
	}

	return t.tools
}
