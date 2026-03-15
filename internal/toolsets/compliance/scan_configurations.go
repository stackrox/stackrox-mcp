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

// listComplianceScanConfigurationsInput defines the input parameters for list_compliance_scan_configurations tool.
type listComplianceScanConfigurationsInput struct {
	Query string `json:"query,omitempty"`
}

// ScanConfigInfo represents a single compliance scan configuration.
type ScanConfigInfo struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Profiles    []string `json:"profiles"`
	Clusters    []string `json:"clusters"`
}

// listComplianceScanConfigurationsOutput defines the output structure for list_compliance_scan_configurations tool.
type listComplianceScanConfigurationsOutput struct {
	Configurations []ScanConfigInfo `json:"configurations"`
	TotalCount     int                        `json:"totalCount"`
}

// listComplianceScanConfigurationsTool implements the list_compliance_scan_configurations tool.
type listComplianceScanConfigurationsTool struct {
	name   string
	client *client.Client
}

// NewListComplianceScanConfigurationsTool creates a new list_compliance_scan_configurations tool.
func NewListComplianceScanConfigurationsTool(c *client.Client) toolsets.Tool {
	return &listComplianceScanConfigurationsTool{
		name:   "list_compliance_scan_configurations",
		client: c,
	}
}

// IsReadOnly returns true as this tool only reads data.
func (t *listComplianceScanConfigurationsTool) IsReadOnly() bool {
	return true
}

// GetName returns the tool name.
func (t *listComplianceScanConfigurationsTool) GetName() string {
	return t.name
}

// GetTool returns the MCP Tool definition.
func (t *listComplianceScanConfigurationsTool) GetTool() *mcp.Tool {
	return &mcp.Tool{
		Name: t.name,
		Description: "List compliance scan configurations." +
			" Returns configured compliance scans including their associated profiles and clusters." +
			" Use this to discover what compliance scans are set up in the system.",
		InputSchema: listComplianceScanConfigurationsInputSchema(),
	}
}

func listComplianceScanConfigurationsInputSchema() *jsonschema.Schema {
	schema, err := jsonschema.For[listComplianceScanConfigurationsInput](nil)
	if err != nil {
		logging.Fatal("Could not get jsonschema for list_compliance_scan_configurations input", err)

		return nil
	}

	schema.Properties["query"].Description = "Optional StackRox search query to filter scan configurations."

	return schema
}

// RegisterWith registers the list_compliance_scan_configurations tool handler with the MCP server.
func (t *listComplianceScanConfigurationsTool) RegisterWith(server *mcp.Server) {
	mcp.AddTool(server, t.GetTool(), t.handle)
}

// handle is the handler for list_compliance_scan_configurations tool.
func (t *listComplianceScanConfigurationsTool) handle(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input listComplianceScanConfigurationsInput,
) (*mcp.CallToolResult, *listComplianceScanConfigurationsOutput, error) {
	conn, err := t.client.ReadyConn(ctx)
	if err != nil {
		return nil, nil, errors.Wrap(err, "unable to connect to server")
	}

	callCtx := auth.WithMCPRequestContext(ctx, req)

	scanConfigClient := v2.NewComplianceScanConfigurationServiceClient(conn)

	resp, err := scanConfigClient.ListComplianceScanConfigurations(callCtx, &v2.RawQuery{
		Query: input.Query,
	})
	if err != nil {
		return nil, nil, client.NewError(err, "ListComplianceScanConfigurations")
	}

	configs := make([]ScanConfigInfo, 0, len(resp.GetConfigurations()))
	for _, cfg := range resp.GetConfigurations() {
		info := ScanConfigInfo{
			ID:   cfg.GetId(),
			Name: cfg.GetScanName(),
		}

		if scanCfg := cfg.GetScanConfig(); scanCfg != nil {
			info.Description = scanCfg.GetDescription()
			info.Profiles = scanCfg.GetProfiles()
		}

		clusterStatuses := cfg.GetClusterStatus()

		clusters := make([]string, 0, len(clusterStatuses))
		for _, cs := range clusterStatuses {
			clusters = append(clusters, cs.GetClusterId())
		}

		info.Clusters = clusters

		configs = append(configs, info)
	}

	output := &listComplianceScanConfigurationsOutput{
		Configurations: configs,
		TotalCount:     int(resp.GetTotalCount()),
	}

	return nil, output, nil
}
