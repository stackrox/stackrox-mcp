package mock

import (
	"testing"

	"github.com/stackrox/stackrox-mcp/internal/toolsets"
)

func TestNewTool(t *testing.T) {
	t.Run("creates tool with provided values", func(t *testing.T) {
		name := "test-tool"
		readOnly := true

		tool := NewTool(name, readOnly)

		if tool.NameValue != name {
			t.Errorf("Expected NameValue %q, got %q", name, tool.NameValue)
		}

		if tool.ReadOnlyValue != readOnly {
			t.Errorf("Expected ReadOnlyValue %v, got %v", readOnly, tool.ReadOnlyValue)
		}

		if tool.RegisterCalled {
			t.Error("Expected RegisterCalled to be false initially")
		}
	})

	t.Run("creates read-only tool", func(t *testing.T) {
		tool := NewTool("readonly", true)

		if !tool.ReadOnlyValue {
			t.Error("Expected tool to be read-only")
		}
	})

	t.Run("creates writable tool", func(t *testing.T) {
		tool := NewTool("writable", false)

		if tool.ReadOnlyValue {
			t.Error("Expected tool to be writable")
		}
	})

	t.Run("creates tool with empty name", func(t *testing.T) {
		tool := NewTool("", true)

		if tool.NameValue != "" {
			t.Errorf("Expected empty name, got %q", tool.NameValue)
		}
	})
}

func TestTool_GetName(t *testing.T) {
	t.Run("returns configured name", func(t *testing.T) {
		name := "my-tool"
		tool := NewTool(name, true)

		if tool.GetName() != name {
			t.Errorf("Expected name %q, got %q", name, tool.GetName())
		}
	})

	t.Run("returns empty string if configured", func(t *testing.T) {
		tool := NewTool("", false)

		if tool.GetName() != "" {
			t.Errorf("Expected empty name, got %q", tool.GetName())
		}
	})

	t.Run("name with special characters", func(t *testing.T) {
		name := "tool-name_123!@#"
		tool := NewTool(name, true)

		if tool.GetName() != name {
			t.Errorf("Expected name %q, got %q", name, tool.GetName())
		}
	})
}

func TestTool_IsReadOnly(t *testing.T) {
	t.Run("returns true when read-only", func(t *testing.T) {
		tool := NewTool("readonly", true)

		if !tool.IsReadOnly() {
			t.Error("Expected tool to be read-only")
		}
	})

	t.Run("returns false when writable", func(t *testing.T) {
		tool := NewTool("writable", false)

		if tool.IsReadOnly() {
			t.Error("Expected tool to be writable")
		}
	})

	t.Run("can toggle read-only state", func(t *testing.T) {
		tool := NewTool("toggle", true)

		if !tool.IsReadOnly() {
			t.Error("Expected initially read-only")
		}

		tool.ReadOnlyValue = false

		if tool.IsReadOnly() {
			t.Error("Expected writable after toggle")
		}
	})
}

func TestTool_GetTool(t *testing.T) {
	t.Run("returns MCP tool definition", func(t *testing.T) {
		name := "test-tool"
		tool := NewTool(name, true)

		mcpTool := tool.GetTool()

		if mcpTool == nil {
			t.Fatal("Expected non-nil MCP tool")
		}

		if mcpTool.Name != name {
			t.Errorf("Expected MCP tool name %q, got %q", name, mcpTool.Name)
		}

		if mcpTool.Description == "" {
			t.Error("Expected non-empty description")
		}

		expectedDesc := "Mock tool for testing"
		if mcpTool.Description != expectedDesc {
			t.Errorf("Expected description %q, got %q", expectedDesc, mcpTool.Description)
		}
	})

	t.Run("returns new tool instance each time", func(t *testing.T) {
		tool := NewTool("test", true)

		mcpTool1 := tool.GetTool()
		mcpTool2 := tool.GetTool()

		// Should be different instances
		if mcpTool1 == mcpTool2 {
			t.Error("Expected different instances, got same pointer")
		}

		// But with same values
		if mcpTool1.Name != mcpTool2.Name {
			t.Error("Expected same name in both instances")
		}
	})

	t.Run("MCP tool has correct structure", func(t *testing.T) {
		tool := NewTool("structured-tool", false)

		mcpTool := tool.GetTool()

		if mcpTool.Name == "" {
			t.Error("MCP tool should have a name")
		}

		if mcpTool.Description == "" {
			t.Error("MCP tool should have a description")
		}
	})
}

