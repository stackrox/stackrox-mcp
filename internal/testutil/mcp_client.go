package testutil

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// MCPTestClient wraps the official MCP SDK client for testing purposes.
type MCPTestClient struct {
	session *mcp.ClientSession
	cancel  context.CancelFunc
	errCh   chan error
	t       *testing.T
}

// ServerRunFunc is a function that runs the MCP server with the given context and I/O streams.
// This allows tests to inject the server run function without creating circular dependencies.
type ServerRunFunc func(ctx context.Context, stdin io.ReadCloser, stdout io.WriteCloser) error

// NewMCPTestClient creates a new MCP test client that starts the MCP server in-process with stdio transport.
// The runFunc parameter should be a function that starts the MCP server (typically app.Run wrapped with config).
func NewMCPTestClient(t *testing.T, runFunc ServerRunFunc) (*MCPTestClient, error) {
	t.Helper()

	// Create pipes for bidirectional communication
	// Server reads from serverStdin, client writes to clientStdout (same pipe)
	serverStdin, clientStdout := io.Pipe()
	// Server writes to serverStdout, client reads from clientStdin (same pipe)
	clientStdin, serverStdout := io.Pipe()

	// Start server in goroutine
	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)

	go func() {
		err := runFunc(ctx, serverStdin, serverStdout)
		if err != nil && !errors.Is(err, context.Canceled) {
			t.Logf("MCP server error: %v", err)

			errCh <- err
		}
	}()

	// Create MCP client using official SDK
	client := mcp.NewClient(
		&mcp.Implementation{
			Name:    "mcp-test-client",
			Version: "1.0.0",
		},
		nil, // No custom options needed for basic testing
	)

	// Create IO transport for client
	transport := &mcp.IOTransport{
		Reader: clientStdin,  // Client reads from this pipe (server writes)
		Writer: clientStdout, // Client writes to this pipe (server reads)
	}

	// Connect and initialize
	session, err := client.Connect(ctx, transport, &mcp.ClientSessionOptions{})
	if err != nil {
		cancel()

		_ = clientStdout.Close()
		_ = clientStdin.Close()

		return nil, errors.New("failed to connect to MCP server: " + err.Error())
	}

	return &MCPTestClient{
		session: session,
		cancel:  cancel,
		errCh:   errCh,
		t:       t,
	}, nil
}

// Close stops the MCP server and cleans up resources.
func (c *MCPTestClient) Close() error {
	if err := c.session.Close(); err != nil {
		c.t.Logf("Error closing session: %v", err)
	}

	c.cancel()

	// Wait for server to finish (with timeout)
	select {
	case err := <-c.errCh:
		if err != nil && !errors.Is(err, context.Canceled) {
			return err
		}
	default:
		// Server is still running or finished cleanly
	}

	return nil
}

// ListTools returns all available tools from the server.
func (c *MCPTestClient) ListTools(ctx context.Context) (*mcp.ListToolsResult, error) {
	result, err := c.session.ListTools(ctx, nil)
	if err != nil {
		return nil, errors.New("failed to list tools: " + err.Error())
	}

	return result, nil
}

// CallTool invokes a tool with the given name and arguments.
func (c *MCPTestClient) CallTool(
	ctx context.Context,
	toolName string,
	args map[string]any,
) (*mcp.CallToolResult, error) {
	result, err := c.session.CallTool(ctx, &mcp.CallToolParams{
		Name:      toolName,
		Arguments: args,
	})
	if err != nil {
		return nil, errors.New("failed to call tool: " + err.Error())
	}

	return result, nil
}

// RequireNoError asserts that the tool call result does not contain an error.
func RequireNoError(t *testing.T, result *mcp.CallToolResult) {
	t.Helper()

	if !result.IsError {
		return
	}

	// Extract error message from content
	errMsg := "unknown error"

	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(*mcp.TextContent); ok {
			errMsg = textContent.Text
		}
	}

	t.Fatalf("expected no error, got: %s", errMsg)
}
