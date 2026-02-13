# WireMock Mock StackRox Central

WireMock standalone mock service for StackRox Central. Docs: [WireMock](https://wiremock.org/docs/), [gRPC Extension](https://github.com/wiremock/wiremock-grpc-extension)

## Quick Start

```bash
make mock-download        # Download WireMock JARs
make proto-setup          # Setup proto files from github.com/stackrox/rox
./scripts/generate-proto-descriptors.sh  # Generate proto descriptors
make mock-start           # Start on port 8081
```

## Usage

Commands: `make mock-start|stop|restart|status|logs|test`

Connect MCP server to `localhost:8081` with `STACKROX_MCP__CENTRAL__API_TOKEN=test-token-admin` and `INSECURE_SKIP_TLS_VERIFY=true`

## Test Data

**Auth:** Any `test-token-*` accepted (e.g., `test-token-admin`)

**CVE Queries:**
- `CVE-2021-44228` → 3 deployments (log4j)
- `CVE-2021-31805` → 3 deployments, 2 clusters
- `CVE-2016-1000031` → 2 deployments, 1 cluster
- Other CVEs → 1 deployment
- No CVE → all clusters (3)

## Adding Scenarios

1. Add fixture to `fixtures/` (see `fixtures/README.md`)
2. Add mapping to `mappings/` ([docs](https://wiremock.org/docs/request-matching/))
3. `make mock-restart`

## Troubleshooting

- Check logs: `cat wiremock/wiremock.log`
- Missing proto descriptors: `./scripts/generate-proto-descriptors.sh`
- Debug requests: `curl http://localhost:8081/__admin/requests` ([Admin API docs](https://wiremock.org/docs/api/))
