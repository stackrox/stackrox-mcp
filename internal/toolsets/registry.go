package toolsets

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stackrox/stackrox-mcp/internal/config"
)

// Registry manages all available toolsets and collects their tools.
type Registry struct {
	cfg      *config.Config
	toolsets []Toolset
}

// NewRegistry creates a new registry with the given config and toolsets.
func NewRegistry(cfg *config.Config, toolsets []Toolset) *Registry {
	return &Registry{
		cfg:      cfg,
		toolsets: toolsets,
	}
}

// GetAllTools returns all enabled tools from all enabled toolsets.
// Each toolset handles its own enabled check.
func (r *Registry) GetAllTools() []*mcp.Tool {
	tools := make([]*mcp.Tool, 0)

	for _, toolset := range r.toolsets {
		for _, tool := range toolset.GetTools() {
			if r.cfg.Global.ReadOnlyTools && !tool.IsReadOnly() {
				continue
			}

			tools = append(tools, tool.GetTool())
		}
	}

	return tools
}

// GetToolsets returns all registered toolsets (for debugging/testing).
func (r *Registry) GetToolsets() []Toolset {
	return r.toolsets
}
