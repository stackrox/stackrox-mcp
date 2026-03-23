//go:build integration

package integration

import (
	"context"
	"testing"

	"github.com/stackrox/stackrox-mcp/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegration_ListTools verifies that all expected tools are registered.
func TestIntegration_ListTools(t *testing.T) {
	client := testutil.SetupInitializedClient(t, testutil.CreateIntegrationMCPClient)

	ctx := context.Background()
	result, err := client.ListTools(ctx)
	require.NoError(t, err)

	// Verify we have tools registered
	assert.NotEmpty(t, result.Tools, "should have tools registered")

	// Check for specific tools we expect
	toolNames := make([]string, 0, len(result.Tools))
	for _, tool := range result.Tools {
		toolNames = append(toolNames, tool.Name)
	}

	assert.Contains(t, toolNames, "get_deployments_for_cve", "should have get_deployments_for_cve tool")
	assert.Contains(t, toolNames, "list_clusters", "should have list_clusters tool")
}

// TestIntegration_ToolCalls tests successful tool calls using table-driven tests.
func TestIntegration_ToolCalls(t *testing.T) {
	tests := map[string]struct {
		toolName     string
		args         map[string]any
		expectedJSON string // expected JSON response for exact comparison
	}{
		"get_deployments_for_cve with Log4Shell": {
			toolName:     "get_deployments_for_cve",
			args:         map[string]any{"cveName": Log4ShellFixture.CVEName},
			expectedJSON: Log4ShellFixture.ExpectedJSON,
		},
		"get_deployments_for_cve with non-existent CVE": {
			toolName:     "get_deployments_for_cve",
			args:         map[string]any{"cveName": "CVE-9999-99999"},
			expectedJSON: EmptyDeploymentsJSON,
		},
		"list_clusters": {
			toolName:     "list_clusters",
			args:         map[string]any{},
			expectedJSON: AllClustersFixture.ExpectedJSON,
		},
		"get_clusters_with_orchestrator_cve": {
			toolName:     "get_clusters_with_orchestrator_cve",
			args:         map[string]any{"cveName": "CVE-2099-00001"},
			expectedJSON: EmptyClustersForCVEJSON,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			client := testutil.SetupInitializedClient(t, testutil.CreateIntegrationMCPClient)
			result := testutil.CallToolAndGetResult(t, client, tt.toolName, tt.args)

			responseText := testutil.GetTextContent(t, result)
			assert.JSONEq(t, tt.expectedJSON, responseText)
		})
	}
}

// TestIntegration_ToolCallErrors tests error handling using table-driven tests.
func TestIntegration_ToolCallErrors(t *testing.T) {
	tests := map[string]struct {
		toolName         string
		args             map[string]any
		expectedErrorMsg string
	}{
		"get_deployments_for_cve missing CVE name": {
			toolName:         "get_deployments_for_cve",
			args:             map[string]any{},
			expectedErrorMsg: "cveName",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			client := testutil.SetupInitializedClient(t, testutil.CreateIntegrationMCPClient)

			ctx := context.Background()
			_, err := client.CallTool(ctx, tt.toolName, tt.args)

			// Validation errors are returned as protocol errors, not tool errors
			require.Error(t, err, "should receive protocol error for invalid params")
			assert.Contains(t, err.Error(), tt.expectedErrorMsg)
		})
	}
}
