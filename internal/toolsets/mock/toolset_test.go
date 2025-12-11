package mock

import (
	"testing"

	"github.com/stackrox/stackrox-mcp/internal/toolsets"
)

func TestNewToolset(t *testing.T) {
	t.Run("creates toolset with provided values", func(t *testing.T) {
		name := "test-toolset"
		enabled := true
		tools := []toolsets.Tool{
			NewTool("tool1", true),
			NewTool("tool2", false),
		}

		ts := NewToolset(name, enabled, tools)

		if ts.NameValue != name {
			t.Errorf("Expected NameValue %q, got %q", name, ts.NameValue)
		}

		if ts.EnabledValue != enabled {
			t.Errorf("Expected EnabledValue %v, got %v", enabled, ts.EnabledValue)
		}

		if len(ts.ToolsValue) != len(tools) {
			t.Errorf("Expected %d tools, got %d", len(tools), len(ts.ToolsValue))
		}
	})

	t.Run("creates disabled toolset", func(t *testing.T) {
		ts := NewToolset("disabled-toolset", false, nil)

		if ts.EnabledValue {
			t.Error("Expected toolset to be disabled")
		}
	})

	t.Run("creates toolset with empty tools", func(t *testing.T) {
		ts := NewToolset("empty-toolset", true, []toolsets.Tool{})

		if ts.ToolsValue == nil {
			t.Error("Expected non-nil ToolsValue, got nil")
		}

		if len(ts.ToolsValue) != 0 {
			t.Errorf("Expected 0 tools, got %d", len(ts.ToolsValue))
		}
	})

	t.Run("creates toolset with nil tools", func(t *testing.T) {
		ts := NewToolset("nil-tools", true, nil)

		if ts.ToolsValue != nil {
			t.Errorf("Expected nil ToolsValue, got %v", ts.ToolsValue)
		}
	})
}

func TestToolset_GetName(t *testing.T) {
	t.Run("returns configured name", func(t *testing.T) {
		name := "my-toolset"
		ts := NewToolset(name, true, nil)

		if ts.GetName() != name {
			t.Errorf("Expected name %q, got %q", name, ts.GetName())
		}
	})

	t.Run("returns empty string if configured", func(t *testing.T) {
		ts := NewToolset("", true, nil)

		if ts.GetName() != "" {
			t.Errorf("Expected empty name, got %q", ts.GetName())
		}
	})
}

func TestToolset_IsEnabled(t *testing.T) {
	t.Run("returns true when enabled", func(t *testing.T) {
		ts := NewToolset("enabled", true, nil)

		if !ts.IsEnabled() {
			t.Error("Expected toolset to be enabled")
		}
	})

	t.Run("returns false when disabled", func(t *testing.T) {
		ts := NewToolset("disabled", false, nil)

		if ts.IsEnabled() {
			t.Error("Expected toolset to be disabled")
		}
	})

	t.Run("can toggle enabled state", func(t *testing.T) {
		ts := NewToolset("toggle", true, nil)

		if !ts.IsEnabled() {
			t.Error("Expected initially enabled")
		}

		ts.EnabledValue = false

		if ts.IsEnabled() {
			t.Error("Expected disabled after toggle")
		}
	})
}

func TestToolset_GetTools(t *testing.T) {
	t.Run("returns tools when enabled", func(t *testing.T) {
		tools := []toolsets.Tool{
			NewTool("tool1", true),
			NewTool("tool2", false),
			NewTool("tool3", true),
		}
		ts := NewToolset("enabled-toolset", true, tools)

		result := ts.GetTools()

		if len(result) != len(tools) {
			t.Errorf("Expected %d tools, got %d", len(tools), len(result))
		}

		for i, tool := range result {
			if tool != tools[i] {
				t.Errorf("Tool at index %d doesn't match", i)
			}
		}
	})

	t.Run("returns empty slice when disabled", func(t *testing.T) {
		tools := []toolsets.Tool{
			NewTool("tool1", true),
			NewTool("tool2", false),
		}
		ts := NewToolset("disabled-toolset", false, tools)

		result := ts.GetTools()

		if result == nil {
			t.Error("Expected non-nil slice, got nil")
		}

		if len(result) != 0 {
			t.Errorf("Expected empty slice when disabled, got %d tools", len(result))
		}
	})

	t.Run("returns empty slice when enabled with no tools", func(t *testing.T) {
		ts := NewToolset("no-tools", true, []toolsets.Tool{})

		result := ts.GetTools()

		if result == nil {
			t.Error("Expected non-nil slice, got nil")
		}

		if len(result) != 0 {
			t.Errorf("Expected 0 tools, got %d", len(result))
		}
	})

	t.Run("returns nil when enabled with nil tools", func(t *testing.T) {
		ts := NewToolset("nil-tools", true, nil)

		result := ts.GetTools()

		if result != nil {
			t.Errorf("Expected nil, got %v", result)
		}
	})

	t.Run("changing enabled state changes returned tools", func(t *testing.T) {
		tools := []toolsets.Tool{NewTool("tool1", true)}
		ts := NewToolset("toggle-toolset", true, tools)

		// Initially enabled - should return tools
		result1 := ts.GetTools()
		if len(result1) != 1 {
			t.Errorf("Expected 1 tool when enabled, got %d", len(result1))
		}

		// Disable - should return empty slice
		ts.EnabledValue = false
		result2 := ts.GetTools()
		if len(result2) != 0 {
			t.Errorf("Expected 0 tools when disabled, got %d", len(result2))
		}

		// Re-enable - should return tools again
		ts.EnabledValue = true
		result3 := ts.GetTools()
		if len(result3) != 1 {
			t.Errorf("Expected 1 tool when re-enabled, got %d", len(result3))
		}
	})
}

func TestToolset_InterfaceCompliance(t *testing.T) {
	t.Run("implements toolsets.Toolset interface", func(t *testing.T) {
		var _ toolsets.Toolset = (*Toolset)(nil)
	})

	t.Run("can be used as toolsets.Toolset", func(t *testing.T) {
		var ts toolsets.Toolset = NewToolset("interface-test", true, nil)

		if ts.GetName() != "interface-test" {
			t.Errorf("Expected name 'interface-test', got %q", ts.GetName())
		}

		if !ts.IsEnabled() {
			t.Error("Expected toolset to be enabled")
		}

		tools := ts.GetTools()
		if tools != nil {
			t.Errorf("Expected nil tools, got %v", tools)
		}
	})
}

func TestToolset_EdgeCases(t *testing.T) {
	t.Run("toolset with single tool", func(t *testing.T) {
		tools := []toolsets.Tool{NewTool("only-tool", true)}
		ts := NewToolset("single", true, tools)

		result := ts.GetTools()
		if len(result) != 1 {
			t.Errorf("Expected 1 tool, got %d", len(result))
		}
	})

	t.Run("toolset with many tools", func(t *testing.T) {
		tools := make([]toolsets.Tool, 100)
		for i := 0; i < 100; i++ {
			tools[i] = NewTool("tool", i%2 == 0)
		}

		ts := NewToolset("many-tools", true, tools)

		result := ts.GetTools()
		if len(result) != 100 {
			t.Errorf("Expected 100 tools, got %d", len(result))
		}
	})

	t.Run("toolset name with special characters", func(t *testing.T) {
		name := "tool-set_123!@#"
		ts := NewToolset(name, true, nil)

		if ts.GetName() != name {
			t.Errorf("Expected name %q, got %q", name, ts.GetName())
		}
	})
}
