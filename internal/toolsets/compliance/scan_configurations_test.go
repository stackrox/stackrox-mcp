package compliance

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	v2 "github.com/stackrox/rox/generated/api/v2"
	"github.com/stackrox/stackrox-mcp/internal/client"
	"github.com/stackrox/stackrox-mcp/internal/toolsets/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestNewListComplianceScanConfigurationsTool(t *testing.T) {
	tool := NewListComplianceScanConfigurationsTool(&client.Client{})

	require.NotNil(t, tool)
	assert.Equal(t, "list_compliance_scan_configurations", tool.GetName())
}

func TestListComplianceScanConfigurationsTool_IsReadOnly(t *testing.T) {
	tool := NewListComplianceScanConfigurationsTool(&client.Client{})

	assert.True(t, tool.IsReadOnly(), "list_compliance_scan_configurations should be read-only")
}

func TestListComplianceScanConfigurationsTool_GetTool(t *testing.T) {
	tool := NewListComplianceScanConfigurationsTool(&client.Client{})

	mcpTool := tool.GetTool()

	require.NotNil(t, mcpTool)
	assert.Equal(t, "list_compliance_scan_configurations", mcpTool.Name)
	assert.NotEmpty(t, mcpTool.Description)
	require.NotNil(t, mcpTool.InputSchema, "InputSchema should be defined")
}

func TestListComplianceScanConfigurationsTool_RegisterWith(t *testing.T) {
	tool := NewListComplianceScanConfigurationsTool(&client.Client{})
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

func TestListComplianceScanConfigurations_Success(t *testing.T) {
	mockService := mock.NewComplianceScanConfigurationServiceMock(
		[]*v2.ComplianceScanConfigurationStatus{
			{
				Id:       "config-1",
				ScanName: "daily-cis-scan",
				ScanConfig: &v2.BaseComplianceScanConfigurationSettings{
					Description: "Daily CIS benchmark scan",
					Profiles:    []string{"ocp4-cis"},
				},
				ClusterStatus: []*v2.ClusterScanStatus{
					{ClusterId: "cluster-1"},
					{ClusterId: "cluster-2"},
				},
			},
		},
		nil,
	)

	grpcServer, listener := mock.SetupComplianceServer(nil, nil, mockService)
	defer grpcServer.Stop()

	testClient := createTestClient(t, listener)
	tool, ok := NewListComplianceScanConfigurationsTool(testClient).(*listComplianceScanConfigurationsTool)
	require.True(t, ok)

	ctx := context.Background()
	req := &mcp.CallToolRequest{}
	input := listComplianceScanConfigurationsInput{}

	result, output, err := tool.handle(ctx, req, input)

	require.NoError(t, err)
	require.NotNil(t, output)
	assert.Nil(t, result)

	require.Len(t, output.Configurations, 1)
	assert.Equal(t, "config-1", output.Configurations[0].ID)
	assert.Equal(t, "daily-cis-scan", output.Configurations[0].Name)
	assert.Equal(t, "Daily CIS benchmark scan", output.Configurations[0].Description)
	require.Len(t, output.Configurations[0].Profiles, 1)
	assert.Equal(t, "ocp4-cis", output.Configurations[0].Profiles[0])
	require.Len(t, output.Configurations[0].Clusters, 2)
}

func TestListComplianceScanConfigurations_EmptyResults(t *testing.T) {
	mockService := mock.NewComplianceScanConfigurationServiceMock(
		[]*v2.ComplianceScanConfigurationStatus{},
		nil,
	)

	grpcServer, listener := mock.SetupComplianceServer(nil, nil, mockService)
	defer grpcServer.Stop()

	testClient := createTestClient(t, listener)
	tool, ok := NewListComplianceScanConfigurationsTool(testClient).(*listComplianceScanConfigurationsTool)
	require.True(t, ok)

	ctx := context.Background()
	req := &mcp.CallToolRequest{}
	input := listComplianceScanConfigurationsInput{}

	result, output, err := tool.handle(ctx, req, input)

	require.NoError(t, err)
	require.NotNil(t, output)
	assert.Nil(t, result)
	assert.Empty(t, output.Configurations)
}

func TestListComplianceScanConfigurations_GRPCError(t *testing.T) {
	mockService := mock.NewComplianceScanConfigurationServiceMock(
		[]*v2.ComplianceScanConfigurationStatus{},
		status.Error(codes.Internal, "test"),
	)

	grpcServer, listener := mock.SetupComplianceServer(nil, nil, mockService)
	defer grpcServer.Stop()

	testClient := createTestClient(t, listener)
	tool, ok := NewListComplianceScanConfigurationsTool(testClient).(*listComplianceScanConfigurationsTool)
	require.True(t, ok)

	ctx := context.Background()
	req := &mcp.CallToolRequest{}
	input := listComplianceScanConfigurationsInput{}

	result, output, err := tool.handle(ctx, req, input)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Nil(t, output)
	assert.Contains(t, err.Error(), "Internal server error")
}
