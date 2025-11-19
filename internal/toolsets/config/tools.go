package config

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pkg/errors"
	v1 "github.com/stackrox/rox/generated/api/v1"
	"github.com/stackrox/stackrox-mcp/internal/client"
	"github.com/stackrox/stackrox-mcp/internal/client/auth"
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
	name   string
	client *client.Client
}

// NewListClustersTool creates a new list_clusters tool.
func NewListClustersTool(c *client.Client) toolsets.Tool {
	return &listClustersTool{
		name:   "list_clusters",
		client: c,
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
	ctx context.Context,
	req *mcp.CallToolRequest,
	_ listClustersInput,
) (*mcp.CallToolResult, *listClustersOutput, error) {
	conn, err := t.client.ReadyConn(ctx)
	if err != nil {
		return nil, nil, errors.Wrap(err, "unable to connect to server")
	}

	callCtx := auth.WithMCPRequestContext(ctx, req)

	// Create ClustersService client
	clustersClient := v1.NewClustersServiceClient(conn)

	// Call GetClusters
	resp, err := clustersClient.GetClusters(callCtx, &v1.GetClustersRequest{})
	if err != nil {
		// Convert gRPC error to client error
		clientErr := client.NewError(err, "GetClusters")

		return nil, nil, clientErr
	}

	// Extract cluster information
	clusters := make([]string, 0, len(resp.GetClusters()))
	for _, cluster := range resp.GetClusters() {
		// Format: "ID: <id>, Name: <name>, Type: <type>"
		clusterInfo := fmt.Sprintf("ID: %s, Name: %s, Type: %s",
			cluster.GetId(),
			cluster.GetName(),
			cluster.GetType().String())
		clusters = append(clusters, clusterInfo)
	}

	output := &listClustersOutput{
		Clusters: clusters,
	}

	// Return result with text content
	result := &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Found %d cluster(s)", len(clusters)),
			},
		},
	}

	return result, output, nil
}
