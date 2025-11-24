// Package config provides configuration handling for StackRox MCP server.
package config

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

// Config represents the complete application configuration.
type Config struct {
	Central CentralConfig `mapstructure:"central"`
	Global  GlobalConfig  `mapstructure:"global"`
	Tools   ToolsConfig   `mapstructure:"tools"`
}

// CentralConfig contains StackRox Central connection configuration.
type CentralConfig struct {
	URL        string `mapstructure:"url"`
	Insecure   bool   `mapstructure:"insecure"`
	ForceHTTP1 bool   `mapstructure:"force_http1"`
}

// GlobalConfig contains global MCP server configuration.
type GlobalConfig struct {
	ReadOnlyTools bool `mapstructure:"read_only_tools"`
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
	viper.SetDefault("central.url", "central.stackrox:8443")
	viper.SetDefault("central.insecure", false)
	viper.SetDefault("central.force_http1", false)

	viper.SetDefault("global.read_only_tools", true)

	viper.SetDefault("tools.vulnerability.enabled", false)
	viper.SetDefault("tools.config_manager.enabled", false)
}

var (
	errURLRequired    = errors.New("central.url is required")
	errAtLeastOneTool = errors.New("at least one tool has to be enabled")
)

// Validate validates the configuration.
func (c *Config) Validate() error {
	if c.Central.URL == "" {
		return errURLRequired
	}

	if !c.Tools.Vulnerability.Enabled && !c.Tools.ConfigManager.Enabled {
		return errAtLeastOneTool
	}

	return nil
}
