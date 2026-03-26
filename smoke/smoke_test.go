//go:build smoke

package smoke

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"testing"
	"time"

	"github.com/stackrox/stackrox-mcp/internal/app"
	"github.com/stackrox/stackrox-mcp/internal/config"
	"github.com/stackrox/stackrox-mcp/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func waitForImageScan(t *testing.T, client *testutil.MCPTestClient, cveName string) {
	t.Helper()

	// Skip image scan wait in CI - scanner is too resource-intensive for GitHub Actions
	if os.Getenv("CI") == "true" {
		t.Log("Skipping image scan wait in CI environment")
		return
	}

	assert.Eventually(t, func() bool {
		ctx := context.Background()
		result, err := client.CallTool(ctx, "get_deployments_for_cve", map[string]any{
			"cveName": cveName,
		})

		if err != nil || result.IsError {
			return false
		}

		responseText := testutil.GetTextContent(t, result)
		var data struct {
			Deployments []any `json:"deployments"`
		}

		if err := json.Unmarshal([]byte(responseText), &data); err != nil {
			return false
		}

		if len(data.Deployments) > 0 {
			t.Logf("Image scan completed, found %d deployment(s) with CVE %s", len(data.Deployments), cveName)
			return true
		}

		t.Logf("Waiting for image scan (CVE: %s)...", cveName)
		return false
	}, 10*time.Minute, 5*time.Second, "Image scan did not complete for CVE %s", cveName)
}

func TestSmoke_RealCluster(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping smoke test in short mode")
	}

	endpoint := os.Getenv("ROX_ENDPOINT")
	apiToken := os.Getenv("ROX_API_TOKEN")
	password := os.Getenv("ROX_PASSWORD")

	if endpoint == "" {
		t.Fatal("ROX_ENDPOINT environment variable must be set")
	}

	// Generate token if password provided but no token
	if apiToken == "" && password != "" {
		t.Log("No API token provided, generating one using password...")

		// Wait for Central to be ready
		if err := WaitForCentralReady(endpoint, password, 12); err != nil {
			t.Fatalf("Failed waiting for Central: %v", err)
		}
		t.Log("Central API is ready")

		// Generate token
		token, err := GenerateAPIToken(endpoint, password)
		if err != nil {
			t.Fatalf("Failed to generate API token: %v", err)
		}
		apiToken = token
		t.Log("Successfully generated API token")
	}

	if apiToken == "" {
		t.Fatal("Either ROX_API_TOKEN or ROX_PASSWORD must be set")
	}

	client := createSmokeTestClient(t, endpoint, apiToken)

	// nginx:1.14 has CVE-2019-9511 (HTTP/2 vulnerabilities)
	waitForImageScan(t, client, "CVE-2019-9511")

	tests := map[string]struct {
		toolName     string
		args         map[string]any
		validateFunc func(*testing.T, string)
	}{
		"list_clusters": {
			toolName: "list_clusters",
			args:     map[string]any{},
			validateFunc: func(t *testing.T, result string) {
				t.Helper()
				var data struct {
					Clusters []struct {
						Name string `json:"name"`
					} `json:"clusters"`
				}
				require.NoError(t, json.Unmarshal([]byte(result), &data))
				assert.NotEmpty(t, data.Clusters, "should have at least one cluster")
				t.Logf("Found %d cluster(s)", len(data.Clusters))
			},
		},
		"get_deployments_for_cve with known CVE": {
			toolName: "get_deployments_for_cve",
			args:     map[string]any{"cveName": "CVE-2019-11043"},
			validateFunc: func(t *testing.T, result string) {
				t.Helper()
				var data struct {
					Deployments []struct {
						Name      string `json:"name"`
						Namespace string `json:"namespace"`
					} `json:"deployments"`
				}
				require.NoError(t, json.Unmarshal([]byte(result), &data))

				if len(data.Deployments) == 0 {
					t.Log("Warning: No deployments found with CVE. Deployment may not be scanned yet.")
				} else {
					t.Logf("Found %d deployment(s) with CVE", len(data.Deployments))
				}
			},
		},
		"get_deployments_for_cve with non-existent CVE": {
			toolName: "get_deployments_for_cve",
			args:     map[string]any{"cveName": "CVE-9999-99999"},
			validateFunc: func(t *testing.T, result string) {
				t.Helper()
				var data struct {
					Deployments []any `json:"deployments"`
				}
				require.NoError(t, json.Unmarshal([]byte(result), &data))
				assert.Empty(t, data.Deployments, "should have no deployments for non-existent CVE")
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result := testutil.CallToolAndGetResult(t, client, tt.toolName, tt.args)
			responseText := testutil.GetTextContent(t, result)
			tt.validateFunc(t, responseText)
		})
	}
}

func createSmokeTestClient(t *testing.T, endpoint, apiToken string) *testutil.MCPTestClient {
	t.Helper()

	cfg := &config.Config{
		Central: config.CentralConfig{
			URL:                   endpoint,
			AuthType:              "static",
			APIToken:              apiToken,
			InsecureSkipTLSVerify: true,
			RequestTimeout:        30 * time.Second,
			MaxRetries:            3,
			InitialBackoff:        time.Second,
			MaxBackoff:            10 * time.Second,
		},
		Server: config.ServerConfig{
			Type: "stdio",
		},
		Tools: config.ToolsConfig{
			Vulnerability: config.ToolsetVulnerabilityConfig{
				Enabled: true,
			},
			ConfigManager: config.ToolConfigManagerConfig{
				Enabled: true,
			},
		},
	}

	runFunc := func(ctx context.Context, stdin io.ReadCloser, stdout io.WriteCloser) error {
		return app.Run(ctx, cfg, stdin, stdout)
	}

	client, err := testutil.NewMCPTestClient(t, runFunc)
	require.NoError(t, err, "Failed to create MCP client")
	t.Cleanup(func() { client.Close() })

	return client
}
