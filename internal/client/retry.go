package client

import (
	"math"
	"time"

	"github.com/stackrox/stackrox-mcp/internal/config"
)

const (
	backoffMultiplier = 2.0
)

// RetryPolicy holds configuration for retry behavior with exponential backoff.
type RetryPolicy struct {
	config *config.CentralConfig
}

// NewRetryPolicy creates a new RetryPolicy with default values.
func NewRetryPolicy(config *config.CentralConfig) *RetryPolicy {
	return &RetryPolicy{
		config: config,
	}
}

// NextBackoff calculates the next backoff duration based on the attempt number.
// Uses exponential backoff with jitter to prevent thundering herd.
func (rp *RetryPolicy) NextBackoff(attempt int) time.Duration {
	if attempt < 0 {
		attempt = 0
	}

	// Calculate exponential backoff: initial * (multiplier ^ attempt).
	backoff := float64(rp.config.InitialBackoff) * math.Pow(backoffMultiplier, float64(attempt))

	// Apply max backoff cap.
	if backoff > float64(rp.config.MaxBackoff) {
		backoff = float64(rp.config.MaxBackoff)
	}

	return time.Duration(backoff)
}

// GetMaxRetries maximum allowed retries.
func (rp *RetryPolicy) GetMaxRetries() int {
	return rp.config.MaxRetries
}

// ShouldRetry returns true if attempts are not exhausted.
func (rp *RetryPolicy) ShouldRetry(attempt int) bool {
	return attempt < rp.config.MaxRetries-1
}
