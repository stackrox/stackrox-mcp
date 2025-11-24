package auth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStaticTokenCredentials_Success(t *testing.T) {
	tokenCredentials := NewStaticTokenCredentials("static-token")

	meta, err := tokenCredentials.GetRequestMetadata(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "Bearer static-token", meta["authorization"])
}

func TestStaticTokenCredentials_EmptyToken(t *testing.T) {
	tokenCredentials := NewStaticTokenCredentials("")

	_, err := tokenCredentials.GetRequestMetadata(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "API token is empty")
}

func TestStaticTokenCredentials_RequireTransportSecurity(t *testing.T) {
	tokenCredentials := NewStaticTokenCredentials("static-token")

	assert.True(t, tokenCredentials.RequireTransportSecurity())
}
