package auth

import (
	"context"
	"net/http"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractBearerToken_Failures(t *testing.T) {
	tests := map[string]*mcp.CallToolRequest{
		"nil request":          nil,
		"request extra is nil": &mcp.CallToolRequest{},
		"request header is nil": &mcp.CallToolRequest{
			Extra: &mcp.RequestExtra{},
		},
		"missing auth header": &mcp.CallToolRequest{
			Extra: &mcp.RequestExtra{
				Header: http.Header{},
			},
		},
		"wrong bearer prefix": &mcp.CallToolRequest{
			Extra: &mcp.RequestExtra{
				Header: http.Header{
					"Authorization": []string{"Not-Bearer test"},
				},
			},
		},
		"empty token": &mcp.CallToolRequest{
			Extra: &mcp.RequestExtra{
				Header: http.Header{
					"Authorization": []string{"Bearer   "},
				},
			},
		},
	}

	for testName, testMcpReq := range tests {
		t.Run(testName, func(t *testing.T) {
			token, err := extractBearerToken(testMcpReq)

			require.Error(t, err)
			assert.Empty(t, token)
		})
	}
}

func TestExtractBearerToken_Success(t *testing.T) {
	req := &mcp.CallToolRequest{
		Extra: &mcp.RequestExtra{
			Header: http.Header{
				"Authorization": []string{"Bearer my-token"},
			},
		},
	}

	token, err := extractBearerToken(req)
	require.NoError(t, err)
	assert.Equal(t, "my-token", token)
}

func TestPassthroughTokenCredentials_Success(t *testing.T) {
	mcpReq := &mcp.CallToolRequest{
		Extra: &mcp.RequestExtra{
			Header: http.Header{
				"Authorization": []string{"Bearer token-123"},
			},
		},
	}

	ctx := WithMCPRequestContext(context.Background(), mcpReq)
	tokenCredentials := NewPassthroughTokenCredentials()

	meta, err := tokenCredentials.GetRequestMetadata(ctx)
	require.NoError(t, err)

	assert.Equal(t, "Bearer token-123", meta["authorization"])
}

func TestPassthroughTokenCredentials_NoAuthHeader(t *testing.T) {
	mcpReq := &mcp.CallToolRequest{
		Extra: &mcp.RequestExtra{
			Header: http.Header{},
		},
	}

	ctx := WithMCPRequestContext(context.Background(), mcpReq)
	tokenCredentials := NewPassthroughTokenCredentials()

	_, err := tokenCredentials.GetRequestMetadata(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to extract bearer token from MCP request")
}

func TestPassthroughTokenCredentials_MissingMCPRequest(t *testing.T) {
	ctx := context.Background()
	tokenCredentials := NewPassthroughTokenCredentials()

	_, err := tokenCredentials.GetRequestMetadata(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "MCP request is not found in context")
}

func TestPassthroughTokenCredentials_RequireTransportSecurity(t *testing.T) {
	tokenCredentials := NewPassthroughTokenCredentials()

	assert.True(t, tokenCredentials.RequireTransportSecurity())
}
