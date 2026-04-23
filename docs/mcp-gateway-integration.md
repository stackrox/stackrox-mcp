# MCP Gateway Integration

This guide describes how to integrate the StackRox MCP server with [MCP Gateway](https://github.com/Kuadrant/mcp-gateway), enabling tool aggregation behind a centralized gateway endpoint.

## Overview

MCP Gateway is an Envoy-based system that aggregates multiple MCP servers behind a single endpoint. It uses Gateway API resources and a custom `MCPServerRegistration` CRD to discover and route requests to backend MCP servers.

When enabled, the Helm chart creates:
- **HTTPRoute** — routes `/mcp` traffic from the gateway to the StackRox MCP service
- **MCPServerRegistration** — registers the server with the gateway using a tool prefix to avoid naming conflicts

## Prerequisites

- [MCP Gateway](https://github.com/Kuadrant/mcp-gateway) installed on the cluster ([OpenShift installation guide](https://github.com/Kuadrant/mcp-gateway/tree/main/config/openshift))

The Helm chart validates that the required CRDs are available and fails with a descriptive error if they are not.

## Installation

```bash
helm install stackrox-mcp charts/stackrox-mcp \
  --namespace stackrox-mcp \
  --create-namespace \
  --set mcpGateway.enabled=true \
  --set mcpGateway.hostname=stackrox-mcp.mcp.local \
  --set config.central.url=<your-central-url>
```

## Configuration Reference

| Parameter | Description | Default |
|-----------|-------------|---------|
| `mcpGateway.enabled` | Enable MCP Gateway integration | `false` |
| `mcpGateway.gateway.name` | Name of the MCP Gateway resource | `stackrox-mcp-gateway` |
| `mcpGateway.gateway.namespace` | Namespace of the MCP Gateway resource | `gateway-system` |
| `mcpGateway.hostname` | Hostname for the HTTPRoute | `""` |
| `mcpGateway.toolPrefix` | Prefix for tools exposed via the gateway | `stackrox_` |

## Verification

```bash
curl -X POST https://$(oc get routes -n gateway-system -o jsonpath='{ .items[0].spec.host }')/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"initialize","id":1,"params":{"protocolVersion":"2025-03-26","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}'
```

## Architecture

```
MCP Client
  → MCP Gateway (gateway-system namespace)
    → HTTPRoute (/mcp)
      → StackRox MCP Service (stackrox-mcp namespace)
        → StackRox MCP Pod(s)
          → StackRox Central API
```

The gateway aggregates tools from all registered MCP servers. StackRox MCP tools are exposed with the configured prefix (e.g., `stackrox_get_deployments_for_cve`).
