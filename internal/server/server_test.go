package server

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/stackrox/stackrox-mcp/internal/config"
	"github.com/stackrox/stackrox-mcp/internal/testutil"
	"github.com/stackrox/stackrox-mcp/internal/toolsets"
	"github.com/stackrox/stackrox-mcp/internal/toolsets/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getDefaultConfig returns a default config for tests.
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

func TestNewServer(t *testing.T) {
	cfg := getDefaultConfig()

	registry := toolsets.NewRegistry(cfg, []toolsets.Toolset{})

	srv := NewServer(cfg, registry)

	require.NotNil(t, srv)
	assert.Equal(t, cfg, srv.cfg)
	assert.Equal(t, registry, srv.registry)
	assert.NotNil(t, srv.mcp)
}

func TestServer_registerTools_AllEnabled(t *testing.T) {
	cfg := getDefaultConfig()
	cfg.Global.ReadOnlyTools = false

	readOnlyTestTool := mock.NewTool("test_read_only_tool", true)
	readWriteTestTool := mock.NewTool("test_read_write_tool", false)

	toolsetList := []toolsets.Toolset{
		mock.NewToolset("test_toolset", true, []toolsets.Tool{readOnlyTestTool, readWriteTestTool}),
	}

	registry := toolsets.NewRegistry(cfg, toolsetList)
	srv := NewServer(cfg, registry)

	srv.registerTools()

	assert.True(t, readOnlyTestTool.RegisterCalled, "read-only test tool should be registered")
	assert.True(t, readWriteTestTool.RegisterCalled, "read-write test tool should be registered")
}

func TestServer_registerTools_ReadOnlyMode(t *testing.T) {
	cfg := getDefaultConfig()
	cfg.Global.ReadOnlyTools = true

	readOnlyTestTool := mock.NewTool("test_read_only_tool", true)
	readWriteTestTool := mock.NewTool("test_read_write_tool", false)

	toolsetList := []toolsets.Toolset{
		mock.NewToolset("test_toolset", true, []toolsets.Tool{readOnlyTestTool, readWriteTestTool}),
	}

	registry := toolsets.NewRegistry(cfg, toolsetList)
	srv := NewServer(cfg, registry)

	srv.registerTools()

	assert.True(t, readOnlyTestTool.RegisterCalled, "read-only test tool should be registered")
	assert.False(t, readWriteTestTool.RegisterCalled, "read-write test tool should not be registered in read-only mode")
}

func TestServer_registerTools_DisabledToolset(t *testing.T) {
	cfg := getDefaultConfig()
	cfg.Global.ReadOnlyTools = false

	enabledTestTool := mock.NewTool("test_enabled_tool", true)
	disabledTestTool := mock.NewTool("test_disabled_tool", true)

	toolsetList := []toolsets.Toolset{
		mock.NewToolset("enabled_toolset", true, []toolsets.Tool{enabledTestTool}),
		mock.NewToolset("disabled_toolset", false, []toolsets.Tool{disabledTestTool}),
	}

	registry := toolsets.NewRegistry(cfg, toolsetList)
	srv := NewServer(cfg, registry)

	srv.registerTools()

	assert.True(t, enabledTestTool.RegisterCalled, "tool from enabled toolset should be registered")
	assert.False(t, disabledTestTool.RegisterCalled, "tool from disabled toolset should not be registered")
}

func TestServer_Start(t *testing.T) {
	cfg := getDefaultConfig()
	cfg.Server.Port = testutil.GetPortForTest(t)

	testTool := mock.NewTool("test_tool", true)

	toolsetList := []toolsets.Toolset{
		mock.NewToolset("test_toolset", true, []toolsets.Tool{testTool}),
	}

	registry := toolsets.NewRegistry(cfg, toolsetList)
	srv := NewServer(cfg, registry)

	ctx, cancel := context.WithCancel(context.Background())

	errChan := make(chan error, 1)

	go func() {
		errChan <- srv.Start(ctx)
	}()

	serverURL := "http://" + net.JoinHostPort(cfg.Server.Address, strconv.Itoa(cfg.Server.Port))
	err := testutil.WaitForServerReady(serverURL, 3*time.Second)
	require.NoError(t, err, "Server should start within timeout")

	// Verify tools were registered.
	assert.True(t, testTool.RegisterCalled, "test tool should be registered when server starts")

	// Establish actual HTTP connection to verify server is responding.
	//nolint:gosec,noctx
	resp, err := http.Get(serverURL)
	if err == nil {
		_ = resp.Body.Close()
	}
	// We don't require a successful response, just that we can connect
	require.NoError(t, err, "Should be able to establish HTTP connection to server")

	// Trigger graceful shutdown.
	cancel()

	// Wait for server to shut down.
	select {
	case err := <-errChan:
		// Server should shut down cleanly.
		if err != nil && errors.Is(err, context.Canceled) {
			t.Errorf("Server returned unexpected error: %v", err)
		}
	case <-time.After(shutdownTimeout):
		t.Fatal("Server did not shut down within timeout period")
	}
}
