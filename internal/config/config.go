// Package config provides configuration handling for StackRox MCP server.
package config

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

const (
	defaultPort = 8080

	defaultRequestTimeout = 30 * time.Second
	defaultMaxRetries     = 3
	defaultInitialBackoff = time.Second
	defaultMaxBackoff     = 10 * time.Second
)

// Config represents the complete application configuration.
type Config struct {
	Central CentralConfig `mapstructure:"central"`
	Global  GlobalConfig  `mapstructure:"global"`
	Server  ServerConfig  `mapstructure:"server"`
	Tools   ToolsConfig   `mapstructure:"tools"`
}

type authType string
type serverType string

const (
	// AuthTypePassthrough defines auth flow where API token, used to communicate with MCP server,
	// is passed and used in a communication with StackRox Central API.
	AuthTypePassthrough authType = "passthrough"

	// AuthTypeStatic defines auth flow where API token is statically configured and
	// defined in configuration or environment variable.
	AuthTypeStatic authType = "static"

	// ServerTypeStdio indicates server runs over stdio.
	ServerTypeStdio serverType = "stdio"
	// ServerTypeStreamableHTTP indicates server runs over streamable-http.
	ServerTypeStreamableHTTP serverType = "streamable-http"
)

// CentralConfig contains StackRox Central connection configuration.
type CentralConfig struct {
	URL                   string   `mapstructure:"url"`
	AuthType              authType `mapstructure:"auth_type"`
	APIToken              string   `mapstructure:"api_token"`
	InsecureSkipTLSVerify bool     `mapstructure:"insecure_skip_tls_verify"`
	ForceHTTP1            bool     `mapstructure:"force_http1"`

	// Timeouts and retry settings
	RequestTimeout time.Duration `mapstructure:"request_timeout"`
	MaxRetries     int           `mapstructure:"max_retries"`
	InitialBackoff time.Duration `mapstructure:"initial_backoff"`
	MaxBackoff     time.Duration `mapstructure:"max_backoff"`
}

// GlobalConfig contains global MCP server configuration.
type GlobalConfig struct {
	ReadOnlyTools bool `mapstructure:"read_only_tools"`
}

// ServerConfig contains HTTP server configuration.
type ServerConfig struct {
	Type    serverType `mapstructure:"type"`
	Address string     `mapstructure:"address"`
	Port    int        `mapstructure:"port"`
}

// ToolsConfig contains configuration for individual MCP tools.
type ToolsConfig struct {
	Vulnerability ToolsetVulnerabilityConfig `mapstructure:"vulnerability"`
	ConfigManager ToolConfigManagerConfig    `mapstructure:"config_manager"`
}

// ToolsetVulnerabilityConfig contains configuration for vulnerability management tools.
type ToolsetVulnerabilityConfig struct {
	Enabled bool `mapstructure:"enabled"`
}

// ToolConfigManagerConfig contains configuration for config management tools.
type ToolConfigManagerConfig struct {
	Enabled bool `mapstructure:"enabled"`
}

// LoadConfig loads configuration from YAML file and environment variables.
// Environment variables take precedence over YAML configuration.
// Env var naming convention: STACKROX_MCP__SECTION__KEY (double underscore as separator).
// configPath: optional path to YAML configuration file (can be empty).
func LoadConfig(configPath string) (*Config, error) {
	viperInstance := viper.New()

	setDefaults(viperInstance)

	// Set up environment variable support.
	// Note: SetEnvPrefix adds a single underscore, so "STACKROX_MCP_" becomes the prefix.
	// We want double underscores between sections, so we use "__" in the replacer.
	//
	// For environment variable mapping to the config to work, we need to define a default for that config option.
	// Every configuration option that can be set via environment variables must be defined in setDefaults().
	viperInstance.SetEnvPrefix("STACKROX_MCP_")
	viperInstance.SetEnvKeyReplacer(strings.NewReplacer(".", "__"))
	viperInstance.AutomaticEnv()

	if configPath != "" {
		viperInstance.SetConfigFile(configPath)

		if err := viperInstance.ReadInConfig(); err != nil {
			return nil, errors.Wrap(err, "failed to read config file")
		}
	}

	var cfg Config
	if err := viperInstance.Unmarshal(&cfg); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal config")
	}

	if err := cfg.Validate(); err != nil {
		return nil, errors.Wrap(err, "invalid configuration")
	}

	return &cfg, nil
}

