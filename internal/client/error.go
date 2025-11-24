package client

import (
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Error provides detailed error information for gRPC errors.
// Includes retry classification and human-readable error messages.
type Error struct {
	// Error classification
	Code      codes.Code // gRPC status code
	Retriable bool       // Whether error should be retried

	// Human-readable information
	Message     string // Human-readable, actionable error message
	OriginalErr error  // Original gRPC error

	// Context
	Operation string // Operation that failed (e.g., "Connect", "GetDeployment")
}

// NewError creates an Error from a gRPC error.
// This function maps gRPC status codes to human-readable messages and determines if the error is retriable.
func NewError(err error, operation string) *Error {
	if err == nil {
		return nil
	}

	grpcStatus, ok := status.FromError(err)
	if !ok {
		// Not a gRPC error.
		return &Error{
			Code:        codes.Unknown,
			Retriable:   false,
			Message:     fmt.Sprintf("Unknown error: %v", err),
			OriginalErr: err,
			Operation:   operation,
		}
	}

	code := grpcStatus.Code()
	message := formatMessage(code, grpcStatus.Message(), operation)

	return &Error{
		Code:        code,
		Retriable:   IsRetriableGRPCError(err),
		Message:     message,
		OriginalErr: err,
		Operation:   operation,
	}
}

// Error returns the error message (implements error interface).
func (e *Error) Error() string {
	return e.Message
}

// IsRetriable returns true if the error should be retried.
func (e *Error) IsRetriable() bool {
	return e.Retriable
}

// IsRetriableGRPCError determines if an error should be retried based on gRPC status code.
// Retriable errors are transient and may succeed on retry.
func IsRetriableGRPCError(err error) bool {
	grpcStatus, ok := status.FromError(err)
	if !ok {
		return false
	}

	return grpcStatus.Code() == codes.Unavailable || grpcStatus.Code() == codes.DeadlineExceeded
}

// formatMessage generates a human-readable, actionable error message
// based on the gRPC status code.
//
//nolint:lll,cyclop
func formatMessage(code codes.Code, detail string, operation string) string {
	baseMsg := fmt.Sprintf("Operation '%s' failed", operation)
	if operation == "" {
		baseMsg = "Operation failed"
	}

	//nolint:exhaustive
	switch code {
	case codes.Unauthenticated:
		return fmt.Sprintf("%s: Authentication failed - invalid or expired API token. Please check your configuration. %s", baseMsg, detail)
	case codes.PermissionDenied:
		return fmt.Sprintf("%s: Permission denied - your API token does not have sufficient permissions for this operation. %s", baseMsg, detail)
	case codes.NotFound:
		return fmt.Sprintf("%s: Resource not found - the requested resource does not exist. %s", baseMsg, detail)
	case codes.InvalidArgument:
		return fmt.Sprintf("%s: Invalid argument - the request contains invalid parameters. %s", baseMsg, detail)
	case codes.Unavailable:
		return fmt.Sprintf("%s: StackRox Central is temporarily unavailable. The request will be retried automatically. %s", baseMsg, detail)
	case codes.DeadlineExceeded:
		return fmt.Sprintf("%s: Request timed out after 30 seconds. StackRox Central may be overloaded. The request will be retried automatically. %s", baseMsg, detail)
	case codes.ResourceExhausted:
		return fmt.Sprintf("%s: Resource exhausted - rate limit exceeded or server overloaded. The request will be retried automatically. %s", baseMsg, detail)
	case codes.Aborted:
		return fmt.Sprintf("%s: Operation aborted due to concurrency conflict. The request will be retried automatically. %s", baseMsg, detail)
	case codes.AlreadyExists:
		return fmt.Sprintf("%s: Resource already exists. %s", baseMsg, detail)
	case codes.FailedPrecondition:
		return fmt.Sprintf("%s: Failed precondition - the system is not in the correct state for this operation. %s", baseMsg, detail)
	case codes.Unimplemented:
		return fmt.Sprintf("%s: Operation not implemented - this method is not available on the StackRox Central server. %s", baseMsg, detail)
	case codes.Canceled:
		return fmt.Sprintf("%s: Operation was cancelled. %s", baseMsg, detail)
	case codes.Unknown:
		return fmt.Sprintf("%s: Unknown error occurred. %s", baseMsg, detail)
	case codes.Internal:
		return fmt.Sprintf("%s: Internal server error - an error occurred on the StackRox Central server. %s", baseMsg, detail)
	default:
		return fmt.Sprintf("%s: Error code %s. %s", baseMsg, code.String(), detail)
	}
}
