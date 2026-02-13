# StackRox MCP E2E Testing

End-to-end tests for the StackRox MCP server using [mcpchecker](https://github.com/mcpchecker/mcpchecker).

## Quick Start

### Smoke Test (No Agent Required)

Validate configuration and build without running actual agents:

```bash
cd e2e-tests
./scripts/smoke-test.sh
```

This is useful for CI and quickly checking that everything compiles.

## Prerequisites

- Go 1.25+
- Google Cloud Project with Vertex AI enabled (for Claude agent)
- OpenAI API Key (for LLM judge)
- StackRox API Token

## Setup

### 1. Build mcpchecker

```bash
cd e2e-tests
./scripts/build-mcpchecker.sh
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

### Mock Mode (Recommended for Development)

Run tests against the WireMock mock service (no credentials required):

```bash
./scripts/run-tests.sh --mock
```

This mode:
- Starts WireMock automatically on localhost:8081
- Uses deterministic test fixtures
- Requires no API tokens or real StackRox instance
- Fast and reliable for local development

### Real Mode

Run tests against a real StackRox Central instance:

```bash
./scripts/run-tests.sh --real
```

This mode:
- Uses the real StackRox Central API (staging.demo.stackrox.com by default)
- Requires valid API token in `.env`
- Tests against actual production data

Results are saved to `mcpchecker/mcpchecker-stackrox-mcp-e2e-out.json`.

### View Results

```bash
# Summary
jq '.[] | {taskName, taskPassed}' mcpchecker/mcpchecker-stackrox-mcp-e2e-out.json

# Tool calls
jq '[.[] | .callHistory.ToolCalls[]? | {name: .request.Params.name, arguments: .request.Params.arguments}]' mcpchecker/mcpchecker-stackrox-mcp-e2e-out.json
```

## Test Cases

| Test | Description | Tool | Eval Coverage |
|------|-------------|------|---------------|
| `list-clusters` | List all clusters | `list_clusters` | - |
| `cve-detected-workloads` | CVE detected in deployments | `get_deployments_for_cve` | Eval 1 |
| `cve-detected-clusters` | CVE detected in clusters | `get_clusters_with_orchestrator_cve` | Eval 1 |
| `cve-nonexistent` | Handle non-existent CVE | `get_clusters_with_orchestrator_cve` | Eval 2 |
| `cve-cluster-does-exist` | CVE with cluster filter | `get_clusters_with_orchestrator_cve` | Eval 4 |
| `cve-cluster-does-not-exist` | CVE with non-existent cluster | `list_clusters` | - |
| `cve-clusters-general` | General CVE query | `get_clusters_with_orchestrator_cve` | Eval 1 |
| `cve-cluster-list` | CVE across clusters | `get_clusters_with_orchestrator_cve` | - |
| `cve-log4shell` | Well-known CVE (log4shell) | `get_deployments_for_cve` | Eval 3 |
| `cve-multiple` | Multiple CVEs in one prompt | `get_deployments_for_cve` | Eval 5 |
| `rhsa-not-supported` | RHSA detection (should fail) | None | Eval 7 |

## Configuration

- **`mcpchecker/eval.yaml`**: Test configuration, agent settings, assertions (for mock mode)
- **`mcpchecker/mcp-config.yaml`**: MCP server configuration
- **`mcpchecker/tasks/*.yaml`**: Individual test task definitions

## How It Works

mcpchecker uses a proxy architecture to intercept MCP tool calls:

1. AI agent receives task prompt
2. Agent calls MCP tool
3. mcpchecker proxy intercepts and records the call
4. Call forwarded to StackRox MCP server
5. Server executes and returns result
6. mcpchecker validates assertions and response quality

## Troubleshooting

**Tests fail - no tools called**
- Verify StackRox Central is accessible
- Check API token permissions

**Build errors**
```bash
go mod tidy
./scripts/build-mcpchecker.sh
```

## Further Reading

- [mcpchecker Documentation](https://github.com/mcpchecker/mcpchecker)
- [StackRox MCP Server](../README.md)
