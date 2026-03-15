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

func TestNewGetComplianceCheckResultTool(t *testing.T) {
	tool := NewGetComplianceCheckResultTool(&client.Client{})

	require.NotNil(t, tool)
	assert.Equal(t, "get_compliance_check_result", tool.GetName())
}

func TestGetComplianceCheckResultTool_IsReadOnly(t *testing.T) {
	tool := NewGetComplianceCheckResultTool(&client.Client{})

	assert.True(t, tool.IsReadOnly(), "get_compliance_check_result should be read-only")
}

func TestGetComplianceCheckResultTool_GetTool(t *testing.T) {
	tool := NewGetComplianceCheckResultTool(&client.Client{})

	mcpTool := tool.GetTool()

	require.NotNil(t, mcpTool)
	assert.Equal(t, "get_compliance_check_result", mcpTool.Name)
	assert.NotEmpty(t, mcpTool.Description)
	require.NotNil(t, mcpTool.InputSchema, "InputSchema should be defined")
}

func TestGetComplianceCheckResultTool_RegisterWith(t *testing.T) {
	tool := NewGetComplianceCheckResultTool(&client.Client{})
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

//nolint:funlen
func TestGetComplianceCheckResult_Success(t *testing.T) {
	mockService := mock.NewComplianceResultsServiceMock(
		nil, nil, nil,
	)
	mockService.SetCheckResult(&v2.ListComplianceCheckClusterResponse{
		ProfileName: "ocp4-cis",
		CheckName:   "ocp4-api-server-encryption",
		CheckResults: []*v2.ClusterCheckStatus{
			{
				Cluster: &v2.ComplianceScanCluster{
					ClusterId:   "cluster-1",
					ClusterName: "production",
				},
				Status: v2.ComplianceCheckStatus_PASS,
			},
			{
				Cluster: &v2.ComplianceScanCluster{
					ClusterId:   "cluster-2",
					ClusterName: "staging",
				},
				Status: v2.ComplianceCheckStatus_FAIL,
			},
		},
		Controls: []*v2.ComplianceControl{
			{
				Standard: "CIS",
				Control:  "1.2.1",
			},
		},
		TotalCount: 2,
	})

	grpcServer, listener := mock.SetupComplianceServer(nil, mockService, nil)
	defer grpcServer.Stop()

	testClient := createTestClient(t, listener)
	tool, ok := NewGetComplianceCheckResultTool(testClient).(*getComplianceCheckResultTool)
	require.True(t, ok)

	ctx := context.Background()
	req := &mcp.CallToolRequest{}
	input := getComplianceCheckResultInput{
		ProfileName: "ocp4-cis",
		CheckName:   "ocp4-api-server-encryption",
	}

	result, output, err := tool.handle(ctx, req, input)

	require.NoError(t, err)
	require.NotNil(t, output)
	assert.Nil(t, result)

	assert.Equal(t, "ocp4-cis", output.ProfileName)
	assert.Equal(t, "ocp4-api-server-encryption", output.CheckName)
	require.Len(t, output.CheckResults, 2)
	assert.Equal(t, "cluster-1", output.CheckResults[0].ClusterID)
	assert.Equal(t, "production", output.CheckResults[0].ClusterName)
	assert.Equal(t, "PASS", output.CheckResults[0].Status)
	assert.Equal(t, "cluster-2", output.CheckResults[1].ClusterID)
	assert.Equal(t, "FAIL", output.CheckResults[1].Status)

	require.Len(t, output.Controls, 1)
	assert.Equal(t, "CIS", output.Controls[0].Standard)
	assert.Equal(t, "1.2.1", output.Controls[0].Control)
	assert.Equal(t, 2, output.TotalCount)
}

func TestGetComplianceCheckResult_MissingProfileName(t *testing.T) {
	tool, ok := NewGetComplianceCheckResultTool(&client.Client{}).(*getComplianceCheckResultTool)
	require.True(t, ok)

	ctx := context.Background()
	req := &mcp.CallToolRequest{}
	input := getComplianceCheckResultInput{
		ProfileName: "",
		CheckName:   "some-check",
	}

	_, _, err := tool.handle(ctx, req, input)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "profileName is required")
}

func TestGetComplianceCheckResult_MissingCheckName(t *testing.T) {
	tool, ok := NewGetComplianceCheckResultTool(&client.Client{}).(*getComplianceCheckResultTool)
	require.True(t, ok)

	ctx := context.Background()
	req := &mcp.CallToolRequest{}
	input := getComplianceCheckResultInput{
		ProfileName: "ocp4-cis",
		CheckName:   "",
	}

	_, _, err := tool.handle(ctx, req, input)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "checkName is required")
}

func TestGetComplianceCheckResult_GRPCError(t *testing.T) {
	mockService := mock.NewComplianceResultsServiceMock(nil, nil, nil)
	mockService.SetCheckResultError(status.Error(codes.Internal, "test"))

	grpcServer, listener := mock.SetupComplianceServer(nil, mockService, nil)
	defer grpcServer.Stop()

	testClient := createTestClient(t, listener)
	tool, ok := NewGetComplianceCheckResultTool(testClient).(*getComplianceCheckResultTool)
	require.True(t, ok)

	ctx := context.Background()
	req := &mcp.CallToolRequest{}
	input := getComplianceCheckResultInput{
		ProfileName: "ocp4-cis",
		CheckName:   "some-check",
	}

	result, output, err := tool.handle(ctx, req, input)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Nil(t, output)
	assert.Contains(t, err.Error(), "Internal server error")
}
