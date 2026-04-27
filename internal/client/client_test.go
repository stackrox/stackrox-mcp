package client

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/stackrox/stackrox-mcp/internal/config"
	"github.com/stackrox/stackrox-mcp/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
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

// generateTestCAWithKey creates a CA certificate and returns the PEM, parsed cert, and private key.
func generateTestCAWithKey(t *testing.T) ([]byte, *x509.Certificate, *ecdsa.PrivateKey) {
	t.Helper()

	caKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	template := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "Test CA"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		IsCA:                  true,
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &caKey.PublicKey, caKey)
	require.NoError(t, err)

	caCert, err := x509.ParseCertificate(certDER)
	require.NoError(t, err)

	caCertPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	return caCertPEM, caCert, caKey
}

// generateTestCACert creates a self-signed CA certificate PEM for testing.
func generateTestCACert(t *testing.T) []byte {
	t.Helper()

	encode, _, _ := generateTestCAWithKey(t)

	return encode
}

// writeTestFile creates a file with the given content in a temp directory and returns the path.
func writeTestFile(t *testing.T, name string, content []byte) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), name)
	err := os.WriteFile(path, content, 0o600)
	require.NoError(t, err)

	return path
}

func TestLoadCACertPool_NonexistentFile(t *testing.T) {
	_, err := loadCACertPool("/nonexistent/path/to/ca.crt")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to access CA certificate")
	// Error should include the file path for debuggability.
	assert.Contains(t, err.Error(), "/nonexistent/path/to/ca.crt")
}

func TestLoadCACertPool_EmptyFile(t *testing.T) {
	path := writeTestFile(t, "empty-ca.crt", []byte{})

	_, err := loadCACertPool(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "is empty")
}

func TestLoadCACertPool_FileTooLarge(t *testing.T) {
	// Create file just over the 1MB limit.
	oversized := make([]byte, maxCACertFileSize+1)
	path := writeTestFile(t, "oversized-ca.crt", oversized)

	_, err := loadCACertPool(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "too large")
	// Error message should include both actual size and max size for debuggability.
	assert.Contains(t, err.Error(), fmt.Sprintf("%d bytes", maxCACertFileSize+1))
	assert.Contains(t, err.Error(), fmt.Sprintf("max %d", maxCACertFileSize))
}

