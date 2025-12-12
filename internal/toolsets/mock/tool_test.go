package mock

import (
	"testing"

	"github.com/stackrox/stackrox-mcp/internal/toolsets"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTool(t *testing.T) {
	t.Run("creates tool with provided values", func(t *testing.T) {
		name := "test-tool"
		readOnly := true

		tool := NewTool(name, readOnly)

		assert.Equal(t, name, tool.NameValue)
		assert.Equal(t, readOnly, tool.ReadOnlyValue)
		assert.False(t, tool.RegisterCalled)
	})

	t.Run("creates read-only tool", func(t *testing.T) {
		tool := NewTool("readonly", true)

		assert.True(t, tool.ReadOnlyValue)
	})

	t.Run("creates writable tool", func(t *testing.T) {
		tool := NewTool("writable", false)

		assert.False(t, tool.ReadOnlyValue)
	})

	t.Run("creates tool with empty name", func(t *testing.T) {
		tool := NewTool("", true)

		assert.Equal(t, "", tool.NameValue)
	})
}

func TestTool_GetName(t *testing.T) {
	t.Run("returns configured name", func(t *testing.T) {
		name := "my-tool"
		tool := NewTool(name, true)

		assert.Equal(t, name, tool.GetName())
	})

	t.Run("returns empty string if configured", func(t *testing.T) {
		tool := NewTool("", false)

		assert.Equal(t, "", tool.GetName())
	})

	t.Run("name with special characters", func(t *testing.T) {
		name := "tool-name_123!@#"
		tool := NewTool(name, true)

		assert.Equal(t, name, tool.GetName())
	})
}

func TestTool_IsReadOnly(t *testing.T) {
	t.Run("returns true when read-only", func(t *testing.T) {
		tool := NewTool("readonly", true)

		assert.True(t, tool.IsReadOnly())
	})

	t.Run("returns false when writable", func(t *testing.T) {
		tool := NewTool("writable", false)

		assert.False(t, tool.IsReadOnly())
	})

	t.Run("can toggle read-only state", func(t *testing.T) {
		tool := NewTool("toggle", true)

		assert.True(t, tool.IsReadOnly())

		tool.ReadOnlyValue = false

		assert.False(t, tool.IsReadOnly())
	})
}

func TestTool_GetTool(t *testing.T) {
	t.Run("returns MCP tool definition", func(t *testing.T) {
		name := "test-tool"
		tool := NewTool(name, true)

		mcpTool := tool.GetTool()

		require.NotNil(t, mcpTool)
		assert.Equal(t, name, mcpTool.Name)
		assert.NotEmpty(t, mcpTool.Description)

		expectedDesc := "Mock tool for testing"
		assert.Equal(t, expectedDesc, mcpTool.Description)
	})

	t.Run("returns new tool instance each time", func(t *testing.T) {
		tool := NewTool("test", true)

		mcpTool1 := tool.GetTool()
		mcpTool2 := tool.GetTool()

		// Should be different instances
		assert.NotSame(t, mcpTool1, mcpTool2)

		// But with same values
		assert.Equal(t, mcpTool1.Name, mcpTool2.Name)
	})

	t.Run("MCP tool has correct structure", func(t *testing.T) {
		tool := NewTool("structured-tool", false)

		mcpTool := tool.GetTool()

		assert.NotEmpty(t, mcpTool.Name)
		assert.NotEmpty(t, mcpTool.Description)
	})
}

func TestTool_RegisterWith(t *testing.T) {
	t.Run("sets RegisterCalled flag", func(t *testing.T) {
		tool := NewTool("register-test", true)

		assert.False(t, tool.RegisterCalled)

		tool.RegisterWith(nil)

		assert.True(t, tool.RegisterCalled)
	})

	t.Run("can be called multiple times", func(t *testing.T) {
		tool := NewTool("multi-register", false)

		tool.RegisterWith(nil)

		assert.True(t, tool.RegisterCalled)

		tool.RegisterWith(nil)

		assert.True(t, tool.RegisterCalled)
	})

	t.Run("accepts nil server", func(t *testing.T) {
		tool := NewTool("nil-server", true)

		// Should not panic
		tool.RegisterWith(nil)

		assert.True(t, tool.RegisterCalled)
	})

	t.Run("can track registration state", func(t *testing.T) {
		tool := NewTool("track-registration", true)

		// Reset the flag
		tool.RegisterCalled = false

		assert.False(t, tool.RegisterCalled)

		tool.RegisterWith(nil)

		assert.True(t, tool.RegisterCalled)
	})
}

func TestTool_InterfaceCompliance(t *testing.T) {
	t.Run("implements toolsets.Tool interface", func(*testing.T) {
		var _ toolsets.Tool = (*Tool)(nil)
	})
}

func TestTool_AsInterface(t *testing.T) {
	var toolInstance toolsets.Tool = NewTool("interface-test", true)

	assert.Equal(t, "interface-test", toolInstance.GetName())
	assert.True(t, toolInstance.IsReadOnly())

	mcpTool := toolInstance.GetTool()
	assert.NotNil(t, mcpTool)

	toolInstance.RegisterWith(nil)
}

func TestTool_EdgeCases(t *testing.T) {
	t.Run("tool with very long name", func(t *testing.T) {
		longName := "very-long-tool-name-that-might-be-used-in-some-edge-case-scenario-for-testing-purposes"
		tool := NewTool(longName, true)

		assert.Equal(t, longName, tool.GetName())

		mcpTool := tool.GetTool()
		assert.Equal(t, longName, mcpTool.Name)
	})

	t.Run("tool state is mutable", func(t *testing.T) {
		tool := NewTool("mutable", true)

		// Change name
		tool.NameValue = "new-name"
		assert.Equal(t, "new-name", tool.GetName())

		// Change read-only
		tool.ReadOnlyValue = false
		assert.False(t, tool.IsReadOnly())

		// Change register flag
		tool.RegisterCalled = true
		assert.True(t, tool.RegisterCalled)
	})

	t.Run("multiple tools with same name", func(t *testing.T) {
		tool1 := NewTool("same-name", true)
		tool2 := NewTool("same-name", false)

		assert.Equal(t, tool1.GetName(), tool2.GetName())
		assert.NotSame(t, tool1, tool2)
		assert.NotEqual(t, tool1.IsReadOnly(), tool2.IsReadOnly())
	})
}

func TestTool_ReadOnlyWorkflow(t *testing.T) {
	tool := NewTool("read-tool", true)

	assert.True(t, tool.IsReadOnly())
	assert.False(t, tool.RegisterCalled)

	mcpTool := tool.GetTool()
	assert.Equal(t, "read-tool", mcpTool.Name)

	tool.RegisterWith(nil)

	assert.True(t, tool.RegisterCalled)
}

func TestTool_WritableWorkflow(t *testing.T) {
	tool := NewTool("write-tool", false)

	assert.False(t, tool.IsReadOnly())

	_ = tool.GetTool()
	tool.RegisterWith(nil)

	assert.True(t, tool.RegisterCalled)
}

func TestTool_InToolset(t *testing.T) {
	tool1 := NewTool("tool1", true)
	tool2 := NewTool("tool2", false)

	tools := []toolsets.Tool{tool1, tool2}

	for _, toolInstance := range tools {
		assert.NotEmpty(t, toolInstance.GetName())

		_ = toolInstance.GetTool()
		toolInstance.RegisterWith(nil)
	}

	assert.True(t, tool1.RegisterCalled)
	assert.True(t, tool2.RegisterCalled)
}
