package testutil

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

// MCPClient is a client for sending MCP JSON-RPC requests over stdio.
type MCPClient struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
	stderr io.ReadCloser
	reader *bufio.Reader
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

// NewMCPClient creates a new MCP client that starts the MCP server as a subprocess with stdio transport.
func NewMCPClient(t *testing.T, binaryPath string, configPath string) (*MCPClient, error) {
	t.Helper()

	cmd := exec.Command(binaryPath, "--config", configPath)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start logging stderr in background
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			t.Logf("MCP stderr: %s", scanner.Text())
		}
	}()

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start MCP server: %w", err)
	}

	return &MCPClient{
		cmd:    cmd,
		stdin:  stdin,
		stdout: stdout,
		stderr: stderr,
		reader: bufio.NewReader(stdout),
		nextID: 1,
		t:      t,
	}, nil
}

// Close stops the MCP server and cleans up resources.
func (c *MCPClient) Close() error {
	c.stdin.Close()
	c.stdout.Close()
	c.stderr.Close()
	return c.cmd.Wait()
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
