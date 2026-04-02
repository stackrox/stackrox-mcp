package prompts

import (
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stackrox/stackrox-mcp/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockPrompt struct {
	name string
}

func (m *mockPrompt) GetName() string {
	return m.name
}

func (m *mockPrompt) GetPrompt() *mcp.Prompt {
	return &mcp.Prompt{
		Name:        m.name,
		Description: "Mock prompt",
	}
}

func (m *mockPrompt) GetMessages(_ map[string]any) ([]*mcp.PromptMessage, error) {
	return []*mcp.PromptMessage{
		{
			Role: "user",
			Content: &mcp.TextContent{
				Text: "mock message",
			},
		},
	}, nil
}

func (m *mockPrompt) RegisterWith(_ *mcp.Server) {}

type mockPromptset struct {
	name    string
	enabled bool
	prompts []Prompt
}

func (m *mockPromptset) GetName() string {
	return m.name
}

func (m *mockPromptset) IsEnabled() bool {
	return m.enabled
}

func (m *mockPromptset) GetPrompts() []Prompt {
	return m.prompts
}

func TestNewRegistry(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{}
	promptsets := []Promptset{
		&mockPromptset{name: "test", enabled: true},
	}

	registry := NewRegistry(cfg, promptsets)

	require.NotNil(t, registry)
	assert.Equal(t, cfg, registry.cfg)
	assert.Equal(t, promptsets, registry.promptsets)
}

func TestRegistry_GetPromptsets(t *testing.T) {
	t.Parallel()

	promptsets := []Promptset{
		&mockPromptset{name: "test1", enabled: true},
		&mockPromptset{name: "test2", enabled: false},
	}

	registry := NewRegistry(&config.Config{}, promptsets)

	result := registry.GetPromptsets()

	assert.Equal(t, promptsets, result)
}

func TestRegistry_GetAllPrompts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		promptsets []Promptset
		wantCount  int
	}{
		{
			name: "all enabled promptsets",
			promptsets: []Promptset{
				&mockPromptset{
					name:    "set1",
					enabled: true,
					prompts: []Prompt{
						&mockPrompt{name: "prompt1"},
						&mockPrompt{name: "prompt2"},
					},
				},
				&mockPromptset{
					name:    "set2",
					enabled: true,
					prompts: []Prompt{
						&mockPrompt{name: "prompt3"},
					},
				},
			},
			wantCount: 3,
		},
		{
			name: "some disabled promptsets",
			promptsets: []Promptset{
				&mockPromptset{
					name:    "enabled",
					enabled: true,
					prompts: []Prompt{
						&mockPrompt{name: "prompt1"},
					},
				},
				&mockPromptset{
					name:    "disabled",
					enabled: false,
					prompts: []Prompt{
						&mockPrompt{name: "prompt2"},
					},
				},
			},
			wantCount: 1,
		},
		{
			name:       "no promptsets",
			promptsets: []Promptset{},
			wantCount:  0,
		},
		{
			name: "all disabled promptsets",
			promptsets: []Promptset{
				&mockPromptset{
					name:    "disabled1",
					enabled: false,
					prompts: []Prompt{
						&mockPrompt{name: "prompt1"},
					},
				},
			},
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			registry := NewRegistry(&config.Config{}, tt.promptsets)

			prompts := registry.GetAllPrompts()

			assert.Len(t, prompts, tt.wantCount)

			for _, prompt := range prompts {
				assert.NotNil(t, prompt)
				assert.NotEmpty(t, prompt.Name)
			}
		})
	}
}
