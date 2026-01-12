package config

import (
	"context"
	"fmt"
	"net"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stackrox/rox/generated/storage"
	"github.com/stackrox/stackrox-mcp/internal/client"
	"github.com/stackrox/stackrox-mcp/internal/config"
	"github.com/stackrox/stackrox-mcp/internal/toolsets/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

func TestNewListClustersTool(t *testing.T) {
	tool := NewListClustersTool(&client.Client{})

	require.NotNil(t, tool)
	assert.Equal(t, "list_clusters", tool.GetName())
}

func TestListClustersTool_IsReadOnly(t *testing.T) {
	tool := NewListClustersTool(&client.Client{})

	assert.True(t, tool.IsReadOnly(), "list_clusters should be read-only")
}

func TestListClustersTool_GetTool(t *testing.T) {
	tool := NewListClustersTool(&client.Client{})

	mcpTool := tool.GetTool()

	require.NotNil(t, mcpTool)
	assert.Equal(t, "list_clusters", mcpTool.Name)
	assert.NotEmpty(t, mcpTool.Description)
	require.NotNil(t, mcpTool.InputSchema, "InputSchema should be defined")
}

func TestListClustersTool_RegisterWith(t *testing.T) {
	tool := NewListClustersTool(&client.Client{})
	server := mcp.NewServer(
		&mcp.Implementation{
			Name:    "test-server",
			Version: "1.0.0",
		},
		&mcp.ServerOptions{},
	)

	// Should not panic
	assert.NotPanics(t, func() {
		tool.RegisterWith(server)
	})
}

// bufDialer creates a dialer function for bufconn.
func bufDialer(listener *bufconn.Listener) func(context.Context, string) (net.Conn, error) {
	return func(_ context.Context, _ string) (net.Conn, error) {
		return listener.Dial()
	}
}

// createTestClient creates a client connected to the mock server.
func createTestClient(t *testing.T, listener *bufconn.Listener) *client.Client {
	t.Helper()

	conn, err := grpc.NewClient(
		"passthrough://buffer",
		grpc.WithLocalDNSResolution(),
		grpc.WithContextDialer(bufDialer(listener)),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)

	stackroxClient, err := client.NewClient(&config.CentralConfig{
		URL: "buffer",
	})
	require.NoError(t, err)

	// Inject mock connection for testing.
	stackroxClient.SetConnForTesting(t, conn)

	return stackroxClient
}

func TestHandle_DefaultLimit(t *testing.T) {
	mockService := mock.NewClustersServiceMock(
		[]*storage.Cluster{
			{Id: "c1", Name: "Cluster 1", Type: storage.ClusterType_KUBERNETES_CLUSTER},
			{Id: "c2", Name: "Cluster 2", Type: storage.ClusterType_KUBERNETES_CLUSTER},
			{Id: "c3", Name: "Cluster 3", Type: storage.ClusterType_KUBERNETES_CLUSTER},
			{Id: "c4", Name: "Cluster 4", Type: storage.ClusterType_KUBERNETES_CLUSTER},
			{Id: "c5", Name: "Cluster 5", Type: storage.ClusterType_KUBERNETES_CLUSTER},
		},
		nil,
	)

	grpcServer, listener := mock.SetupClusterServer(mockService)
	defer grpcServer.Stop()

	testClient := createTestClient(t, listener)
	tool, ok := NewListClustersTool(testClient).(*listClustersTool)
	require.True(t, ok)

	ctx := context.Background()
	req := &mcp.CallToolRequest{}
	input := listClustersInput{
		Offset: defaultOffset,
		Limit:  defaultOffset,
	}

	result, output, err := tool.handle(ctx, req, input)

	require.NoError(t, err)
	require.NotNil(t, output)
	assert.Nil(t, result) // MCP SDK handles result creation

	require.Len(t, output.Clusters, 5)
	assert.Equal(t, 5, output.TotalCount)
	assert.Equal(t, 0, output.Offset)
	assert.Equal(t, 0, output.Limit)
	assert.Equal(t, "Cluster 1", output.Clusters[0].Name)
	assert.Equal(t, "Cluster 5", output.Clusters[4].Name)
}

//nolint:funlen
func TestHandle_WithPagination(t *testing.T) {
	totalClusters := 10

	clusters := make([]*storage.Cluster, totalClusters)
	for i := range totalClusters {
		clusters[i] = &storage.Cluster{
			Id:   fmt.Sprintf("cluster-%d", i),
			Name: fmt.Sprintf("Cluster %d", i),
			Type: storage.ClusterType_KUBERNETES_CLUSTER,
		}
	}

	mockService := mock.NewClustersServiceMock(clusters, nil)

	grpcServer, listener := mock.SetupClusterServer(mockService)
	defer grpcServer.Stop()

	testClient := createTestClient(t, listener)
	tool, ok := NewListClustersTool(testClient).(*listClustersTool)
	require.True(t, ok)

	tests := map[string]struct {
		offset        int
		limit         int
		expectedCount int
		expectedFirst string
		expectedLast  string
	}{
		"first page": {
			offset:        0,
			limit:         3,
			expectedCount: 3,
			expectedFirst: "cluster-0",
			expectedLast:  "cluster-2",
		},
		"middle page": {
			offset:        2,
			limit:         3,
			expectedCount: 3,
			expectedFirst: "cluster-2",
			expectedLast:  "cluster-4",
		},
		"partial page": {
			offset:        8,
			limit:         10,
			expectedCount: 2,
			expectedFirst: "cluster-8",
			expectedLast:  "cluster-9",
		},
		"offset beyond total": {
			offset:        100,
			limit:         10,
			expectedCount: 0,
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			ctx := context.Background()
			req := &mcp.CallToolRequest{}
			input := listClustersInput{
				Offset: testCase.offset,
				Limit:  testCase.limit,
			}

			result, output, err := tool.handle(ctx, req, input)

			require.NoError(t, err)
			require.NotNil(t, output)
			assert.Nil(t, result) // MCP SDK handles result creation.

			assert.Len(t, output.Clusters, testCase.expectedCount)
			assert.Equal(t, totalClusters, output.TotalCount)
			assert.Equal(t, testCase.offset, output.Offset)
			assert.Equal(t, testCase.limit, output.Limit)

			if testCase.expectedCount > 0 {
				assert.Equal(t, testCase.expectedFirst, output.Clusters[0].ID)
				assert.Equal(t, testCase.expectedLast, output.Clusters[testCase.expectedCount-1].ID)
			}
		})
	}
}

func TestHandle_GetClustersError(t *testing.T) {
	mockService := mock.NewClustersServiceMock(
		[]*storage.Cluster{},
		status.Error(codes.Internal, "test"),
	)

	grpcServer, listener := mock.SetupClusterServer(mockService)
	defer grpcServer.Stop()

	testClient := createTestClient(t, listener)
	tool, ok := NewListClustersTool(testClient).(*listClustersTool)
	require.True(t, ok)

	ctx := context.Background()
	req := &mcp.CallToolRequest{}
	input := listClustersInput{
		Offset: 0,
		Limit:  10,
	}

	result, output, err := tool.handle(ctx, req, input)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Nil(t, output)
	assert.Contains(t, err.Error(), "Internal server error")
}