// setDefaults sets default values for configuration.
func setDefaults(viper *viper.Viper) {
	viper.SetDefault("central.url", "central.stackrox:443")
	viper.SetDefault("central.auth_type", "passthrough")
	viper.SetDefault("central.api_token", "")
	viper.SetDefault("central.insecure_skip_tls_verify", false)
	viper.SetDefault("central.force_http1", false)

	viper.SetDefault("central.request_timeout", defaultRequestTimeout)
	viper.SetDefault("central.max_retries", defaultMaxRetries)
	viper.SetDefault("central.initial_backoff", defaultInitialBackoff)
	viper.SetDefault("central.max_backoff", defaultMaxBackoff)

	viper.SetDefault("global.read_only_tools", true)

	viper.SetDefault("server.address", "0.0.0.0")
	viper.SetDefault("server.port", defaultPort)
	viper.SetDefault("server.type", ServerTypeStreamableHTTP)

	viper.SetDefault("tools.vulnerability.enabled", false)
	viper.SetDefault("tools.config_manager.enabled", false)
}

// GetURLHostname returns URL hostname.
func (cc *CentralConfig) GetURLHostname() (string, error) {
	parsedURL, err := url.Parse(cc.URL)
	if err == nil && parsedURL.Hostname() != "" {
		return parsedURL.Hostname(), nil
	}

	// Many StackRox configurations use hostname:port format without a scheme,
	// so we add a scheme if missing to ensure proper parsing.
	parsedURL, err = url.Parse("https://" + cc.URL)
	if err != nil {
		return "", errors.Wrapf(err, "failed to parse URL %q", cc.URL)
	}

	return parsedURL.Hostname(), nil
}

//nolint:cyclop
func (cc *CentralConfig) validate() error {
	if cc.URL == "" {
		return errors.New("central.url is required")
	}

	_, err := cc.GetURLHostname()
	if err != nil {
		return errors.Wrap(err, "central.url is not a valid URL")
	}

	if cc.AuthType != AuthTypePassthrough && cc.AuthType != AuthTypeStatic {
		return errors.New("central.auth_type must be either passthrough or static")
	}

	if cc.AuthType == AuthTypeStatic && cc.APIToken == "" {
		return fmt.Errorf("central.api_token is required for %q auth type", AuthTypeStatic)
	}

	if cc.AuthType == AuthTypePassthrough && cc.APIToken != "" {
		return fmt.Errorf("central.api_token can not be set for %q auth type", AuthTypePassthrough)
	}

	if cc.RequestTimeout <= 0 {
		return errors.New("central.request_timeout must be positive")
	}

	if cc.MaxRetries < 0 || cc.MaxRetries > 10 {
		return errors.New("central.max_retries must be between 0 and 10")
	}

	if cc.InitialBackoff <= 0 {
		return errors.New("central.initial_backoff must be positive")
	}

	if cc.MaxBackoff <= 0 {
		return errors.New("central.max_backoff must be positive")
	}

	if cc.MaxBackoff < cc.InitialBackoff {
		return errors.New("central.max_backoff has to be greater than or equal to central.initial_backoff")
	}

	return nil
}

func (sc *ServerConfig) validate() error {
	if sc.Type != ServerTypeStreamableHTTP && sc.Type != ServerTypeStdio {
		return errors.New("server.type must be either streamable-http or stdio")
	}

	if sc.Type == ServerTypeStdio {
		return nil
	}

	if sc.Address == "" {
		return errors.New("server.address is required")
	}

	if sc.Port < 1 || sc.Port > 65535 {
		return errors.New("server.port must be between 1 and 65535")
	}

	return nil
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	if err := c.Central.validate(); err != nil {
		return err
	}

	if err := c.Server.validate(); err != nil {
		return err
	}

	if !c.Tools.Vulnerability.Enabled && !c.Tools.ConfigManager.Enabled {
		return errors.New("at least one tool has to be enabled")
	}

	if c.Server.Type == ServerTypeStdio && c.Central.AuthType != AuthTypeStatic {
		return errors.New("stdio server does require static auth type")
	}

	return nil
}

const redacted = "***REDACTED***"

// Redacted returns a copy of the configuration with sensitive data redacted.
// This is useful for logging configuration without exposing secrets.
func (c *Config) Redacted() *Config {
	redactedConfig := *c
	redactedConfig.Central = c.Central.redacted()

	return &redactedConfig
}

// redacted returns a copy of CentralConfig with sensitive data redacted.
func (cc *CentralConfig) redacted() CentralConfig {
	redactedCentral := *cc
	if cc.APIToken != "" {
		redactedCentral.APIToken = redacted
	}

	return redactedCentral
}
