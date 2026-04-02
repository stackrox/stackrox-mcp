// Package config provides MCP prompts for configuration management.
package config

import (
	"github.com/stackrox/stackrox-mcp/internal/config"
	"github.com/stackrox/stackrox-mcp/internal/prompts"
)

type promptset struct {
	cfg *config.Config
}

// NewPromptset creates a new config management promptset.
func NewPromptset(cfg *config.Config) prompts.Promptset {
	return &promptset{
		cfg: cfg,
	}
}

func (p *promptset) GetName() string {
	return "config"
}

func (p *promptset) IsEnabled() bool {
	return p.cfg.Prompts.ConfigManager.Enabled
}

func (p *promptset) GetPrompts() []prompts.Prompt {
	return []prompts.Prompt{
		NewListClusterPrompt(),
	}
}
