package testutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPortForTest_ReturnsPortInValidRange(t *testing.T) {
	port := GetPortForTest(t)

	assert.GreaterOrEqual(t, port, minPort)
	assert.Less(t, port, maxPort)
}

func TestGetPortForTest_ReturnsDeterministicPort(t *testing.T) {
	port1 := GetPortForTest(t)
	port2 := GetPortForTest(t)

	assert.Equal(t, port1, port2)
}

func TestGetPortForTest_DifferentTestNamesGetDifferentPorts(t *testing.T) {
	ports := make(map[int]bool)

	// Create subtests with different names
	for range 10 {
		t.Run("subtest", func(t *testing.T) {
			port := GetPortForTest(t)
			ports[port] = true
		})
	}

	assert.Len(t, ports, 10)
}

func TestPortRangeConstants(t *testing.T) {
	assert.Greater(t, minPort, 1024)
	assert.LessOrEqual(t, maxPort, 65536)
	assert.Equal(t, maxPort-minPort, 50000)
}
