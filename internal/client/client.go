// Package client holds implementation of StachRock Central API client.
package client

import (
	"context"
	"crypto/tls"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stackrox/rox/pkg/grpc/alpn"
	"github.com/stackrox/stackrox-mcp/internal/client/auth"
	"github.com/stackrox/stackrox-mcp/internal/config"
	http1client "golang.stackrox.io/grpc-http1/client"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
)

const (
	minConnectTimeout = 5 * time.Second
	backoffJitter     = 0.2
)

// Client provides gRPC connection to StackRox Central API.
type Client struct {
	config *config.CentralConfig

	mu        sync.RWMutex
	conn      *grpc.ClientConn
	connected bool
}

// NewClient creates a new client with the given configuration and options.
func NewClient(config *config.CentralConfig) (*Client, error) {
	if config == nil {
		return nil, errors.New("config cannot be nil")
	}

	return &Client{
		config:    config,
		connected: false,
	}, nil
}

// Connect establishes a connection to StackRox Central.
// Must be called before any API requests.
func (c *Client) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.shouldRedialNoLock() {
		return nil
	}

	c.resetConnectionNoLock()

	dialOpts, err := c.buildDialOptions()
	if err != nil {
		return err
	}

	tlsConfig, err := c.tlsConfig()
	if err != nil {
		return err
	}

	var conn *grpc.ClientConn
	if c.config.ForceHTTP1 {
		conn, err = c.connectHTTP1(ctx, dialOpts, tlsConfig)
	} else {
		transportDailOpt := grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig))
		dialOpts = append([]grpc.DialOption{transportDailOpt}, dialOpts...)
		conn, err = grpc.NewClient(c.config.URL, dialOpts...)
	}

	if err != nil {
		return NewError(err, "Connect")
	}

	c.conn = conn
	c.connected = true

	return nil
}

// Close gracefully closes the connection to StackRox Central.
// Safe to call multiple times.
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected || c.conn == nil {
		return nil
	}

	err := c.conn.Close()
	c.conn = nil
	c.connected = false

	return errors.Wrap(err, "failed to close connection")
}

// IsConnected returns true if the client is connected to Central.
func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.connected && c.conn != nil
}

// Conn returns the underlying gRPC connection for creating service clients.
// Tools use this to instantiate their own typed service clients.
//
// Example usage:
//
//	conn := client.Conn()
//	deploymentClient := v1.NewDeploymentServiceClient(conn)
//
// Returns nil if client is not connected.
func (c *Client) Conn() *grpc.ClientConn {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.conn
}

// ReadyConn ensures the connection to Central is healthy and returns it.
func (c *Client) ReadyConn(ctx context.Context) (*grpc.ClientConn, error) {
	if err := c.Connect(ctx); err != nil {
		return nil, err
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.conn == nil {
		return nil, errors.New("client is not connected to StackRox Central")
	}

	return c.conn, nil
}

// SetConnForTesting sets a gRPC connection for testing purposes.
// This should only be used in tests.
func (c *Client) SetConnForTesting(t *testing.T, conn *grpc.ClientConn) {
	t.Helper()

	c.mu.Lock()
	defer c.mu.Unlock()

	c.conn = conn
	c.connected = true
}

func (c *Client) shouldRedialNoLock() bool {
	if !c.connected || c.conn == nil {
		return true
	}

	state := c.conn.GetState()
	//nolint:exhaustive
	switch state {
	case connectivity.TransientFailure, connectivity.Shutdown:
		return true
	default:
		return false
	}
}

func (c *Client) resetConnectionNoLock() {
	if c.conn != nil {
		_ = c.conn.Close()
		c.conn = nil
	}

	c.connected = false
}

func (c *Client) buildDialOptions() ([]grpc.DialOption, error) {
	retryPolicy := NewRetryPolicy(c.config)

	dialOpts := []grpc.DialOption{
		grpc.WithChainUnaryInterceptor(
			createLoggingInterceptor(),
			createRetryInterceptor(retryPolicy, c.config.RequestTimeout),
		),
		grpc.WithConnectParams(grpc.ConnectParams{
			Backoff: backoff.Config{
				BaseDelay:  c.config.InitialBackoff,
				Multiplier: backoffMultiplier,
				Jitter:     backoffJitter,
				MaxDelay:   c.config.MaxBackoff,
			},
			MinConnectTimeout: minConnectTimeout,
		}),
	}

	authOpt, err := c.perRPCCredentialsOption()
	if err != nil {
		return nil, err
	}

	dialOpts = append(dialOpts, authOpt)

	return dialOpts, nil
}

func (c *Client) perRPCCredentialsOption() (grpc.DialOption, error) {
	switch c.config.AuthType {
	case config.AuthTypeStatic:
		return grpc.WithPerRPCCredentials(auth.NewStaticTokenCredentials(c.config.APIToken)), nil
	case config.AuthTypePassthrough:
		return grpc.WithPerRPCCredentials(auth.NewPassthroughTokenCredentials()), nil
	default:
		return nil, fmt.Errorf("unsupported auth type: %s", c.config.AuthType)
	}
}

func (c *Client) tlsConfig() (*tls.Config, error) {
	// Extract hostname for TLS verification.
	// This is especially important for force_http1 mode where the gRPC-HTTP/1 bridge
	// needs explicit ServerName for certificate validation.
	hostname, err := c.config.GetURLHostname()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get central URL hostname")
	}

	return &tls.Config{
		InsecureSkipVerify: c.config.InsecureSkipTLSVerify, //nolint:gosec
		MinVersion:         tls.VersionTLS12,
		ServerName:         hostname,
	}, nil
}

func (c *Client) connectHTTP1(
	ctx context.Context,
	dialOpts []grpc.DialOption,
	tlsConfig *tls.Config,
) (*grpc.ClientConn, error) {
	connectOpts := []http1client.ConnectOption{
		http1client.ForceDowngrade(true),
		http1client.ExtraH2ALPNs(alpn.PureGRPCALPNString),
		http1client.DialOpts(dialOpts...),
	}

	http1Client, err := http1client.ConnectViaProxy(ctx, c.config.URL, tlsConfig, connectOpts...)

	return http1Client, errors.Wrap(err, "unable to connect via http1")
}
