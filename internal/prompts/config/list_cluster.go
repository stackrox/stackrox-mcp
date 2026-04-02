package config

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stackrox/stackrox-mcp/internal/prompts"
)

type listClusterPrompt struct {
	name string
}

// NewListClusterPrompt creates a new list-cluster prompt.
func NewListClusterPrompt() prompts.Prompt {
	return &listClusterPrompt{
		name: "list-cluster",
	}
}

func (p *listClusterPrompt) GetName() string {
	return p.name
}

func (p *listClusterPrompt) GetPrompt() *mcp.Prompt {
	return &mcp.Prompt{
		Name:        p.name,
		Description: "List all Kubernetes/OpenShift clusters secured by StackRox Central.",
		Arguments:   nil,
	}
}

func (p *listClusterPrompt) GetMessages(_ map[string]interface{}) ([]*mcp.PromptMessage, error) {
	content := `You are helping list all Kubernetes/OpenShift clusters secured by StackRox Central.

Use the list_clusters tool to retrieve all managed clusters.

The tool will return:
- Cluster ID
- Cluster name
- Cluster type (e.g., KUBERNETES_CLUSTER, OPENSHIFT_CLUSTER)

Present the clusters in a clear, readable format.`

	return []*mcp.PromptMessage{
		{
			Role: "user",
			Content: &mcp.TextContent{
				Text: content,
			},
		},
	}, nil
}

func (p *listClusterPrompt) RegisterWith(server *mcp.Server) {
	server.AddPrompt(p.GetPrompt(), p.handle)
}

func (p *listClusterPrompt) handle(
	_ context.Context,
	_ *mcp.GetPromptRequest,
) (*mcp.GetPromptResult, error) {
	messages, err := p.GetMessages(nil)
	if err != nil {
		return nil, err
	}

	return &mcp.GetPromptResult{
		Messages: messages,
	}, nil
}
