package config

// Branding variables are set at build time via ldflags.
// Red Hat builds override these defaults:
//
//	-X github.com/stackrox/stackrox-mcp/internal/config.serverName=acs-mcp-server
//	-X github.com/stackrox/stackrox-mcp/internal/config.productDisplayName=Red Hat Advanced Cluster Security (ACS)
//	-X github.com/stackrox/stackrox-mcp/internal/config.version=<version>
var (
	// serverName is the MCP server name reported to clients.
	//nolint:gochecknoglobals
	serverName = "stackrox-mcp"

	// productDisplayName is the product name used in tool descriptions.
	//nolint:gochecknoglobals
	productDisplayName = "StackRox"

	// version is the application version.
	version = "dev"
)

// GetServerName returns the MCP server name.
func GetServerName() string {
	return serverName
}

// GetProductDisplayName returns the product display name used in tool descriptions.
func GetProductDisplayName() string {
	return productDisplayName
}

// GetVersion returns the application version.
func GetVersion() string {
	return version
}
