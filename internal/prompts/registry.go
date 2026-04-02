package prompts

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stackrox/stackrox-mcp/internal/config"
)

// Registry manages all promptsets and provides access to prompts.
type Registry struct {
	cfg        *config.Config
	promptsets []Promptset
}

// NewRegistry creates a new prompt registry with the given configuration and promptsets.
func NewRegistry(cfg *config.Config, promptsets []Promptset) *Registry {
	return &Registry{
		cfg:        cfg,
		promptsets: promptsets,
	}
}

// GetAllPrompts returns all prompt definitions from all enabled promptsets.
func (r *Registry) GetAllPrompts() []*mcp.Prompt {
	prompts := make([]*mcp.Prompt, 0)
	for _, promptset := range r.promptsets {
		if !promptset.IsEnabled() {
			continue
		}
		for _, prompt := range promptset.GetPrompts() {
			prompts = append(prompts, prompt.GetPrompt())
		}
	}
	return prompts
}

// GetPromptsets returns all registered promptsets.
func (r *Registry) GetPromptsets() []Promptset {
	return r.promptsets
}
