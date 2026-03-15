package compliance

import (
	"context"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pkg/errors"
	v2 "github.com/stackrox/rox/generated/api/v2"
	"github.com/stackrox/stackrox-mcp/internal/client"
	"github.com/stackrox/stackrox-mcp/internal/client/auth"
	"github.com/stackrox/stackrox-mcp/internal/logging"
	"github.com/stackrox/stackrox-mcp/internal/toolsets"
)

// getComplianceCheckResultInput defines the input parameters for get_compliance_check_result tool.
type getComplianceCheckResultInput struct {
	ProfileName string `json:"profileName"`
	CheckName   string `json:"checkName"`
	Query       string `json:"query,omitempty"`
}

// ControlInfo represents a compliance control reference.
type ControlInfo struct {
	Standard string `json:"standard"`
	Control  string `json:"control"`
}

// ClusterCheckResultInfo represents the check result for a specific cluster.
type ClusterCheckResultInfo struct {
	ClusterID   string `json:"clusterId"`
	ClusterName string `json:"clusterName"`
	Status      string `json:"status"`
}

// getComplianceCheckResultOutput defines the output structure for get_compliance_check_result tool.
type getComplianceCheckResultOutput struct {
	ProfileName  string                   `json:"profileName"`
	CheckName    string                   `json:"checkName"`
	CheckResults []ClusterCheckResultInfo `json:"checkResults"`
	Controls     []ControlInfo  `json:"controls"`
	TotalCount   int                      `json:"totalCount"`
}

// getComplianceCheckResultTool implements the get_compliance_check_result tool.
type getComplianceCheckResultTool struct {
	name   string
	client *client.Client
}

// NewGetComplianceCheckResultTool creates a new get_compliance_check_result tool.
func NewGetComplianceCheckResultTool(c *client.Client) toolsets.Tool {
	return &getComplianceCheckResultTool{
		name:   "get_compliance_check_result",
		client: c,
	}
}

// IsReadOnly returns true as this tool only reads data.
func (t *getComplianceCheckResultTool) IsReadOnly() bool {
	return true
}

// GetName returns the tool name.
func (t *getComplianceCheckResultTool) GetName() string {
	return t.name
}

// GetTool returns the MCP Tool definition.
func (t *getComplianceCheckResultTool) GetTool() *mcp.Tool {
	return &mcp.Tool{
		Name: t.name,
		Description: "Get the result of a specific compliance check across clusters for a given profile." +
			" Returns per-cluster pass/fail status and associated compliance controls." +
			" Use list_compliance_profiles to discover available profiles and check names.",
		InputSchema: getComplianceCheckResultInputSchema(),
	}
}

func getComplianceCheckResultInputSchema() *jsonschema.Schema {
	schema, err := jsonschema.For[getComplianceCheckResultInput](nil)
	if err != nil {
		logging.Fatal("Could not get jsonschema for get_compliance_check_result input", err)

		return nil
	}

	schema.Required = []string{"profileName", "checkName"}

	schema.Properties["profileName"].Description = "The name of the compliance profile (e.g., 'ocp4-cis')."
	schema.Properties["checkName"].Description = "The name of the specific compliance check to query."
	schema.Properties["query"].Description = "Optional StackRox search query to filter results" +
		" (e.g., 'Cluster:mycluster')."

	return schema
}

// RegisterWith registers the get_compliance_check_result tool handler with the MCP server.
func (t *getComplianceCheckResultTool) RegisterWith(server *mcp.Server) {
	mcp.AddTool(server, t.GetTool(), t.handle)
}

// handle is the handler for get_compliance_check_result tool.
//
//nolint:funlen
func (t *getComplianceCheckResultTool) handle(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input getComplianceCheckResultInput,
) (*mcp.CallToolResult, *getComplianceCheckResultOutput, error) {
	if input.ProfileName == "" {
		return nil, nil, errors.New("profileName is required")
	}

	if input.CheckName == "" {
		return nil, nil, errors.New("checkName is required")
	}

	conn, err := t.client.ReadyConn(ctx)
	if err != nil {
		return nil, nil, errors.Wrap(err, "unable to connect to server")
	}

	callCtx := auth.WithMCPRequestContext(ctx, req)

	resultsClient := v2.NewComplianceResultsServiceClient(conn)

	checkReq := &v2.ComplianceProfileCheckRequest{
		ProfileName: input.ProfileName,
		CheckName:   input.CheckName,
	}

	if input.Query != "" {
		checkReq.Query = &v2.RawQuery{Query: input.Query}
	}

	resp, err := resultsClient.GetComplianceProfileCheckResult(callCtx, checkReq)
	if err != nil {
		return nil, nil, client.NewError(err, "GetComplianceProfileCheckResult")
	}

	checkResults := make([]ClusterCheckResultInfo, 0, len(resp.GetCheckResults()))
	for _, cr := range resp.GetCheckResults() {
		checkResults = append(checkResults, ClusterCheckResultInfo{
			ClusterID:   cr.GetCluster().GetClusterId(),
			ClusterName: cr.GetCluster().GetClusterName(),
			Status:      cr.GetStatus().String(),
		})
	}

	controls := make([]ControlInfo, 0, len(resp.GetControls()))
	for _, ctrl := range resp.GetControls() {
		controls = append(controls, ControlInfo{
			Standard: ctrl.GetStandard(),
			Control:  ctrl.GetControl(),
		})
	}

	output := &getComplianceCheckResultOutput{
		ProfileName:  resp.GetProfileName(),
		CheckName:    resp.GetCheckName(),
		CheckResults: checkResults,
		Controls:     controls,
		TotalCount:   int(resp.GetTotalCount()),
	}

	return nil, output, nil
}
