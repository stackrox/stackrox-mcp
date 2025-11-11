package mock

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Tool is a mock implementation of the toolsets.Tool interface for testing.
type Tool struct {
	NameValue      string
	ReadOnlyValue  bool
	RegisterCalled bool
}

// NewTool creates a new mock tool with the given name and read-only status.
func NewTool(name string, readOnly bool) *Tool {
	return &Tool{
		NameValue:     name,
		ReadOnlyValue: readOnly,
	}
}

// IsReadOnly returns whether this tool is read-only.
func (m *Tool) IsReadOnly() bool {
	return m.ReadOnlyValue
}

// GetName returns the tool name.
func (m *Tool) GetName() string {
	return m.NameValue
}

// GetTool returns the MCP Tool definition.
func (m *Tool) GetTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        m.NameValue,
		Description: "Mock tool for testing",
	}
}

// RegisterWith tracks that the tool was registered with the MCP server.
func (m *Tool) RegisterWith(_ *mcp.Server) {
	m.RegisterCalled = true
}
