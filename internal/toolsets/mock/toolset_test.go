package mock

import (
	"testing"

	"github.com/stackrox/stackrox-mcp/internal/toolsets"
	"github.com/stretchr/testify/assert"
)

func TestNewToolset(t *testing.T) {
	t.Run("creates toolset with provided values", func(t *testing.T) {
		name := "test-toolset"
		enabled := true
		tools := []toolsets.Tool{
			NewTool("tool1", true),
			NewTool("tool2", false),
		}

		toolset := NewToolset(name, enabled, tools)

		assert.Equal(t, name, toolset.NameValue)
		assert.Equal(t, enabled, toolset.EnabledValue)
		assert.Len(t, toolset.ToolsValue, len(tools))
	})

	t.Run("creates disabled toolset", func(t *testing.T) {
		toolset := NewToolset("disabled-toolset", false, nil)

		assert.False(t, toolset.EnabledValue)
	})

	t.Run("creates toolset with empty tools", func(t *testing.T) {
		toolset := NewToolset("empty-toolset", true, []toolsets.Tool{})

		assert.NotNil(t, toolset.ToolsValue)
		assert.Len(t, toolset.ToolsValue, 0)
	})

	t.Run("creates toolset with nil tools", func(t *testing.T) {
		toolset := NewToolset("nil-tools", true, nil)

		assert.Nil(t, toolset.ToolsValue)
	})
}

func TestToolset_GetName(t *testing.T) {
	t.Run("returns configured name", func(t *testing.T) {
		name := "my-toolset"
		toolset := NewToolset(name, true, nil)

		assert.Equal(t, name, toolset.GetName())
	})

	t.Run("returns empty string if configured", func(t *testing.T) {
		toolset := NewToolset("", true, nil)

		assert.Equal(t, "", toolset.GetName())
	})
}

func TestToolset_IsEnabled(t *testing.T) {
	t.Run("returns true when enabled", func(t *testing.T) {
		toolset := NewToolset("enabled", true, nil)

		assert.True(t, toolset.IsEnabled())
	})

	t.Run("returns false when disabled", func(t *testing.T) {
		toolset := NewToolset("disabled", false, nil)

		assert.False(t, toolset.IsEnabled())
	})

	t.Run("can toggle enabled state", func(t *testing.T) {
		toolset := NewToolset("toggle", true, nil)

		assert.True(t, toolset.IsEnabled())

		toolset.EnabledValue = false

		assert.False(t, toolset.IsEnabled())
	})
}

func TestToolset_GetTools_Enabled(t *testing.T) {
	tools := []toolsets.Tool{
		NewTool("tool1", true),
		NewTool("tool2", false),
		NewTool("tool3", true),
	}
	toolset := NewToolset("enabled-toolset", true, tools)

	result := toolset.GetTools()

	assert.Len(t, result, len(tools))

	for i, tool := range result {
		assert.Same(t, tools[i], tool)
	}
}

func TestToolset_GetTools_Disabled(t *testing.T) {
	tools := []toolsets.Tool{
		NewTool("tool1", true),
		NewTool("tool2", false),
	}
	toolset := NewToolset("disabled-toolset", false, tools)

	result := toolset.GetTools()

	assert.NotNil(t, result)
	assert.Len(t, result, 0)
}

func TestToolset_GetTools_EmptyList(t *testing.T) {
	toolset := NewToolset("no-tools", true, []toolsets.Tool{})

	result := toolset.GetTools()

	assert.NotNil(t, result)
	assert.Len(t, result, 0)
}

func TestToolset_GetTools_NilTools(t *testing.T) {
	toolset := NewToolset("nil-tools", true, nil)

	result := toolset.GetTools()

	assert.Nil(t, result)
}

func TestToolset_GetTools_ToggleState(t *testing.T) {
	tools := []toolsets.Tool{NewTool("tool1", true)}
	toolset := NewToolset("toggle-toolset", true, tools)

	// Initially enabled - should return tools
	result1 := toolset.GetTools()
	assert.Len(t, result1, 1)

	// Disable - should return empty slice
	toolset.EnabledValue = false

	result2 := toolset.GetTools()
	assert.Len(t, result2, 0)

	// Re-enable - should return tools again
	toolset.EnabledValue = true

	result3 := toolset.GetTools()
	assert.Len(t, result3, 1)
}

func TestToolset_InterfaceCompliance(t *testing.T) {
	t.Run("implements toolsets.Toolset interface", func(*testing.T) {
		var _ toolsets.Toolset = (*Toolset)(nil)
	})
}

func TestToolset_AsInterface(t *testing.T) {
	var toolsetInstance toolsets.Toolset = NewToolset("interface-test", true, nil)

	assert.Equal(t, "interface-test", toolsetInstance.GetName())
	assert.True(t, toolsetInstance.IsEnabled())

	tools := toolsetInstance.GetTools()
	assert.Nil(t, tools)
}

func TestToolset_EdgeCases(t *testing.T) {
	t.Run("toolset with single tool", func(t *testing.T) {
		tools := []toolsets.Tool{NewTool("only-tool", true)}
		toolset := NewToolset("single", true, tools)

		result := toolset.GetTools()
		assert.Len(t, result, 1)
	})

	t.Run("toolset with many tools", func(t *testing.T) {
		tools := make([]toolsets.Tool, 100)
		for i := range 100 {
			tools[i] = NewTool("tool", i%2 == 0)
		}

		toolset := NewToolset("many-tools", true, tools)

		result := toolset.GetTools()
		assert.Len(t, result, 100)
	})

	t.Run("toolset name with special characters", func(t *testing.T) {
		name := "tool-set_123!@#"
		toolset := NewToolset(name, true, nil)

		assert.Equal(t, name, toolset.GetName())
	})
}
