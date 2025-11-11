// Package mock holds mocks for Tool and Toolset interfaces.
package mock

import (
	"github.com/stackrox/stackrox-mcp/internal/toolsets"
)

// Toolset is a mock implementation of the toolsets.Toolset interface for testing.
type Toolset struct {
	NameValue    string
	EnabledValue bool
	ToolsValue   []toolsets.Tool
}

// NewToolset creates a new mock toolset with the given parameters.
func NewToolset(name string, enabled bool, tools []toolsets.Tool) *Toolset {
	return &Toolset{
		NameValue:    name,
		EnabledValue: enabled,
		ToolsValue:   tools,
	}
}

// GetName returns the toolset name.
func (m *Toolset) GetName() string {
	return m.NameValue
}

// IsEnabled returns whether this toolset is enabled.
func (m *Toolset) IsEnabled() bool {
	return m.EnabledValue
}

// GetTools returns the tools in this toolset.
// If the toolset is disabled, returns an empty slice.
func (m *Toolset) GetTools() []toolsets.Tool {
	if !m.EnabledValue {
		return []toolsets.Tool{}
	}

	return m.ToolsValue
}
