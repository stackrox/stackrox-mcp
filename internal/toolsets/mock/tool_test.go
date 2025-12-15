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
}

func TestTool_RegisterWith(t *testing.T) {
	t.Run("sets RegisterCalled flag", func(t *testing.T) {
		tool := NewTool("register-test", true)

		assert.False(t, tool.RegisterCalled)

		tool.RegisterWith(nil)

		assert.True(t, tool.RegisterCalled)

		// Can be called multiple times
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
