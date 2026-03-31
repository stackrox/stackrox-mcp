//go:build smoke

package smoke

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	tokenGenerationTimeout = 30 * time.Second
	pingTimeout            = 5 * time.Second
)

type generateTokenRequest struct {
	Name string `json:"name"`
	Role string `json:"role,omitempty"`
}

type generateTokenResponse struct {
	Token string `json:"token"`
}

type clusterHealthResponse struct {
	Clusters []struct {
		HealthStatus struct {
			OverallHealthStatus string `json:"overallHealthStatus"`
		} `json:"healthStatus"`
	} `json:"clusters"`
}

// GenerateAPIToken generates an API token using basic authentication.
func GenerateAPIToken(t *testing.T, endpoint, password string) string {
	t.Helper()

	tokenReq := generateTokenRequest{
		Name: "smoke-test-token",
		Role: "Admin",
	}

	reqBody, err := json.Marshal(tokenReq)
	require.NoError(t, err, "Failed to marshal token request")

	url := fmt.Sprintf("https://%s/v1/apitokens/generate", endpoint)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, bytes.NewReader(reqBody))
	require.NoError(t, err, "Failed to create request")

	req.SetBasicAuth("admin", password)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: tokenGenerationTimeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
		},
	}

	resp, err := client.Do(req)
	require.NoError(t, err, "Failed to make token generation request")
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "Failed to read response body")

	require.Equal(t, http.StatusOK, resp.StatusCode, "Token generation failed: %s", string(body))

	var tokenResp generateTokenResponse
	require.NoError(t, json.Unmarshal(body, &tokenResp), "Failed to parse token response")
	require.NotEmpty(t, tokenResp.Token, "Received empty token in response")

	return tokenResp.Token
}

// WaitForCentralReady waits for Central API to be ready by polling /v1/ping.
func WaitForCentralReady(t *testing.T, endpoint, password string) {
	t.Helper()

	require.Eventually(t, func() bool {
		return isCentralReady(endpoint, password)
	}, 2*time.Minute, 2*time.Second, "Central API did not become ready")
}

// isCentralReady checks if Central API responds to /v1/ping.
func isCentralReady(endpoint, password string) bool {
	url := fmt.Sprintf("https://%s/v1/ping", endpoint)

	client := &http.Client{
		Timeout: pingTimeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
		},
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return false
	}

	req.SetBasicAuth("admin", password)

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer func() { _ = resp.Body.Close() }()

	return resp.StatusCode == http.StatusOK
}

// IsClusterHealthy checks if the first cluster registered with Central is in HEALTHY status.
func IsClusterHealthy(endpoint, apiToken string) bool {
	url := fmt.Sprintf("https://%s/v1/clusters", endpoint)

	client := &http.Client{
		Timeout: pingTimeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
		},
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return false
	}

	req.Header.Set("Authorization", "Bearer "+apiToken)

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return false
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false
	}

	var healthResp clusterHealthResponse
	if err := json.Unmarshal(body, &healthResp); err != nil {
		return false
	}

	return len(healthResp.Clusters) > 0 &&
		healthResp.Clusters[0].HealthStatus.OverallHealthStatus == "HEALTHY"
}
