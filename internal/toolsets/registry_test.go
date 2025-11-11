package toolsets_test

import (
	"testing"

	"github.com/stackrox/stackrox-mcp/internal/config"
	"github.com/stackrox/stackrox-mcp/internal/toolsets"
	"github.com/stackrox/stackrox-mcp/internal/toolsets/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRegistry(t *testing.T) {
	cfg := &config.Config{}
	toolsetList := []toolsets.Toolset{
		mock.NewToolset("mock_toolset_1", true, []toolsets.Tool{
			mock.NewTool("tool_1", true),
		}),
		mock.NewToolset("mock_toolset_2", true, []toolsets.Tool{
			mock.NewTool("tool_2", true),
		}),
	}

	registry := toolsets.NewRegistry(cfg, toolsetList)

	require.NotNil(t, registry)
	assert.Len(t, registry.GetToolsets(), 2)
}

func TestRegistry_GetAllTools_AllToolsetsEnabled(t *testing.T) {
	cfg := &config.Config{
		Global: config.GlobalConfig{
			ReadOnlyTools: false,
		},
	}

	toolsetList := []toolsets.Toolset{
		mock.NewToolset("mock_toolset_1", true, []toolsets.Tool{
			mock.NewTool("read_only_tool", true),
		}),
		mock.NewToolset("mock_toolset_2", true, []toolsets.Tool{
			mock.NewTool("read_write_tool", false),
		}),
	}

	registry := toolsets.NewRegistry(cfg, toolsetList)

	tools := registry.GetAllTools()

	require.NotEmpty(t, tools)
	assert.Len(t, tools, 2, "Should have tools from both toolsets")

	toolNames := make(map[string]bool)
	for _, tool := range tools {
		toolNames[tool.Name] = true
	}

	assert.True(t, toolNames["read_only_tool"], "Should have read_only_tool")
	assert.True(t, toolNames["read_write_tool"], "Should have read_write_tool")
}

func TestRegistry_GetAllTools_OneToolsetDisabled(t *testing.T) {
	cfg := &config.Config{
		Global: config.GlobalConfig{
			ReadOnlyTools: false,
		},
	}

	toolsetList := []toolsets.Toolset{
		mock.NewToolset("enabled_toolset", true, []toolsets.Tool{
			mock.NewTool("enabled_tool", true),
		}),
		mock.NewToolset("disabled_toolset", false, []toolsets.Tool{
			mock.NewTool("disabled_tool", true),
		}),
	}

	registry := toolsets.NewRegistry(cfg, toolsetList)

	tools := registry.GetAllTools()

	require.NotEmpty(t, tools)
	require.Len(t, tools, 1, "Should only have tools from enabled toolset")
	assert.Equal(t, "enabled_tool", tools[0].Name)
}

func TestRegistry_GetAllTools_AllToolsetsDisabled(t *testing.T) {
	cfg := &config.Config{}

	toolsetList := []toolsets.Toolset{
		mock.NewToolset("disabled_toolset_1", false, []toolsets.Tool{
			mock.NewTool("tool_1", true),
		}),
		mock.NewToolset("disabled_toolset_2", false, []toolsets.Tool{
			mock.NewTool("tool_2", true),
		}),
	}

	registry := toolsets.NewRegistry(cfg, toolsetList)

	tools := registry.GetAllTools()

	assert.Empty(t, tools, "Should return empty list when all toolsets are disabled")
}

func TestRegistry_GetAllTools_FiltersReadWriteTools(t *testing.T) {
	cfg := &config.Config{
		Global: config.GlobalConfig{
			ReadOnlyTools: true,
		},
	}

	toolsetList := []toolsets.Toolset{
		mock.NewToolset("mixed_toolset", true, []toolsets.Tool{
			mock.NewTool("read_only_1", true),
			mock.NewTool("read_write_1", false),
			mock.NewTool("read_only_2", true),
			mock.NewTool("read_write_2", false),
		}),
	}

	registry := toolsets.NewRegistry(cfg, toolsetList)
	tools := registry.GetAllTools()

	require.Len(t, tools, 2, "Should filter out read-write tools")

	toolNames := make(map[string]bool)
	for _, tool := range tools {
		toolNames[tool.Name] = true
	}

	assert.True(t, toolNames["read_only_1"], "Should have read-only tools")
	assert.True(t, toolNames["read_only_2"], "Should have read-only tools")
}

func TestRegistry_GetToolsets(t *testing.T) {
	cfg := &config.Config{}

	mockToolset1 := mock.NewToolset("toolset_1", true, []toolsets.Tool{
		mock.NewTool("tool_1", true),
	})
	mockToolset2 := mock.NewToolset("toolset_2", true, []toolsets.Tool{
		mock.NewTool("tool_2", true),
	})

	toolsetList := []toolsets.Toolset{mockToolset1, mockToolset2}
	registry := toolsets.NewRegistry(cfg, toolsetList)

	retrievedToolsets := registry.GetToolsets()

	require.Len(t, retrievedToolsets, 2)
	assert.Contains(t, retrievedToolsets, mockToolset1)
	assert.Contains(t, retrievedToolsets, mockToolset2)
}

func TestRegistry_EmptyToolsets(t *testing.T) {
	cfg := &config.Config{}
	registry := toolsets.NewRegistry(cfg, []toolsets.Toolset{})

	tools := registry.GetAllTools()
	toolsetList := registry.GetToolsets()

	assert.Empty(t, tools)
	assert.Empty(t, toolsetList)
}
