package config

import (
	"os"
	"testing"

	"github.com/stackrox/stackrox-mcp/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getDefaultConfig returns a default config for testing validation logic.
func getDefaultConfig() *Config {
	return &Config{
		Central: CentralConfig{
			URL: "central.example.com:8443",
		},
		Server: ServerConfig{
			Address: "localhost",
			Port:    8080,
		},
		Tools: ToolsConfig{
			Vulnerability: ToolsetVulnerabilityConfig{
				Enabled: true,
			},
		},
	}
}

func TestLoadConfig_FromYAML(t *testing.T) {
	yamlContent := `
central:
  url: central.example.com:8443
  insecure: true
  force_http1: true
global:
  read_only_tools: false
tools:
  vulnerability:
    enabled: true
  config_manager:
    enabled: False
`

	configPath := testutil.WriteYAMLFile(t, yamlContent)

	defer func() { assert.NoError(t, os.Remove(configPath)) }()

	cfg, err := LoadConfig(configPath)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "central.example.com:8443", cfg.Central.URL)
	assert.True(t, cfg.Central.Insecure)
	assert.True(t, cfg.Central.ForceHTTP1)
	assert.False(t, cfg.Global.ReadOnlyTools)
	assert.True(t, cfg.Tools.Vulnerability.Enabled)
	assert.False(t, cfg.Tools.ConfigManager.Enabled)
}

func TestLoadConfig_EnvVarOverride(t *testing.T) {
	yamlContent := `
central:
  url: central.example.com:8443
tools:
  vulnerability:
    enabled: false
`

	configPath := testutil.WriteYAMLFile(t, yamlContent)

	defer func() { assert.NoError(t, os.Remove(configPath)) }()

	t.Setenv("STACKROX_MCP__CENTRAL__URL", "override.example.com:443")
	t.Setenv("STACKROX_MCP__TOOLS__VULNERABILITY__ENABLED", "true")

	cfg, err := LoadConfig(configPath)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "override.example.com:443", cfg.Central.URL)
	assert.True(t, cfg.Tools.Vulnerability.Enabled)
}

func TestLoadConfig_EnvVarOnly(t *testing.T) {
	t.Setenv("STACKROX_MCP__CENTRAL__URL", "env.example.com:8443")
	t.Setenv("STACKROX_MCP__CENTRAL__INSECURE", "true")
	t.Setenv("STACKROX_MCP__CENTRAL__FORCE_HTTP1", "true")
	t.Setenv("STACKROX_MCP__GLOBAL__READ_ONLY_TOOLS", "false")
	t.Setenv("STACKROX_MCP__TOOLS__VULNERABILITY__ENABLED", "true")
	t.Setenv("STACKROX_MCP__TOOLS__CONFIG_MANAGER__ENABLED", "true")

	cfg, err := LoadConfig("")
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "env.example.com:8443", cfg.Central.URL)
	assert.True(t, cfg.Central.Insecure)
	assert.True(t, cfg.Central.ForceHTTP1)
	assert.False(t, cfg.Global.ReadOnlyTools)
	assert.True(t, cfg.Tools.Vulnerability.Enabled)
	assert.True(t, cfg.Tools.ConfigManager.Enabled)
}

func TestLoadConfig_Defaults(t *testing.T) {
	// Set only required field.
	t.Setenv("STACKROX_MCP__TOOLS__CONFIG_MANAGER__ENABLED", "true")

	cfg, err := LoadConfig("")
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "central.stackrox:8443", cfg.Central.URL)
	assert.False(t, cfg.Central.Insecure)
	assert.False(t, cfg.Central.ForceHTTP1)
	assert.True(t, cfg.Global.ReadOnlyTools)
	assert.False(t, cfg.Tools.Vulnerability.Enabled)
	assert.True(t, cfg.Tools.ConfigManager.Enabled)

	// Check another tools default.
	t.Setenv("STACKROX_MCP__TOOLS__CONFIG_MANAGER__ENABLED", "")
	t.Setenv("STACKROX_MCP__TOOLS__VULNERABILITY__ENABLED", "true")

	cfg, err = LoadConfig("")
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.True(t, cfg.Tools.Vulnerability.Enabled)
	assert.False(t, cfg.Tools.ConfigManager.Enabled)
}

func TestLoadConfig_MissingFile(t *testing.T) {
	_, err := LoadConfig("/non/existent/config.yaml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read config file")
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
	invalidYAML := `
central:
  url: central.example.com:8443
  invalid yaml syntax here: [[[
`

	configPath := testutil.WriteYAMLFile(t, invalidYAML)

	defer func() { assert.NoError(t, os.Remove(configPath)) }()

	_, err := LoadConfig(configPath)
	assert.Error(t, err)
}

func TestLoadConfig_UnmarshalFailure(t *testing.T) {
	// YAML with type mismatch - port should be int.
	invalidTypeYAML := `
server:
  port: "not-a-number"
`
	configPath := testutil.WriteYAMLFile(t, invalidTypeYAML)
	_, err := LoadConfig(configPath)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal config")
}

func TestLoadConfig_ValidationFailure(t *testing.T) {
	// Valid YAML but fails on central URL validation (no URL).
	validYAMLInvalidConfig := `
central:
  url: ""
server:
  address: localhost
  port: 8080
tools:
  vulnerability:
    enabled: true
`

	configPath := testutil.WriteYAMLFile(t, validYAMLInvalidConfig)
	_, err := LoadConfig(configPath)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid configuration")
	assert.Contains(t, err.Error(), "central.url is required")
}

func TestValidate_MissingURL(t *testing.T) {
	cfg := getDefaultConfig()
	cfg.Central.URL = ""

	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "central.url is required")
}

func TestValidate_AtLeastOneTool(t *testing.T) {
	cfg := getDefaultConfig()
	cfg.Tools.Vulnerability.Enabled = false

	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "at least one tool has to be enabled")
}

func TestValidate_ValidConfig(t *testing.T) {
	cfg := getDefaultConfig()

	err := cfg.Validate()
	assert.NoError(t, err)
}

func TestValidate_MissingServerAddress(t *testing.T) {
	cfg := getDefaultConfig()
	cfg.Server.Address = ""

	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "server.address is required")
}

func TestValidate_InvalidServerPort(t *testing.T) {
	tests := map[string]struct {
		port int
	}{
		"zero port":     {port: 0},
		"negative port": {port: -1},
		"port too high": {port: 65536},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			cfg := getDefaultConfig()
			cfg.Server.Port = tt.port

			err := cfg.Validate()
			require.Error(t, err)
			assert.Contains(t, err.Error(), "server.port must be between 1 and 65535")
		})
	}
}

func TestLoadConfig_ServerDefaults(t *testing.T) {
	// Set only required fields.
	t.Setenv("STACKROX_MCP__TOOLS__CONFIG_MANAGER__ENABLED", "true")

	cfg, err := LoadConfig("")
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "localhost", cfg.Server.Address)
	assert.Equal(t, 8080, cfg.Server.Port)
}
