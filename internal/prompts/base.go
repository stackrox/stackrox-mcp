package prompts

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pkg/errors"
)

// RegisterWithStandardHandler registers a prompt using the standard handler pattern.
// This eliminates boilerplate by providing a common implementation for most prompts.
func RegisterWithStandardHandler(server *mcp.Server, prompt Prompt) {
	handler := func(_ context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		args := make(map[string]any)

		if req.Params.Arguments != nil {
			for key, value := range req.Params.Arguments {
				args[key] = value
			}
		}

		messages, err := prompt.GetMessages(args)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get prompt messages")
		}

		return &mcp.GetPromptResult{
			Messages: messages,
		}, nil
	}

	server.AddPrompt(prompt.GetPrompt(), handler)
}
