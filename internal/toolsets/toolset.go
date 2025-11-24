// Package toolsets handles tools and toolsets registration.
package toolsets

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Tool represents a single MCP tool with metadata.
type Tool interface {
	// IsReadOnly returns true if the tool only performs read operations.
	IsReadOnly() bool

	// GetTool returns the MCP SDK Tool definition.
	GetTool() *mcp.Tool

	// GetName returns the tool name for logging/debugging.
	GetName() string

	// RegisterWith registers the tool's handler with the MCP server.
	RegisterWith(server *mcp.Server)
}

// Toolset represents a collection of related tools.
type Toolset interface {
	// GetName returns the toolset name.
	GetName() string

	// IsEnabled checks if this toolset is enabled in configuration.
	IsEnabled() bool

	// GetTools returns available tools based on configuration.
	GetTools() []Tool
}
