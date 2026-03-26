//go:build smoke

// Package smoke provides smoke test utilities for testing StackRox MCP server.
package smoke

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	// Timeouts for HTTP requests.
	tokenGenerationTimeout = 30 * time.Second
	pingTimeout            = 5 * time.Second
	maxSleepTime           = 30
)

// GenerateTokenRequest represents the request body for API token generation.
type GenerateTokenRequest struct {
	Name string `json:"name"`
	Role string `json:"role,omitempty"`
}

// GenerateTokenResponse represents the response from API token generation.
type GenerateTokenResponse struct {
	Token string `json:"token"`
}

// ClusterHealthResponse represents the response from /v1/clusters endpoint.
type ClusterHealthResponse struct {
	Clusters []struct {
		HealthStatus struct {
			OverallHealthStatus string `json:"overallHealthStatus"`
		} `json:"healthStatus"`
	} `json:"clusters"`
}

// GenerateAPIToken generates an API token using basic authentication.
// It calls the /v1/apitokens/generate endpoint with admin credentials.
func GenerateAPIToken(endpoint, password string) (string, error) {
	tokenReq := GenerateTokenRequest{
		Name: "smoke-test-token",
		Role: "Admin",
	}

	reqBody, err := json.Marshal(tokenReq)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("https://%s/v1/apitokens/generate", endpoint)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, bytes.NewReader(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth("admin", password)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: tokenGenerationTimeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec // Testing with self-signed certificates
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token generation failed (status %d): %s", resp.StatusCode, string(body))
	}

	var tokenResp GenerateTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if tokenResp.Token == "" {
		return "", errors.New("received empty token in response")
	}

	return tokenResp.Token, nil
}

// WaitForCentralReady polls the /v1/ping endpoint until Central is ready.
func WaitForCentralReady(endpoint, password string, maxAttempts int) error {
	url := fmt.Sprintf("https://%s/v1/ping", endpoint)

	client := &http.Client{
		Timeout: pingTimeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec // Testing with self-signed certificates
		},
	}

	for attempt := range maxAttempts {
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		req.SetBasicAuth("admin", password)

		resp, err := client.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			_ = resp.Body.Close()

			return nil
		}

		if resp != nil {
			_ = resp.Body.Close()
		}

		// Exponential backoff: 2, 4, 8, 16... seconds (max 30)
		sleepTime := min(1<<attempt, maxSleepTime)

		time.Sleep(time.Duration(sleepTime) * time.Second)
	}

	return fmt.Errorf("central did not become ready after %d attempts", maxAttempts)
}

// IsClusterHealthy checks if the first cluster registered with Central is in HEALTHY status.
// Returns true if healthy, false otherwise.
func IsClusterHealthy(endpoint, password string) bool {
	url := fmt.Sprintf("https://%s/v1/clusters", endpoint)

	client := &http.Client{
		Timeout: pingTimeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec // Testing with self-signed certificates
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

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return false
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false
	}

	var healthResp ClusterHealthResponse
	if err := json.Unmarshal(body, &healthResp); err != nil {
		return false
	}

	return len(healthResp.Clusters) > 0 &&
		healthResp.Clusters[0].HealthStatus.OverallHealthStatus == "HEALTHY"
}
