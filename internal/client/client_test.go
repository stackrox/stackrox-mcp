package client

import (
	"context"
	"crypto/tls"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/stackrox/stackrox-mcp/internal/config"
	"github.com/stackrox/stackrox-mcp/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func TestClientReconnectsAfterServerRestart(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	listener, server := startTestGRPCServer(t, "127.0.0.1:"+strconv.Itoa(testutil.GetPortForTest(t)))

	cfg := &config.CentralConfig{
		URL:                   listener.Addr().String(),
		AuthType:              config.AuthTypeStatic,
		APIToken:              "dummy",
		InsecureSkipTLSVerify: true,
		RequestTimeout:        time.Second,
		MaxRetries:            3,
		InitialBackoff:        time.Millisecond,
		MaxBackoff:            5 * time.Millisecond,
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)

	defer func() { assert.NoError(t, client.Close()) }()

	require.NoError(t, client.Connect(ctx))
	initialConn := client.Conn()
	require.NotNil(t, initialConn)

	// Simulate server failure.
	server.Stop()
	// Note: server.Stop() already closes the listener, so we don't need to close it explicitly.

	waitCtx, waitCancel := context.WithTimeout(ctx, time.Second)
	initialConn.WaitForStateChange(waitCtx, initialConn.GetState())
	waitCancel()

	// Restart server on the same address.
	_, server2 := startTestGRPCServer(t, cfg.URL)

	defer server2.Stop()

	require.NoError(t, client.Connect(ctx))
	reconnected := client.Conn()
	require.NotNil(t, reconnected)
}

func startTestGRPCServer(t *testing.T, addr string) (net.Listener, *grpc.Server) {
	t.Helper()

	//nolint:noctx
	lis, err := net.Listen("tcp", addr)
	require.NoError(t, err)

	srv := grpc.NewServer()

	go func() {
		_ = srv.Serve(lis)
	}()

	return lis, srv
}

func TestClient_tlsConfig(t *testing.T) {
	tests := map[string]struct {
		url            string
		expectedServer string
	}{
		"hostname": {
			url:            "central.stackrox.io:8443",
			expectedServer: "central.stackrox.io",
		},
		"https scheme": {
			url:            "https://central.stackrox.io:8443",
			expectedServer: "central.stackrox.io",
		},
		"IP address": {
			url:            "192.168.1.100:8443",
			expectedServer: "192.168.1.100",
		},
		"service name": {
			url:            "central.stackrox:443",
			expectedServer: "central.stackrox",
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			client := &Client{
				config: &config.CentralConfig{
					URL: testCase.url,
				},
			}

			tlsCfg, err := client.tlsConfig()

			require.NoError(t, err)
			require.NotNil(t, tlsCfg)
			assert.Equal(t, testCase.expectedServer, tlsCfg.ServerName)
			assert.Equal(t, uint16(tls.VersionTLS12), tlsCfg.MinVersion) // TLS 1.2
		})
	}
}

func TestClient_tlsConfig_insecureSkipVerify(t *testing.T) {
	client := &Client{
		config: &config.CentralConfig{
			URL:                   "central.stackrox.io:8443",
			InsecureSkipTLSVerify: true,
		},
	}

	tlsCfg, err := client.tlsConfig()

	require.NoError(t, err)
	require.NotNil(t, tlsCfg)
	assert.True(t, tlsCfg.InsecureSkipVerify)

	client.config.InsecureSkipTLSVerify = false

	tlsCfg, err = client.tlsConfig()

	require.NoError(t, err)
	require.NotNil(t, tlsCfg)
	assert.False(t, tlsCfg.InsecureSkipVerify)
}
