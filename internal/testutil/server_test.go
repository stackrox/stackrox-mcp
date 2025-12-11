package testutil

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestWaitForServerReady(t *testing.T) {
	t.Run("returns nil when server is ready immediately", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		err := WaitForServerReady(server.URL, 1*time.Second)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
	})

	t.Run("returns nil when server becomes ready after a delay", func(t *testing.T) {
		ready := false
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if ready {
				w.WriteHeader(http.StatusOK)
			} else {
				w.WriteHeader(http.StatusServiceUnavailable)
			}
		}))
		defer server.Close()

		// Make server ready after a short delay
		go func() {
			time.Sleep(200 * time.Millisecond)
			ready = true
		}()

		err := WaitForServerReady(server.URL, 2*time.Second)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
	})

	t.Run("returns error when server never becomes ready", func(t *testing.T) {
		// Use an address that won't have a server
		err := WaitForServerReady("http://localhost:59999", 300*time.Millisecond)
		if err == nil {
			t.Error("Expected error when server is not ready, got nil")
		}

		expectedSubstring := "did not become ready"
		if !strings.Contains(err.Error(), expectedSubstring) {
			t.Errorf("Expected error to contain %q, got: %v", expectedSubstring, err)
		}
	})

	t.Run("respects timeout parameter", func(t *testing.T) {
		start := time.Now()
		timeout := 300 * time.Millisecond

		err := WaitForServerReady("http://localhost:59998", timeout)

		elapsed := time.Since(start)

		if err == nil {
			t.Error("Expected error when server is not ready")
		}

		// Allow some margin for timing (timeout + 200ms for final attempt)
		maxExpected := timeout + 300*time.Millisecond
		if elapsed > maxExpected {
			t.Errorf("Expected to wait approximately %v, but waited %v", timeout, elapsed)
		}

		if elapsed < timeout {
			t.Errorf("Expected to wait at least %v, but only waited %v", timeout, elapsed)
		}
	})

	t.Run("error message includes timeout duration", func(t *testing.T) {
		timeout := 250 * time.Millisecond
		err := WaitForServerReady("http://localhost:59997", timeout)

		if err == nil {
			t.Fatal("Expected error, got nil")
		}

		if !strings.Contains(err.Error(), "250ms") {
			t.Errorf("Expected error message to include timeout duration, got: %v", err)
		}
	})

	t.Run("works with different HTTP status codes", func(t *testing.T) {
		tests := []struct {
			name       string
			statusCode int
		}{
			{"200 OK", http.StatusOK},
			{"201 Created", http.StatusCreated},
			{"404 Not Found", http.StatusNotFound},
			{"500 Internal Server Error", http.StatusInternalServerError},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(tt.statusCode)
				}))
				defer server.Close()

				err := WaitForServerReady(server.URL, 1*time.Second)
				if err != nil {
					t.Errorf("Expected no error for status %d, got: %v", tt.statusCode, err)
				}
			})
		}
	})

	t.Run("handles invalid URL gracefully", func(t *testing.T) {
		err := WaitForServerReady("http://invalid-host-that-does-not-exist.local:12345", 200*time.Millisecond)
		if err == nil {
			t.Error("Expected error for invalid URL, got nil")
		}
	})
}

func TestWaitForServerReadyIntegration(t *testing.T) {
	t.Run("simulates real server startup scenario", func(t *testing.T) {
		var server *httptest.Server
		serverReady := make(chan struct{})

		// Simulate server starting up in background
		go func() {
			time.Sleep(150 * time.Millisecond)
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))
			close(serverReady)
		}()

		// Wait for server to be created
		<-serverReady
		defer server.Close()

		err := WaitForServerReady(server.URL, 2*time.Second)
		if err != nil {
			t.Errorf("Expected server to become ready, got error: %v", err)
		}
	})
}

func ExampleWaitForServerReady() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	err := WaitForServerReady(server.URL, 5*time.Second)
	if err != nil {
		fmt.Printf("Server not ready: %v\n", err)
		return
	}

	fmt.Println("Server is ready!")
	// Output: Server is ready!
}
