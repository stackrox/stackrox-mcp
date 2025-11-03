package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig_FromYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

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
	err := os.WriteFile(configPath, []byte(yamlContent), 0600)
	require.NoError(t, err)

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

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	err := os.WriteFile(configPath, []byte(yamlContent), 0600)
	require.NoError(t, err)

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
	// Set only required field
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
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	invalidYAML := `
central:
  url: central.example.com:8443
  invalid yaml syntax here: [[[
`

	err := os.WriteFile(configPath, []byte(invalidYAML), 0600)
	require.NoError(t, err)

	_, err = LoadConfig(configPath)
	assert.Error(t, err)
}

func TestValidate_MissingURL(t *testing.T) {
	cfg := &Config{
		Central: CentralConfig{
			URL: "",
		},
		Tools: ToolsConfig{
			Vulnerability: ToolsetVulnerabilityConfig{
				Enabled: true,
			},
		},
	}

	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "central.url is required")
}

func TestValidate_AtLeastOneTool(t *testing.T) {
	cfg := &Config{
		Central: CentralConfig{
			URL: "central.example.com:8443",
		},
	}

	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "at least one tool has to be enabled")
}

func TestValidate_ValidConfig(t *testing.T) {
	cfg := &Config{
		Central: CentralConfig{
			URL:        "central.example.com:8443",
			Insecure:   false,
			ForceHTTP1: false,
		},
		Global: GlobalConfig{
			ReadOnlyTools: true,
		},
		Tools: ToolsConfig{
			Vulnerability: ToolsetVulnerabilityConfig{
				Enabled: true,
			},
			ConfigManager: ToolConfigManagerConfig{
				Enabled: false,
			},
		},
	}

	err := cfg.Validate()
	assert.NoError(t, err)
}
