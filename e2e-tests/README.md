# StackRox MCP E2E Testing

End-to-end tests for the StackRox MCP server using [gevals](https://github.com/genmcp/gevals).

## Prerequisites

- Go 1.25+
- Google Cloud Project with Vertex AI enabled (for Claude agent)
- OpenAI API Key (for LLM judge)
- StackRox API Token

## Setup

### 1. Build gevals

```bash
cd e2e-tests
./scripts/build-gevals.sh
```

### 2. Configure Environment

Create `.env` file:

```bash
# Required: GCP Project for Vertex AI (Claude agent)
ANTHROPIC_VERTEX_PROJECT_ID=<GCP Project ID>

# Required: StackRox Central API Token
STACKROX_MCP__CENTRAL__API_TOKEN=<StackRox API Token>

# Required: OpenAI API Key (for LLM judge)
OPENAI_API_KEY=<OpenAI API Key>

# Optional: Vertex AI region (defaults to us-east5)
CLOUD_ML_REGION=us-east5

# Optional: Judge configuration (defaults to OpenAI)
JUDGE_MODEL_NAME=gpt-5-nano
```

## Running Tests

```bash
./scripts/run-tests.sh
```

Results are saved to `gevals/gevals-stackrox-mcp-e2e-out.json`.

### View Results

```bash
# Summary
jq '.[] | {taskName, taskPassed}' gevals/gevals-stackrox-mcp-e2e-out.json

# Tool calls
jq '.[].callHistory[] | {toolName, arguments}' gevals/gevals-stackrox-mcp-e2e-out.json
```

## Test Cases

| Test | Description | Tool |
|------|-------------|------|
| `list-clusters` | List all clusters | `list_clusters` |
| `cve-detected-workloads` | CVE detected in deployments | `get_deployments_for_cve` |
| `cve-detected-clusters` | CVE detected in clusters | `get_clusters_with_orchestrator_cve` |
| `cve-nonexistent` | Handle non-existent CVE | `get_clusters_with_orchestrator_cve` |
| `cve-cluster-does-exist` | CVE with cluster filter | `get_clusters_with_orchestrator_cve` |
| `cve-cluster-does-not-exist` | CVE with cluster filter | `get_clusters_with_orchestrator_cve` |
| `cve-clusters-general` | General CVE query | `get_clusters_with_orchestrator_cve` |
| `cve-cluster-list` | CVE across clusters | `get_clusters_with_orchestrator_cve` |

## Configuration

- **`gevals/eval.yaml`**: Main test configuration, agent settings, assertions
- **`gevals/mcp-config.yaml`**: MCP server configuration
- **`gevals/tasks/*.yaml`**: Individual test task definitions

## How It Works

Gevals uses a proxy architecture to intercept MCP tool calls:

1. AI agent receives task prompt
2. Agent calls MCP tool
3. Gevals proxy intercepts and records the call
4. Call forwarded to StackRox MCP server
5. Server executes and returns result
6. Gevals validates assertions and response quality

## Troubleshooting

**Tests fail - no tools called**
- Verify StackRox Central is accessible
- Check API token permissions

**Build errors**
```bash
go mod tidy
./scripts/build-gevals.sh
```

## Further Reading

- [Gevals Documentation](https://github.com/genmcp/gevals)
- [StackRox MCP Server](../README.md)
