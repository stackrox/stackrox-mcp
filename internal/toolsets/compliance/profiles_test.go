package compliance

import (
	"context"
	"net"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	v2 "github.com/stackrox/rox/generated/api/v2"
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

func bufDialer(listener *bufconn.Listener) func(context.Context, string) (net.Conn, error) {
	return func(_ context.Context, _ string) (net.Conn, error) {
		return listener.Dial()
	}
}

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

	stackroxClient.SetConnForTesting(t, conn)

	return stackroxClient
}

func TestNewListComplianceProfilesTool(t *testing.T) {
	tool := NewListComplianceProfilesTool(&client.Client{})

	require.NotNil(t, tool)
	assert.Equal(t, "list_compliance_profiles", tool.GetName())
}

func TestListComplianceProfilesTool_IsReadOnly(t *testing.T) {
	tool := NewListComplianceProfilesTool(&client.Client{})

	assert.True(t, tool.IsReadOnly(), "list_compliance_profiles should be read-only")
}

func TestListComplianceProfilesTool_GetTool(t *testing.T) {
	tool := NewListComplianceProfilesTool(&client.Client{})

	mcpTool := tool.GetTool()

	require.NotNil(t, mcpTool)
	assert.Equal(t, "list_compliance_profiles", mcpTool.Name)
	assert.NotEmpty(t, mcpTool.Description)
	require.NotNil(t, mcpTool.InputSchema, "InputSchema should be defined")
}

func TestListComplianceProfilesTool_RegisterWith(t *testing.T) {
	tool := NewListComplianceProfilesTool(&client.Client{})
	server := mcp.NewServer(
		&mcp.Implementation{
			Name:    "test-server",
			Version: "1.0.0",
		},
		&mcp.ServerOptions{},
	)

	assert.NotPanics(t, func() {
		tool.RegisterWith(server)
	})
}

func TestListComplianceProfiles_Success(t *testing.T) {
	mockService := mock.NewComplianceProfileServiceMock(
		[]*v2.ComplianceProfile{
			{
				Id:             "profile-1",
				Name:           "ocp4-cis",
				ProfileVersion: "1.4.0",
				Description:    "CIS Benchmark for OpenShift",
				Title:          "CIS OpenShift Benchmark",
				Rules:          []*v2.ComplianceRule{{Name: "rule-1"}, {Name: "rule-2"}},
			},
			{
				Id:             "profile-2",
				Name:           "ocp4-nist",
				ProfileVersion: "1.0.0",
				Description:    "NIST 800-53 for OpenShift",
				Title:          "NIST OpenShift Profile",
				Rules:          []*v2.ComplianceRule{{Name: "rule-3"}},
			},
		},
		nil,
	)

	grpcServer, listener := mock.SetupComplianceServer(mockService, nil, nil)
	defer grpcServer.Stop()

	testClient := createTestClient(t, listener)
	tool, ok := NewListComplianceProfilesTool(testClient).(*listComplianceProfilesTool)
	require.True(t, ok)

	ctx := context.Background()
	req := &mcp.CallToolRequest{}
	input := listComplianceProfilesInput{ClusterID: "cluster-1"}

	result, output, err := tool.handle(ctx, req, input)

	require.NoError(t, err)
	require.NotNil(t, output)
	assert.Nil(t, result)

	require.Len(t, output.Profiles, 2)
	assert.Equal(t, "ocp4-cis", output.Profiles[0].Name)
	assert.Equal(t, "1.4.0", output.Profiles[0].ProfileVersion)
	assert.Equal(t, 2, output.Profiles[0].RuleCount)
	assert.Equal(t, "ocp4-nist", output.Profiles[1].Name)
	assert.Equal(t, 1, output.Profiles[1].RuleCount)
}

func TestListComplianceProfiles_EmptyClusterID(t *testing.T) {
	tool, ok := NewListComplianceProfilesTool(&client.Client{}).(*listComplianceProfilesTool)
	require.True(t, ok)

	ctx := context.Background()
	req := &mcp.CallToolRequest{}
	input := listComplianceProfilesInput{ClusterID: ""}

	_, _, err := tool.handle(ctx, req, input)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "clusterId is required")
}

func TestListComplianceProfiles_EmptyResults(t *testing.T) {
	mockService := mock.NewComplianceProfileServiceMock([]*v2.ComplianceProfile{}, nil)

	grpcServer, listener := mock.SetupComplianceServer(mockService, nil, nil)
	defer grpcServer.Stop()

	testClient := createTestClient(t, listener)
	tool, ok := NewListComplianceProfilesTool(testClient).(*listComplianceProfilesTool)
	require.True(t, ok)

	ctx := context.Background()
	req := &mcp.CallToolRequest{}
	input := listComplianceProfilesInput{ClusterID: "cluster-1"}

	result, output, err := tool.handle(ctx, req, input)

	require.NoError(t, err)
	require.NotNil(t, output)
	assert.Nil(t, result)
	assert.Empty(t, output.Profiles)
}

func TestListComplianceProfiles_GRPCError(t *testing.T) {
	mockService := mock.NewComplianceProfileServiceMock(
		[]*v2.ComplianceProfile{},
		status.Error(codes.Internal, "test"),
	)

	grpcServer, listener := mock.SetupComplianceServer(mockService, nil, nil)
	defer grpcServer.Stop()

	testClient := createTestClient(t, listener)
	tool, ok := NewListComplianceProfilesTool(testClient).(*listComplianceProfilesTool)
	require.True(t, ok)

	ctx := context.Background()
	req := &mcp.CallToolRequest{}
	input := listComplianceProfilesInput{ClusterID: "cluster-1"}

	result, output, err := tool.handle(ctx, req, input)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Nil(t, output)
	assert.Contains(t, err.Error(), "Internal server error")
}
