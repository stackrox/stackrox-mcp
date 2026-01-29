# WireMock Mock StackRox Central

This directory contains a WireMock standalone mock service for StackRox Central, allowing you to develop and test the MCP server without requiring an actual StackRox Central instance.

## Directory Structure

```
wiremock/
├── lib/                    # WireMock JARs (downloaded via script)
├── proto/                  # Proto files (copied from stackrox repo)
│   ├── stackrox/           # StackRox proto files
│   ├── googleapis/         # Google API proto files
│   └── descriptors/        # Generated proto descriptors
├── grpc/                   # gRPC descriptor files for WireMock
├── mappings/               # WireMock stub definitions
│   ├── auth.json           # Authentication validation
│   ├── deployments.json    # DeploymentService mappings
│   ├── images.json         # ImageService mappings
│   ├── nodes.json          # NodeService mappings
│   └── clusters.json       # ClustersService mappings
├── fixtures/               # Response data (easy to edit!)
│   ├── deployments/
│   ├── images/
│   ├── nodes/
│   └── clusters/
└── __files -> fixtures     # Symlink for WireMock compatibility
```

## Initial Setup

### 1. Download WireMock JARs

```bash
make mock-download
```

This downloads:
- `wiremock-standalone.jar` (~17MB)
- `wiremock-grpc-extension.jar` (~24MB)

### 2. Copy Proto Files

Copy proto files from the stackrox repository:

```bash
./scripts/setup-proto-files.sh
```

This requires the stackrox repo cloned as a sibling directory or set via `STACKROX_REPO_PATH`.

### 3. Generate Proto Descriptors

```bash
./scripts/generate-proto-descriptors.sh
```

This creates `wiremock/proto/descriptors/stackrox.pb` used by WireMock gRPC extension.

### 4. Start the Mock Service

```bash
make mock-start
```

WireMock will start on port 8081 (both HTTP and gRPC).

## Usage

### Starting/Stopping

```bash
# Start mock Central
make mock-start

# Check status
make mock-status

# View logs
make mock-logs

# Restart
make mock-restart

# Stop
make mock-stop

# Run smoke tests
make mock-test
```

### Connecting MCP Server

```bash
# Configure environment
export STACKROX_MCP__SERVER__TYPE=stdio
export STACKROX_MCP__CENTRAL__URL=localhost:8081
export STACKROX_MCP__CENTRAL__AUTH_TYPE=static
export STACKROX_MCP__CENTRAL__API_TOKEN=test-token-admin
export STACKROX_MCP__CENTRAL__INSECURE_SKIP_TLS_VERIFY=true
export STACKROX_MCP__TOOLS__VULNERABILITY__ENABLED=true

# Run MCP server
./stackrox-mcp
```

### Test with curl

```bash
# Test with valid token - should return log4j deployments
curl -X POST \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer test-token-admin" \
  -d '{"query":{"query":"CVE:\"CVE-2021-44228\""}}' \
  http://localhost:8081/v1.DeploymentService/ListDeployments

# Test without token - should return 401
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{}' \
  http://localhost:8081/v1.DeploymentService/ListDeployments
```

## Test Scenarios

### Deployments

| Query Contains | Returns |
|---|---|
| `CVE-2021-44228` | 3 deployments (Log4j scenario) |
| `CVE-2024-1234` | 1 deployment (custom scenario) |
| Anything else | Empty results |

### Authentication

| Header | Result |
|---|---|
| `Authorization: Bearer test-token-admin` | ✓ Accepted |
| `Authorization: Bearer test-token-readonly` | ✓ Accepted |
| `Authorization: Bearer test-token-*` | ✓ Accepted (any test-token) |
| Missing or invalid | 401 Unauthenticated |

## Adding New Test Scenarios

See `fixtures/README.md` for detailed instructions on adding new test data.

Quick example:

1. Create fixture file:
```bash
cat > fixtures/deployments/my_scenario.json <<EOF
{
  "deployments": [
    {
      "id": "dep-001",
      "name": "my-deployment",
      "namespace": "production"
    }
  ]
}
EOF
```

2. Update mapping in `mappings/deployments.json`:
```json
{
  "priority": 15,
  "request": {
    "method": "POST",
    "urlPath": "/v1.DeploymentService/ListDeployments",
    "bodyPatterns": [
      {"matchesJsonPath": "$.query[?(@.query =~ /.*CVE-2024-5678.*/)]"}
    ]
  },
  "response": {
    "status": 200,
    "bodyFileName": "deployments/my_scenario.json"
  }
}
```

3. Restart WireMock:
```bash
make mock-restart
```

## Testing

### Smoke Tests

Run smoke tests to verify WireMock integration:

```bash
make mock-test
```

The smoke test verifies:
- ✓ WireMock service starts and responds
- ✓ Authentication validation works
- ✓ CVE queries return correct data
- ✓ MCP server integrates with mock Central

**CI Integration**: Smoke tests run automatically in GitHub Actions on all PRs.

## Troubleshooting

### WireMock fails to start

Check logs:
```bash
cat wiremock/wiremock.log
```

Common issues:
- Proto descriptors missing: Run `./scripts/generate-proto-descriptors.sh`
- Port 8081 in use: Stop other services or change port in `start-mock-central.sh`

### Mappings not matching

View requests:
```bash
curl http://localhost:8081/__admin/requests
```

View unmatched requests:
```bash
curl http://localhost:8081/__admin/requests/unmatched
```

### Fixture file not found

Ensure `__files` symlink exists:
```bash
ls -la wiremock/ | grep __files
# Should show: __files -> fixtures
```

If missing, recreate:
```bash
ln -s fixtures wiremock/__files
```

## Admin API

WireMock provides an admin API at `http://localhost:8081/__admin`

Useful endpoints:
- `GET /__admin/mappings` - List all mappings
- `GET /__admin/requests` - List all requests
- `POST /__admin/reset` - Reset request journal
- `GET /__admin/version` - WireMock version

## Architecture

WireMock serves both HTTP/JSON and gRPC on port 8081:
1. Loads proto descriptors from `grpc/` directory
2. Matches requests using mappings in `mappings/`
3. Returns response data from `__files/` (symlink to `fixtures/`)
4. Validates authentication tokens via regex patterns
