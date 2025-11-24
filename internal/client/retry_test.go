package client

import (
	"testing"
	"time"

	"github.com/stackrox/stackrox-mcp/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestRetryPolicy_ShouldRetry(t *testing.T) {
	policy := NewRetryPolicy(&config.CentralConfig{
		MaxRetries: 3,
	})

	assert.True(t, policy.ShouldRetry(-1))
	assert.True(t, policy.ShouldRetry(0))
	assert.True(t, policy.ShouldRetry(policy.config.MaxRetries-2))
	assert.False(t, policy.ShouldRetry(policy.config.MaxRetries-1))
}

func TestRetryPolicy_NextBackoff(t *testing.T) {
	tests := map[string]struct {
		initialBackoff time.Duration
		maxBackoff     time.Duration
		attempt        int
		expected       time.Duration
	}{
		"zero attempt": {
			initialBackoff: 500 * time.Millisecond,
			maxBackoff:     5 * time.Second,
			attempt:        0,
			expected:       500 * time.Millisecond, // 500ms * (2^0) = 500ms
		},
		"first attempt": {
			initialBackoff: time.Second,
			maxBackoff:     10 * time.Second,
			attempt:        0,
			expected:       time.Second, // 1s * (2^0) = 1s
		},
		"second attempt": {
			initialBackoff: time.Second,
			maxBackoff:     10 * time.Second,
			attempt:        1,
			expected:       2 * time.Second, // 1s * (2^1) = 2s
		},
		"third attempt": {
			initialBackoff: time.Second,
			maxBackoff:     10 * time.Second,
			attempt:        2,
			expected:       4 * time.Second, // 1s * (2^2) = 4s
		},
		"capped at max backoff": {
			initialBackoff: time.Second,
			maxBackoff:     5 * time.Second,
			attempt:        3,
			expected:       5 * time.Second, // 8s capped to 5s
		},
		"negative attempt treated as zero": {
			initialBackoff: time.Second,
			maxBackoff:     10 * time.Second,
			attempt:        -1,
			expected:       time.Second, // negative becomes 0
		},
		"max equals initial backoff": {
			initialBackoff: time.Second,
			maxBackoff:     time.Second,
			attempt:        5,
			expected:       time.Second, // always capped at 1s
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			policy := &RetryPolicy{
				config: &config.CentralConfig{
					InitialBackoff: testCase.initialBackoff,
					MaxBackoff:     testCase.maxBackoff,
				},
			}

			result := policy.NextBackoff(testCase.attempt)
			assert.Equal(t, testCase.expected, result)
		})
	}
}
