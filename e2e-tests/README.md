# StackRox MCP E2E Testing

End-to-end tests for the StackRox MCP server using [gevals](https://github.com/genmcp/gevals).

## Prerequisites

- Go 1.25+
- OpenAI API Key (for AI agent and LLM judge)
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
OPENAI_API_KEY=sk-your-key-here
STACKROX_API_TOKEN=your-token-here
```

## Running Tests

```bash
./scripts/run-tests.sh
```

Results are saved to `gevals-stackrox-mcp-e2e-out.json`.

### View Results

```bash
# Summary
jq '.tasks[] | {name, passed}' gevals-stackrox-mcp-e2e-out.json

# Tool calls
jq '.tasks[].callHistory[] | {toolName, arguments}' gevals-stackrox-mcp-e2e-out.json
```

## Test Cases

| Test | Description | Tool |
|------|-------------|------|
| `list-clusters` | List all clusters | `list_clusters` |
| `cve-affecting-workloads` | CVE impact on deployments | `get_deployments_for_cve` |
| `cve-affecting-clusters` | CVE impact on clusters | `get_clusters_for_cve` |
| `cve-nonexistent` | Handle non-existent CVE | `get_clusters_for_cve` |
| `cve-cluster-scooby` | CVE with cluster filter | `get_clusters_for_cve` |
| `cve-cluster-maria` | CVE with cluster filter | `get_clusters_for_cve` |
| `cve-clusters-general` | General CVE query | `get_clusters_for_cve` |
| `cve-cluster-list` | CVE across clusters | `get_clusters_for_cve` |

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
