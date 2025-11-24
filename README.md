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
export STACKROX_MCP__CENTRAL__URL=central.stackrox:8443
export STACKROX_MCP__TOOLS__VULNERABILITY__ENABLED=true
./stackrox-mcp
```

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
export STACKROX_MCP__CENTRAL__URL=central.stackrox:8443
export STACKROX_MCP__GLOBAL__READ_ONLY_TOOLS=true
export STACKROX_MCP__TOOLS__CONFIG_MANAGER__ENABLED=true
```

### Configuration Options

#### Central Configuration

Configuration for connecting to StackRox Central.

| Option | Environment Variable | Type | Required | Default | Description |
|--------|---------------------|------|----------|---------|-------------|
| `central.url` | `STACKROX_MCP__CENTRAL__URL` | string | Yes | central.stackrox:8443 | URL of StackRox Central instance |
| `central.insecure` | `STACKROX_MCP__CENTRAL__INSECURE` | bool | No | `false` | Skip TLS certificate verification |
| `central.force_http1` | `STACKROX_MCP__CENTRAL__FORCE_HTTP1` | bool | No | `false` | Force HTTP/1.1 instead of HTTP/2 |

#### Global Configuration

Global MCP server settings.

| Option | Environment Variable | Type | Required | Default | Description |
|--------|---------------------|------|----------|---------|-------------|
| `global.read_only_tools` | `STACKROX_MCP__GLOBAL__READ_ONLY_TOOLS` | bool | No | `true` | Only allow read-only tools |

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
