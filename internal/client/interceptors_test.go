package client

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stackrox/stackrox-mcp/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// mockInvoker is a helper to create mock invokers for testing.
type mockInvoker struct {
	calls     int
	responses []error
}

func (m *mockInvoker) invoke(
	_ context.Context,
	_ string,
	_, _ any,
	_ *grpc.ClientConn,
	_ ...grpc.CallOption,
) error {
	if m.calls >= len(m.responses) {
		return errors.New("no more mock responses available")
	}

	err := m.responses[m.calls]
	m.calls++

	return err
}

func TestCreateRetryInterceptor_Success(t *testing.T) {
	centralConfig := &config.CentralConfig{
		MaxRetries:     3,
		InitialBackoff: time.Nanosecond,
		MaxBackoff:     time.Nanosecond,
	}
	policy := NewRetryPolicy(centralConfig)
	interceptor := createRetryInterceptor(policy, time.Second)

	mock := &mockInvoker{
		responses: []error{nil}, // Success on first attempt.
	}

	err := interceptor(
		context.Background(),
		"test.Method",
		nil,
		nil,
		nil,
		mock.invoke,
	)

	require.NoError(t, err)
	assert.Equal(t, 1, mock.calls)
}

func TestCreateRetryInterceptor_RetryableError(t *testing.T) {
	centralConfig := &config.CentralConfig{
		MaxRetries:     3,
		InitialBackoff: time.Nanosecond,
		MaxBackoff:     time.Nanosecond,
	}
	policy := NewRetryPolicy(centralConfig)
	interceptor := createRetryInterceptor(policy, time.Second)

	mock := &mockInvoker{
		responses: []error{
			status.Error(codes.Unavailable, "service unavailable"),    // Retry 1.
			status.Error(codes.DeadlineExceeded, "deadline exceeded"), // Retry 2.
			nil, // Success on third attempt.
		},
	}

	startTime := time.Now()
	err := interceptor(
		context.Background(),
		"test.Method",
		nil,
		nil,
		nil,
		mock.invoke,
	)
	duration := time.Since(startTime)

	require.NoError(t, err)
	assert.Equal(t, 3, mock.calls)

	// Should have at least some delay due to backoff (allow for jitter and variability).
	assert.Greater(t, duration, centralConfig.InitialBackoff)
}

func TestCreateRetryInterceptor_NonRetriableError(t *testing.T) {
	centralConfig := &config.CentralConfig{
		MaxRetries:     3,
		InitialBackoff: time.Nanosecond,
		MaxBackoff:     time.Nanosecond,
	}
	policy := NewRetryPolicy(centralConfig)
	interceptor := createRetryInterceptor(policy, time.Second)

	mock := &mockInvoker{
		responses: []error{
			status.Error(codes.Unauthenticated, "invalid token"), // Non-retriable.
		},
	}

	err := interceptor(
		context.Background(),
		"test.Method",
		nil,
		nil,
		nil,
		mock.invoke,
	)

	require.Error(t, err)
	assert.Equal(t, codes.Unauthenticated, status.Code(err))

	// Should not retry.
	assert.Equal(t, 1, mock.calls)
}

func TestCreateRetryInterceptor_MaxAttemptsReached(t *testing.T) {
	centralConfig := &config.CentralConfig{
		MaxRetries:     3,
		InitialBackoff: time.Nanosecond,
		MaxBackoff:     time.Nanosecond,
	}
	policy := NewRetryPolicy(centralConfig)
	interceptor := createRetryInterceptor(policy, time.Second)

	mock := &mockInvoker{
		responses: []error{
			status.Error(codes.Unavailable, "service unavailable"),
			status.Error(codes.Unavailable, "service unavailable"),
			status.Error(codes.Unavailable, "service unavailable"),
		},
	}

	err := interceptor(
		context.Background(),
		"test.Method",
		nil,
		nil,
		nil,
		mock.invoke,
	)

	require.Error(t, err)
	assert.Equal(t, codes.Unavailable, status.Code(err))
	// All attempts used.
	assert.Equal(t, 3, mock.calls)
}

func TestCreateRetryInterceptor_ContextCancellation(t *testing.T) {
	centralConfig := &config.CentralConfig{
		MaxRetries:     3,
		InitialBackoff: 5 * time.Millisecond,
		MaxBackoff:     5 * time.Millisecond,
	}
	policy := NewRetryPolicy(centralConfig)
	interceptor := createRetryInterceptor(policy, time.Second)

	ctx, cancel := context.WithCancel(context.Background())

	mock := &mockInvoker{
		responses: []error{
			status.Error(codes.Unavailable, "service unavailable"),
			status.Error(codes.Unavailable, "service unavailable"),
		},
	}

	// Cancel context after first attempt.
	go func() {
		time.Sleep(time.Millisecond)
		cancel()
	}()

	err := interceptor(
		ctx,
		"test.Method",
		nil,
		nil,
		nil,
		mock.invoke,
	)

	require.Error(t, err)
	assert.Equal(t, context.Canceled, err)
	// Should have made first attempt, then cancelled during backoff.
	assert.Equal(t, 1, mock.calls)
}