func TestTool_RegisterWith(t *testing.T) {
	t.Run("sets RegisterCalled flag", func(t *testing.T) {
		tool := NewTool("register-test", true)

		if tool.RegisterCalled {
			t.Error("Expected RegisterCalled to be false initially")
		}

		tool.RegisterWith(nil)

		if !tool.RegisterCalled {
			t.Error("Expected RegisterCalled to be true after calling RegisterWith")
		}
	})

	t.Run("can be called multiple times", func(t *testing.T) {
		tool := NewTool("multi-register", false)

		tool.RegisterWith(nil)
		if !tool.RegisterCalled {
			t.Error("Expected RegisterCalled to be true after first call")
		}

		tool.RegisterWith(nil)
		if !tool.RegisterCalled {
			t.Error("Expected RegisterCalled to remain true after second call")
		}
	})

	t.Run("accepts nil server", func(t *testing.T) {
		tool := NewTool("nil-server", true)

		// Should not panic
		tool.RegisterWith(nil)

		if !tool.RegisterCalled {
			t.Error("Expected RegisterCalled to be true")
		}
	})

	t.Run("can track registration state", func(t *testing.T) {
		tool := NewTool("track-registration", true)

		// Reset the flag
		tool.RegisterCalled = false

		if tool.RegisterCalled {
			t.Error("Expected RegisterCalled to be false after reset")
		}

		tool.RegisterWith(nil)

		if !tool.RegisterCalled {
			t.Error("Expected RegisterCalled to be true after registration")
		}
	})
}

func TestTool_InterfaceCompliance(t *testing.T) {
	t.Run("implements toolsets.Tool interface", func(t *testing.T) {
		var _ toolsets.Tool = (*Tool)(nil)
	})

	t.Run("can be used as toolsets.Tool", func(t *testing.T) {
		var tool toolsets.Tool = NewTool("interface-test", true)

		if tool.GetName() != "interface-test" {
			t.Errorf("Expected name 'interface-test', got %q", tool.GetName())
		}

		if !tool.IsReadOnly() {
			t.Error("Expected tool to be read-only")
		}

		mcpTool := tool.GetTool()
		if mcpTool == nil {
			t.Error("Expected non-nil MCP tool")
		}

		tool.RegisterWith(nil)
		// Can't check RegisterCalled through interface, but shouldn't panic
	})
}

func TestTool_EdgeCases(t *testing.T) {
	t.Run("tool with very long name", func(t *testing.T) {
		longName := "very-long-tool-name-that-might-be-used-in-some-edge-case-scenario-for-testing-purposes"
		tool := NewTool(longName, true)

		if tool.GetName() != longName {
			t.Errorf("Expected long name to be preserved")
		}

		mcpTool := tool.GetTool()
		if mcpTool.Name != longName {
			t.Error("Expected MCP tool to have long name")
		}
	})

	t.Run("tool state is mutable", func(t *testing.T) {
		tool := NewTool("mutable", true)

		// Change name
		tool.NameValue = "new-name"
		if tool.GetName() != "new-name" {
			t.Error("Expected name to be mutable")
		}

		// Change read-only
		tool.ReadOnlyValue = false
		if tool.IsReadOnly() {
			t.Error("Expected read-only to be mutable")
		}

		// Change register flag
		tool.RegisterCalled = true
		if !tool.RegisterCalled {
			t.Error("Expected register flag to be mutable")
		}
	})

	t.Run("multiple tools with same name", func(t *testing.T) {
		tool1 := NewTool("same-name", true)
		tool2 := NewTool("same-name", false)

		if tool1.GetName() != tool2.GetName() {
			t.Error("Expected both tools to have same name")
		}

		if tool1 == tool2 {
			t.Error("Expected different tool instances")
		}

		if tool1.IsReadOnly() == tool2.IsReadOnly() {
			t.Error("Expected different read-only values")
		}
	})
}

func TestTool_UsageScenarios(t *testing.T) {
	t.Run("typical read-only tool workflow", func(t *testing.T) {
		tool := NewTool("read-tool", true)

		// Check initial state
		if !tool.IsReadOnly() {
			t.Error("Expected read-only tool")
		}

		if tool.RegisterCalled {
			t.Error("Should not be registered initially")
		}

		// Get MCP definition
		mcpTool := tool.GetTool()
		if mcpTool.Name != "read-tool" {
			t.Error("MCP tool should have correct name")
		}

		// Register with server
		tool.RegisterWith(nil)
		if !tool.RegisterCalled {
			t.Error("Should be registered after RegisterWith call")
		}
	})

	t.Run("typical writable tool workflow", func(t *testing.T) {
		tool := NewTool("write-tool", false)

		// Check initial state
		if tool.IsReadOnly() {
			t.Error("Expected writable tool")
		}

		// Get tool definition and register
		_ = tool.GetTool()
		tool.RegisterWith(nil)

		if !tool.RegisterCalled {
			t.Error("Should be registered")
		}
	})

	t.Run("tool in toolset context", func(t *testing.T) {
		tool1 := NewTool("tool1", true)
		tool2 := NewTool("tool2", false)

		tools := []toolsets.Tool{tool1, tool2}

		for _, tool := range tools {
			if tool.GetName() == "" {
				t.Error("Tool in toolset should have name")
			}

			_ = tool.GetTool()
			tool.RegisterWith(nil)
		}

		// Check that both were registered
		if !tool1.RegisterCalled || !tool2.RegisterCalled {
			t.Error("All tools should be registered")
		}
	})
}
