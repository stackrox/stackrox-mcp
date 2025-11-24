// Package auth handles tokens required for StackRox Central API communication.
package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pkg/errors"
	"google.golang.org/grpc/credentials"
)

const (
	authorizationHeader = "Authorization"
	bearerPrefix        = "Bearer "
)

// extractBearerToken returns the bearer token provided in the MCP call metadata header.
// Returns an error if the header or token are missing or malformed.
func extractBearerToken(mcpReq *mcp.CallToolRequest) (string, error) {
	if mcpReq == nil {
		return "", errors.New("MCP request is nil")
	}

	extra := mcpReq.GetExtra()
	if extra == nil {
		return "", errors.New("MCP request metadata is missing")
	}

	token, err := tokenFromHeader(extra.Header)
	if err != nil {
		return "", err
	}

	return token, nil
}

func tokenFromHeader(header http.Header) (string, error) {
	if header == nil {
		return "", errors.New("headers are missing")
	}

	raw := header.Get(authorizationHeader)
	if raw == "" {
		return "", errors.New("authorization header is missing")
	}

	if !strings.HasPrefix(raw, bearerPrefix) {
		return "", errors.New("authorization header must contain a bearer token")
	}

	token := strings.TrimSpace(raw[len(bearerPrefix):])
	if token == "" {
		return "", errors.New("authorization token is empty")
	}

	return token, nil
}

// passthroughTokenCredentials implements credentials.PerRPCCredentials for context-based API token authentication.
// It reads the API token from the request context using the tokenContextKey.
// This is used when auth_type is "passthrough" in the configuration, allowing tools
// to provide their own tokens on a per-request basis.
type passthroughTokenCredentials struct{}

// NewPassthroughTokenCredentials creates a new passthroughTokenCredentials.
func NewPassthroughTokenCredentials() credentials.PerRPCCredentials {
	return &passthroughTokenCredentials{}
}

// GetRequestMetadata implements credentials.PerRPCCredentials.
// It reads the token from the context and returns the authorization metadata.
func (t *passthroughTokenCredentials) GetRequestMetadata(ctx context.Context, _ ...string) (map[string]string, error) {
	mcpReq := mcpRequestFromContext(ctx)
	if mcpReq == nil {
		return nil, errors.New("MCP request is not found in context")
	}

	token, err := extractBearerToken(mcpReq)
	if err != nil {
		return nil, errors.Wrap(err, "failed to extract bearer token from MCP request")
	}

	return map[string]string{
		"authorization": "Bearer " + token,
	}, nil
}

// RequireTransportSecurity implements credentials.PerRPCCredentials.
func (t *passthroughTokenCredentials) RequireTransportSecurity() bool {
	return true
}
