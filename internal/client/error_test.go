package client

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestNewError_WithGRPCError(t *testing.T) {
	grpcErr := status.Error(codes.Unauthenticated, "invalid token")
	err := NewError(grpcErr, "GetDeployment")

	require.NotNil(t, err)
	assert.Equal(t, codes.Unauthenticated, err.Code)
	assert.False(t, err.Retriable)
	assert.Contains(t, err.Message, "Authentication failed")
	assert.Contains(t, err.Message, "GetDeployment")
	assert.Equal(t, grpcErr, err.OriginalErr)
	assert.Equal(t, "GetDeployment", err.Operation)
}

func TestNewError_WithNonGRPCError(t *testing.T) {
	nonGRPCErr := errors.New("connection refused")
	err := NewError(nonGRPCErr, "Connect")

	require.NotNil(t, err)
	assert.Equal(t, codes.Unknown, err.Code)
	assert.False(t, err.Retriable)
	assert.Contains(t, err.Message, "Unknown error")
	assert.Contains(t, err.Message, "connection refused")
	assert.Equal(t, nonGRPCErr, err.OriginalErr)
	assert.Equal(t, "Connect", err.Operation)
}

func TestNewError_WithNilError(t *testing.T) {
	err := NewError(nil, "SomeOperation")

	assert.Nil(t, err)
}

func TestError_ErrorMethod(t *testing.T) {
	grpcErr := status.Error(codes.NotFound, "deployment not found")
	err := NewError(grpcErr, "GetDeployment")

	assert.Equal(t, err.Message, err.Error())
}

func TestIsRetriableGRPCError(t *testing.T) {
	tests := map[string]struct {
		err      error
		expected bool
	}{
		"Unavailable is retriable": {
			err:      status.Error(codes.Unavailable, "service unavailable"),
			expected: true,
		},
		"DeadlineExceeded is retriable": {
			err:      status.Error(codes.DeadlineExceeded, "timeout"),
			expected: true,
		},
		"Unauthenticated is not retriable": {
			err:      status.Error(codes.Unauthenticated, "invalid credentials"),
			expected: false,
		},
		"NotFound is not retriable": {
			err:      status.Error(codes.NotFound, "not found"),
			expected: false,
		},
		"PermissionDenied is not retriable": {
			err:      status.Error(codes.PermissionDenied, "permission denied"),
			expected: false,
		},
		"Non-gRPC error is not retriable": {
			err:      errors.New("regular error"),
			expected: false,
		},
	}

	for testName, tt := range tests {
		t.Run(testName, func(t *testing.T) {
			result := IsRetriableGRPCError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

//nolint:funlen
func TestFormatMessage_AllCodes(t *testing.T) {
	tests := map[string]struct {
		code            codes.Code
		detail          string
		operation       string
		expectedMessage string
	}{
		"Unauthenticated": {
			code:            codes.Unauthenticated,
			detail:          "token expired",
			operation:       "GetDeployment",
			expectedMessage: "Authentication failed",
		},
		"PermissionDenied": {
			code:            codes.PermissionDenied,
			detail:          "insufficient permissions",
			operation:       "DeleteDeployment",
			expectedMessage: "Permission denied",
		},
		"NotFound": {
			code:            codes.NotFound,
			detail:          "deployment not found",
			operation:       "GetDeployment",
			expectedMessage: "Resource not found",
		},
		"InvalidArgument": {
			code:            codes.InvalidArgument,
			detail:          "invalid deployment ID",
			operation:       "GetDeployment",
			expectedMessage: "Invalid argument",
		},
		"Unavailable": {
			code:            codes.Unavailable,
			detail:          "service unavailable",
			operation:       "GetDeployment",
			expectedMessage: "StackRox Central is temporarily unavailable",
		},
		"DeadlineExceeded": {
			code:            codes.DeadlineExceeded,
			detail:          "timeout",
			operation:       "GetDeployment",
			expectedMessage: "Request timed out after 30 seconds",
		},
		"ResourceExhausted": {
			code:            codes.ResourceExhausted,
			detail:          "rate limit exceeded",
			operation:       "GetDeployment",
			expectedMessage: "Resource exhausted",
		},
		"Aborted": {
			code:            codes.Aborted,
			detail:          "transaction aborted",
			operation:       "UpdateDeployment",
			expectedMessage: "Operation aborted",
		},
		"AlreadyExists": {
			code:            codes.AlreadyExists,
			detail:          "deployment already exists",
			operation:       "CreateDeployment",
			expectedMessage: "Resource already exists",
		},
		"FailedPrecondition": {
			code:            codes.FailedPrecondition,
			detail:          "system not ready",
			operation:       "CreateDeployment",
			expectedMessage: "Failed precondition",
		},
		"Unimplemented": {
			code:            codes.Unimplemented,
			detail:          "method not supported",
			operation:       "GetDeployment",
			expectedMessage: "Operation not implemented",
		},
		"Canceled": {
			code:            codes.Canceled,
			detail:          "operation cancelled",
			operation:       "GetDeployment",
			expectedMessage: "Operation was cancelled",
		},
		"Unknown": {
			code:            codes.Unknown,
			detail:          "unknown error",
			operation:       "GetDeployment",
			expectedMessage: "Unknown error occurred",
		},
		"Internal": {
			code:            codes.Internal,
			detail:          "internal server error",
			operation:       "GetDeployment",
			expectedMessage: "Internal server error",
		},
		"Empty operation": {
			code:            codes.Unknown,
			detail:          "test",
			operation:       "",
			expectedMessage: "Operation failed",
		},
		"Default case": {
			code:            codes.DataLoss,
			detail:          "data loss",
			operation:       "GetDeployment",
			expectedMessage: "Error code DataLoss",
		},
	}

	for testName, tt := range tests {
		t.Run(testName, func(t *testing.T) {
			message := formatMessage(tt.code, tt.detail, tt.operation)
			assert.Contains(t, message, tt.expectedMessage)
			assert.Contains(t, message, tt.detail)
		})
	}
}

func TestFormatMessage_ContainsOperationName(t *testing.T) {
	message := formatMessage(codes.Unknown, "test error", "GetDeployment")
	assert.Contains(t, message, "GetDeployment")
}

func TestError_Fields(t *testing.T) {
	grpcErr := status.Error(codes.Unavailable, "service down")
	err := NewError(grpcErr, "Connect")

	assert.Equal(t, codes.Unavailable, err.Code)
	assert.True(t, err.Retriable)
	assert.NotEmpty(t, err.Message)
	assert.Equal(t, grpcErr, err.OriginalErr)
	assert.Equal(t, "Connect", err.Operation)
}
