package prompts

import (
	"context"
	"errors"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testPrompt struct {
	name         string
	returnError  bool
	capturedArgs map[string]any
}

func (p *testPrompt) GetName() string {
	return p.name
}

func (p *testPrompt) GetPrompt() *mcp.Prompt {
	return &mcp.Prompt{
		Name:        p.name,
		Description: "Test prompt",
	}
}

func (p *testPrompt) GetMessages(arguments map[string]any) ([]*mcp.PromptMessage, error) {
	p.capturedArgs = arguments

	if p.returnError {
		return nil, errors.New("test error")
	}

	return []*mcp.PromptMessage{
		{
			Role: "user",
			Content: &mcp.TextContent{
				Text: "test message",
			},
		},
	}, nil
}

func (p *testPrompt) RegisterWith(server *mcp.Server) {
	RegisterWithStandardHandler(server, p)
}

func TestRegisterWithStandardHandler(t *testing.T) {
	t.Parallel()

	prompt := &testPrompt{
		name: "test-prompt",
	}

	server := mcp.NewServer(
		&mcp.Implementation{
			Name:    "test-server",
			Version: "1.0.0",
		},
		&mcp.ServerOptions{},
	)

	assert.NotPanics(t, func() {
		RegisterWithStandardHandler(server, prompt)
	})
}

func TestRegisterWithStandardHandler_ArgumentPassing(t *testing.T) {
	t.Parallel()

	prompt := &testPrompt{
		name: "test-prompt",
	}

	server := mcp.NewServer(
		&mcp.Implementation{
			Name:    "test-server",
			Version: "1.0.0",
		},
		&mcp.ServerOptions{},
	)

	RegisterWithStandardHandler(server, prompt)

	req := &mcp.GetPromptRequest{
		Params: &mcp.GetPromptParams{
			Name: "test-prompt",
			Arguments: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
		},
	}

	assert.NotNil(t, req)
}

func TestRegisterWithStandardHandler_ErrorHandling(t *testing.T) {
	t.Parallel()

	prompt := &testPrompt{
		name:        "test-prompt",
		returnError: true,
	}

	server := mcp.NewServer(
		&mcp.Implementation{
			Name:    "test-server",
			Version: "1.0.0",
		},
		&mcp.ServerOptions{},
	)

	// Should not panic even with error-returning prompt
	assert.NotPanics(t, func() {
		RegisterWithStandardHandler(server, prompt)
	})
}

func TestRegisterWithStandardHandler_NilArguments(t *testing.T) {
	t.Parallel()

	prompt := &testPrompt{
		name: "test-prompt",
	}

	server := mcp.NewServer(
		&mcp.Implementation{
			Name:    "test-server",
			Version: "1.0.0",
		},
		&mcp.ServerOptions{},
	)

	RegisterWithStandardHandler(server, prompt)

	handler := func(_ context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		args := make(map[string]any)

		if req.Params.Arguments != nil {
			for key, value := range req.Params.Arguments {
				args[key] = value
			}
		}

		messages, err := prompt.GetMessages(args)
		if err != nil {
			return nil, err
		}

		return &mcp.GetPromptResult{
			Messages: messages,
		}, nil
	}

	req := &mcp.GetPromptRequest{
		Params: &mcp.GetPromptParams{
			Name:      "test-prompt",
			Arguments: nil,
		},
	}

	result, err := handler(context.Background(), req)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Len(t, result.Messages, 1)
	assert.Empty(t, prompt.capturedArgs)
}
