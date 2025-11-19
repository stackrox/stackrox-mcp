package client

import (
	"context"
	"log/slog"
	"time"

	"google.golang.org/grpc"
)

// createRetryInterceptor creates a retry interceptor closure that captures the retry policy and timeout.
func createRetryInterceptor(policy *RetryPolicy, requestTimeout time.Duration) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply any,
		clientConn *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		var lastErr error

		for attempt := range policy.GetMaxRetries() {
			// Create context with timeout for this attempt.
			attemptCtx, cancel := context.WithTimeout(ctx, requestTimeout)
			err := invoker(attemptCtx, method, req, reply, clientConn, opts...)

			cancel()

			if err == nil {
				if attempt > 0 {
					slog.Info("Request succeeded after retry", "method", method, "attempt", attempt+1)
				}

				return nil
			}

			if !IsRetriableGRPCError(err) {
				return err
			}

			lastErr = err

			if !policy.ShouldRetry(attempt) {
				break
			}

			backoff := policy.NextBackoff(attempt)

			slog.Info("Request failed, retrying",
				"method", method,
				"attempt", attempt+1,
				"backoff", backoff,
				"error", err,
			)

			// Wait for backoff duration or context cancellation.
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
				// Continue to next attempt.
			}
		}

		slog.Warn("Request failed after all retries",
			"method", method,
			"attempts", policy.GetMaxRetries(),
			"error", lastErr,
		)

		return lastErr
	}
}

// createLoggingInterceptor creates a logging interceptor closure.
func createLoggingInterceptor() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply any,
		clientConn *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		slog.Debug("API request started", "method", method)

		startTime := time.Now()
		err := invoker(ctx, method, req, reply, clientConn, opts...)
		duration := time.Since(startTime)

		if err != nil {
			slog.Error("API request failed", "method", method, "duration", duration, "error", err)

			return err
		}

		slog.Debug("API request completed", "method", method, "duration", duration)

		return nil
	}
}
