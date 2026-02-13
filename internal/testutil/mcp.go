package testutil

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

// MCPClient is a client for sending MCP JSON-RPC requests over stdio.
type MCPClient struct {
	stdin  io.WriteCloser // Client writes to this (server reads from it)
	stdout io.ReadCloser  // Client reads from this (server writes to it)
	reader *bufio.Reader
	cancel context.CancelFunc
	errCh  chan error
	mu     sync.Mutex
	nextID int
	t      *testing.T
}

// MCPRequest represents a JSON-RPC 2.0 request.
type MCPRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// MCPResponse represents a JSON-RPC 2.0 response.
type MCPResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *MCPError       `json:"error,omitempty"`
}

// MCPError represents a JSON-RPC 2.0 error object.
type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// ServerRunFunc is a function that runs the MCP server with the given context, config, and I/O streams.
// This allows tests to inject the main.Run function without creating circular dependencies.
type ServerRunFunc func(ctx context.Context, stdin io.ReadCloser, stdout io.WriteCloser) error

// NewMCPClient creates a new MCP client that starts the MCP server in-process with stdio transport.
// The runFunc parameter should be a function that starts the MCP server (typically main.Run wrapped with config).
// This is more efficient than subprocess execution and allows for better code coverage.
func NewMCPClient(t *testing.T, runFunc ServerRunFunc) (*MCPClient, error) {
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
		if err != nil {
			t.Logf("MCP server error: %v", err)
			errCh <- err
		}
	}()

	return &MCPClient{
		stdin:  clientStdout, // Client writes to this pipe (server reads)
		stdout: clientStdin,  // Client reads from this pipe (server writes)
		reader: bufio.NewReader(clientStdin),
		cancel: cancel,
		errCh:  errCh,
		nextID: 1,
		t:      t,
	}, nil
}

// Close stops the MCP server and cleans up resources.
func (c *MCPClient) Close() error {
	c.cancel()
	c.stdin.Close()
	c.stdout.Close()

	// Wait for server to finish (with timeout)
	select {
	case err := <-c.errCh:
		if err != nil && err != context.Canceled {
			return err
		}
	default:
		// Server is still running or finished cleanly
	}

	return nil
}

// sendRequest sends a JSON-RPC request and returns the response.
func (c *MCPClient) sendRequest(method string, params any) (*MCPResponse, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var paramsJSON json.RawMessage
	if params != nil {
		var err error
		paramsJSON, err = json.Marshal(params)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal params: %w", err)
		}
	}

	req := MCPRequest{
		JSONRPC: "2.0",
		ID:      c.nextID,
		Method:  method,
		Params:  paramsJSON,
	}
	c.nextID++

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Write request to stdin
	if _, err := c.stdin.Write(append(reqBody, '\n')); err != nil {
		return nil, fmt.Errorf("failed to write request: %w", err)
	}

	c.t.Logf("MCP request: %s", string(reqBody))

	// Read response from stdout
	respLine, err := c.reader.ReadBytes('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	c.t.Logf("MCP response: %s", string(respLine))

	var resp MCPResponse
	if err := json.Unmarshal(respLine, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &resp, nil
}

// Initialize sends an initialize request to the MCP server.
func (c *MCPClient) Initialize() (*MCPResponse, error) {
	params := map[string]any{
		"protocolVersion": "2024-11-05",
		"capabilities":    map[string]any{},
		"clientInfo": map[string]any{
			"name":    "mcp-test-client",
			"version": "1.0.0",
		},
	}
	return c.sendRequest("initialize", params)
}

// ListTools sends a tools/list request to the MCP server.
func (c *MCPClient) ListTools() (*MCPResponse, error) {
	return c.sendRequest("tools/list", nil)
}

// CallTool sends a tools/call request to the MCP server.
func (c *MCPClient) CallTool(toolName string, args map[string]any) (*MCPResponse, error) {
	params := map[string]any{
		"name":      toolName,
		"arguments": args,
	}
	return c.sendRequest("tools/call", params)
}

// RequireNoError asserts that the MCP response does not contain an error.
func RequireNoError(t *testing.T, resp *MCPResponse) {
	t.Helper()
	if resp.Error != nil {
		t.Fatalf("expected no error, got: code=%d, message=%s, data=%v",
			resp.Error.Code, resp.Error.Message, resp.Error.Data)
	}
}

// UnmarshalResult unmarshals the result field of an MCP response into the target.
func UnmarshalResult(t *testing.T, resp *MCPResponse, target any) {
	t.Helper()
	require.NotNil(t, resp.Result, "result should not be nil")
	err := json.Unmarshal(resp.Result, target)
	require.NoError(t, err, "failed to unmarshal result")
}
