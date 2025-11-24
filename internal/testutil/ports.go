package testutil

import (
	"hash/fnv"
	"testing"
)

const (
	minPort = 10000
	maxPort = 60000
)

// GetPortForTest returns a deterministic port number based on the test name.
// This ensures that each test gets a unique, reproducible port number for parallel execution.
// The port is calculated by hashing the test name and mapping it to a safe range.
func GetPortForTest(t *testing.T) int {
	t.Helper()

	// Hash the test name using FNV-1a
	h := fnv.New32a()
	_, _ = h.Write([]byte(t.Name()))
	hash := h.Sum32()

	// Map the hash to the port range [minPort, maxPort)
	portRange := maxPort - minPort
	port := minPort + int(hash%uint32(portRange))

	return port
}
