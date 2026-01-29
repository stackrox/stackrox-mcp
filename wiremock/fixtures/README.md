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

### Step 3: Restart and Test

```bash
make mock-restart
# Then test via MCP server or curl
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

- Use realistic IDs and timestamps (ISO 8601 format)
- For empty results, use `empty.json` files with empty arrays
- Check logs with `make mock-logs` to debug mapping issues
- View unmatched requests: `curl http://localhost:8081/__admin/requests/unmatched`

## Best Practices

- Use descriptive snake_case names: `log4j_cve.json`, not `scenario1.json`
- Validate JSON before committing: `cat file.json | jq .`
- Keep fixtures small and focused on specific test scenarios
- Always provide an `empty.json` for each service
