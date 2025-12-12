package testutil

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWaitForServerReady_Immediate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	err := WaitForServerReady(server.URL, 1*time.Second)
	assert.NoError(t, err)
}

func TestWaitForServerReady_AfterDelay(t *testing.T) {
	var ready atomic.Bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if ready.Load() {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
	}))
	defer server.Close()

	// Make server ready after a short delay
	go func() {
		time.Sleep(200 * time.Millisecond)
		ready.Store(true)
	}()

	err := WaitForServerReady(server.URL, 2*time.Second)
	assert.NoError(t, err)
}

func TestWaitForServerReady_NeverReady(t *testing.T) {
	// Use an address that won't have a server
	err := WaitForServerReady("http://localhost:59999", 300*time.Millisecond)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "did not become ready")
}

func TestWaitForServerReady_RespectsTimeout(t *testing.T) {
	start := time.Now()
	timeout := 300 * time.Millisecond

	err := WaitForServerReady("http://localhost:59998", timeout)

	elapsed := time.Since(start)

	require.Error(t, err)

	// Allow some margin for timing (timeout + 200ms for final attempt)
	maxExpected := timeout + 300*time.Millisecond
	assert.LessOrEqual(t, elapsed, maxExpected)
	assert.GreaterOrEqual(t, elapsed, timeout)
}

func TestWaitForServerReady_ErrorMessage(t *testing.T) {
	timeout := 250 * time.Millisecond

	err := WaitForServerReady("http://localhost:59997", timeout)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "250ms")
}

func TestWaitForServerReady_StatusCodes(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
	}{
		{"200 OK", http.StatusOK},
		{"201 Created", http.StatusCreated},
		{"404 Not Found", http.StatusNotFound},
		{"500 Internal Server Error", http.StatusInternalServerError},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(testCase.statusCode)
			}))
			defer server.Close()

			err := WaitForServerReady(server.URL, 1*time.Second)
			assert.NoError(t, err)
		})
	}
}

func TestWaitForServerReady_InvalidURL(t *testing.T) {
	err := WaitForServerReady("http://invalid-host-that-does-not-exist.local:12345", 200*time.Millisecond)
	assert.Error(t, err)
}

func TestWaitForServerReadyIntegration(t *testing.T) {
	t.Run("simulates real server startup scenario", func(t *testing.T) {
		var server *httptest.Server

		serverReady := make(chan struct{})

		// Simulate server starting up in background
		go func() {
			time.Sleep(150 * time.Millisecond)

			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			close(serverReady)
		}()

		// Wait for server to be created
		<-serverReady

		defer server.Close()

		err := WaitForServerReady(server.URL, 2*time.Second)
		assert.NoError(t, err)
	})
}

func ExampleWaitForServerReady() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
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