func TestCreateRetryInterceptor_BackoffProgression(t *testing.T) {
	centralConfig := &config.CentralConfig{
		MaxRetries:     4,
		InitialBackoff: time.Millisecond,
		MaxBackoff:     time.Millisecond,
	}
	policy := NewRetryPolicy(centralConfig)
	interceptor := createRetryInterceptor(policy, time.Second)

	mock := &mockInvoker{
		responses: []error{
			status.Error(codes.Unavailable, "service unavailable"),
			status.Error(codes.Unavailable, "service unavailable"),
			status.Error(codes.Unavailable, "service unavailable"),
			nil, // Success on 4th attempt.
		},
	}

	startTime := time.Now()
	err := interceptor(
		context.Background(),
		"test.Method",
		nil,
		nil,
		nil,
		mock.invoke,
	)
	duration := time.Since(startTime)

	require.NoError(t, err)
	assert.Equal(t, 4, mock.calls)
	// With exponential backoff and jitter, we should see some delay but exact timing is variable.
	assert.Greater(t, duration, 2*centralConfig.InitialBackoff)
}

func TestCreateRetryInterceptor_RequestTimeout(t *testing.T) {
	centralConfig := &config.CentralConfig{
		MaxRetries:     3,
		InitialBackoff: 5 * time.Millisecond,
		MaxBackoff:     5 * time.Millisecond,
	}
	policy := NewRetryPolicy(centralConfig)

	// Very short request timeout.
	interceptor := createRetryInterceptor(policy, time.Millisecond)

	attemptCount := 0
	slowInvoker := func(
		ctx context.Context,
		_ string,
		_, _ any,
		_ *grpc.ClientConn,
		_ ...grpc.CallOption,
	) error {
		attemptCount++

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(2 * centralConfig.MaxBackoff):
			return nil
		}
	}

	err := interceptor(
		context.Background(),
		"test.Method",
		nil,
		nil,
		nil,
		slowInvoker,
	)

	// Should timeout on each attempt.
	require.Error(t, err)
	assert.GreaterOrEqual(t, attemptCount, 1)
}

func TestCreateRetryInterceptor_ZeroRetries(t *testing.T) {
	centralConfig := &config.CentralConfig{
		MaxRetries:     1, // No retries.
		InitialBackoff: time.Nanosecond,
		MaxBackoff:     time.Nanosecond,
	}
	policy := NewRetryPolicy(centralConfig)
	interceptor := createRetryInterceptor(policy, time.Second)

	mock := &mockInvoker{
		responses: []error{
			status.Error(codes.Unavailable, "service unavailable"),
		},
	}

	err := interceptor(
		context.Background(),
		"test.Method",
		nil,
		nil,
		nil,
		mock.invoke,
	)

	require.Error(t, err)
	assert.Equal(t, codes.Unavailable, status.Code(err))
	assert.Equal(t, 1, mock.calls)
}

func TestCreateRetryInterceptor_ContextTimeoutDuringAttempt(t *testing.T) {
	centralConfig := &config.CentralConfig{
		MaxRetries:     3,
		InitialBackoff: time.Nanosecond,
		MaxBackoff:     time.Nanosecond,
	}
	policy := NewRetryPolicy(centralConfig)

	// Request timeout of 10ms.
	interceptor := createRetryInterceptor(policy, time.Millisecond)

	callCount := 0
	slowInvoker := func(
		_ context.Context,
		_ string,
		_, _ any,
		_ *grpc.ClientConn,
		_ ...grpc.CallOption,
	) error {
		callCount++
		// First attempt times out, second succeeds quickly.
		if callCount == 1 {
			time.Sleep(5 * time.Millisecond)

			return status.Error(codes.DeadlineExceeded, "timeout")
		}

		return nil
	}

	err := interceptor(
		context.Background(),
		"test.Method",
		nil,
		nil,
		nil,
		slowInvoker,
	)

	require.NoError(t, err)
	assert.Equal(t, 2, callCount)
}

func TestCreateLoggingInterceptor_Success(t *testing.T) {
	interceptor := createLoggingInterceptor()

	mock := &mockInvoker{
		responses: []error{
			nil,
		},
	}

	err := interceptor(
		context.Background(),
		"test.Method",
		nil,
		nil,
		nil,
		mock.invoke,
	)

	require.NoError(t, err)
	assert.Equal(t, 1, mock.calls)
}

func TestCreateLoggingInterceptor_Error(t *testing.T) {
	interceptor := createLoggingInterceptor()

	expectedErr := status.Error(codes.NotFound, "not found")
	mock := &mockInvoker{
		responses: []error{
			expectedErr,
		},
	}

	err := interceptor(
		context.Background(),
		"test.Method",
		nil,
		nil,
		nil,
		mock.invoke,
	)

	require.Error(t, err)
	assert.Equal(t, expectedErr, err)
}
