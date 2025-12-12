package testutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPortForTest(t *testing.T) {
	t.Run("returns port in valid range", func(t *testing.T) {
		port := GetPortForTest(t)

		assert.GreaterOrEqual(t, port, minPort)
		assert.Less(t, port, maxPort)
	})

	t.Run("returns deterministic port for same test name", func(t *testing.T) {
		port1 := GetPortForTest(t)
		port2 := GetPortForTest(t)

		assert.Equal(t, port1, port2)
	})

	t.Run("different test names get different ports", func(t *testing.T) {
		ports := make(map[int]bool)

		// Create subtests with different names
		for range 10 {
			t.Run("subtest", func(t *testing.T) {
				port := GetPortForTest(t)
				ports[port] = true
			})
		}

		assert.Len(t, ports, 10)
	})
}

func TestPortRangeConstants(t *testing.T) {
	assert.Greater(t, minPort, 1024)
	assert.LessOrEqual(t, maxPort, 65536)
	assert.Equal(t, maxPort-minPort, 50000)
}
