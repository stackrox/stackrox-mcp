package config

import (
	"os"
	"testing"
	"time"

	"github.com/stackrox/stackrox-mcp/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getDefaultConfig returns a default config for testing validation logic.
func getDefaultConfig() *Config {
	return &Config{
		Central: CentralConfig{
			URL:                   "central.example.com:8443",
			AuthType:              AuthTypeStatic,
			APIToken:              "test-token",
			InsecureSkipTLSVerify: false,
			ForceHTTP1:            false,
			RequestTimeout:        defaultRequestTimeout,
			MaxRetries:            defaultMaxRetries,
			InitialBackoff:        defaultInitialBackoff,
			MaxBackoff:            defaultMaxBackoff,
		},
		Global: GlobalConfig{
			ReadOnlyTools: false,
		},
		Server: ServerConfig{
			Address: "localhost",
			Port:    8080,
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
}

func TestLoadConfig_FromYAML(t *testing.T) {
	yamlContent := `
central:
  url: central.example.com:8443
  insecure_skip_tls_verify: true
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
	assert.True(t, cfg.Central.InsecureSkipTLSVerify)
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
	t.Setenv("STACKROX_MCP__CENTRAL__AUTH_TYPE", string(AuthTypeStatic))
	t.Setenv("STACKROX_MCP__CENTRAL__API_TOKEN", "test-token")
	t.Setenv("STACKROX_MCP__CENTRAL__INSECURE_SKIP_TLS_VERIFY", "true")
	t.Setenv("STACKROX_MCP__CENTRAL__FORCE_HTTP1", "true")
	t.Setenv("STACKROX_MCP__GLOBAL__READ_ONLY_TOOLS", "false")
	t.Setenv("STACKROX_MCP__TOOLS__VULNERABILITY__ENABLED", "true")
	t.Setenv("STACKROX_MCP__TOOLS__CONFIG_MANAGER__ENABLED", "true")

	cfg, err := LoadConfig("")
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "env.example.com:8443", cfg.Central.URL)
	assert.Equal(t, AuthTypeStatic, cfg.Central.AuthType)
	assert.Equal(t, "test-token", cfg.Central.APIToken)
	assert.True(t, cfg.Central.InsecureSkipTLSVerify)
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
	assert.False(t, cfg.Central.InsecureSkipTLSVerify)
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

	for testName, tt := range tests {
		t.Run(testName, func(t *testing.T) {
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

	assert.Equal(t, "0.0.0.0", cfg.Server.Address)
	assert.Equal(t, 8080, cfg.Server.Port)
}

func TestValidate_AuthType_Invalid(t *testing.T) {
	cfg := getDefaultConfig()
	cfg.Central.AuthType = "bad-type"

	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "central.auth_type must be either passthrough or static")
}

func TestValidate_AuthTypeStatic_RequiresAPIToken(t *testing.T) {
	cfg := getDefaultConfig()
	cfg.Central.AuthType = AuthTypeStatic
	cfg.Central.APIToken = ""

	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "central.api_token is required for")
	assert.Contains(t, err.Error(), "static")
}

func TestValidate_AuthTypePassthrough_ForbidsAPIToken(t *testing.T) {
	cfg := getDefaultConfig()
	cfg.Central.AuthType = AuthTypePassthrough
	cfg.Central.APIToken = "some-token"

	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "central.api_token can not be set for")
	assert.Contains(t, err.Error(), "passthrough")
}

func TestValidate_AuthTypePassthrough_Success(t *testing.T) {
	cfg := getDefaultConfig()
	cfg.Central.AuthType = AuthTypePassthrough
	cfg.Central.APIToken = ""

	err := cfg.Validate()
	assert.NoError(t, err)
}

func TestValidate_RequestTimeout_MustBePositive(t *testing.T) {
	tests := map[string]struct {
		timeout time.Duration
	}{
		"zero timeout": {
			timeout: 0,
		},
		"negative timeout": {
			timeout: -1 * time.Second,
		},
	}

	for testName, tt := range tests {
		t.Run(testName, func(t *testing.T) {
			cfg := getDefaultConfig()
			cfg.Central.RequestTimeout = tt.timeout

			err := cfg.Validate()
			require.Error(t, err)
			assert.Contains(t, err.Error(), "central.request_timeout must be positive")
		})
	}
}

func TestValidate_RequestTimeout_PositiveValue(t *testing.T) {
	cfg := getDefaultConfig()
	cfg.Central.RequestTimeout = 10 * time.Second

	err := cfg.Validate()
	assert.NoError(t, err)
}

func TestValidate_MaxRetries_Range(t *testing.T) {
	tests := map[string]struct {
		maxRetries int
		shouldFail bool
	}{
		"negative retries": {
			maxRetries: -1,
			shouldFail: true,
		},
		"zero retries": {
			maxRetries: 0,
			shouldFail: false,
		},
		"valid retries": {
			maxRetries: 5,
			shouldFail: false,
		},
		"max retries": {
			maxRetries: 10,
			shouldFail: false,
		},
		"over max retries": {
			maxRetries: 11,
			shouldFail: true,
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			cfg := getDefaultConfig()
			cfg.Central.MaxRetries = testCase.maxRetries

			err := cfg.Validate()
			if testCase.shouldFail {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "central.max_retries must be between 0 and 10")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidate_InitialBackoff_MustBePositive(t *testing.T) {
	tests := map[string]struct {
		backoff time.Duration
	}{
		"zero backoff": {
			backoff: 0,
		},
		"negative backoff": {
			backoff: -1 * time.Second,
		},
	}

	for testName, tt := range tests {
		t.Run(testName, func(t *testing.T) {
			cfg := getDefaultConfig()
			cfg.Central.InitialBackoff = tt.backoff

			err := cfg.Validate()
			require.Error(t, err)
			assert.Contains(t, err.Error(), "central.initial_backoff must be positive")
		})
	}
}

func TestValidate_InitialBackoff_PositiveValue(t *testing.T) {
	cfg := getDefaultConfig()
	cfg.Central.InitialBackoff = 2 * time.Second

	err := cfg.Validate()
	assert.NoError(t, err)
}

func TestValidate_MaxBackoff_MustBePositive(t *testing.T) {
	tests := map[string]struct {
		backoff time.Duration
	}{
		"zero backoff": {
			backoff: 0,
		},
		"negative backoff": {
			backoff: -1 * time.Second,
		},
	}

	for testName, tt := range tests {
		t.Run(testName, func(t *testing.T) {
			cfg := getDefaultConfig()
			cfg.Central.MaxBackoff = tt.backoff

			err := cfg.Validate()
			require.Error(t, err)
			assert.Contains(t, err.Error(), "central.max_backoff must be positive")
		})
	}
}

func TestValidate_MaxBackoff_PositiveValue(t *testing.T) {
	cfg := getDefaultConfig()
	cfg.Central.MaxBackoff = 30 * time.Second

	err := cfg.Validate()
	assert.NoError(t, err)
}

func TestValidate_MaxBackoff_MustBeGreaterThanInitialBackoff(t *testing.T) {
	cfg := getDefaultConfig()
	cfg.Central.InitialBackoff = 10 * time.Second
	cfg.Central.MaxBackoff = 5 * time.Second

	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "central.max_backoff has to be greater than or equal to central.initial_backoff")
}

func TestValidate_MaxBackoff_EqualToInitialBackoff(t *testing.T) {
	cfg := getDefaultConfig()
	cfg.Central.InitialBackoff = 5 * time.Second
	cfg.Central.MaxBackoff = 5 * time.Second

	err := cfg.Validate()
	assert.NoError(t, err)
}

func TestLoadConfig_TimeoutAndRetryDefaults(t *testing.T) {
	t.Setenv("STACKROX_MCP__TOOLS__CONFIG_MANAGER__ENABLED", "true")

	cfg, err := LoadConfig("")
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, defaultRequestTimeout, cfg.Central.RequestTimeout)
	assert.Equal(t, defaultMaxRetries, cfg.Central.MaxRetries)
	assert.Equal(t, defaultInitialBackoff, cfg.Central.InitialBackoff)
	assert.Equal(t, defaultMaxBackoff, cfg.Central.MaxBackoff)
}

func TestConfig_Redacted(t *testing.T) {
	cfg := getDefaultConfig()
	cfg.Central.APIToken = "super-secret-token"

	redactedConfig := cfg.Redacted()

	// Verify sensitive data is redacted.
	assert.Equal(t, "***REDACTED***", redactedConfig.Central.APIToken)

	// Verify non-sensitive data is preserved.
	assert.Equal(t, cfg.Central.URL, redactedConfig.Central.URL)
	assert.Equal(t, cfg.Central.AuthType, redactedConfig.Central.AuthType)
	assert.Equal(t, cfg.Central.InsecureSkipTLSVerify, redactedConfig.Central.InsecureSkipTLSVerify)
	assert.Equal(t, cfg.Central.ForceHTTP1, redactedConfig.Central.ForceHTTP1)
	assert.Equal(t, cfg.Central.RequestTimeout, redactedConfig.Central.RequestTimeout)
	assert.Equal(t, cfg.Central.MaxRetries, redactedConfig.Central.MaxRetries)
	assert.Equal(t, cfg.Central.InitialBackoff, redactedConfig.Central.InitialBackoff)
	assert.Equal(t, cfg.Central.MaxBackoff, redactedConfig.Central.MaxBackoff)

	// Verify other config sections are preserved.
	assert.Equal(t, cfg.Global, redactedConfig.Global)
	assert.Equal(t, cfg.Server, redactedConfig.Server)
	assert.Equal(t, cfg.Tools, redactedConfig.Tools)

	// Verify original config is unchanged.
	assert.Equal(t, "super-secret-token", cfg.Central.APIToken)
}

func TestConfig_Redacted_EmptyToken(t *testing.T) {
	cfg := getDefaultConfig()
	cfg.Central.APIToken = ""

	redactedConfig := cfg.Redacted()

	// Empty token should remain empty, not be replaced with redacted marker.
	assert.Empty(t, redactedConfig.Central.APIToken)
}
