package mock

import (
	"testing"

	"github.com/stackrox/stackrox-mcp/internal/toolsets"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToolset_GetName(t *testing.T) {
	toolset := NewToolset("my-toolset", true, nil)

	require.NotNil(t, toolset)
	assert.Equal(t, "my-toolset", toolset.GetName())
}

func TestToolset_IsEnabled_True(t *testing.T) {
	tools := []toolsets.Tool{&Tool{NameValue: "test-tool"}}
	toolset := NewToolset("test", true, tools)

	require.NotNil(t, toolset)
	assert.True(t, toolset.IsEnabled())
	assert.Equal(t, tools, toolset.GetTools(), "Should return tools when enabled")
}

func TestToolset_IsEnabled_False(t *testing.T) {
	tools := []toolsets.Tool{&Tool{NameValue: "test-tool"}}
	toolset := NewToolset("test", false, tools)

	require.NotNil(t, toolset)
	assert.False(t, toolset.IsEnabled())
	assert.Empty(t, toolset.GetTools(), "Should return empty slice when disabled")
}
