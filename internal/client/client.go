// Package client holds implementation of StachRock Central API client.
package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log/slog"
	"os"
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
	maxCACertFileSize = 1 << 20 // 1MB
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

	tlsCfg := &tls.Config{
		InsecureSkipVerify: c.config.InsecureSkipTLSVerify, //nolint:gosec
		MinVersion:         tls.VersionTLS12,
		ServerName:         hostname,
	}

	// There is no reason to load certificates if we allow InsecureSkipTLSVerify.
	if !c.config.InsecureSkipTLSVerify && c.config.CACertPath != "" {
		certPool, err := loadCACertPool(c.config.CACertPath)
		if err != nil {
			return nil, err
		}

		tlsCfg.RootCAs = certPool
	}

	return tlsCfg, nil
}

func loadCACertPool(caCertPath string) (*x509.CertPool, error) {
	// File size guard
	fileInfo, err := os.Stat(caCertPath)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to access CA certificate at %s", caCertPath)
	}

	if !fileInfo.Mode().IsRegular() {
		return nil, errors.Errorf("CA certificate path %s is not a regular file", caCertPath)
	}

	if fileInfo.Size() == 0 {
		return nil, errors.Errorf("CA certificate file %s is empty", caCertPath)
	}

	if fileInfo.Size() > maxCACertFileSize {
		return nil, errors.Errorf(
			"CA certificate file %s is too large (%d bytes, max %d)",
			caCertPath, fileInfo.Size(),
			maxCACertFileSize,
		)
	}

	//nolint: gosec
	caCert, err := os.ReadFile(caCertPath)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read CA certificate from %s", caCertPath)
	}

	// Get system cert pool, warn on fallback
	certPool, err := x509.SystemCertPool()
	if err != nil {
		slog.Warn("Failed to load system CA pool, using custom CA only", "error", err)

		certPool = x509.NewCertPool()
	}

	if !certPool.AppendCertsFromPEM(caCert) {
		return nil, errors.Errorf("failed to parse CA certificate from %s: no valid PEM data found", caCertPath)
	}

	showCertInfo(caCert)

	return certPool, nil
}

// showCertInfo parses and logs certificate metadata.
func showCertInfo(caCert []byte) {
	block, _ := pem.Decode(caCert)
	if block == nil {
		slog.Warn("Unable to decode CA certificate")

		return
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		slog.Warn("Failed to parse CA certificate", "error", err)

		return
	}

	slog.Info("Loaded CA certificate",
		"subject", cert.Subject.CommonName,
		"issuer", cert.Issuer.CommonName,
		"notAfter", cert.NotAfter,
		"isCA", cert.IsCA,
	)

	if !cert.IsCA {
		slog.Warn("Provided certificate does not have the CA basic constraint set — TLS verification may fail",
			"subject", cert.Subject.CommonName,
		)
	}

	if time.Now().After(cert.NotAfter) {
		slog.Warn("CA certificate is expired — TLS verification will fail",
			"subject", cert.Subject.CommonName,
			"expiredAt", cert.NotAfter,
		)
	}

	if time.Now().Before(cert.NotBefore) {
		slog.Warn("CA certificate is not yet valid",
			"subject", cert.Subject.CommonName,
			"validFrom", cert.NotBefore,
		)
	}
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
