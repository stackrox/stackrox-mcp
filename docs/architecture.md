# StackRox MCP Architecture

## Overview

StackRox MCP Server is a Model Context Protocol (MCP) server that exposes StackRox Central's security capabilities through a standardized interface. It enables AI assistants to query vulnerability data.

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────────┐
│                           MCP Client                                    │
│                    (Claude Code, goose, etc.)                           │
└───────────────┬─────────────────────────────────────────────────────────┘
                │ HTTP/SSE or stdio
                │ (includes Authorization header)
                ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                        StackRox MCP Server                              │
│  ┌────────────────────────────────────────────────────────────────────┐ │
│  │                          MCP Server                                │ │
│  │              (go-sdk/mcp.Server with HTTP/stdio transport)         │ │
│  └────────────┬───────────────────────────────────────────────────────┘ │
│               │                                                         │
│               ▼                                                         │
│  ┌────────────────────────────────────────────────────────────────────┐ │
│  │                      Toolsets Registry                             │ │
│  │  ┌──────────────────┐      ┌──────────────────┐                    │ │
│  │  │   Vulnerability  │      │  Config Manager  │                    │ │
│  │  │     Toolset      │      │     Toolset      │                    │ │
│  │  └──────────────────┘      └──────────────────┘                    │ │
│  └────────────┬───────────────────────────────────────────────────────┘ │
│               │                                                         │
│               ▼                                                         │
│  ┌────────────────────────────────────────────────────────────────────┐ │
│  │                      StackRox Client                               │ │
│  │  ┌──────────────┐  ┌─────────────┐  ┌──────────────────┐           │ │
│  │  │ Auth Handler │  │ Interceptors│  │  Retry Policy    │           │ │
│  │  │(passthrough/ │  │(logging/    │  │(exponential      │           │ │
│  │  │   static)    │  │ retry)      │  │ backoff)         │           │ │
│  │  └──────────────┘  └─────────────┘  └──────────────────┘           │ │
│  └────────────┬───────────────────────────────────────────────────────┘ │
└───────────────┼─────────────────────────────────────────────────────────┘
                │ gRPC (HTTP/2 or HTTP/1 bridge)
                │ TLS with Bearer token
                ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                       StackRox Central                                  │
│  ┌────────────────────────────────────────────────────────────────────┐ │
│  │                        gRPC API Services                           │ │
│  │  • DeploymentService    • ImageService                             │ │
│  │  • NodeService          • ClustersService                          │ │
│  └────────────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────┘
```

## Core Components

### MCP Server

The MCP server handles client connections and routes tool invocations to the appropriate toolsets.

**Responsibilities**:
- Serves MCP protocol over HTTP with Stream-HTTP or stdio transport
- Routes tool calls to registered toolsets
- Provides health check endpoint
- Manages graceful shutdown

**Transport Modes**:
- **Stream-HTTP**: Streaming responses over HTTP, supports both auth modes
- **stdio**: Standard input/output, requires static authentication

### Toolsets Registry

Central registry that manages all available toolsets and their tools.

**Responsibilities**:
- Manages toolset registration
- Applies global read-only filtering when configured
- Provides unified tool discovery

**Available Toolsets**:

1. **Vulnerability Toolset**: Query resources where CVEs are detected
   - `get_deployments_for_cve`: Find deployments where CVE is detected
   - `get_nodes_for_cve`: Find nodes where CVE is detected (aggregated by cluster and OS)
   - `get_clusters_with_orchestrator_cve`: Find clusters where CVE is detected in orchestrator components

2. **Config Manager Toolset**: Manage cluster configurations
   - `list_clusters`: List all managed clusters with pagination

### StackRox Client

Manages the gRPC connection to StackRox Central API.

**Responsibilities**:
- Establishes and maintains gRPC connections
- Handles authentication (static or passthrough)
- Applies interceptors for logging and retry
- Manages connection lifecycle and automatic reconnection

**Connection Features**:
- Lazy connection initialization
- Automatic reconnection on transient failures
- Support for both HTTP/2 (native gRPC) and HTTP/1 bridge mode
- Configurable request timeouts (default: 30 seconds)

### Authentication

Two authentication modes are supported:

**Passthrough Authentication**:
- Token extracted from incoming MCP request headers
- Enables per-user authentication when MCP server is shared
- Token passed directly to StackRox Central for each API call
- Supports multi-tenant deployments

**Static Authentication**:
- Single API token configured at server startup
- All API calls use the same credentials
- Required for stdio transport mode
- Simpler setup for single-user scenarios

### Configuration

Centralized configuration with multiple sources (in precedence order):
1. Default values
2. YAML configuration file
3. Environment variables (prefix: `STACKROX_MCP__`)

**Key Configuration Areas**:
- `central`: StackRox Central connection settings (endpoint, auth, TLS)
- `global`: Server-wide settings (read-only mode)
- `server`: HTTP server configuration (port, timeouts)
- `tools`: Individual toolset enable/disable flags

## Request Flow

```
MCP Client
    │
    ├─> 1. HTTP POST with Authorization header
    │
    ▼
