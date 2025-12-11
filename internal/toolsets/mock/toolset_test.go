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

		toolset := NewToolset(name, enabled, tools)

		if toolset.NameValue != name {
			t.Errorf("Expected NameValue %q, got %q", name, toolset.NameValue)
		}

		if toolset.EnabledValue != enabled {
			t.Errorf("Expected EnabledValue %v, got %v", enabled, toolset.EnabledValue)
		}

		if len(toolset.ToolsValue) != len(tools) {
			t.Errorf("Expected %d tools, got %d", len(tools), len(toolset.ToolsValue))
		}
	})

	t.Run("creates disabled toolset", func(t *testing.T) {
		toolset := NewToolset("disabled-toolset", false, nil)

		if toolset.EnabledValue {
			t.Error("Expected toolset to be disabled")
		}
	})

	t.Run("creates toolset with empty tools", func(t *testing.T) {
		toolset := NewToolset("empty-toolset", true, []toolsets.Tool{})

		if toolset.ToolsValue == nil {
			t.Error("Expected non-nil ToolsValue, got nil")
		}

		if len(toolset.ToolsValue) != 0 {
			t.Errorf("Expected 0 tools, got %d", len(toolset.ToolsValue))
		}
	})

	t.Run("creates toolset with nil tools", func(t *testing.T) {
		toolset := NewToolset("nil-tools", true, nil)

		if toolset.ToolsValue != nil {
			t.Errorf("Expected nil ToolsValue, got %v", toolset.ToolsValue)
		}
	})
}

func TestToolset_GetName(t *testing.T) {
	t.Run("returns configured name", func(t *testing.T) {
		name := "my-toolset"
		toolset := NewToolset(name, true, nil)

		if toolset.GetName() != name {
			t.Errorf("Expected name %q, got %q", name, toolset.GetName())
		}
	})

	t.Run("returns empty string if configured", func(t *testing.T) {
		toolset := NewToolset("", true, nil)

		if toolset.GetName() != "" {
			t.Errorf("Expected empty name, got %q", toolset.GetName())
		}
	})
}

func TestToolset_IsEnabled(t *testing.T) {
	t.Run("returns true when enabled", func(t *testing.T) {
		toolset := NewToolset("enabled", true, nil)

		if !toolset.IsEnabled() {
			t.Error("Expected toolset to be enabled")
		}
	})

	t.Run("returns false when disabled", func(t *testing.T) {
		toolset := NewToolset("disabled", false, nil)

		if toolset.IsEnabled() {
			t.Error("Expected toolset to be disabled")
		}
	})

	t.Run("can toggle enabled state", func(t *testing.T) {
		toolset := NewToolset("toggle", true, nil)

		if !toolset.IsEnabled() {
			t.Error("Expected initially enabled")
		}

		toolset.EnabledValue = false

		if toolset.IsEnabled() {
			t.Error("Expected disabled after toggle")
		}
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

	if len(result) != len(tools) {
		t.Errorf("Expected %d tools, got %d", len(tools), len(result))
	}

	for i, tool := range result {
		if tool != tools[i] {
			t.Errorf("Tool at index %d doesn't match", i)
		}
	}
}

func TestToolset_GetTools_Disabled(t *testing.T) {
	tools := []toolsets.Tool{
		NewTool("tool1", true),
		NewTool("tool2", false),
	}
	toolset := NewToolset("disabled-toolset", false, tools)

	result := toolset.GetTools()

	if result == nil {
		t.Error("Expected non-nil slice, got nil")
	}

	if len(result) != 0 {
		t.Errorf("Expected empty slice when disabled, got %d tools", len(result))
	}
}

func TestToolset_GetTools_EmptyList(t *testing.T) {
	toolset := NewToolset("no-tools", true, []toolsets.Tool{})

	result := toolset.GetTools()

	if result == nil {
		t.Error("Expected non-nil slice, got nil")
	}

	if len(result) != 0 {
		t.Errorf("Expected 0 tools, got %d", len(result))
	}
}

func TestToolset_GetTools_NilTools(t *testing.T) {
	toolset := NewToolset("nil-tools", true, nil)

	result := toolset.GetTools()

	if result != nil {
		t.Errorf("Expected nil, got %v", result)
	}
}

func TestToolset_GetTools_ToggleState(t *testing.T) {
	tools := []toolsets.Tool{NewTool("tool1", true)}
	toolset := NewToolset("toggle-toolset", true, tools)

	// Initially enabled - should return tools
	result1 := toolset.GetTools()
	if len(result1) != 1 {
		t.Errorf("Expected 1 tool when enabled, got %d", len(result1))
	}

	// Disable - should return empty slice
	toolset.EnabledValue = false

	result2 := toolset.GetTools()
	if len(result2) != 0 {
		t.Errorf("Expected 0 tools when disabled, got %d", len(result2))
	}

	// Re-enable - should return tools again
	toolset.EnabledValue = true

	result3 := toolset.GetTools()
	if len(result3) != 1 {
		t.Errorf("Expected 1 tool when re-enabled, got %d", len(result3))
	}
}

func TestToolset_InterfaceCompliance(t *testing.T) {
	t.Run("implements toolsets.Toolset interface", func(*testing.T) {
		var _ toolsets.Toolset = (*Toolset)(nil)
	})
}

func TestToolset_AsInterface(t *testing.T) {
	var toolsetInstance toolsets.Toolset = NewToolset("interface-test", true, nil)

	if toolsetInstance.GetName() != "interface-test" {
		t.Errorf("Expected name 'interface-test', got %q", toolsetInstance.GetName())
	}

	if !toolsetInstance.IsEnabled() {
		t.Error("Expected toolset to be enabled")
	}

	tools := toolsetInstance.GetTools()
	if tools != nil {
		t.Errorf("Expected nil tools, got %v", tools)
	}
}

func TestToolset_EdgeCases(t *testing.T) {
	t.Run("toolset with single tool", func(t *testing.T) {
		tools := []toolsets.Tool{NewTool("only-tool", true)}
		toolset := NewToolset("single", true, tools)

		result := toolset.GetTools()
		if len(result) != 1 {
			t.Errorf("Expected 1 tool, got %d", len(result))
		}
	})

	t.Run("toolset with many tools", func(t *testing.T) {
		tools := make([]toolsets.Tool, 100)
		for i := range 100 {
			tools[i] = NewTool("tool", i%2 == 0)
		}

		toolset := NewToolset("many-tools", true, tools)

		result := toolset.GetTools()
		if len(result) != 100 {
			t.Errorf("Expected 100 tools, got %d", len(result))
		}
	})

	t.Run("toolset name with special characters", func(t *testing.T) {
		name := "tool-set_123!@#"
		toolset := NewToolset(name, true, nil)

		if toolset.GetName() != name {
			t.Errorf("Expected name %q, got %q", name, toolset.GetName())
		}
	})
}