func TestLoadCACertPool_InvalidPEM(t *testing.T) {
	path := writeTestFile(t, "invalid-ca.crt", []byte("this is not a PEM certificate"))

	_, err := loadCACertPool(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no valid PEM data found")
}

func TestLoadCACertPool_ValidCACert(t *testing.T) {
	certPEM := generateTestCACert(t)
	path := writeTestFile(t, "valid-ca.crt", certPEM)

	pool, err := loadCACertPool(path)
	require.NoError(t, err)
	require.NotNil(t, pool)
}

func TestLoadCACertPool_UnreadableFile(t *testing.T) {
	certPEM := generateTestCACert(t)
	path := writeTestFile(t, "unreadable-ca.crt", certPEM)

	// Remove read permission.
	err := os.Chmod(path, 0o000)
	require.NoError(t, err)

	_, err = loadCACertPool(path)
	require.Error(t, err)
	// Depending on the platform, os.Stat or os.ReadFile will fail.
	// The error should reference the file path.
	assert.Contains(t, err.Error(), "unreadable-ca.crt")
}

func TestLoadCACertPool_DirectoryPath(t *testing.T) {
	// Pointing ca_cert_path at a directory should produce a clear error, not a panic.
	dir := t.TempDir()

	_, err := loadCACertPool(dir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not a regular file")
	assert.Contains(t, err.Error(), dir)
}

func TestClient_tlsConfig_WithCACertPath(t *testing.T) {
	certPEM := generateTestCACert(t)
	path := writeTestFile(t, "ca.crt", certPEM)

	client := &Client{
		config: &config.CentralConfig{
			URL:        "central.stackrox.io:8443",
			CACertPath: path,
		},
	}

	tlsCfg, err := client.tlsConfig()
	require.NoError(t, err)
	require.NotNil(t, tlsCfg)
	assert.NotNil(t, tlsCfg.RootCAs, "RootCAs should be set when CACertPath is provided")
}

func TestClient_tlsConfig_EmptyCACertPath(t *testing.T) {
	client := &Client{
		config: &config.CentralConfig{
			URL:        "central.stackrox.io:8443",
			CACertPath: "",
		},
	}

	tlsCfg, err := client.tlsConfig()
	require.NoError(t, err)
	require.NotNil(t, tlsCfg)
	assert.Nil(t, tlsCfg.RootCAs, "RootCAs should be nil when CACertPath is empty")
}

func TestClient_tlsConfig_NonexistentCACertPath(t *testing.T) {
	client := &Client{
		config: &config.CentralConfig{
			URL:        "central.stackrox.io:8443",
			CACertPath: "/nonexistent/ca.crt",
		},
	}

	_, err := client.tlsConfig()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to access CA certificate")
}

// generateTestCert creates a certificate PEM with the given options, signed by the given CA.
// If ca/caKey are nil, the cert is self-signed.
func generateTestCert(
	t *testing.T, template *x509.Certificate, caCert *x509.Certificate, caKey *ecdsa.PrivateKey,
) []byte {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	parent := template
	signingKey := key

	if caCert != nil && caKey != nil {
		parent = caCert
		signingKey = caKey
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, parent, &key.PublicKey, signingKey)
	require.NoError(t, err)

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	return certPEM
}

func TestLoadCACertPool_MultiCertBundle(t *testing.T) {
	// A PEM certs with multiple CA certificates should add ALL certs to the pool,
	// even though diagnostic logging only inspects the first PEM block.
	ca1Template := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "Test CA 1"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		IsCA:                  true,
		BasicConstraintsValid: true,
	}

	ca2Template := &x509.Certificate{
		SerialNumber:          big.NewInt(2),
		Subject:               pkix.Name{CommonName: "Test CA 2"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		IsCA:                  true,
		BasicConstraintsValid: true,
	}

	cert1 := generateTestCert(t, ca1Template, nil, nil)
	cert2 := generateTestCert(t, ca2Template, nil, nil)

	certs := append(cert1, cert2...) //nolint: gocritic
	path := writeTestFile(t, "certs-ca.crt", certs)

	pool, err := loadCACertPool(path)
	require.NoError(t, err)
	require.NotNil(t, pool)

	expectedPool, err := x509.SystemCertPool()
	require.NoError(t, err)

	expectedPool.AppendCertsFromPEM(cert1)
	expectedPool.AppendCertsFromPEM(cert2)

	assert.True(t, expectedPool.Equal(pool))
}

// generateTestServerCert creates a server certificate signed by the given CA, with SANs for 127.0.0.1.
func generateTestServerCert(t *testing.T, caCert *x509.Certificate, caKey *ecdsa.PrivateKey) ([]byte, []byte) {
	t.Helper()

	serverKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	template := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject:      pkix.Name{CommonName: "localhost"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1)},
		DNSNames:     []string{"localhost"},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, caCert, &serverKey.PublicKey, caKey)
	require.NoError(t, err)

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	keyDER, err := x509.MarshalECPrivateKey(serverKey)
	require.NoError(t, err)

	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})

	return certPEM, keyPEM
}

// startTLSGRPCServer starts a gRPC server with TLS on a random port.
func startTLSGRPCServer(t *testing.T, certPEM, keyPEM []byte) net.Listener {
	t.Helper()

	serverCert, err := tls.X509KeyPair(certPEM, keyPEM)
	require.NoError(t, err)

	tlsCfg := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		MinVersion:   tls.VersionTLS12,
		NextProtos:   []string{"h2"},
	}

	lis, err := tls.Listen("tcp", "127.0.0.1:0", tlsCfg)
	require.NoError(t, err)

	srv := grpc.NewServer()

	go func() { _ = srv.Serve(lis) }()

	t.Cleanup(func() { srv.Stop() })

	return lis
}

