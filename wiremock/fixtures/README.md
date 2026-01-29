# WireMock Fixtures

This directory contains JSON fixture files that WireMock uses to respond to gRPC requests.

## Directory Structure

```
fixtures/
├── deployments/     # DeploymentService responses
├── images/          # ImageService responses
├── nodes/           # NodeService responses
└── clusters/        # ClustersService responses
```

## How Fixtures Work

1. **WireMock mappings** (in `../mappings/`) define request matching rules
2. **Fixtures** (in this directory) provide the response data
3. Mappings reference fixtures via `bodyFileName` field

Example mapping:
```json
{
  "request": {
    "urlPath": "/v1.DeploymentService/ListDeployments",
    "bodyPatterns": [
      {"matchesJsonPath": "$.query[?(@.query =~ /.*CVE-2021-44228.*/)]"}
    ]
  },
  "response": {
    "bodyFileName": "deployments/log4j_cve.json"
  }
}
```

## Adding New Test Scenarios

### Step 1: Create a Fixture File

Create a new JSON file with realistic response data:

```bash
# Example: Create a fixture for a specific CVE
cat > deployments/my_cve_scenario.json <<EOF
{
  "deployments": [
    {
      "id": "dep-123",
      "name": "my-deployment",
      "namespace": "production",
      "clusterId": "cluster-prod-01",
      "cluster": "Production Cluster",
      "created": "2024-01-15T10:30:00Z",
      "labels": {
        "app": "my-app"
      }
    }
  ]
}
EOF
```

### Step 2: Add a Mapping Rule

Edit the corresponding mapping file (e.g., `../mappings/deployments.json`):

```json
{
  "priority": 15,
  "request": {
    "method": "POST",
    "urlPath": "/v1.DeploymentService/ListDeployments",
    "bodyPatterns": [
      {
        "matchesJsonPath": "$.query[?(@.query =~ /.*CVE-2024-5678.*/)]"
      }
    ]
  },
  "response": {
    "status": 200,
    "bodyFileName": "deployments/my_cve_scenario.json",
    "headers": {
      "Content-Type": "application/grpc+json",
      "grpc-status": "0"
    }
  }
}
```

### Step 3: Restart WireMock

```bash
make mock-restart
```

### Step 4: Test Your Scenario

```bash
# Configure MCP to use mock Central
export STACKROX_MCP__CENTRAL__URL=localhost:9000
export STACKROX_MCP__CENTRAL__API_TOKEN=test-token-admin
export STACKROX_MCP__CENTRAL__INSECURE_SKIP_TLS_VERIFY=true

# Test via MCP server
./bin/stackrox-mcp
# Then call get_deployments_for_cve with your CVE
```

## Fixture Data Format

### Deployments

Must match the `storage.ListDeploymentsResponse` protobuf structure:

```json
{
  "deployments": [
    {
      "id": "unique-deployment-id",
      "name": "deployment-name",
      "namespace": "kubernetes-namespace",
      "clusterId": "cluster-id",
      "cluster": "cluster-name",
      "created": "2024-01-15T10:30:00Z",
      "labels": {
        "key": "value"
      },
      "containers": [
        {
          "name": "container-name",
          "image": {
            "name": {
              "fullName": "registry.io/image:tag"
            }
          }
        }
      ]
    }
  ]
}
```

### Images

Must match the `storage.ListImagesResponse` protobuf structure:

```json
{
  "images": [
    {
      "id": "unique-image-id",
      "name": {
        "registry": "docker.io",
        "remote": "library/nginx",
        "tag": "1.19",
        "fullName": "docker.io/library/nginx:1.19"
      }
    }
  ]
}
```

### Nodes

Must match the `storage.Node` protobuf structure:

```json
{
  "nodes": [
    {
      "id": "unique-node-id",
      "name": "node-name",
      "clusterId": "cluster-id",
      "clusterName": "cluster-name",
      "osImage": "Ubuntu 20.04.3 LTS",
      "containerRuntimeVersion": "containerd://1.5.8",
      "kernelVersion": "5.4.0-90-generic",
      "kubeletVersion": "v1.22.4"
    }
  ]
}
```

### Clusters

Must match the `storage.Cluster` protobuf structure:

```json
{
  "clusters": [
    {
      "id": "unique-cluster-id",
      "name": "cluster-name",
      "type": "KUBERNETES_CLUSTER",
      "healthStatus": {
        "overallHealthStatus": "HEALTHY"
      }
    }
  ]
}
```

## Tips

### Realistic Data

- Use realistic IDs (e.g., UUIDs or descriptive strings)
- Include timestamps in ISO 8601 format
- Add labels and annotations that match real Kubernetes resources
- Use actual CVE numbers for testing

### Empty Responses

For scenarios with no results, use the `empty.json` files:

```json
{
  "deployments": []
}
```

### Parameter Matching

You can create different fixtures based on query parameters:

```json
{
  "bodyPatterns": [
    {"matchesJsonPath": "$.query[?(@.query =~ /.*Cluster:\"prod\".*/)]"}
  ]
}
```

This matches queries containing `Cluster:"prod"`.

### Debugging

Check WireMock logs to see which mappings are being matched:

```bash
make mock-logs
```

View unmatched requests via admin API:

```bash
curl http://localhost:8081/__admin/requests/unmatched
```

## Naming Conventions

- Use descriptive names: `log4j_cve.json`, not `scenario1.json`
- Use snake_case for file names
- Group related scenarios in subdirectories if needed
- Always provide an `empty.json` for each service

## Validation

Validate your JSON before adding it:

```bash
cat deployments/my_scenario.json | jq .
```

If `jq` returns an error, fix the JSON syntax.

## Version Control

- Commit fixture files to git
- Document what each fixture is testing
- Keep fixtures small and focused on specific scenarios
- Add comments in this README when adding complex scenarios
