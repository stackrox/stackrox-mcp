package auth

import (
	"context"
	"net/http"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
)

func TestWithMCPRequestContext(t *testing.T) {
	mcpReq := &mcp.CallToolRequest{
		Extra: &mcp.RequestExtra{
			Header: http.Header{
				"Authorization": []string{"Bearer test-token"},
			},
		},
	}

	ctxWithReq := WithMCPRequestContext(context.Background(), mcpReq)
	assert.Equal(t, mcpReq, ctxWithReq.Value(requestContextKey))
}

func TestMCPRequestFromContext_WithoutRequest(t *testing.T) {
	ctx := context.Background()
	assert.Nil(t, mcpRequestFromContext(ctx))
}

func TestMCPRequestFromContext_WithWrongType(t *testing.T) {
	ctx := context.WithValue(context.Background(), requestContextKey, "not a request")
	assert.Nil(t, mcpRequestFromContext(ctx))
}

func TestMCPRequestFromContext_WithNilRequest(t *testing.T) {
	var nilReq *mcp.CallToolRequest

	ctx := WithMCPRequestContext(context.Background(), nilReq)
	assert.Nil(t, mcpRequestFromContext(ctx))
}
