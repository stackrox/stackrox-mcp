package testutil

import (
	"testing"
)

func TestGetPortForTest(t *testing.T) {
	t.Run("returns port in valid range", func(t *testing.T) {
		port := GetPortForTest(t)

		if port < minPort || port >= maxPort {
			t.Errorf("Port %d is outside valid range [%d, %d)", port, minPort, maxPort)
		}
	})

	t.Run("returns deterministic port for same test name", func(t *testing.T) {
		port1 := GetPortForTest(t)
		port2 := GetPortForTest(t)

		if port1 != port2 {
			t.Errorf("Expected same port for same test name, got %d and %d", port1, port2)
		}
	})

	t.Run("different test names get different ports", func(t *testing.T) {
		ports := make(map[int]bool)
		testCount := 0

		// Create subtests with different names
		for range 10 {
			t.Run("subtest", func(t *testing.T) {
				port := GetPortForTest(t)
				ports[port] = true
				testCount++
			})
		}

		// We should have different ports for different test paths
		if len(ports) < 2 {
			t.Errorf("Expected multiple different ports, got %d unique ports from %d tests", len(ports), testCount)
		}
	})

	t.Run("port is within safe range avoiding privileged ports", func(t *testing.T) {
		port := GetPortForTest(t)

		if port < 1024 {
			t.Errorf("Port %d is in privileged range (< 1024)", port)
		}

		if port >= 65536 {
			t.Errorf("Port %d exceeds maximum valid port number", port)
		}
	})
}

func TestPortRangeConstants(t *testing.T) {
	t.Run("minPort is above privileged range", func(t *testing.T) {
		if minPort < 1024 {
			t.Errorf("minPort %d should be above 1024 to avoid privileged ports", minPort)
		}
	})

	t.Run("maxPort is below max valid port", func(t *testing.T) {
		if maxPort > 65536 {
			t.Errorf("maxPort %d should be below 65536", maxPort)
		}
	})

	t.Run("port range is reasonable", func(t *testing.T) {
		portRange := maxPort - minPort
		if portRange < 1000 {
			t.Errorf("Port range %d is too small, may cause collisions", portRange)
		}
	})
}