func checkRawConn(t *testing.T, lis net.Listener, caCertPEM []byte) {
	t.Helper()

	// Verify the TLS handshake works at the raw TCP level first.
	caCertPool := x509.NewCertPool()
	require.True(t, caCertPool.AppendCertsFromPEM(caCertPEM))

	dialer := tls.Dialer{
		Config: &tls.Config{
			RootCAs:    caCertPool,
			MinVersion: tls.VersionTLS12,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rawConn, err := dialer.DialContext(ctx, "tcp", lis.Addr().String())
	require.NoError(t, err, "raw TLS dial should succeed with CA cert")
	require.NoError(t, rawConn.Close())
}

func TestClient_ConnectWithCACert_Positive(t *testing.T) {
	caCertPEM, caCert, caKey := generateTestCAWithKey(t)
	serverCertPEM, serverKeyPEM := generateTestServerCert(t, caCert, caKey)

	lis := startTLSGRPCServer(t, serverCertPEM, serverKeyPEM)
	caCertPath := writeTestFile(t, "ca.crt", caCertPEM)

	checkRawConn(t, lis, caCertPEM)

	cfg := &config.CentralConfig{
		URL:            lis.Addr().String(),
		AuthType:       config.AuthTypeStatic,
		APIToken:       "dummy",
		CACertPath:     caCertPath,
		RequestTimeout: 2 * time.Second,
		MaxRetries:     0,
		InitialBackoff: time.Millisecond,
		MaxBackoff:     5 * time.Millisecond,
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)

	defer func() { assert.NoError(t, client.Close()) }()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	require.NoError(t, client.Connect(ctx))

	conn := client.Conn()
	require.NotNil(t, conn)

	conn.Connect()

	for {
		state := conn.GetState()
		if state == connectivity.Ready {
			break
		}

		if state == connectivity.TransientFailure {
			t.Fatalf("connection entered TransientFailure — TLS handshake likely failed (addr=%s)", lis.Addr().String())
		}

		if !conn.WaitForStateChange(ctx, state) {
			t.Fatalf("timeout waiting for connection to become Ready (last state: %s)", state)
		}
	}
}

func TestClient_ConnectWithoutCACert_Negative(t *testing.T) {
	_, caCert, caKey := generateTestCAWithKey(t)
	serverCertPEM, serverKeyPEM := generateTestServerCert(t, caCert, caKey)

	lis := startTLSGRPCServer(t, serverCertPEM, serverKeyPEM)

	cfg := &config.CentralConfig{
		URL:                   lis.Addr().String(),
		AuthType:              config.AuthTypeStatic,
		APIToken:              "dummy",
		InsecureSkipTLSVerify: false,
		RequestTimeout:        2 * time.Second,
		MaxRetries:            0,
		InitialBackoff:        time.Millisecond,
		MaxBackoff:            5 * time.Millisecond,
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)

	defer func() { assert.NoError(t, client.Close()) }()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	require.NoError(t, client.Connect(ctx))

	conn := client.Conn()
	require.NotNil(t, conn)

	conn.Connect()

	for {
		state := conn.GetState()
		if state == connectivity.TransientFailure {
			break
		}

		if !conn.WaitForStateChange(ctx, state) {
			break
		}
	}

	assert.NotEqual(t, connectivity.Ready, conn.GetState(),
		"connection must not reach Ready state without CA cert for self-signed server")
}

func TestLoadCACertPool_MixedPEMContent(t *testing.T) {
	// A PEM file containing a valid CERTIFICATE block plus a PRIVATE KEY block.
	// AppendCertsFromPEM ignores non-CERTIFICATE blocks and returns true if at least
	// one cert was found. The pool should be functional.
	mixedCerts := generateTestCACert(t)

	// Generate a throwaway private key PEM block.
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	keyDER, err := x509.MarshalECPrivateKey(key)
	require.NoError(t, err)

	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})

	// Bundle cert + key into one file.
	mixedCerts = append(mixedCerts, keyPEM...)
	path := writeTestFile(t, "mixed-ca.crt", mixedCerts)

	pool, err := loadCACertPool(path)
	require.NoError(t, err, "mixed PEM content with at least one valid cert should succeed")
	require.NotNil(t, pool)
}
