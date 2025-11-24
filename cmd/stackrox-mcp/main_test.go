package main

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/stackrox/stackrox-mcp/internal/config"
	"github.com/stackrox/stackrox-mcp/internal/server"
	"github.com/stackrox/stackrox-mcp/internal/testutil"
	"github.com/stackrox/stackrox-mcp/internal/toolsets"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getDefaultConfig() *config.Config {
	return &config.Config{
		Global: config.GlobalConfig{
			ReadOnlyTools: false,
		},
		Central: config.CentralConfig{
			URL: "central.example.com:8443",
		},
		Server: config.ServerConfig{
			Address: "localhost",
			Port:    8080,
		},
		Tools: config.ToolsConfig{
			Vulnerability: config.ToolsetVulnerabilityConfig{
				Enabled: true,
			},
			ConfigManager: config.ToolConfigManagerConfig{
				Enabled: false,
			},
		},
	}
}

func TestGetToolsets(t *testing.T) {
	cfg := getDefaultConfig()
	cfg.Tools.ConfigManager.Enabled = true

	allToolsets := getToolsets(cfg)

	require.NotNil(t, allToolsets)
	assert.Len(t, allToolsets, 2, "Should have 2 allToolsets")
	assert.Equal(t, "config_manager", allToolsets[0].GetName())
	assert.Equal(t, "vulnerability", allToolsets[1].GetName())
}

func TestGracefulShutdown(t *testing.T) {
	// Set up minimal valid config.
	t.Setenv("STACKROX_MCP__TOOLS__VULNERABILITY__ENABLED", "true")

	cfg, err := config.LoadConfig("")
	require.NoError(t, err)
	require.NotNil(t, cfg)
	cfg.Server.Port = testutil.GetPortForTest(t)

	registry := toolsets.NewRegistry(cfg, getToolsets(cfg))
	srv := server.NewServer(cfg, registry)
	ctx, cancel := context.WithCancel(context.Background())

	errChan := make(chan error, 1)

	go func() {
		errChan <- srv.Start(ctx)
	}()

	serverURL := "http://" + net.JoinHostPort(cfg.Server.Address, strconv.Itoa(cfg.Server.Port))
	err = testutil.WaitForServerReady(serverURL, 3*time.Second)
	require.NoError(t, err, "Server should start within timeout")

	// Establish actual HTTP connection to verify server is responding.
	//nolint:gosec,noctx
	resp, err := http.Get(serverURL)
	if err == nil {
		_ = resp.Body.Close()
	}

	require.NoError(t, err, "Should be able to establish HTTP connection to server")

	// Simulate shutdown signal by canceling context.
	cancel()

	// Wait for server to shut down.
	select {
	case err := <-errChan:
		// Server should shut down cleanly (either nil or context.Canceled).
		if err != nil && errors.Is(err, context.Canceled) {
			t.Errorf("Server returned unexpected error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Server did not shut down within timeout period")
	}
}
