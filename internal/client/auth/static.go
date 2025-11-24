package auth

import (
	"context"
	"errors"

	"google.golang.org/grpc/credentials"
)

// staticTokenCredentials implements credentials.PerRPCCredentials for static API token authentication.
// It adds a fixed API token as a Bearer token in the authorization header for each RPC call.
// This is used when auth_type is "static" in the configuration.
type staticTokenCredentials struct {
	token string
}

// NewStaticTokenCredentials creates a new staticTokenCredentials with the given API token.
func NewStaticTokenCredentials(token string) credentials.PerRPCCredentials {
	return &staticTokenCredentials{
		token: token,
	}
}

// GetRequestMetadata implements credentials.PerRPCCredentials.
// It returns the authorization metadata to be added to each RPC request.
func (t *staticTokenCredentials) GetRequestMetadata(_ context.Context, _ ...string) (map[string]string, error) {
	if t.token == "" {
		return nil, errors.New("API token is empty")
	}

	return map[string]string{
		"authorization": "Bearer " + t.token,
	}, nil
}

// RequireTransportSecurity implements credentials.PerRPCCredentials.
func (t *staticTokenCredentials) RequireTransportSecurity() bool {
	return true
}
