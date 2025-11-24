package testutil

import (
	"fmt"
	"net/http"
	"time"
)

const timeoutDuration = 100 * time.Millisecond

// WaitForServerReady polls the server until it's ready to accept connections.
// This function is useful for integration tests where you need to wait for
// a server to start before making requests to it.
func WaitForServerReady(address string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	client := &http.Client{Timeout: timeoutDuration}

	for time.Now().Before(deadline) {
		//nolint:noctx
		resp, err := client.Get(address)
		if err == nil {
			_ = resp.Body.Close()

			return nil
		}

		time.Sleep(timeoutDuration)
	}

	return fmt.Errorf("server did not become ready within %v", timeout)
}
