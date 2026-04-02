package config

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stackrox/stackrox-mcp/internal/prompts"
)

type listClusterPrompt struct{}

// NewListClusterPrompt creates a new list-cluster prompt.
func NewListClusterPrompt() prompts.Prompt {
	return &listClusterPrompt{}
}

func (p *listClusterPrompt) GetName() string {
	return "list-cluster"
}

func (p *listClusterPrompt) GetPrompt() *mcp.Prompt {
	return &mcp.Prompt{
		Name:        p.GetName(),
		Description: "List all Kubernetes/OpenShift clusters secured by StackRox Central.",
		Arguments:   nil,
	}
}

func (p *listClusterPrompt) GetMessages(_ map[string]any) ([]*mcp.PromptMessage, error) {
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
	prompts.RegisterWithStandardHandler(server, p)
}
