package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetServerName(t *testing.T) {
	assert.Equal(t, "stackrox-mcp", GetServerName())

	original := serverName

	t.Cleanup(func() { serverName = original })

	serverName = "acs-mcp-server"

	assert.Equal(t, "acs-mcp-server", GetServerName())
}

func TestGetProductDisplayName(t *testing.T) {
	assert.Equal(t, "StackRox", GetProductDisplayName())

	original := productDisplayName

	t.Cleanup(func() { productDisplayName = original })

	productDisplayName = "Red Hat Advanced Cluster Security (ACS)"

	assert.Equal(t, "Red Hat Advanced Cluster Security (ACS)", GetProductDisplayName())
}

func TestGetVersion(t *testing.T) {
	assert.Equal(t, "dev", GetVersion())

	original := version

	t.Cleanup(func() { version = original })

	version = "1.2.3"

	assert.Equal(t, "1.2.3", GetVersion())
}
