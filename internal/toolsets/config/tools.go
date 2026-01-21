package config

import (
	"context"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pkg/errors"
	v1 "github.com/stackrox/rox/generated/api/v1"
	"github.com/stackrox/stackrox-mcp/internal/client"
	"github.com/stackrox/stackrox-mcp/internal/client/auth"
	"github.com/stackrox/stackrox-mcp/internal/logging"
	"github.com/stackrox/stackrox-mcp/internal/toolsets"
)

const (
	defaultOffset = 0

	// 0 = no limit.
	defaultLimit = 0
)

// listClustersInput defines the input parameters for list_clusters tool.
type listClustersInput struct {
	Offset int `json:"offset,omitempty"`
	Limit  int `json:"limit,omitempty"`
}

// ClusterInfo represents information about a single cluster.
type ClusterInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// listClustersOutput defines the output structure for list_clusters tool.
type listClustersOutput struct {
	Clusters   []ClusterInfo `json:"clusters"`
	TotalCount int           `json:"totalCount"`
	Offset     int           `json:"offset"`
	Limit      int           `json:"limit"`
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
		Name: t.name,
		Description: "List all clusters managed by StackRox with their IDs, names, and types." +
			" Use this tool to get cluster information," +
			" or when you need to map a cluster name to its cluster ID for use in other tools.",
		InputSchema: listClustersInputSchema(),
	}
}

func listClustersInputSchema() *jsonschema.Schema {
	schema, err := jsonschema.For[listClustersInput](nil)
	if err != nil {
		logging.Fatal("Could not get jsonschema for list_clusters input", err)

		return nil
	}

	schema.Properties["offset"].Minimum = jsonschema.Ptr(0.0)
	schema.Properties["offset"].Default = toolsets.MustJSONMarshal(defaultOffset)
	schema.Properties["offset"].Description = "Starting index for pagination (0-based)." +
		" When using pagination, always provide both offset and limit together. Default: 0."

	schema.Properties["limit"].Minimum = jsonschema.Ptr(0.0)
	schema.Properties["limit"].Default = toolsets.MustJSONMarshal(defaultLimit)
	schema.Properties["limit"].Description = "Maximum number of clusters to return." +
		" When using pagination, always provide both limit and offset together. Use 0 for unlimited (default)."

	return schema
}

// RegisterWith registers the list_clusters tool handler with the MCP server.
func (t *listClustersTool) RegisterWith(server *mcp.Server) {
	mcp.AddTool(server, t.GetTool(), t.handle)
}

func (t *listClustersTool) getClusters(ctx context.Context, req *mcp.CallToolRequest) ([]ClusterInfo, error) {
	conn, err := t.client.ReadyConn(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to connect to server")
	}

	callCtx := auth.WithMCPRequestContext(ctx, req)

	// Create ClustersService client
	clustersClient := v1.NewClustersServiceClient(conn)

	// Call GetClusters to fetch all clusters
	resp, err := clustersClient.GetClusters(callCtx, &v1.GetClustersRequest{})
	if err != nil {
		// Convert gRPC error to client error
		clientErr := client.NewError(err, "GetClusters")

		return nil, clientErr
	}

	// Convert all clusters to ClusterInfo objects
	allClusters := make([]ClusterInfo, 0, len(resp.GetClusters()))
	for _, cluster := range resp.GetClusters() {
		clusterInfo := ClusterInfo{
			ID:   cluster.GetId(),
			Name: cluster.GetName(),
			Type: cluster.GetType().String(),
		}
		allClusters = append(allClusters, clusterInfo)
	}

	return allClusters, nil
}

// handle is the handler for list_clusters tool.
func (t *listClustersTool) handle(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input listClustersInput,
) (*mcp.CallToolResult, *listClustersOutput, error) {
	clusters, err := t.getClusters(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	totalCount := len(clusters)

	// 0 = unlimited.
	limit := input.Limit
	if limit == 0 {
		limit = totalCount
	}

	// Apply client-side pagination.
	var paginatedClusters []ClusterInfo
	if input.Offset >= totalCount {
		paginatedClusters = []ClusterInfo{}
	} else {
		end := min(input.Offset+limit, totalCount)
		if end < 0 {
			end = totalCount
		}

		paginatedClusters = clusters[input.Offset:end]
	}

	output := &listClustersOutput{
		Clusters:   paginatedClusters,
		TotalCount: totalCount,
		Offset:     input.Offset,
		Limit:      input.Limit,
	}

	return nil, output, nil
}
