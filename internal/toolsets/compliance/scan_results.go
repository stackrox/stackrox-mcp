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

// getComplianceScanResultsInput defines the input parameters for get_compliance_scan_results tool.
type getComplianceScanResultsInput struct {
	ScanConfigName string `json:"scanConfigName,omitempty"`
	Query          string `json:"query,omitempty"`
}

// ScanResultInfo represents a single compliance scan result.
type ScanResultInfo struct {
	ClusterID   string `json:"clusterId"`
	ScanName    string `json:"scanName"`
	CheckID     string `json:"checkId"`
	CheckName   string `json:"checkName"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

// getComplianceScanResultsOutput defines the output structure for get_compliance_scan_results tool.
type getComplianceScanResultsOutput struct {
	ScanResults []ScanResultInfo `json:"scanResults"`
	TotalCount  int                        `json:"totalCount"`
}

// getComplianceScanResultsTool implements the get_compliance_scan_results tool.
type getComplianceScanResultsTool struct {
	name   string
	client *client.Client
}

// NewGetComplianceScanResultsTool creates a new get_compliance_scan_results tool.
func NewGetComplianceScanResultsTool(c *client.Client) toolsets.Tool {
	return &getComplianceScanResultsTool{
		name:   "get_compliance_scan_results",
		client: c,
	}
}

// IsReadOnly returns true as this tool only reads data.
func (t *getComplianceScanResultsTool) IsReadOnly() bool {
	return true
}

// GetName returns the tool name.
func (t *getComplianceScanResultsTool) GetName() string {
	return t.name
}

// GetTool returns the MCP Tool definition.
func (t *getComplianceScanResultsTool) GetTool() *mcp.Tool {
	return &mcp.Tool{
		Name: t.name,
		Description: "Get compliance scan results." +
			" If scanConfigName is provided, returns results for that specific scan configuration." +
			" Otherwise, returns general compliance scan results filtered by an optional query.",
		InputSchema: getComplianceScanResultsInputSchema(),
	}
}

func getComplianceScanResultsInputSchema() *jsonschema.Schema {
	schema, err := jsonschema.For[getComplianceScanResultsInput](nil)
	if err != nil {
		logging.Fatal("Could not get jsonschema for get_compliance_scan_results input", err)

		return nil
	}

	schema.Properties["scanConfigName"].Description = "Name of the scan configuration to get results for." +
		" When provided, results are scoped to this scan configuration."
	schema.Properties["query"].Description = "Optional StackRox search query to filter results" +
		" (e.g., 'Cluster:mycluster')."

	return schema
}

// RegisterWith registers the get_compliance_scan_results tool handler with the MCP server.
func (t *getComplianceScanResultsTool) RegisterWith(server *mcp.Server) {
	mcp.AddTool(server, t.GetTool(), t.handle)
}

// handle is the handler for get_compliance_scan_results tool.
func (t *getComplianceScanResultsTool) handle(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input getComplianceScanResultsInput,
) (*mcp.CallToolResult, *getComplianceScanResultsOutput, error) {
	conn, err := t.client.ReadyConn(ctx)
	if err != nil {
		return nil, nil, errors.Wrap(err, "unable to connect to server")
	}

	callCtx := auth.WithMCPRequestContext(ctx, req)

	resultsClient := v2.NewComplianceResultsServiceClient(conn)

	var resp *v2.ListComplianceResultsResponse

	if input.ScanConfigName != "" {
		resp, err = resultsClient.GetComplianceScanConfigurationResults(callCtx, &v2.ComplianceScanResultsRequest{
			ScanConfigName: input.ScanConfigName,
			Query:          &v2.RawQuery{Query: input.Query},
		})
	} else {
		resp, err = resultsClient.GetComplianceScanResults(callCtx, &v2.RawQuery{
			Query: input.Query,
		})
	}

	if err != nil {
		return nil, nil, client.NewError(err, "GetComplianceScanResults")
	}

	scanResults := make([]ScanResultInfo, 0, len(resp.GetScanResults()))
	for _, sr := range resp.GetScanResults() {
		result := sr.GetResult()

		info := ScanResultInfo{
			ClusterID: sr.GetClusterId(),
			ScanName:  sr.GetScanName(),
		}

		if result != nil {
			info.CheckID = result.GetCheckId()
			info.CheckName = result.GetCheckName()
			info.Description = result.GetDescription()
			info.Status = result.GetStatus().String()
		}

		scanResults = append(scanResults, info)
	}

	output := &getComplianceScanResultsOutput{
		ScanResults: scanResults,
		TotalCount:  int(resp.GetTotalCount()),
	}

	return nil, output, nil
}
