//go:build integration

package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stackrox/stackrox-mcp/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	binaryPath string
	configPath string
)

// TestMain ensures WireMock is running and builds the MCP binary before any integration tests.
func TestMain(m *testing.M) {
	// Check WireMock is running
	if err := testutil.WaitForWireMockReady(10 * time.Second); err != nil {
		fmt.Fprintf(os.Stderr, "Integration tests require WireMock.\n")
		fmt.Fprintf(os.Stderr, "Start with: make mock-start\n")
		os.Exit(1)
	}

	// Build the MCP binary for testing
	tmpDir, err := os.MkdirTemp("", "mcp-integration-test-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create temp directory: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)

	binaryPath = filepath.Join(tmpDir, "stackrox-mcp")

	// Build the binary
	if err := buildMCPBinary(binaryPath); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to build MCP binary: %v\n", err)
		os.Exit(1)
	}

	// Create test config file
	configPath = filepath.Join(tmpDir, "config.yaml")
	if err := createTestConfig(configPath); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create test config: %v\n", err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

// buildMCPBinary builds the stackrox-mcp binary for testing.
func buildMCPBinary(outputPath string) error {
	// Run: go build -o <outputPath> ../cmd/stackrox-mcp
	// The .. is because integration tests run from integration/ directory
	cmd := fmt.Sprintf("go build -o %s ../cmd/stackrox-mcp", outputPath)
	if output, err := testutil.RunCommand(cmd); err != nil {
		return fmt.Errorf("build failed: %v\nOutput: %s", err, output)
	}
	return nil
}

// createTestConfig creates a test configuration file for the MCP server.
func createTestConfig(configPath string) error {
	config := `
central:
  url: localhost:8081
  auth_type: static
  api_token: test-token-admin
  insecure_skip_tls_verify: true

server:
  type: stdio

tools:
  vulnerability:
    enabled: true
  config_manager:
    enabled: true
`
	return os.WriteFile(configPath, []byte(config), 0600)
}

// TestIntegration_Initialize verifies the MCP initialize handshake.
func TestIntegration_Initialize(t *testing.T) {
	client, err := testutil.NewMCPClient(t, binaryPath, configPath)
	require.NoError(t, err, "Failed to create MCP client")
	defer client.Close()

	resp, err := client.Initialize()
	require.NoError(t, err)

	testutil.RequireNoError(t, resp)

	var result map[string]any
	testutil.UnmarshalResult(t, resp, &result)

	assert.Contains(t, result, "protocolVersion", "initialize response should include protocolVersion")
	assert.Contains(t, result, "serverInfo", "initialize response should include serverInfo")
	assert.Contains(t, result, "capabilities", "initialize response should include capabilities")
}

// TestIntegration_ListTools verifies that all expected tools are registered.
func TestIntegration_ListTools(t *testing.T) {
	client, err := testutil.NewMCPClient(t, binaryPath, configPath)
	require.NoError(t, err, "Failed to create MCP client")
	defer client.Close()

	// Initialize first
	initResp, err := client.Initialize()
	require.NoError(t, err)
	testutil.RequireNoError(t, initResp)

	// List tools
	resp, err := client.ListTools()
	require.NoError(t, err)
	testutil.RequireNoError(t, resp)

	var result struct {
		Tools []map[string]any `json:"tools"`
	}
	testutil.UnmarshalResult(t, resp, &result)

	// Verify we have tools registered
	assert.NotEmpty(t, result.Tools, "should have tools registered")

	// Check for specific tools we expect
	toolNames := make([]string, 0, len(result.Tools))
	for _, tool := range result.Tools {
		if name, ok := tool["name"].(string); ok {
			toolNames = append(toolNames, name)
		}
	}

	assert.Contains(t, toolNames, "get_deployments_for_cve", "should have get_deployments_for_cve tool")
	assert.Contains(t, toolNames, "list_clusters", "should have list_clusters tool")
}

// TestIntegration_GetDeploymentsForCVE_Log4Shell tests successful retrieval of deployments for Log4Shell CVE.
func TestIntegration_GetDeploymentsForCVE_Log4Shell(t *testing.T) {

	client, err := testutil.NewMCPClient(t, binaryPath, configPath)
	require.NoError(t, err, "Failed to create MCP client")
	defer client.Close()

	// Initialize
	initResp, err := client.Initialize()
	require.NoError(t, err)
	testutil.RequireNoError(t, initResp)

	// Call get_deployments_for_cve with Log4Shell CVE
	args := map[string]any{
		"cveName": Log4ShellFixture.CVEName,
	}
	resp, err := client.CallTool("get_deployments_for_cve", args)
	require.NoError(t, err)
	testutil.RequireNoError(t, resp)

	var result struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	testutil.UnmarshalResult(t, resp, &result)

	// Verify we got content
	require.NotEmpty(t, result.Content, "should have content in response")

	// Verify deployment names appear in the response text
	responseText := result.Content[0].Text
	for _, deploymentName := range Log4ShellFixture.DeploymentNames {
		assert.Contains(t, responseText, deploymentName,
			"response should contain deployment %s", deploymentName)
	}
}

// TestIntegration_GetDeploymentsForCVE_NotFound tests handling of non-existent CVE.
func TestIntegration_GetDeploymentsForCVE_NotFound(t *testing.T) {
	client, err := testutil.NewMCPClient(t, binaryPath, configPath)
	require.NoError(t, err, "Failed to create MCP client")
	defer client.Close()

	// Initialize
	initResp, err := client.Initialize()
	require.NoError(t, err)
	testutil.RequireNoError(t, initResp)

	// Call with non-existent CVE
	args := map[string]any{
		"cveName": "CVE-9999-99999",
	}
	resp, err := client.CallTool("get_deployments_for_cve", args)
	require.NoError(t, err)
	testutil.RequireNoError(t, resp)

	var result struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	testutil.UnmarshalResult(t, resp, &result)

	// Should get a response with empty deployments
	require.NotEmpty(t, result.Content, "should have content in response")
	responseText := result.Content[0].Text
	assert.Contains(t, responseText, "\"deployments\":[]", "response should contain empty deployments array")
}

// TestIntegration_GetDeploymentsForCVE_InvalidInput tests handling of missing required parameter.
func TestIntegration_GetDeploymentsForCVE_InvalidInput(t *testing.T) {
	client, err := testutil.NewMCPClient(t, binaryPath, configPath)
	require.NoError(t, err, "Failed to create MCP client")
	defer client.Close()

	// Initialize
	initResp, err := client.Initialize()
	require.NoError(t, err)
	testutil.RequireNoError(t, initResp)

	// Call without required cveName parameter
	args := map[string]any{}
	resp, err := client.CallTool("get_deployments_for_cve", args)
	require.NoError(t, err)

	// Should get an MCP error response
	assert.NotNil(t, resp.Error, "should receive error for missing required parameter")
	assert.Contains(t, resp.Error.Message, "cveName", "error message should mention missing cveName parameter")
}

// TestIntegration_ListClusters tests listing all clusters.
func TestIntegration_ListClusters(t *testing.T) {
	client, err := testutil.NewMCPClient(t, binaryPath, configPath)
	require.NoError(t, err, "Failed to create MCP client")
	defer client.Close()

	// Initialize
	initResp, err := client.Initialize()
	require.NoError(t, err)
	testutil.RequireNoError(t, initResp)

	// Call list_clusters
	resp, err := client.CallTool("list_clusters", map[string]any{})
	require.NoError(t, err)
	testutil.RequireNoError(t, resp)

	var result struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	testutil.UnmarshalResult(t, resp, &result)

	// Verify we got content
	require.NotEmpty(t, result.Content, "should have content in response")

	// Verify cluster names appear in the response
	responseText := result.Content[0].Text
	for _, clusterName := range AllClustersFixture.ClusterNames {
		assert.Contains(t, responseText, clusterName,
			"response should contain cluster %s", clusterName)
	}
}

// TestIntegration_GetClustersWithOrchestratorCVE tests getting clusters with orchestrator CVE.
func TestIntegration_GetClustersWithOrchestratorCVE(t *testing.T) {
	client, err := testutil.NewMCPClient(t, binaryPath, configPath)
	require.NoError(t, err, "Failed to create MCP client")
	defer client.Close()

	// Initialize
	initResp, err := client.Initialize()
	require.NoError(t, err)
	testutil.RequireNoError(t, initResp)

	// Call get_clusters_with_orchestrator_cve
	args := map[string]any{
		"cveName": "CVE-2099-00001",
	}
	resp, err := client.CallTool("get_clusters_with_orchestrator_cve", args)
	require.NoError(t, err)
	testutil.RequireNoError(t, resp)

	var result struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	testutil.UnmarshalResult(t, resp, &result)

	// Verify we got content
	require.NotEmpty(t, result.Content, "should have content in response")
	assert.NotEmpty(t, result.Content[0].Text, "response text should not be empty")
}
