package config

import (
	"testing"

	"github.com/stackrox/stackrox-mcp/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPromptset(t *testing.T) {
	cfg := &config.Config{
		Prompts: config.PromptsConfig{
			ConfigManager: config.PromptsConfigManagerConfig{
				Enabled: true,
			},
		},
	}

	promptset := NewPromptset(cfg)

	require.NotNil(t, promptset)
	assert.Equal(t, "config", promptset.GetName())
}

func TestPromptset_IsEnabled(t *testing.T) {
	tests := []struct {
		name    string
		enabled bool
	}{
		{
			name:    "enabled",
			enabled: true,
		},
		{
			name:    "disabled",
			enabled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				Prompts: config.PromptsConfig{
					ConfigManager: config.PromptsConfigManagerConfig{
						Enabled: tt.enabled,
					},
				},
			}

			promptset := NewPromptset(cfg)

			assert.Equal(t, tt.enabled, promptset.IsEnabled())
		})
	}
}

func TestPromptset_GetPrompts(t *testing.T) {
	cfg := &config.Config{
		Prompts: config.PromptsConfig{
			ConfigManager: config.PromptsConfigManagerConfig{
				Enabled: true,
			},
		},
	}

	promptset := NewPromptset(cfg)

	prompts := promptset.GetPrompts()

	require.Len(t, prompts, 1)
	assert.Equal(t, "list-cluster", prompts[0].GetName())
}
