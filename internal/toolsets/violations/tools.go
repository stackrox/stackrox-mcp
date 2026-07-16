package violations

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pkg/errors"
	v1 "github.com/stackrox/rox/generated/api/v1"
	"github.com/stackrox/stackrox-mcp/internal/client"
	"github.com/stackrox/stackrox-mcp/internal/client/auth"
	"github.com/stackrox/stackrox-mcp/internal/config"
	"github.com/stackrox/stackrox-mcp/internal/cursor"
	"github.com/stackrox/stackrox-mcp/internal/logging"
	"github.com/stackrox/stackrox-mcp/internal/toolsets"
	"github.com/stackrox/stackrox-mcp/internal/toolsets/vulnerability"
)

const defaultLimit = 100

type filterSeverity string

const (
	filterSeverityNoFilter filterSeverity = "NO_FILTER"
	filterSeverityLow      filterSeverity = "LOW"
	filterSeverityMedium   filterSeverity = "MEDIUM"
	filterSeverityHigh     filterSeverity = "HIGH"
	filterSeverityCritical filterSeverity = "CRITICAL"
)

type filterState string

const (
	filterStateNoFilter  filterState = "NO_FILTER"
	filterStateActive    filterState = "ACTIVE"
	filterStateResolved  filterState = "RESOLVED"
	filterStateAttempted filterState = "ATTEMPTED"
)

// listViolationsInput defines the input parameters for list_violations tool.
type listViolationsInput struct {
	FilterClusterID   string         `json:"filterClusterId,omitempty"`
	FilterClusterName string         `json:"filterClusterName,omitempty"`
	FilterNamespace   string         `json:"filterNamespace,omitempty"`
	FilterPolicyName  string         `json:"filterPolicyName,omitempty"`
	FilterSeverity    filterSeverity `json:"filterSeverity,omitempty"`
	FilterState       filterState    `json:"filterState,omitempty"`
	Cursor            string         `json:"cursor,omitempty"`
}

func (input *listViolationsInput) validate() error {
	if input.FilterClusterID != "" && input.FilterClusterName != "" {
		return errors.New("cannot specify both filterClusterId and filterClusterName")
	}

	return nil
}

// PolicyInfo represents policy information for a violation.
type PolicyInfo struct {
	Name        string   `json:"name"`
	Severity    string   `json:"severity"`
	Description string   `json:"description"`
	Categories  []string `json:"categories"`
}

// ViolationResult represents a single policy violation.
type ViolationResult struct {
	ID             string     `json:"id"`
	LifecycleStage string     `json:"lifecycleStage"`
	Time           string     `json:"time"`
	Policy         PolicyInfo `json:"policy"`
	State          string     `json:"state"`
	ClusterName    string     `json:"clusterName"`
	Namespace      string     `json:"namespace"`
	DeploymentName string     `json:"deploymentName,omitempty"`
	ResourceName   string     `json:"resourceName,omitempty"`
}

// listViolationsOutput defines the output structure for list_violations tool.
type listViolationsOutput struct {
	Violations []ViolationResult `json:"violations"`
	NextCursor string            `json:"nextCursor"`
}

// listViolationsTool implements the list_violations tool.
type listViolationsTool struct {
	name   string
	client *client.Client
}

// NewListViolationsTool creates a new list_violations tool.
func NewListViolationsTool(c *client.Client) toolsets.Tool {
	return &listViolationsTool{
		name:   "list_violations",
		client: c,
	}
}

// IsReadOnly returns true as this tool only reads data.
func (t *listViolationsTool) IsReadOnly() bool {
	return true
}

// GetName returns the tool name.
func (t *listViolationsTool) GetName() string {
	return t.name
}

// GetTool returns the MCP Tool definition.
func (t *listViolationsTool) GetTool() *mcp.Tool {
	return &mcp.Tool{
		Name: t.name,
		Description: "List policy violations (alerts) detected by " + config.GetProductDisplayName() + "." +
			" Returns violations with policy details, severity, state, and affected deployment/resource information." +
			" Supports filtering by cluster, namespace, policy name, severity, and state.",
		InputSchema: listViolationsInputSchema(),
	}
}

func listViolationsInputSchema() *jsonschema.Schema {
	schema, err := jsonschema.For[listViolationsInput](nil)
	if err != nil {
		logging.Fatal("Could not get jsonschema for list_violations input", err)

		return nil
	}

	schema.Properties["filterClusterId"].Description = "Optional cluster ID to filter violations." +
		" Cannot be used together with filterClusterName."
	schema.Properties["filterClusterName"].Description = "Optional cluster name to filter violations." +
		" Cannot be used together with filterClusterId."
	schema.Properties["filterNamespace"].Description = "Optional namespace to filter violations."
	schema.Properties["filterPolicyName"].Description = "Optional policy name to filter violations."

	schema.Properties["filterSeverity"].Description = "Optional severity filter: LOW, MEDIUM, HIGH, or CRITICAL."
	schema.Properties["filterSeverity"].Default = toolsets.MustJSONMarshal(filterSeverityNoFilter)
	schema.Properties["filterSeverity"].Enum = []any{
		filterSeverityNoFilter,
		filterSeverityLow,
		filterSeverityMedium,
		filterSeverityHigh,
		filterSeverityCritical,
	}

	schema.Properties["filterState"].Description = "Optional state filter: ACTIVE, RESOLVED, or ATTEMPTED."
	schema.Properties["filterState"].Default = toolsets.MustJSONMarshal(filterStateNoFilter)
	schema.Properties["filterState"].Enum = []any{
		filterStateNoFilter,
		filterStateActive,
		filterStateResolved,
		filterStateAttempted,
	}

	schema.Properties["cursor"].Description = "Cursor for next page provided by server."

	return schema
}

