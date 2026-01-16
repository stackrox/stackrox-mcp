# StackRox MCP Helm Chart

A Helm chart for deploying the StackRox Model Context Protocol (MCP) Server to Kubernetes and OpenShift clusters.

## Prerequisites

- Kubernetes 1.19+ or OpenShift 4.x+
- Helm 3.0+
- Access to a StackRox Central instance

## Installing the Chart

To install the chart with the release name `stackrox-mcp`:

```bash
helm install stackrox-mcp charts/stackrox-mcp \
  --namespace stackrox-mcp \
  --create-namespace \
  --set config.central.url=<your-central-url>
```

## Upgrading the Chart

To upgrade an existing release:

```bash
helm upgrade stackrox-mcp charts/stackrox-mcp \
  --namespace stackrox-mcp \
  --reuse-values
```

## Uninstalling the Chart

To uninstall/delete the `stackrox-mcp` release:

```bash
helm uninstall stackrox-mcp --namespace stackrox-mcp
```

## Configuration

The following table lists the configurable parameters of the StackRox MCP chart and their default values.

### Image Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `image.registry` | Container image registry | `quay.io` |
| `image.repository` | Container image repository | `stackrox-io/mcp` |
| `image.tag` | Container image tag (overrides appVersion) | `""` |
| `image.pullPolicy` | Image pull policy | `IfNotPresent` |
| `imagePullSecrets` | Image pull secrets for private registries | `[]` |

### Deployment Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `replicaCount` | Number of replicas | `2` |
| `annotations` | Annotations for Deployment and Pod metadata | `{}` |

### Service Account

| Parameter | Description | Default |
|-----------|-------------|---------|
| `serviceAccount.annotations` | Annotations for the service account | `{}` |
| `serviceAccount.automountServiceAccountToken` | Automount service account token | `false` |

### Security Contexts

| Parameter | Description | Default |
|-----------|-------------|---------|
| `podSecurityContext.runAsNonRoot` | Run as non-root user | `true` |
| `podSecurityContext.runAsUser` | User ID to run as (auto-omitted on OpenShift) | `4000` |
| `podSecurityContext.runAsGroup` | Group ID to run as (auto-omitted on OpenShift) | `4000` |
| `podSecurityContext.fsGroup` | Filesystem group ID (auto-omitted on OpenShift) | `4000` |
| `podSecurityContext.seccompProfile.type` | Seccomp profile type | `RuntimeDefault` |
| `securityContext.allowPrivilegeEscalation` | Allow privilege escalation | `false` |
| `securityContext.readOnlyRootFilesystem` | Read-only root filesystem | `true` |
| `securityContext.runAsNonRoot` | Run as non-root user | `true` |
| `securityContext.runAsUser` | User ID to run as (auto-omitted on OpenShift) | `4000` |
| `securityContext.capabilities.drop` | List of capabilities to drop | `["ALL"]` |

**Note:** On OpenShift, `runAsUser`, `runAsGroup`, and `fsGroup` are automatically omitted to allow the SecurityContextConstraints (SCC) to assign UIDs from the namespace's valid range (e.g., 1000760000-1000769999).

### Service Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `service.type` | Service type | `ClusterIP` |
| `service.port` | Service port | `8080` |
| `service.targetPort` | Target port | `8080` |
| `service.annotations` | Service annotations | `{}` |

### Resource Limits

| Parameter | Description | Default |
|-----------|-------------|---------|
| `resources.requests.cpu` | CPU request | `200m` |
| `resources.requests.memory` | Memory request | `500Mi` |
| `resources.limits.cpu` | CPU limit | *not defined* |
| `resources.limits.memory` | Memory limit | `500Mi` |

### Probes

| Parameter | Description | Default |
|-----------|-------------|---------|
| `livenessProbe.enabled` | Enable liveness probe | `true` |
| `livenessProbe.httpGet.path` | Path for liveness probe | `/health` |
| `livenessProbe.httpGet.port` | Port for liveness probe | `http` |
| `livenessProbe.initialDelaySeconds` | Initial delay for liveness probe | `10` |
| `livenessProbe.periodSeconds` | Period for liveness probe | `10` |
| `livenessProbe.timeoutSeconds` | Timeout for liveness probe | `5` |
| `livenessProbe.successThreshold` | Success threshold for liveness probe | `1` |
| `livenessProbe.failureThreshold` | Failure threshold for liveness probe | `3` |
| `readinessProbe.enabled` | Enable readiness probe | `true` |
| `readinessProbe.httpGet.path` | Path for readiness probe | `/health` |
| `readinessProbe.httpGet.port` | Port for readiness probe | `http` |
| `readinessProbe.initialDelaySeconds` | Initial delay for readiness probe | `5` |
| `readinessProbe.periodSeconds` | Period for readiness probe | `5` |
| `readinessProbe.timeoutSeconds` | Timeout for readiness probe | `3` |
| `readinessProbe.successThreshold` | Success threshold for readiness probe | `1` |
| `readinessProbe.failureThreshold` | Failure threshold for readiness probe | `3` |

### OpenShift Route Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `openshift.route.host` | Route hostname | `""` |
| `openshift.route.tls.enabled` | Enable TLS for route | `true` |
| `openshift.route.tls.termination` | TLS termination type | `edge` |
| `openshift.route.tls.insecureEdgeTerminationPolicy` | Policy for insecure edge traffic | `Redirect` |

### Scheduling

