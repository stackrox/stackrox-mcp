# StackRox MCP

## Project Overview

StackRox MCP is a Model Context Protocol (MCP) server that provides AI assistants with access to StackRox.

## Quick Start

Clone the repository:
```bash
git clone https://github.com/stackrox/stackrox-mcp.git
cd stackrox-mcp
```

Build the project:
```bash
make build
```

Run the server:
```bash
# With configuration file
./stackrox-mcp --config=examples/config-read-only.yaml

# Or using environment variables only
export STACKROX_MCP__CENTRAL__URL=central.stackrox:443
export STACKROX_MCP__TOOLS__VULNERABILITY__ENABLED=true
./stackrox-mcp
```

The server will start on `http://0.0.0.0:8080` by default. See the [Testing the MCP Server](#testing-the-mcp-server) section for instructions on connecting with Claude Code.

## Configuration

The StackRox MCP server supports configuration through both YAML files and environment variables. Environment variables take precedence over YAML configuration.

### Configuration File

Specify a configuration file using the `--config` flag:

```bash
./stackrox-mcp --config=/path/to/config.yaml
```

See [examples/config-read-only.yaml](examples/config-read-only.yaml) for a complete configuration example.

### Environment Variables

All configuration options can be set via environment variables using the naming convention:

```
STACKROX_MCP__SECTION__KEY
```

Note the double underscore (`__`) separator between sections and keys.

#### Examples

```bash
export STACKROX_MCP__CENTRAL__URL=central.stackrox:443
export STACKROX_MCP__GLOBAL__READ_ONLY_TOOLS=true
export STACKROX_MCP__TOOLS__CONFIG_MANAGER__ENABLED=true
```

### Configuration Options

#### Central Configuration

Configuration for connecting to StackRox Central.

| Option | Environment Variable | Type | Required | Default | Description |
|--------|---------------------|------|----------|---------|-------------|
| `central.url` | `STACKROX_MCP__CENTRAL__URL` | string | Yes | central.stackrox:443 | URL of StackRox Central instance |
| `central.auth_type` | `STACKROX_MCP__CENTRAL__AUTH_TYPE` | string | No | `passthrough` | Authentication type: `passthrough` (use token from MCP client headers) or `static` (use configured token). **Note**: Helm chart only supports `passthrough` mode. |
| `central.api_token` | `STACKROX_MCP__CENTRAL__API_TOKEN` | string | Conditional | - | API token for static authentication (required when `auth_type` is `static`, must not be set when `passthrough`) |
| `central.insecure_skip_tls_verify` | `STACKROX_MCP__CENTRAL__INSECURE_SKIP_TLS_VERIFY` | bool | No | `false` | Skip TLS certificate verification (use only for testing) |
| `central.force_http1` | `STACKROX_MCP__CENTRAL__FORCE_HTTP1` | bool | No | `false` | Route gRPC traffic through the HTTP/1 bridge (gRPC-Web/WebSockets) for environments that block HTTP/2 |
| `central.request_timeout` | `STACKROX_MCP__CENTRAL__REQUEST_TIMEOUT` | duration | No | `30s` | Maximum time to wait for a single request to complete (must be positive) |
| `central.max_retries` | `STACKROX_MCP__CENTRAL__MAX_RETRIES` | int | No | `3` | Maximum number of retry attempts (must be 0-10) |
| `central.initial_backoff` | `STACKROX_MCP__CENTRAL__INITIAL_BACKOFF` | duration | No | `1s` | Initial backoff duration for retries (must be positive) |
| `central.max_backoff` | `STACKROX_MCP__CENTRAL__MAX_BACKOFF` | duration | No | `10s` | Maximum backoff duration for retries (must be positive and >= initial_backoff) |

When `central.force_http1` is enabled, the client uses the [StackRox gRPC-over-HTTP/1 bridge](https://github.com/stackrox/go-grpc-http1) to downgrade requests. This should only be turned on when Central is reached through an HTTP/1-only proxy or load balancer, as client-side streaming remains unsupported in downgrade mode.

#### Global Configuration

Global MCP server settings.

| Option | Environment Variable | Type | Required | Default | Description |
|--------|---------------------|------|----------|---------|-------------|
| `global.read_only_tools` | `STACKROX_MCP__GLOBAL__READ_ONLY_TOOLS` | bool | No | `true` | Only allow read-only tools |

#### Server Configuration

HTTP server settings for the MCP server.

| Option | Environment Variable | Type | Required | Default | Description |
|--------|---------------------|------|----------|---------|-------------|
| `server.type` | `STACKROX_MCP__SERVER__TYPE` | string | No | `streamable-http` | Server transport type: `streamable-http` (HTTP server) or `stdio` (stdio transport). **Note**: stdio transport requires `central.auth_type` to be set to `static`. Helm chart only supports `streamable-http`. |
| `server.address` | `STACKROX_MCP__SERVER__ADDRESS` | string | No | `0.0.0.0` | HTTP server listen address (only applies when `server.type` is `http`) |
| `server.port` | `STACKROX_MCP__SERVER__PORT` | int | No | `8080` | HTTP server listen port (must be 1-65535, only applies when `server.type` is `http`) |

#### Tools Configuration

Enable or disable individual MCP tools. At least one tool has to be enabled.

| Option | Environment Variable | Type | Required | Default | Description |
|--------|---------------------|------|----------|---------|-------------|
| `tools.vulnerability.enabled` | `STACKROX_MCP__TOOLS__VULNERABILITY__ENABLED` | bool | No | `false` | Enable vulnerability management tools |
| `tools.config_manager.enabled` | `STACKROX_MCP__TOOLS__CONFIG_MANAGER__ENABLED` | bool | No | `false` | Enable configuration management tools |

### Configuration Precedence

Configuration values are loaded in the following order (later sources override earlier ones):

1. Default values
2. YAML configuration file (if provided via `--config`)
3. Environment variables (highest precedence)

## Testing the MCP Server

### Starting the Server

Start the server with a configuration file:

```bash
./stackrox-mcp --config examples/config-read-only.yaml
```

Or using environment variables:

```bash
export STACKROX_MCP__CENTRAL__URL="central.example.com:443"
export STACKROX_MCP__TOOLS__VULNERABILITY__ENABLED="true"
./stackrox-mcp
```

The server will start on `http://0.0.0.0:8080` by default (configurable via `server.address` and `server.port`).

### Connecting with Claude Code CLI

#### HTTP Transport

Add the MCP server to Claude Code using HTTP transport:

```bash
claude mcp add --transport http stackrox-mcp http://localhost:8080/mcp \
  --header "Authorization: Bearer <YOUR_STACKROX_API_TOKEN>"
```

#### Stdio Transport

Add the MCP server to Claude Code using stdio transport with static authentication:

```bash
claude mcp add --transport stdio stackrox-mcp \
  --env STACKROX_MCP__SERVER__TYPE=stdio \
  --env STACKROX_MCP__CENTRAL__AUTH_TYPE=static \
  --env STACKROX_MCP__CENTRAL__API_TOKEN="${ROX_TOKEN}" \
  --env STACKROX_MCP__CENTRAL__URL=central.stackrox:443 \
  --env STACKROX_MCP__TOOLS__CONFIG_MANAGER__ENABLED=true \
  -- /path/to/stackrox-mcp
```

**Important**: Stdio transport requires static authentication (`central.auth_type=static`). Passthrough authentication is not supported with stdio transport.

### Verifying Connection

List configured MCP servers:

```bash
claude mcp list
```

Get details for a specific server:

```bash
claude mcp get stackrox-mcp
```

Within a Claude Code session, use the `/mcp` command to view available tools from connected servers.

### Example Usage

Once connected, interact with the tools using natural language:

#### List all clusters
```
Can you list all the clusters secured by StackRox?
```

#### Check for a specific CVE
```
Is CVE-2021-44228 detected in any of my clusters?
```

#### CVE analysis in specific namespace
```
Check if CVE-2021-44228 is present in deployments in namespace "backend"
```

#### Filter by cluster
```
Show me all deployments affected by CVE-2021-44228 in the dev-cluster
```

## Container Images

### Registry

Official images are published to Quay.io:

```
quay.io/stackrox-io/mcp
```

### Supported Architectures

Multi-architecture images support the following platforms:

- `linux/amd64` - Standard x86_64 architecture
- `linux/arm64` - ARM 64-bit (Apple Silicon, AWS Graviton, etc.)
- `linux/ppc64le` - IBM POWER architecture
- `linux/s390x` - IBM Z mainframe architecture

Docker/Podman will automatically pull the correct image for your platform.

### Available Tags

| Tag Pattern | Description | Example |
|-------------|-------------|---------|
| `latest` | Latest release version | `quay.io/stackrox-io/mcp:latest` |
| `v{version}` | Specific release version | `quay.io/stackrox-io/mcp:v1.0.0` |
| `{commit-sha}` | Specific commit from main branch | `quay.io/stackrox-io/mcp:a1b2c3d` |

### Usage

#### Pull Image

```bash
docker pull quay.io/stackrox-io/mcp:latest
# or
podman pull quay.io/stackrox-io/mcp:latest
```

#### Run Container

```bash
docker run -p 8080:8080 \
  --env STACKROX_MCP__CENTRAL__URL=central.stackrox:443 \
  --env STACKROX_MCP__TOOLS__CONFIG_MANAGER__ENABLED=true \
  quay.io/stackrox-io/mcp:latest
```

### Building Images Locally

Build a single-platform image:
```bash
VERSION=dev make image
```

### Build Arguments

- `TARGETOS` - Target operating system (default: `linux`)
- `TARGETARCH` - Target architecture (default: `amd64`)
- `VERSION` - Application version (default: auto-detected from git)

### Image Details

- **Base Image**: Red Hat UBI10-micro (minimal, secure)
- **User**: Non-root user (UID/GID 4000)
- **Port**: 8080
- **Health Check**: Built-in health endpoint at `/health`

### Automated Builds

Images are automatically built and pushed on:

- **Main branch commits**: Tagged with commit SHA
- **Version tags**: Tagged with version number and `latest`

See [.github/workflows/build.yml](.github/workflows/build.yml) for build pipeline details.

## Kubernetes Deployment

Deploy the StackRox MCP server to Kubernetes or OpenShift clusters using Helm.

> **For Kubernetes/OpenShift deployments, see the [Helm Chart README](charts/stackrox-mcp/README.md) for complete configuration options, authentication details, and deployment examples.**

### Prerequisites

- Kubernetes 1.19+ or OpenShift 4.x+
- Helm 3.0+
- Access to a StackRox Central instance

### Installing with Helm

**Basic installation:**

```bash
helm install stackrox-mcp charts/stackrox-mcp \
  --namespace stackrox-mcp \
  --create-namespace \
  --set-file tlsSecret.cert=/path/to/tls.crt \
  --set-file tlsSecret.key=/path/to/tls.key \
  --set-file openshift.route.tls.destinationCACertificate=/path/to/tls.crt \
  --set config.central.url=central.stackrox:443
```

**With custom values file:**

Create a `values.yaml` file:

```yaml
config:
  central:
    url: "central.example.com:443"
```

Install with custom values:

```bash
helm install stackrox-mcp charts/stackrox-mcp \
  --namespace stackrox-mcp \
  --create-namespace \
  --set-file tlsSecret.cert=/path/to/tls.crt \
  --set-file tlsSecret.key=/path/to/tls.key \
  --set-file openshift.route.tls.destinationCACertificate=/path/to/tls.crt \
  --values values.yaml
```

**OpenShift deployment with Route:**

```bash
helm install stackrox-mcp charts/stackrox-mcp \
  --namespace stackrox-mcp \
  --create-namespace \
  --set config.central.url=central.stackrox:443 \
  --set-file tlsSecret.cert=/path/to/tls.crt \
  --set-file tlsSecret.key=/path/to/tls.key \
  --set-file openshift.route.tls.destinationCACertificate=/path/to/tls.crt \
  --set openshift.route.host=stackrox-mcp.apps.example.com
```

**Important Notes for Helm Deployments:**
- The Helm chart uses **passthrough authentication** only (`central.auth_type=passthrough`). Clients must provide API tokens.
- The Helm chart uses **streamable-http transport** only (`server.type=streamable-http`). Stdio transport is not supported in Kubernetes.
- See the [Helm Chart README](charts/stackrox-mcp/README.md) for details on authentication configuration, TLS setup, and advanced options.

### Managing the Deployment

**Upgrade an existing release:**

```bash
helm upgrade stackrox-mcp charts/stackrox-mcp \
  --namespace stackrox-mcp \
  --reuse-values
```

**Uninstall:**

```bash
helm uninstall stackrox-mcp --namespace stackrox-mcp
```

### Configuration

For complete configuration options including:
- Security contexts and pod security standards
- Resource limits and requests
- High availability setup
- OpenShift-specific configuration

See the [Helm Chart README](charts/stackrox-mcp/README.md).

### Integrations

**OpenShift Lightspeed Integration:**

For integrating StackRox MCP with OpenShift Lightspeed, see the [OpenShift Lightspeed Integration Guide](docs/lightspeed-integration.md).

## Development

For detailed development guidelines, testing standards, and contribution workflows, see [CONTRIBUTING.md](.github/CONTRIBUTING.md).

### Quick Reference

View all available commands:
```bash
make help
```

Common commands:
- `make build` - Build the binary
- `make test` - Run tests
- `make fmt` - Format code
- `make lint` - Run linter
