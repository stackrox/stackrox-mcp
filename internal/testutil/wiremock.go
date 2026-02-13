package testutil

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"time"
)

// WaitForWireMockReady polls WireMock's admin API until it's ready or timeout occurs.
// Returns nil if WireMock is ready, error otherwise.
func WaitForWireMockReady(timeout time.Duration) error {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		Timeout: 2 * time.Second,
	}

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := client.Get("https://localhost:8081/__admin/mappings")
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return nil
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(100 * time.Millisecond)
	}

	return fmt.Errorf("WireMock not ready after %v. Start with: make mock-start", timeout)
}

// IsWireMockRunning checks if WireMock is currently running.
func IsWireMockRunning() bool {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		Timeout: 2 * time.Second,
	}

	resp, err := client.Get("https://localhost:8081/__admin/mappings")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}
