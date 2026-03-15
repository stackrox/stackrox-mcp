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

func TestNewGetComplianceScanResultsTool(t *testing.T) {
	tool := NewGetComplianceScanResultsTool(&client.Client{})

	require.NotNil(t, tool)
	assert.Equal(t, "get_compliance_scan_results", tool.GetName())
}

func TestGetComplianceScanResultsTool_IsReadOnly(t *testing.T) {
	tool := NewGetComplianceScanResultsTool(&client.Client{})

	assert.True(t, tool.IsReadOnly(), "get_compliance_scan_results should be read-only")
}

func TestGetComplianceScanResultsTool_GetTool(t *testing.T) {
	tool := NewGetComplianceScanResultsTool(&client.Client{})

	mcpTool := tool.GetTool()

	require.NotNil(t, mcpTool)
	assert.Equal(t, "get_compliance_scan_results", mcpTool.Name)
	assert.NotEmpty(t, mcpTool.Description)
	require.NotNil(t, mcpTool.InputSchema, "InputSchema should be defined")
}

func TestGetComplianceScanResultsTool_RegisterWith(t *testing.T) {
	tool := NewGetComplianceScanResultsTool(&client.Client{})
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

func TestGetComplianceScanResults_GeneralQuery(t *testing.T) {
	mockService := mock.NewComplianceResultsServiceMock(
		&v2.ListComplianceResultsResponse{
			ScanResults: []*v2.ComplianceCheckData{
				{
					ClusterId: "cluster-1",
					ScanName:  "daily-scan",
					Result: &v2.ComplianceCheckResult{
						CheckId:     "check-1",
						CheckName:   "ocp4-api-server-encryption",
						Description: "Ensure API server encryption is enabled",
						Status:      v2.ComplianceCheckStatus_PASS,
					},
				},
				{
					ClusterId: "cluster-1",
					ScanName:  "daily-scan",
					Result: &v2.ComplianceCheckResult{
						CheckId:     "check-2",
						CheckName:   "ocp4-etcd-encryption",
						Description: "Ensure etcd encryption is enabled",
						Status:      v2.ComplianceCheckStatus_FAIL,
					},
				},
			},
			TotalCount: 2,
		},
		nil, nil,
	)

	grpcServer, listener := mock.SetupComplianceServer(nil, mockService, nil)
	defer grpcServer.Stop()

	testClient := createTestClient(t, listener)
	tool, ok := NewGetComplianceScanResultsTool(testClient).(*getComplianceScanResultsTool)
	require.True(t, ok)

	ctx := context.Background()
	req := &mcp.CallToolRequest{}
	input := getComplianceScanResultsInput{Query: "Cluster:cluster-1"}

	result, output, err := tool.handle(ctx, req, input)

	require.NoError(t, err)
	require.NotNil(t, output)
	assert.Nil(t, result)

	require.Len(t, output.ScanResults, 2)
	assert.Equal(t, "check-1", output.ScanResults[0].CheckID)
	assert.Equal(t, "PASS", output.ScanResults[0].Status)
	assert.Equal(t, "check-2", output.ScanResults[1].CheckID)
	assert.Equal(t, "FAIL", output.ScanResults[1].Status)
	assert.Equal(t, 2, output.TotalCount)
}

func TestGetComplianceScanResults_WithScanConfigName(t *testing.T) {
	mockService := mock.NewComplianceResultsServiceMock(
		nil,
		&v2.ListComplianceResultsResponse{
			ScanResults: []*v2.ComplianceCheckData{
				{
					ClusterId: "cluster-1",
					ScanName:  "my-scan",
					Result: &v2.ComplianceCheckResult{
						CheckId:   "check-1",
						CheckName: "ocp4-api-server-encryption",
						Status:    v2.ComplianceCheckStatus_PASS,
					},
				},
			},
			TotalCount: 1,
		},
		nil,
	)

	grpcServer, listener := mock.SetupComplianceServer(nil, mockService, nil)
	defer grpcServer.Stop()

	testClient := createTestClient(t, listener)
	tool, ok := NewGetComplianceScanResultsTool(testClient).(*getComplianceScanResultsTool)
	require.True(t, ok)

	ctx := context.Background()
	req := &mcp.CallToolRequest{}
	input := getComplianceScanResultsInput{ScanConfigName: "my-scan"}

	result, output, err := tool.handle(ctx, req, input)

	require.NoError(t, err)
	require.NotNil(t, output)
	assert.Nil(t, result)

	require.Len(t, output.ScanResults, 1)
	assert.Equal(t, "my-scan", output.ScanResults[0].ScanName)
}

func TestGetComplianceScanResults_EmptyResults(t *testing.T) {
	mockService := mock.NewComplianceResultsServiceMock(
		&v2.ListComplianceResultsResponse{
			ScanResults: []*v2.ComplianceCheckData{},
			TotalCount:  0,
		},
		nil, nil,
	)

	grpcServer, listener := mock.SetupComplianceServer(nil, mockService, nil)
	defer grpcServer.Stop()

	testClient := createTestClient(t, listener)
	tool, ok := NewGetComplianceScanResultsTool(testClient).(*getComplianceScanResultsTool)
	require.True(t, ok)

	ctx := context.Background()
	req := &mcp.CallToolRequest{}
	input := getComplianceScanResultsInput{}

	result, output, err := tool.handle(ctx, req, input)

	require.NoError(t, err)
	require.NotNil(t, output)
	assert.Nil(t, result)
	assert.Empty(t, output.ScanResults)
	assert.Equal(t, 0, output.TotalCount)
}

func TestGetComplianceScanResults_GRPCError(t *testing.T) {
	mockErr := status.Error(codes.Internal, "test")
	mockService := mock.NewComplianceResultsServiceMock(nil, nil, mockErr)

	grpcServer, listener := mock.SetupComplianceServer(nil, mockService, nil)
	defer grpcServer.Stop()

	testClient := createTestClient(t, listener)
	tool, ok := NewGetComplianceScanResultsTool(testClient).(*getComplianceScanResultsTool)
	require.True(t, ok)

	ctx := context.Background()
	req := &mcp.CallToolRequest{}
	input := getComplianceScanResultsInput{}

	result, output, err := tool.handle(ctx, req, input)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Nil(t, output)
	assert.Contains(t, err.Error(), "Internal server error")
}
