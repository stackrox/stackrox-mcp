// Package server represents MCP server.
package server

import (
	"context"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pkg/errors"
	"github.com/stackrox/stackrox-mcp/internal/config"
	"github.com/stackrox/stackrox-mcp/internal/toolsets"
)

const (
	// ShutdownTimeout represents allowed timeout for graceful shutdown to finish.
	ShutdownTimeout = 5 * time.Second

	readHeaderTimeout = 5 * time.Second
)

// version is set at build time via ldflags (ldflags can't modify constants).
var version = "dev"

// Server represents the MCP HTTP server.
type Server struct {
	cfg      *config.Config
	registry *toolsets.Registry
	mcp      *mcp.Server
}

// NewServer creates a new MCP server instance.
func NewServer(cfg *config.Config, registry *toolsets.Registry) *Server {
	mcpServer := mcp.NewServer(
		&mcp.Implementation{
			Name:    "stackrox-mcp",
			Version: version,
		},
		nil,
	)

	return &Server{
		cfg:      cfg,
		registry: registry,
		mcp:      mcpServer,
	}
}

// Start starts the HTTP server with Streamable HTTP transport.
// If stdin/stdout are provided (non-nil), they will be used for stdio transport.
// If they are nil, os.Stdin/os.Stdout will be used (production mode).
func (s *Server) Start(ctx context.Context, stdin io.ReadCloser, stdout io.WriteCloser) error {
	s.registerTools()

	if s.cfg.Server.Type == config.ServerTypeStdio {
		var transport mcp.Transport
		if stdin != nil && stdout != nil {
			// Use custom stdin/stdout (for testing)
			transport = &mcp.IOTransport{
				Reader: stdin,
				Writer: stdout,
			}
		} else {
			// Use os.Stdin/os.Stdout (production)
			transport = &mcp.StdioTransport{}
		}
		return errors.Wrap(s.mcp.Run(ctx, transport), "running mcp over stdio")
	}

	// Create a new ServeMux for routing.
	mux := http.NewServeMux()
	s.registerRouteHealth(mux)
	s.registerRouteDefault(mux)

	addr := net.JoinHostPort(s.cfg.Server.Address, strconv.Itoa(s.cfg.Server.Port))
	httpServer := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: readHeaderTimeout,
	}

	protocol := "HTTP"
	if s.cfg.Server.TLSEnabled {
		protocol = "HTTPS"
	}

	slog.Info("Starting MCP server", "protocol", protocol, "address", s.cfg.Server.Address, "port", s.cfg.Server.Port)

	// Start server in a goroutine.
	errChan := make(chan error, 1)

	go func() {
		var err error

		slog.Info("Starting HTTP/HTTPS server")

		if s.cfg.Server.TLSEnabled {
			err = httpServer.ListenAndServeTLS(s.cfg.Server.TLSCertPath, s.cfg.Server.TLSKeyPath)
		} else {
			err = httpServer.ListenAndServe()
		}

		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- errors.Wrap(err, "HTTP/HTTPS server error")
		}
	}()

	// Wait for context cancellation or server error.
	select {
	case <-ctx.Done():
		slog.Info("Shutting down HTTP server")
		// Create a context with timeout for graceful shutdown.
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), ShutdownTimeout)
		defer shutdownCancel()
		//nolint:contextcheck
		return errors.Wrap(httpServer.Shutdown(shutdownCtx), "server shutting down failed")
	case err := <-errChan:
		return err
	}
}

func (s *Server) registerRouteHealth(mux *http.ServeMux) {
	mux.HandleFunc("/health", func(responseWriter http.ResponseWriter, _ *http.Request) {
		responseWriter.Header().Set("Content-Type", "application/json")
		responseWriter.WriteHeader(http.StatusOK)

		_, err := responseWriter.Write([]byte(`{"status":"ok"}`))
		if err != nil {
			slog.Error("Failed to write health response", "error", err)
		}
	})
}

func (s *Server) registerRouteDefault(mux *http.ServeMux) {
	mcpHandler := mcp.NewStreamableHTTPHandler(
		func(*http.Request) *mcp.Server {
			return s.mcp
		},
		nil,
	)

	mux.Handle("/", mcpHandler)
}

// registerTools registers all tools from the registry with the MCP server.
func (s *Server) registerTools() {
	slog.Info("Registering MCP tools")

	for _, toolset := range s.registry.GetToolsets() {
		if !toolset.IsEnabled() {
			slog.Info("Skipping disabled toolset", "toolset", toolset.GetName())

			continue
		}

		for _, tool := range toolset.GetTools() {
			if s.cfg.Global.ReadOnlyTools && !tool.IsReadOnly() {
				slog.Info("Skipping read-write tool (read-only mode enabled)", "tool", tool.GetName())

				continue
			}

			slog.Info("Registering tool", "toolset", toolset.GetName(), "tool", tool.GetName(), "read_only", tool.IsReadOnly())

			tool.RegisterWith(s.mcp)
		}
	}

	slog.Info("Tools registration complete")
}