// RegisterWith registers the list_violations tool handler with the MCP server.
func (t *listViolationsTool) RegisterWith(server *mcp.Server) {
	mcp.AddTool(server, t.GetTool(), t.handle)
}

func buildAlertsQuery(input listViolationsInput) string {
	var queryParts []string

	if input.FilterClusterID != "" {
		queryParts = append(queryParts, fmt.Sprintf("Cluster ID:%q", input.FilterClusterID))
	}

	if input.FilterNamespace != "" {
		queryParts = append(queryParts, fmt.Sprintf("Namespace:%q", input.FilterNamespace))
	}

	if input.FilterPolicyName != "" {
		queryParts = append(queryParts, fmt.Sprintf("Policy:%q", input.FilterPolicyName))
	}

	switch input.FilterSeverity {
	case filterSeverityLow:
		queryParts = append(queryParts, "Severity:LOW_SEVERITY")
	case filterSeverityMedium:
		queryParts = append(queryParts, "Severity:MEDIUM_SEVERITY")
	case filterSeverityHigh:
		queryParts = append(queryParts, "Severity:HIGH_SEVERITY")
	case filterSeverityCritical:
		queryParts = append(queryParts, "Severity:CRITICAL_SEVERITY")
	case filterSeverityNoFilter:
	}

	switch input.FilterState {
	case filterStateActive:
		queryParts = append(queryParts, "Violation State:ACTIVE")
	case filterStateResolved:
		queryParts = append(queryParts, "Violation State:RESOLVED")
	case filterStateAttempted:
		queryParts = append(queryParts, "Violation State:ATTEMPTED")
	case filterStateNoFilter:
	}

	return strings.Join(queryParts, "+")
}

func getAlertsCursor(input *listViolationsInput) (*cursor.Cursor, error) {
	if input.Cursor == "" {
		startCursor, err := cursor.New(0)

		return startCursor, errors.Wrap(err, "error creating starting cursor")
	}

	currCursor, err := cursor.Decode(input.Cursor)
	if err != nil {
		return nil, errors.Wrap(err, "error decoding cursor")
	}

	return currCursor, nil
}

func mapSeverity(severity string) string {
	switch severity {
	case "LOW_SEVERITY":
		return "LOW"
	case "MEDIUM_SEVERITY":
		return "MEDIUM"
	case "HIGH_SEVERITY":
		return "HIGH"
	case "CRITICAL_SEVERITY":
		return "CRITICAL"
	default:
		return severity
	}
}

// handle is the handler for list_violations tool.
func (t *listViolationsTool) handle(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input listViolationsInput,
) (*mcp.CallToolResult, *listViolationsOutput, error) {
	err := input.validate()
	if err != nil {
		return nil, nil, err
	}

	currCursor, err := getAlertsCursor(&input)
	if err != nil {
		return nil, nil, err
	}

	conn, err := t.client.ReadyConn(ctx)
	if err != nil {
		return nil, nil, errors.Wrap(err, "unable to connect to server")
	}

	callCtx := auth.WithMCPRequestContext(ctx, req)

	resolvedClusterID, err := vulnerability.ResolveClusterID(callCtx, conn, input.FilterClusterID, input.FilterClusterName)
	if err != nil {
		return nil, nil, err
	}

	alertClient := v1.NewAlertServiceClient(conn)

	queryInput := listViolationsInput{
		FilterClusterID:  resolvedClusterID,
		FilterNamespace:  input.FilterNamespace,
		FilterPolicyName: input.FilterPolicyName,
		FilterSeverity:   input.FilterSeverity,
		FilterState:      input.FilterState,
	}

	listReq := &v1.ListAlertsRequest{
		Query: buildAlertsQuery(queryInput),
		Pagination: &v1.Pagination{
			Offset: currCursor.GetOffset(),
			Limit:  defaultLimit + 1,
		},
	}

	resp, err := alertClient.ListAlerts(callCtx, listReq)
	if err != nil {
		return nil, nil, client.NewError(err, "ListAlerts")
	}

	rawAlerts := resp.GetAlerts()

	violations := make([]ViolationResult, len(rawAlerts))
	for i, alert := range rawAlerts {
		v := ViolationResult{
			ID:             alert.GetId(),
			LifecycleStage: alert.GetLifecycleStage().String(),
			State:          alert.GetState().String(),
		}

		if alert.GetTime() != nil {
			v.Time = alert.GetTime().AsTime().UTC().Format("2006-01-02T15:04:05Z")
		}

		if p := alert.GetPolicy(); p != nil {
			v.Policy = PolicyInfo{
				Name:        p.GetName(),
				Severity:    mapSeverity(p.GetSeverity().String()),
				Description: p.GetDescription(),
				Categories:  p.GetCategories(),
			}
		}

		if info := alert.GetCommonEntityInfo(); info != nil {
			v.ClusterName = info.GetClusterName()
			v.Namespace = info.GetNamespace()
		}

		if d := alert.GetDeployment(); d != nil {
			v.DeploymentName = d.GetName()
		}

		if r := alert.GetResource(); r != nil {
			v.ResourceName = r.GetName()
		}

		violations[i] = v
	}

	if len(violations) <= defaultLimit {
		return nil, &listViolationsOutput{Violations: violations}, nil
	}

	nextCursorStr, err := currCursor.GetNextCursor(defaultLimit).Encode()
	if err != nil {
		return nil, nil, errors.Wrap(err, "unable to create next cursor")
	}

	output := &listViolationsOutput{
		Violations: violations[:len(violations)-1],
		NextCursor: nextCursorStr,
	}

	return nil, output, nil
}
