// Package violations provides functionality for violations toolset.
package violations

import (
	"github.com/stackrox/stackrox-mcp/internal/client"
	"github.com/stackrox/stackrox-mcp/internal/config"
	"github.com/stackrox/stackrox-mcp/internal/toolsets"
)

// Toolset implements the violations management toolset.
type Toolset struct {
	cfg   *config.Config
	tools []toolsets.Tool
}

// NewToolset creates a new violations management toolset.
func NewToolset(cfg *config.Config, c *client.Client) *Toolset {
	return &Toolset{
		cfg: cfg,
		tools: []toolsets.Tool{
			NewListViolationsTool(c),
		},
	}
}

// GetName returns the toolset name.
func (t *Toolset) GetName() string {
	return "violations"
}

// IsEnabled checks if this toolset is enabled in configuration.
func (t *Toolset) IsEnabled() bool {
	return t.cfg.Tools.Violations.Enabled
}

// GetTools returns all tools.
func (t *Toolset) GetTools() []toolsets.Tool {
	if !t.IsEnabled() {
		return []toolsets.Tool{}
	}

	return t.tools
}
