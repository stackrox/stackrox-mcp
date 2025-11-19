package auth

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type contextKey string

const (
	// requestContextKey is the context key for storing MCP request.
	requestContextKey contextKey = "mcp-request"
)

// WithMCPRequestContext returns a new context with the MCP request.
func WithMCPRequestContext(ctx context.Context, mcpReq *mcp.CallToolRequest) context.Context {
	return context.WithValue(ctx, requestContextKey, mcpReq)
}

// mcpRequestFromContext extracts the MCP request from the context.
// Returns nil if no MCP request is found.
func mcpRequestFromContext(ctx context.Context) *mcp.CallToolRequest {
	mcpReq, ok := ctx.Value(requestContextKey).(*mcp.CallToolRequest)
	if !ok {
		return nil
	}

	return mcpReq
}