| Parameter | Description | Default |
|-----------|-------------|---------|
| `nodeSelector` | Node labels for pod assignment | `{}` |
| `tolerations` | Tolerations for pod assignment | `[]` |
| `affinity` | Affinity rules for pod assignment | `{}` |
| `priorityClassName` | Priority class name for pod scheduling | `""` |

### Advanced Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `extraEnv` | Additional environment variables | `[]` |

### StackRox MCP Configuration

#### Central Connection

| Parameter | Description | Default |
|-----------|-------------|---------|
| `config.central.url` | StackRox Central URL | `central.stackrox:443` |
| `config.central.insecureSkipTLSVerify` | Skip TLS verification (testing only) | `false` |
| `config.central.forceHTTP1` | Force HTTP/1 bridge | `false` |
| `config.central.requestTimeout` | Request timeout | `30s` |
| `config.central.maxRetries` | Maximum retry attempts | `3` |
| `config.central.initialBackoff` | Initial retry backoff | `1s` |
| `config.central.maxBackoff` | Maximum retry backoff | `10s` |

#### Global Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `config.global.readOnlyTools` | Restrict to read-only operations | `true` |

#### Server Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `config.server.address` | Server listen address | `0.0.0.0` |
| `config.server.port` | Server listen port | `8080` |

#### Tools Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `config.tools.vulnerability.enabled` | Enable vulnerability management tools | `true` |
| `config.tools.configManager.enabled` | Enable configuration management tools | `true` |

**Note**: At least one tool must be enabled.

## Common Configurations

### Basic configuration

```yaml
config:
  central:
    url: "central.stackrox:443"
```

### OpenShift Deployment

For OpenShift with external Route:

```yaml
# OpenShift automatically assigns UIDs via SCC
# No need to override security context values

openshift:
  route:
    host: "stackrox-mcp.apps.example.com"
    tls:
      enabled: true
      termination: edge

config:
  central:
    url: "central.stackrox:443"
```

**OpenShift Security Context Constraints (SCC):**

The chart automatically detects OpenShift and omits `runAsUser`, `runAsGroup`, and `fsGroup` settings to allow the cluster's SCC to assign UIDs from valid namespace ranges (e.g., 1000760000-1000769999).

The chart maintains security hardening on OpenShift:
- `runAsNonRoot: true` - Always enforced
- `readOnlyRootFilesystem: true` - Always enforced
- `allowPrivilegeEscalation: false` - Always enforced
- `capabilities: drop: ["ALL"]` - Always enforced

The chart is compatible with the `restricted-v2` SCC on OpenShift.

### High Availability Setup

For high availability with multiple replicas:

```yaml
replicaCount: 3

affinity:
  podAntiAffinity:
    preferredDuringSchedulingIgnoredDuringExecution:
    - weight: 100
      podAffinityTerm:
        labelSelector:
          matchExpressions:
          - key: app.kubernetes.io/name
            operator: In
            values:
            - stackrox-mcp
        topologyKey: kubernetes.io/hostname

config:
  central:
    url: "central.stackrox:443"
```

## Configuration Loading

The StackRox MCP Helm chart uses a YAML configuration file approach for cleaner and more maintainable configuration management.

### How It Works

1. **ConfigMap with YAML File**: The chart creates a ConfigMap containing a complete `config.yaml` file with all configuration
2. **File Mounting**: The ConfigMap is mounted as a file at `/config/config.yaml` in the pod
3. **Command Args**: The application is started with `--config /config/config.yaml` to load the configuration

### Viewing Configuration

To view the current configuration:

```bash
# View the YAML config file in ConfigMap
kubectl get configmap -n stackrox-mcp <release-name>-config -o jsonpath='{.data.config\.yaml}'

# Verify config file is mounted in pod
kubectl exec -n stackrox-mcp deployment/<release-name> -- cat /config/config.yaml
```

### Configuration Precedence

The application loads configuration in this order (highest to lowest precedence):
1. Environment variables (e.g., `STACKROX_MCP__CENTRAL__API_TOKEN`)
2. YAML configuration file (`/config/config.yaml`)
3. Application defaults

This means you can override any YAML configuration value using environment variables via `extraEnv` in values.yaml.

## Troubleshooting

### Deployment fails to start

Check the pod logs:

```bash
kubectl logs -n stackrox-mcp deployment/stackrox-mcp
```

Common issues:
- **Invalid Central URL**: Verify the StackRox Central URL is correct
- **Network connectivity**: Ensure the pod can reach StackRox Central on defined URL and port

### Health check failures

Test the health endpoint:

```bash
kubectl run -i --tty --rm debug --image=curlimages/curl --restart=Never -- \
  curl http://stackrox-mcp.stackrox-mcp:8080/health
```

Expected response: `{"status":"ok"}`

### Configuration not updating

Configuration changes trigger automatic pod restarts via checksum annotations. If changes don't apply:

```bash
kubectl rollout restart deployment/stackrox-mcp -n stackrox-mcp
```

### No tools enabled error

At least one tool must be enabled. Set either:
- `config.tools.vulnerability.enabled: true`
- `config.tools.configManager.enabled: true`

### OpenShift SCC Errors

If you see errors like:
```
.containers[0].runAsUser: Invalid value: 4000: must be in the ranges: [1000760000, 1000769999]
```

**Solution:** The chart automatically detects OpenShift and omits hardcoded UIDs. Ensure you're using Helm 3.1+ which properly passes API capabilities to templates.

To verify OpenShift detection:
```bash
helm template test charts/stackrox-mcp --api-versions route.openshift.io/v1 | grep -A 5 "securityContext:"
```

Expected: No `runAsUser`, `runAsGroup`, or `fsGroup` should appear in the pod security context.
