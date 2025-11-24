package config

import (
	"context"
	"errors"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stackrox/stackrox-mcp/internal/toolsets"
)

// listClustersInput defines the input parameters for list_clusters tool.
type listClustersInput struct{}

// listClustersOutput defines the output structure for list_clusters tool.
type listClustersOutput struct {
	Clusters []string `json:"clusters"`
}

// listClustersTool implements the list_clusters tool.
type listClustersTool struct {
	name string
}

// NewListClustersTool creates a new list_clusters tool.
func NewListClustersTool() toolsets.Tool {
	return &listClustersTool{
		name: "list_clusters",
	}
}

// IsReadOnly returns true as this tool only reads data.
func (t *listClustersTool) IsReadOnly() bool {
	return true
}

// GetName returns the tool name.
func (t *listClustersTool) GetName() string {
	return t.name
}

// GetTool returns the MCP Tool definition.
func (t *listClustersTool) GetTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        t.name,
		Description: "List all clusters managed by StackRox Central with their IDs, names, and types",
	}
}

// RegisterWith registers the list_clusters tool handler with the MCP server.
func (t *listClustersTool) RegisterWith(server *mcp.Server) {
	mcp.AddTool(server, t.GetTool(), t.handle)
}

// handle is the placeholder handler for list_clusters tool.
func (t *listClustersTool) handle(
	_ context.Context,
	_ *mcp.CallToolRequest,
	_ listClustersInput,
) (*mcp.CallToolResult, *listClustersOutput, error) {
	return nil, nil, errors.New("list_clusters tool is not yet implemented")
}
