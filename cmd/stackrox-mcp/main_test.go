package main

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/stackrox/stackrox-mcp/internal/client"
	"github.com/stackrox/stackrox-mcp/internal/config"
	"github.com/stackrox/stackrox-mcp/internal/server"
	"github.com/stackrox/stackrox-mcp/internal/testutil"
	"github.com/stackrox/stackrox-mcp/internal/toolsets"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetToolsets(t *testing.T) {
	allToolsets := getToolsets(&config.Config{}, &client.Client{})

	toolsetNames := []string{}
	for _, toolset := range allToolsets {
		toolsetNames = append(toolsetNames, toolset.GetName())
	}

	assert.Contains(t, toolsetNames, "config_manager")
	assert.Contains(t, toolsetNames, "vulnerability")
}

func TestGracefulShutdown(t *testing.T) {
	// Set up minimal valid config. config.LoadConfig() validates configuration.
	t.Setenv("STACKROX_MCP__TOOLS__VULNERABILITY__ENABLED", "true")

	cfg, err := config.LoadConfig("")
	require.NoError(t, err)
	require.NotNil(t, cfg)
	cfg.Server.Port = testutil.GetPortForTest(t)

	registry := toolsets.NewRegistry(cfg, getToolsets(cfg, &client.Client{}))
	srv := server.NewServer(cfg, registry)
	ctx, cancel := context.WithCancel(context.Background())

	errChan := make(chan error, 1)

	go func() {
		errChan <- srv.Start(ctx)
	}()

	serverURL := "http://" + net.JoinHostPort(cfg.Server.Address, strconv.Itoa(cfg.Server.Port))
	err = testutil.WaitForServerReady(t, serverURL, 3*time.Second)
	require.NoError(t, err, "Server should start within timeout")

	// Establish actual HTTP connection to verify server is responding.
	//nolint:gosec,noctx
	resp, err := http.Get(serverURL)
	require.NoError(t, err, "Should be able to establish HTTP connection to server")

	require.NoError(t, resp.Body.Close())

	// Simulate shutdown signal by canceling context.
	cancel()

	// Wait for server to shut down.
	select {
	case err := <-errChan:
		// Server should shut down cleanly (either nil or context.Canceled).
		if err != nil && errors.Is(err, context.Canceled) {
			t.Errorf("Server returned unexpected error: %v", err)
		}
	case <-time.After(server.ShutdownTimeout):
		t.Fatal("Server did not shut down within timeout period")
	}
}
