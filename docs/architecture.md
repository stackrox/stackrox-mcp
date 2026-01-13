# StackRox MCP Architecture

## Table of Contents

1. [High-Level Architecture](#high-level-architecture)
2. [Component Descriptions](#component-descriptions)
3. [Data Flow Diagrams](#data-flow-diagrams)
4. [Token Passthrough Flow](#token-passthrough-flow)
5. [Error Handling Strategy](#error-handling-strategy)
6. [Tool API Endpoints](#tool-api-endpoints)

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────────┐
│                           MCP Client                                    │
│                    (Claude Code CLI, Desktop, etc.)                     │
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

## Component Descriptions

### 1. MCP Server (`internal/server`)

**Purpose**: Hosts the Model Context Protocol server and handles client connections.

**Key Responsibilities**:

- Serves MCP over HTTP with Server-Sent Events (SSE) or stdio transport
- Routes tool calls to registered toolsets
- Provides health check endpoint (`/health`)
- Manages graceful shutdown with 5-second timeout

**Implementation Details**:

- Uses `modelcontextprotocol/go-sdk/mcp` for MCP protocol
- Supports two transport modes:
  - `streamable-http`: HTTP server with SSE for streaming responses
  - `stdio`: Standard input/output (requires static auth)
- Filters tools based on `read_only_tools` configuration
- Located at: `internal/server/server.go:1`

### 2. Toolsets Registry (`internal/toolsets`)

**Purpose**: Central registry for all available toolsets and their tools.

**Key Responsibilities**:

- Manages toolset registration and lifecycle
- Applies global read-only filtering
- Provides unified tool discovery

**Toolset Interface**:

```go
type Toolset interface {
    GetName() string
    IsEnabled() bool
    GetTools() []Tool
}
```

**Tool Interface**:

```go
type Tool interface {
    IsReadOnly() bool
    GetTool() *mcp.Tool
    GetName() string
    RegisterWith(server *mcp.Server)
}
```

**Available Toolsets**:

1. **Vulnerability Toolset** (`internal/toolsets/vulnerability`):
   - `get_deployments_for_cve`: Query deployments affected by CVE
   - `get_nodes_for_cve`: Query nodes affected by CVE (streaming aggregation)
   - `get_clusters_for_cve`: Query clusters affected by CVE

2. **Config Manager Toolset** (`internal/toolsets/config`):
   - `list_clusters`: List all clusters with pagination

### 3. StackRox Client (`internal/client`)

**Purpose**: Manages gRPC connection to StackRox Central API.

**Key Responsibilities**:

- Establishes and maintains gRPC connection
- Handles authentication (static or passthrough)
- Manages connection lifecycle and reconnection
- Applies interceptors for logging and retry

**Connection Management**:

- Lazy connection with automatic reconnection on failure states
- Connection state monitoring (TransientFailure, Shutdown)
- Support for both HTTP/2 (native gRPC) and HTTP/1 bridge
- Located at: `internal/client/client.go:1`

**HTTP/1 Bridge Support**:
When `force_http1` is enabled, the client uses the StackRox gRPC-over-HTTP/1 bridge to downgrade requests for environments that block HTTP/2:

- Uses `golang.stackrox.io/grpc-http1/client`
- Enables ALPN negotiation for pure gRPC
- Required for HTTP/1-only proxies or load balancers
- Note: Client-side streaming remains unsupported in downgrade mode

### 4. Authentication (`internal/client/auth`)

**Purpose**: Handles API token authentication for StackRox Central.

**Authentication Modes**:

#### Passthrough Authentication

- Token extracted from MCP request headers (`Authorization: Bearer <token>`)
- Token passed through to StackRox Central on each API call
- Enables per-user authentication when MCP server is shared
- Implementation: `internal/client/auth/passthrough.go`

#### Static Authentication

- Token configured in server configuration
- Single token used for all API calls
- Required for stdio transport mode
- Implementation: `internal/client/auth/static.go`

**Implementation**:
Both modes implement `credentials.PerRPCCredentials`:

```go
type PerRPCCredentials interface {
    GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error)
    RequireTransportSecurity() bool
}
```

### 5. Configuration (`internal/config`)

**Purpose**: Centralized configuration management with validation.

**Configuration Sources** (in precedence order):

1. Default values
2. YAML configuration file (via `--config` flag)
3. Environment variables (prefix: `STACKROX_MCP__`)

**Key Configuration Sections**:

- `central`: StackRox Central connection settings
- `global`: Server-wide settings (e.g., read-only mode)
- `server`: HTTP server configuration
- `tools`: Individual toolset enable/disable flags

**Validation**: Configuration is validated at startup with detailed error messages.

### 6. Interceptors (`internal/client/interceptors.go`)

**Purpose**: gRPC interceptors for cross-cutting concerns.

**Interceptors**:

1. **Logging Interceptor**:
   - Logs request start/completion with duration
   - Logs errors with method and duration
   - Debug-level logging for successful requests

2. **Retry Interceptor**:
   - Wraps each RPC call with retry logic
   - Creates per-attempt timeout contexts
   - Implements exponential backoff
   - Only retries retriable errors
   - Located at: `internal/client/interceptors.go:11`

### 7. Error Handling (`internal/client/error.go`)

**Purpose**: Unified error handling with user-friendly messages.

**Error Structure**:

```go
type Error struct {
    Code        codes.Code  // gRPC status code
    Retriable   bool        // Whether error should be retried
    Message     string      // Human-readable error message
    OriginalErr error       // Original gRPC error
    Operation   string      // Operation that failed
}
```

**Retriable Errors**:

- `Unavailable`: Service temporarily unavailable
- `DeadlineExceeded`: Request timeout

**Non-Retriable Errors**:

- `Unauthenticated`: Invalid/expired token
- `PermissionDenied`: Insufficient permissions
- `NotFound`: Resource not found
- `InvalidArgument`: Bad request parameters

Located at: `internal/client/error.go`

### 8. Retry Policy (`internal/client/retry.go`)

**Purpose**: Exponential backoff retry strategy.

**Configuration**:

- `max_retries`: Maximum retry attempts (0-10, default: 3)
- `initial_backoff`: Initial backoff duration (default: 1s)
- `max_backoff`: Maximum backoff duration (default: 10s)
- Backoff multiplier: 2.0 (exponential)

**Backoff Calculation**:

```
backoff = min(initial_backoff * 2^attempt, max_backoff)
```

Located at: `internal/client/retry.go`

## Data Flow Diagrams

### Tool Invocation Flow

```
┌─────────────┐
│ MCP Client  │
│ (Claude)    │
└──────┬──────┘
       │ 1. HTTP POST /
       │    Authorization: Bearer <token>
       │    {name: "get_deployments_for_cve", arguments: {...}}
       ▼
┌─────────────────────┐
│   MCP Server        │
│   (HTTP Handler)    │
└──────┬──────────────┘
       │ 2. Route to tool handler
       ▼
┌─────────────────────┐
│ Tool Handler        │
│ (e.g., get_         │
│  deployments_for_   │
│  cve)               │
└──────┬──────────────┘
       │ 3. Create context with MCP request
       │    ctx = auth.WithMCPRequestContext(ctx, req)
       ▼
┌─────────────────────┐
│  StackRox Client    │
│  (ReadyConn)        │
└──────┬──────────────┘
       │ 4. Extract token from context
       │    (passthrough mode only)
       ▼
┌─────────────────────┐
│ gRPC Interceptors   │
│ (Logging, Retry)    │
└──────┬──────────────┘
       │ 5. gRPC call with retry
       │    Authorization: Bearer <token>
       ▼
┌─────────────────────┐
│  StackRox Central   │
│  (gRPC API)         │
└──────┬──────────────┘
       │ 6. Return response
       ▼
┌─────────────────────┐
│ Tool Handler        │
│ (Process response)  │
└──────┬──────────────┘
       │ 7. Format output
       ▼
┌─────────────────────┐
│   MCP Server        │
│   (Send response)   │
└──────┬──────────────┘
       │ 8. HTTP 200 + SSE
       │    {content: [...], isError: false}
       ▼
┌─────────────┐
│ MCP Client  │
└─────────────┘
```

### Retry Flow

```
┌──────────────┐
│ gRPC Request │
└──────┬───────┘
       │
       ▼
┌────────────────────────┐
│  Retry Interceptor     │
│  (Attempt 1)           │
└──────┬─────────────────┘
       │
       ▼
┌────────────────────────┐
│  Create timeout ctx    │
│  (30s default)         │
└──────┬─────────────────┘
       │
       ▼
┌────────────────────────┐
│  Invoke gRPC call      │
└──────┬─────────────────┘
       │
       ├─── Success ───┐
       │               ▼
       │          ┌─────────┐
       │          │ Return  │
       │          └─────────┘
       │
       ├─── Non-retriable error ───┐
       │                            ▼
       │                       ┌─────────┐
       │                       │ Return  │
       │                       │ Error   │
       │                       └─────────┘
       │
       └─── Retriable error (Unavailable/DeadlineExceeded)
                   │
                   ▼
            ┌──────────────────┐
            │ Check if should  │
            │ retry (attempt < │
            │ max_retries)     │
            └──────┬───────────┘
                   │
                   ├─── No more retries ───┐
                   │                       ▼
                   │                   ┌─────────┐
                   │                   │ Return  │
                   │                   │ Error   │
                   │                   └─────────┘
                   │
                   └─── Can retry
                           │
                           ▼
                    ┌──────────────────┐
                    │ Calculate backoff│
                    │ initial * 2^n    │
                    └──────┬───────────┘
                           │
                           ▼
                    ┌──────────────────┐
                    │ Wait backoff     │
                    │ duration         │
                    └──────┬───────────┘
                           │
                           ▼
                    ┌──────────────────┐
                    │ Retry attempt    │
                    │ (Attempt 2, 3...)│
                    └──────────────────┘
```

## Token Passthrough Flow

### Passthrough Authentication Mode

```
┌─────────────────────────────────────────────────────────────────────┐
│                         MCP Client                                  │
│  • User provides API token via Claude Code CLI config               │
│  • Token included in Authorization header                           │
└───────────────────────────┬─────────────────────────────────────────┘
                            │
                            │ HTTP POST /
                            │ Authorization: Bearer eyJ...token...
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────────┐
│                      MCP Server (HTTP Handler)                      │
│  • Receives HTTP request with Authorization header                  │
│  • MCP SDK extracts headers into CallToolRequest.Extra.Header       │
└───────────────────────────┬─────────────────────────────────────────┘
                            │
                            │ CallToolRequest
                            │ {
                            │   name: "tool_name",
                            │   arguments: {...},
                            │   _meta: {
                            │     extra: {
                            │       header: {
                            │         "Authorization": ["Bearer eyJ..."]
                            │       }
                            │     }
                            │   }
                            │ }
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────────┐
│                         Tool Handler                                │
│  Step 1: Store MCP request in context                               │
│    callCtx = auth.WithMCPRequestContext(ctx, req)                   │
│                                                                     │
│  Located at: internal/client/auth/auth.go:17                        │
└───────────────────────────┬─────────────────────────────────────────┘
                            │
                            │ Context with MCP request
                            │
                            ▼
┌──────────────────────────────────────────────────────────────────────┐
│              StackRox Client (gRPC Connection)                       │
│  Step 2: gRPC dial options include passthroughTokenCredentials       │
│    grpc.WithPerRPCCredentials(auth.NewPassthroughTokenCredentials()) │
│                                                                      │
│  Located at: internal/client/client.go:212                           │
└───────────────────────────┬──────────────────────────────────────────┘
                            │
                            │ Before each gRPC call
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────────┐
│          passthroughTokenCredentials.GetRequestMetadata()           │
│  Step 3: Extract token from context                                 │
│    1. Get MCP request from context                                  │
│       mcpReq := mcpRequestFromContext(ctx)                          │
│    2. Extract Authorization header                                  │
│       header := mcpReq.GetExtra().Header                            │
│       authHeader := header.Get("Authorization")                     │
│    3. Parse Bearer token                                            │
│       token := strings.TrimPrefix(authHeader, "Bearer ")            │
│    4. Return gRPC metadata                                          │
│       return map[string]string{                                     │
│         "authorization": "Bearer " + token                          │
│       }                                                             │
│                                                                     │
│  Located at: internal/client/auth/passthrough.go:74                 │
└───────────────────────────┬─────────────────────────────────────────┘
                            │
                            │ gRPC request with metadata
                            │ metadata: {authorization: "Bearer eyJ..."}
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────────┐
│                      StackRox Central API                           │
│  • Validates Bearer token                                           │
│  • Executes API call with user's permissions                        │
│  • Returns response or error (Unauthenticated, PermissionDenied)    │
└─────────────────────────────────────────────────────────────────────┘
```

### Static Authentication Mode

```
┌─────────────────────────────────────────────────────────────────────┐
│                    MCP Server Configuration                         │
│  • API token configured at server startup                           │
│  • Token stored in CentralConfig.APIToken                           │
│  • MCP client does NOT provide token                                │
└───────────────────────────┬─────────────────────────────────────────┘
                            │
                            │ Configuration
                            │ central.auth_type = "static"
                            │ central.api_token = "eyJ...token..."
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────────┐
│              StackRox Client (gRPC Connection)                      │
│  Step 1: gRPC dial options include staticTokenCredentials           │
│    grpc.WithPerRPCCredentials(                                      │
│      auth.NewStaticTokenCredentials(config.APIToken)                │
│    )                                                                │
│                                                                     │
│  Located at: internal/client/client.go:215                          │
└───────────────────────────┬─────────────────────────────────────────┘
                            │
                            │ Before each gRPC call
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────────┐
│             staticTokenCredentials.GetRequestMetadata()             │
│  Step 2: Return configured token                                    │
│    return map[string]string{                                        │
│      "authorization": "Bearer " + t.token                           │
│    }                                                                │
│                                                                     │
│  Located at: internal/client/auth/static.go:26                      │
└───────────────────────────┬─────────────────────────────────────────┘
                            │
                            │ gRPC request with metadata
                            │ metadata: {authorization: "Bearer eyJ..."}
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────────┐
│                      StackRox Central API                           │
│  • Validates Bearer token                                           │
│  • Executes API call with configured user's permissions             │
│  • Returns response or error                                        │
└─────────────────────────────────────────────────────────────────────┘
```

### Token Validation Error Flow

```
┌────────────────────┐
│ MCP Client sends   │
│ invalid/expired    │
│ token              │
└─────────┬──────────┘
          │
          ▼
┌────────────────────┐
│ StackRox Central   │
│ rejects token      │
└─────────┬──────────┘
          │
          │ gRPC error
          │ code: Unauthenticated
          │ message: "invalid token"
          │
          ▼
┌────────────────────┐
│ Retry Interceptor  │
│ checks if retriable│
└─────────┬──────────┘
          │
          │ IsRetriableGRPCError(err) = false
          │ (Unauthenticated is not retriable)
          │
          ▼
┌────────────────────┐
│ Error Handler      │
│ formats message    │
└─────────┬──────────┘
          │
          │ Error{
          │   Code: Unauthenticated,
          │   Retriable: false,
          │   Message: "Operation 'ListDeployments' failed:
          │            Authentication failed - invalid or
          │            expired API token. Please check your
          │            configuration."
          │ }
          │
          ▼
┌────────────────────┐
│ Return to MCP      │
│ Client with        │
│ user-friendly msg  │
└────────────────────┘
```

## Error Handling Strategy

### Error Classification

The system categorizes errors into two main types:

1. **Retriable Errors** (transient failures):
   - `Unavailable`: Service temporarily unavailable
   - `DeadlineExceeded`: Request timeout
   - Automatically retried with exponential backoff

2. **Non-Retriable Errors** (permanent failures):
   - `Unauthenticated`: Invalid/expired API token
   - `PermissionDenied`: Insufficient permissions
   - `NotFound`: Resource not found
   - `InvalidArgument`: Bad request parameters
   - Returned immediately to client

### Error Handling Flow

```
┌──────────────────┐
│   API Call       │
└────────┬─────────┘
         │
         ▼
┌────────────────────────────┐
│   gRPC Call via            │
│   Retry Interceptor        │
└────────┬───────────────────┘
         │
         ├─── Success ──────────────────────┐
         │                                  │
         ├─── gRPC Error                    │
         │         │                        │
         │         ▼                        │
         │    ┌─────────────────────┐       │
         │    │ client.NewError()   │       │
         │    │ - Extract gRPC code │       │
         │    │ - Map to message    │       │
         │    │ - Set retriable flag│       │
         │    └──────┬──────────────┘       │
         │           │                      │
         │           ▼                      │
         │    ┌─────────────────────┐       │
         │    │ IsRetriableGRPCError│       │
         │    │ (Unavailable or     │       │
         │    │  DeadlineExceeded)  │       │
         │    └──────┬──────────────┘       │
         │           │                      │
         │           ├── Yes (Retriable)    │
         │           │   │                  │
         │           │   ▼                  │
         │           │ ┌──────────────┐     │
         │           │ │ Backoff wait │     │
         │           │ └──────┬───────┘     │
         │           │        │             │
         │           │        ▼             │
         │           │ ┌──────────────┐     │
         │           │ │ Retry call   │     │
         │           │ │ (if attempts │     │
         │           │ │  remaining)  │     │
         │           │ └──────────────┘     │
         │           │                      │
         │           └── No (Non-retriable) │
         │                   │              │
         │                   ▼              │
         └───────────────► Return Error ────┘
                     (with formatted message)
```

### Error Message Formatting

Located at: `internal/client/error.go:77`

Each gRPC status code is mapped to a user-friendly, actionable message:


| gRPC Code          | Message Template                                                        | Example                                                                                                                                                |
| ------------------ | ----------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `Unauthenticated`  | Authentication failed - invalid or expired API token                    | Operation 'ListDeployments' failed: Authentication failed - invalid or expired API token. Please check your configuration.                             |
| `PermissionDenied` | Permission denied - your API token does not have sufficient permissions | Operation 'ListDeployments' failed: Permission denied - your API token does not have sufficient permissions for this operation.                        |
| `NotFound`         | Resource not found - the requested resource does not exist              | Operation 'GetDeployment' failed: Resource not found - the requested resource does not exist.                                                          |
| `InvalidArgument`  | Invalid argument - the request contains invalid parameters              | Operation 'ListDeployments' failed: Invalid argument - the request contains invalid parameters.                                                        |
| `Unavailable`      | Service temporarily unavailable (auto-retry)                            | Operation 'ListDeployments' failed: StackRox Central is temporarily unavailable. The request will be retried automatically.                            |
| `DeadlineExceeded` | Request timed out (auto-retry)                                          | Operation 'ListDeployments' failed: Request timed out after 30 seconds. StackRox Central may be overloaded. The request will be retried automatically. |

### Error Handling Best Practices

1. **Contextual Information**: Each error includes the operation name for debugging
2. **User Guidance**: Messages suggest corrective actions (e.g., "check your configuration")
3. **Retry Transparency**: Auto-retry messages inform users about automatic retry
4. **Original Error Preservation**: Original gRPC error is preserved for debugging
5. **Logging**: Failed requests are logged with method, duration, and error details

### Timeout Configuration

Default request timeout: 30 seconds (configurable via `central.request_timeout`)

Timeout applies to:

- Each individual retry attempt (not cumulative)
- gRPC call execution
- Context deadline enforcement

## Tool API Endpoints

This section documents each MCP tool, the StackRox Central API endpoints it uses, and how query parameters are utilized.

### Vulnerability Toolset

#### 1. get_deployments_for_cve

**Purpose**: Retrieve deployments affected by a specific CVE.

**Tool Parameters**:

- `cveName` (required): CVE identifier (e.g., "CVE-2021-44228")
- `filterClusterId` (optional): Filter by cluster ID
- `filterNamespace` (optional): Filter by namespace
- `filterPlatform` (optional): Filter by platform type
  - `NO_FILTER`: All deployments (default)
  - `USER_WORKLOAD`: User workload deployments only
  - `PLATFORM`: Platform deployments only
- `includeAffectedImages` (optional): Include affected image names (default: false)
- `cursor` (optional): Pagination cursor

**API Endpoint**: `DeploymentService.ListDeployments` - `/v1/deployments`

**gRPC Service**: `v1.DeploymentServiceClient`

**Request Construction**:

```go
// Build query string
query := "CVE:\"CVE-2021-44228\"+Cluster ID:\"cluster-123\"+Namespace:\"default\""

// Create request
req := &v1.RawQuery{
    Query: query,
    Pagination: &v1.Pagination{
        Offset: cursor.GetOffset(),
        Limit:  100 + 1,  // Fetch one extra to detect more pages
    },
}

// Execute
deploymentClient.ListDeployments(ctx, req)
```

**Query String Format**:

- Uses StackRox query syntax with `+` as separator
- Values are quoted for exact matching: `CVE:"CVE-2021-44228"`
- Platform filter: `Platform Component:0` (user workload) or `Platform Component:1` (platform)

**Pagination**:

- Uses offset-based pagination with custom cursor encoding
- Cursor encodes offset as base64-encoded JSON
- Fetches `limit + 1` items to detect if more pages exist
- Returns `nextCursor` in response if more pages available

**Image Enrichment** (when `includeAffectedImages=true`):

After fetching deployments, for each deployment:

- API Endpoint: `ImageService.ListImages`
- gRPC Service: `v1.ImageServiceClient`
- Query: `CVE:"CVE-2021-44228"+Deployment ID:"deployment-123"`
- Concurrent fetching with semaphore (max 10 concurrent requests)

**Implementation**: `internal/toolsets/vulnerability/deployments.go:257`

**Query Building**: `internal/toolsets/vulnerability/deployments.go:142`

#### 2. get_nodes_for_cve

**Purpose**: Retrieve aggregated node groups affected by a CVE, grouped by cluster and OS image.

**Tool Parameters**:

- `cveName` (required): CVE identifier (e.g., "CVE-2020-26159")
- `filterClusterId` (optional): Filter by cluster ID

**API Endpoint**: `NodeService.ExportNodes` (streaming) - `/v1/export/nodes`

**gRPC Service**: `v1.NodeServiceClient`

**Request Construction**:

```go
// Build query string
query := "CVE:\"CVE-2020-26159\"+Cluster ID:\"cluster-123\""

// Create request
req := &v1.ExportNodeRequest{
    Query: query,
}

// Execute streaming call
stream, _ := nodeClient.ExportNodes(ctx, req)
```

**Query String Format**:

- Uses StackRox query syntax with `+` as separator
- Values are quoted for exact matching

**Streaming Aggregation**:

- Consumes entire gRPC stream (server-streaming RPC)
- Aggregates nodes by cluster ID and OS image
- Groups nodes using map key: `"clusterId|osImage"`
- Sorts results by cluster ID, then OS image for deterministic output

**Aggregation Logic**:

```go
// Map key: "clusterId|osImage"
groups := make(map[string]*NodeGroupResult)

for {
    resp, err := stream.Recv()
    if err == io.EOF {
        break  // Stream ended
    }

    node := resp.GetNode()
    key := fmt.Sprintf("%s|%s", node.GetClusterId(), node.GetOsImage())

    if group, exists := groups[key]; exists {
        group.Count++  // Increment count for existing group
    } else {
        // Create new group
        groups[key] = &NodeGroupResult{
            ClusterID:       node.GetClusterId(),
            ClusterName:     node.GetClusterName(),
            OperatingSystem: node.GetOsImage(),
            Count:           1,
        }
    }
}
```

**Implementation**: `internal/toolsets/vulnerability/nodes.go:175`

**Aggregation Logic**: `internal/toolsets/vulnerability/nodes.go:116`

#### 3. get_clusters_for_cve

**Purpose**: Retrieve clusters affected by a specific CVE.

**Tool Parameters**:

- `cveName` (required): CVE identifier (e.g., "CVE-2021-44228")
- `filterClusterId` (optional): Verify if specific cluster is affected

**API Endpoint**: `ClustersService.GetClusters` - `/v1/clusters`

**gRPC Service**: `v1.ClustersServiceClient`

**Request Construction**:

```go
// Build query string
query := "CVE:\"CVE-2021-44228\"+Cluster ID:\"cluster-123\""

// Create request
req := &v1.GetClustersRequest{
    Query: query,
}

// Execute
clustersClient.GetClusters(ctx, req)
```

**Query String Format**:

- Uses StackRox query syntax with `+` as separator
- Values are quoted for exact matching

**Sorting**:

- Results sorted by cluster ID for deterministic output

**Implementation**: `internal/toolsets/vulnerability/clusters.go:113`

**Query Building**: `internal/toolsets/vulnerability/clusters.go:100`

### Config Manager Toolset

#### 4. list_clusters

**Purpose**: List all clusters managed by StackRox.

**Tool Parameters**:

- `offset` (optional): Starting index for pagination (default: 0)
- `limit` (optional): Maximum clusters to return (default: 0 = unlimited)

**API Endpoint**: `ClustersService.GetClusters` - `/v1/clusters`

**gRPC Service**: `v1.ClustersServiceClient`

**Request Construction**:

```go
// Create request (no query, fetch all clusters)
req := &v1.GetClustersRequest{}

// Execute
clustersClient.GetClusters(ctx, req)
```

**Pagination**:

- Client-side pagination (API returns all clusters)
- Offset and limit applied after fetching all clusters
- Returns total count for client pagination UX

**Response**:

```go
{
    "clusters": [...],      // Paginated subset
    "totalCount": 50,       // Total clusters available
    "offset": 0,            // Requested offset
    "limit": 10             // Requested limit
}
```

**Implementation**: `internal/toolsets/config/tools.go:136`

**Cluster Fetching**: `internal/toolsets/config/tools.go:101`

### Query Syntax Summary

StackRox query syntax used across all tools:

| Element         | Format                   | Example                                |
| --------------- | ------------------------ | -------------------------------------- |
| Field filter    | `Field:value`            | `CVE:CVE-2021-44228`                   |
| Exact match     | `Field:"value"`          | `CVE:"CVE-2021-44228"`                 |
| AND operator    | `+`                      | `CVE:"CVE-2021"+Namespace:"default"`   |
| Platform filter | `Platform Component:0/1` | `Platform Component:0` (user workload) |

**Quote Protection**: All CVE names and IDs are quoted to prevent partial matching:

- Without quotes: `CVE:CVE-2025-10` would match `CVE-2025-101`
- With quotes: `CVE:"CVE-2025-10"` matches exactly

### API Error Handling

All tools follow the same error handling pattern:

1. **Connection Errors**:
   - Wrapped with operation context: `"unable to connect to server"`
   - Includes original error for debugging

2. **API Errors**:
   - Converted using `client.NewError(err, "OperationName")`
   - Maps gRPC codes to user-friendly messages
   - Preserves original error for logging

3. **Validation Errors**:
   - Checked before API call
   - Returns immediately without API invocation
   - Example: Empty CVE name

### Performance Considerations

1. **Deployment Image Fetching**:
   - Disabled by default (`includeAffectedImages=false`)
   - When enabled, uses semaphore to limit concurrent requests (max 10)
   - Can significantly increase response time for large deployments

2. **Node Aggregation**:
   - Streams all nodes before returning (no pagination)
   - Memory usage grows with number of affected nodes
   - Aggregation reduces response size (groups vs. individual nodes)

3. **Cluster Listing**:
   - Fetches all clusters (no server-side pagination)
   - Client-side pagination applied after fetch
   - Suitable for typical StackRox deployments (10-1000 clusters)
