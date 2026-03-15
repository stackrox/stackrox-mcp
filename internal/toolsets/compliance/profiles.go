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

// listComplianceProfilesInput defines the input parameters for list_compliance_profiles tool.
type listComplianceProfilesInput struct {
	ClusterID string `json:"clusterId"`
}

// ProfileInfo represents information about a single compliance profile.
type ProfileInfo struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	ProfileVersion string `json:"profileVersion"`
	Description    string `json:"description"`
	Title          string `json:"title"`
	RuleCount      int    `json:"ruleCount"`
}

// listComplianceProfilesOutput defines the output structure for list_compliance_profiles tool.
type listComplianceProfilesOutput struct {
	Profiles   []ProfileInfo `json:"profiles"`
	TotalCount int                     `json:"totalCount"`
}

// listComplianceProfilesTool implements the list_compliance_profiles tool.
type listComplianceProfilesTool struct {
	name   string
	client *client.Client
}

// NewListComplianceProfilesTool creates a new list_compliance_profiles tool.
func NewListComplianceProfilesTool(c *client.Client) toolsets.Tool {
	return &listComplianceProfilesTool{
		name:   "list_compliance_profiles",
		client: c,
	}
}

// IsReadOnly returns true as this tool only reads data.
func (t *listComplianceProfilesTool) IsReadOnly() bool {
	return true
}

// GetName returns the tool name.
func (t *listComplianceProfilesTool) GetName() string {
	return t.name
}

// GetTool returns the MCP Tool definition.
func (t *listComplianceProfilesTool) GetTool() *mcp.Tool {
	return &mcp.Tool{
		Name: t.name,
		Description: "List compliance profiles available for a specific cluster." +
			" Returns profiles such as CIS benchmarks, NIST, and PCI DSS standards." +
			" Use list_clusters first to obtain the cluster ID.",
		InputSchema: listComplianceProfilesInputSchema(),
	}
}

func listComplianceProfilesInputSchema() *jsonschema.Schema {
	schema, err := jsonschema.For[listComplianceProfilesInput](nil)
	if err != nil {
		logging.Fatal("Could not get jsonschema for list_compliance_profiles input", err)

		return nil
	}

	schema.Required = []string{"clusterId"}

	schema.Properties["clusterId"].Description = "The ID of the cluster to list compliance profiles for." +
		" Use list_clusters to find available cluster IDs."

	return schema
}

// RegisterWith registers the list_compliance_profiles tool handler with the MCP server.
func (t *listComplianceProfilesTool) RegisterWith(server *mcp.Server) {
	mcp.AddTool(server, t.GetTool(), t.handle)
}

// handle is the handler for list_compliance_profiles tool.
func (t *listComplianceProfilesTool) handle(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input listComplianceProfilesInput,
) (*mcp.CallToolResult, *listComplianceProfilesOutput, error) {
	if input.ClusterID == "" {
		return nil, nil, errors.New("clusterId is required")
	}

	conn, err := t.client.ReadyConn(ctx)
	if err != nil {
		return nil, nil, errors.Wrap(err, "unable to connect to server")
	}

	callCtx := auth.WithMCPRequestContext(ctx, req)

	profileClient := v2.NewComplianceProfileServiceClient(conn)

	resp, err := profileClient.ListComplianceProfiles(callCtx, &v2.ProfilesForClusterRequest{
		ClusterId: input.ClusterID,
	})
	if err != nil {
		return nil, nil, client.NewError(err, "ListComplianceProfiles")
	}

	profiles := make([]ProfileInfo, 0, len(resp.GetProfiles()))
	for _, profile := range resp.GetProfiles() {
		profiles = append(profiles, ProfileInfo{
			ID:             profile.GetId(),
			Name:           profile.GetName(),
			ProfileVersion: profile.GetProfileVersion(),
			Description:    profile.GetDescription(),
			Title:          profile.GetTitle(),
			RuleCount:      len(profile.GetRules()),
		})
	}

	output := &listComplianceProfilesOutput{
		Profiles:   profiles,
		TotalCount: int(resp.GetTotalCount()),
	}

	return nil, output, nil
}
