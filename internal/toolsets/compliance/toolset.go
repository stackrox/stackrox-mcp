// Package compliance provides functionality for compliance management toolset.
package compliance

import (
	"github.com/stackrox/stackrox-mcp/internal/client"
	"github.com/stackrox/stackrox-mcp/internal/config"
	"github.com/stackrox/stackrox-mcp/internal/toolsets"
)

// Toolset implements the compliance management toolset.
type Toolset struct {
	cfg   *config.Config
	tools []toolsets.Tool
}

// NewToolset creates a new compliance management toolset.
func NewToolset(cfg *config.Config, stackroxClient *client.Client) *Toolset {
	return &Toolset{
		cfg: cfg,
		tools: []toolsets.Tool{
			NewListComplianceProfilesTool(stackroxClient),
			NewGetComplianceScanResultsTool(stackroxClient),
			NewListComplianceScanConfigurationsTool(stackroxClient),
			NewGetComplianceCheckResultTool(stackroxClient),
		},
	}
}

// GetName returns the toolset name.
func (t *Toolset) GetName() string {
	return "compliance"
}

// IsEnabled checks if this toolset is enabled in configuration.
func (t *Toolset) IsEnabled() bool {
	return t.cfg.Tools.Compliance.Enabled
}

// GetTools returns all tools.
func (t *Toolset) GetTools() []toolsets.Tool {
	if !t.IsEnabled() {
		return []toolsets.Tool{}
	}

	return t.tools
}
