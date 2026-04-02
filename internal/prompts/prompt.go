// Package prompts handles MCP prompts and promptset registration.
package prompts

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Prompt represents a single MCP prompt with metadata.
type Prompt interface {
	// GetName returns the prompt name for logging/debugging.
	GetName() string

	// GetPrompt returns the MCP SDK Prompt definition.
	GetPrompt() *mcp.Prompt

	// GetMessages returns PromptMessage objects for the given arguments.
	GetMessages(arguments map[string]any) ([]*mcp.PromptMessage, error)

	// RegisterWith registers the prompt's handler with the MCP server.
	RegisterWith(server *mcp.Server)
}

// Promptset represents a collection of related prompts.
type Promptset interface {
	// GetName returns the promptset name.
	GetName() string

	// IsEnabled checks if this promptset is enabled in configuration.
	IsEnabled() bool

	// GetPrompts returns available prompts based on configuration.
	GetPrompts() []Prompt
}