MCP Server
    │
    ├─> 2. Route to tool handler
    │
    ▼
Tool Handler
    │
    ├─> 3. Store MCP request in context
    │
    ▼
StackRox Client
    │
    ├─> 4. Extract token (passthrough) or use static token
    │
    ▼
gRPC Interceptors
    │
    ├─> 5. Apply logging and retry logic
    │
    ▼
StackRox Central API
    │
    ├─> 6. Process request and return response
    │
    ▼
Tool Handler
    │
    ├─> 7. Format response for MCP
    │
    ▼
MCP Client
```

## Error Handling

The system implements intelligent error handling with retry logic for transient failures.

### Error Classification

**Retriable Errors** (automatically retried with exponential backoff):
- `Unavailable`: Service temporarily unavailable
- `DeadlineExceeded`: Request timeout

**Non-Retriable Errors** (returned immediately):
- `Unauthenticated`: Invalid or expired API token
- `PermissionDenied`: Insufficient permissions
- `NotFound`: Resource not found
- `InvalidArgument`: Bad request parameters

### Retry Strategy

- Maximum retries: 3 (configurable)
- Exponential backoff: starts at 1s, doubles each attempt, capped at 10s
- Timeout per attempt: 30 seconds (configurable)
- Only retriable errors trigger retry logic

### Error Messages

All errors are converted to user-friendly messages with:
- Clear description of what went wrong
- Actionable guidance for resolution
- Context about the failed operation
- Transparency about automatic retries

## Available Tools

### Vulnerability Tools

**get_deployments_for_cve**
- Query deployments where CVE is detected
- Optional filters: cluster, namespace, platform type
- Optional image enrichment (lists container images where CVE is detected)
- Pagination support for large result sets

**get_nodes_for_cve**
- Query nodes where CVE is detected
- Results aggregated by cluster and OS image
- Optional cluster filter
- Streaming API for efficient processing

**get_clusters_with_orchestrator_cve**
- Query clusters where CVE is detected for orchestrator components
- Optional cluster filter for verification
- Sorted results for deterministic output

### Config Management Tools

**list_clusters**
- List all clusters managed by StackRox
- Client-side pagination support
- Returns cluster metadata and status

## Query Syntax

All vulnerability tools use StackRox query syntax:

- **Field filters**: `CVE:"CVE-2021-44228"`
- **Multiple conditions**: `CVE:"CVE-2021"+Namespace:"default"`
- **Exact matching**: Values quoted to prevent partial matches
- **Platform filters**: `Platform Component:0` (user workload) or `Platform Component:1` (platform)

## Performance Considerations

**Deployment Image Enrichment**:
- Disabled by default for faster response times
- When enabled, uses concurrent requests with semaphore limiting
- Can significantly increase response time for large deployments

**Node Aggregation**:
- Streams all nodes before aggregating and returning results
- Groups nodes by cluster and OS for reduced response size
- Memory usage scales with number of nodes

**Cluster Listing**:
- Fetches all clusters from API
- Applies client-side pagination
- Optimized for typical deployments (10-1000 clusters)
