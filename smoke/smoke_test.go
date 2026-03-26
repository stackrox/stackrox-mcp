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

func TestSmoke_RealCluster(t *testing.T) {
	endpoint := os.Getenv("ROX_ENDPOINT")
	apiToken := os.Getenv("ROX_API_TOKEN")
	password := os.Getenv("ROX_PASSWORD")

	require.NotEmpty(t, endpoint, "ROX_ENDPOINT environment variable must be set")

	// Generate token if password provided but no token
	if apiToken == "" && password != "" {
		t.Log("No API token provided, generating one using password...")

		// Wait for Central to be ready
		err := WaitForCentralReady(endpoint, password, 12)
		require.NoError(t, err, "Failed waiting for Central")
		t.Log("Central API is ready")

		// Generate token
		token, err := GenerateAPIToken(endpoint, password)
		require.NoError(t, err, "Failed to generate API token")
		apiToken = token
		t.Log("Successfully generated API token")
	}

	require.NotEmpty(t, apiToken, "Either ROX_API_TOKEN or ROX_PASSWORD must be set")

	// Wait for cluster to be registered and healthy
	assert.Eventually(t, func() bool {
		healthy := IsClusterHealthy(endpoint, password)
		if !healthy {
			t.Log("Waiting for cluster to be registered and healthy...")
		}
		return healthy
	}, 6*time.Minute, 2*time.Second, "Cluster did not become healthy")
	t.Log("Cluster is healthy and ready for testing")

	client := createSmokeTestClient(t, endpoint, apiToken)

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
