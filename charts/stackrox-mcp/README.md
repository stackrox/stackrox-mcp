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
  --set-file tlsSecret.cert=/path/to/tls.crt \
  --set-file tlsSecret.key=/path/to/tls.key \
  --set-file openshift.route.tls.destinationCACertificate=/path/to/tls.crt \
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
| `nameOverride` | Override the chart name used in resource names | `""` |
| `fullnameOverride` | Override the full resource names entirely | `""` |

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
| `service.type` | Service type | `LoadBalancer` |
| `service.port` | Service port | `8080` |
| `service.annotations` | Service annotations | `{}` |

### TLS Secret Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `tlsSecret.existingSecretName` | Name of existing Kubernetes TLS secret | `""` |
| `tlsSecret.cert` | Server TLS certificate in PEM format | `""` |
| `tlsSecret.key` | Server TLS private key in PEM format | `""` |

**Note:** When `existingSecretName` is set, `cert` and `key` are ignored. Both `cert` and `key` must be provided together when not using an existing secret.

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
| `livenessProbe.initialDelaySeconds` | Initial delay for liveness probe | `10` |
| `livenessProbe.periodSeconds` | Period for liveness probe | `10` |
| `livenessProbe.timeoutSeconds` | Timeout for liveness probe | `5` |
| `livenessProbe.successThreshold` | Success threshold for liveness probe | `1` |
| `livenessProbe.failureThreshold` | Failure threshold for liveness probe | `3` |
| `readinessProbe.enabled` | Enable readiness probe | `true` |
| `readinessProbe.initialDelaySeconds` | Initial delay for readiness probe | `5` |
| `readinessProbe.periodSeconds` | Period for readiness probe | `5` |
| `readinessProbe.timeoutSeconds` | Timeout for readiness probe | `3` |
| `readinessProbe.successThreshold` | Success threshold for readiness probe | `1` |
| `readinessProbe.failureThreshold` | Failure threshold for readiness probe | `3` |

### OpenShift Route Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `openshift.route.host` | Route hostname | `""` |
| `openshift.route.annotations` | Additional annotations to add to the OpenShift Route | `{}` |
| `openshift.route.tls.enabled` | Enable TLS for route | `true` |
| `openshift.route.tls.termination` | TLS termination type (auto: reencrypt if TLS enabled) | `reencrypt` |
| `openshift.route.tls.insecureEdgeTerminationPolicy` | Policy for insecure edge traffic | `Redirect` |
| `openshift.route.tls.destinationCACertificate` | CA certificate for pod verification | `""` |

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
| `config.server.address` | Server listen address inside container | `0.0.0.0` |
| `config.server.port` | Server listen port inside container | `8443` |
| `config.server.TLSEnabled` | Enable TLS/HTTPS on server | `true` |
| `config.server.TLSCertPath` | Certificate path in container (**internal, do not modify**) | `/certs/tls.crt` |
| `config.server.TLSKeyPath` | Private key path in container (**internal, do not modify**) | `/certs/tls.key` |

**Port Configuration:**
- `service.port` (default: `443`) - External port exposed by the Kubernetes Service
- `config.server.port` (default: `8443`) - Port the application listens on inside the container
- Traffic flow: External requests hit Service on port `443` → routed to pod targetPort `8443` → container listens on `config.server.port` (`8443`)

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

### End-to-End TLS/HTTPS Configuration

The chart supports end-to-end TLS encryption from client to pod, providing secure communication throughout the entire connection path.

#### Architecture

When TLS is enabled (`config.server.TLSEnabled: true`):
```
Client (HTTPS)
  → LoadBalancer/Route (TLS passthrough/re-encryption)
    → Service (HTTPS port 8443)
      → Pod (HTTPS on port 8443)
```

When TLS is disabled:
```
Client (HTTPS)
  → LoadBalancer/Route (TLS edge termination)
    → Service (HTTP port 8080)
      → Pod (HTTP on port 8080)
```

#### TLS Secret Configuration

The chart provides two ways to configure TLS certificates for the MCP server.

**Important:** The `tlsSecret.cert` and `tlsSecret.key` fields contain the **server's TLS certificate and private key**. These certificates are used by the MCP server pod to serve HTTPS traffic.

##### Certificate Flow

```
tlsSecret.cert & tlsSecret.key (values.yaml)
  → Kubernetes Secret (tls.crt & tls.key)
    → Volume Mount (/certs/)
      → Referenced by config.server.TLSCertPath & TLSKeyPath
        → Used by MCP server for HTTPS
```

##### Option 1: Generate Secret from Certificate Data (Recommended)

Provide certificate data in values.yaml and the chart will create the secret automatically:

Installation:
```bash
helm install stackrox-mcp charts/stackrox-mcp \
  --namespace stackrox-mcp \
  --create-namespace \
  --set-file tlsSecret.cert=/path/to/tls.crt \
  --set-file tlsSecret.key=/path/to/tls.key \
  --set-file openshift.route.tls.destinationCACertificate=/path/to/tls.crt \
  --set config.central.url=<your-central-url>
```

The chart creates a secret named `<release-name>-stackrox-mcp-tls` automatically.

##### Option 2: Use Existing Secret

Prerequisites - create a Kubernetes TLS secret with your server certificate:
```bash
kubectl create secret tls my-existing-tls-secret \
  --cert=/path/to/tls.crt \
  --key=/path/to/tls.key \
  --namespace stackrox-mcp
```

Reference the existing Kubernetes TLS secret:
```bash
helm install stackrox-mcp charts/stackrox-mcp \
  --namespace stackrox-mcp \
  --create-namespace \
  --set tlsSecret.existingSecretName=my-existing-tls-secret \
  --set config.central.url=<your-central-url>
```

**Important Notes:**
- `tlsSecret.cert` and `tlsSecret.key` are the **server's TLS certificate and private key** (PEM format)
- When `tlsSecret.existingSecretName` is set, `tlsSecret.cert` and `tlsSecret.key` are ignored
- Both `cert` and `key` must be provided together when not using an existing secret
- Use `--set-file` to load certificates from files (recommended over embedding in values)

**Internal TLS Paths (Do Not Modify):**
- `config.server.TLSCertPath: "/certs/tls.crt"` - Container-internal mount path for certificate
- `config.server.TLSKeyPath: "/certs/tls.key"` - Container-internal mount path for private key

These paths are automatically managed by the chart's volume mounts and should not be changed.

##### Generating Self-Signed Certificates

For testing purposes, you can create a self-signed certificate:

```bash
# Generate self-signed certificate for testing
openssl req -x509 -newkey rsa:2048 -days 365 -keyout tls.key -out tls.crt -nodes

# Install with self-signed certificate
helm install stackrox-mcp charts/stackrox-mcp \
  --namespace stackrox-mcp \
  --create-namespace \
  --set-file tlsSecret.cert=tls.crt \
  --set-file tlsSecret.key=tls.key \
  --set-file openshift.route.tls.destinationCACertificate=tls.crt \
  --set config.central.url=<your-central-url>
```

**Warning:** Self-signed certificates should only be used for testing. For production, use certificates from a trusted Certificate Authority.

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

## Default Behaviors

The Helm chart includes several default behaviors to improve reliability and maintainability in Kubernetes environments.

### Automatic Pod Restarts on Configuration Changes

The chart automatically injects a checksum annotation of the ConfigMap into the deployment's pod template. This ensures that:

- When you update configuration via `helm upgrade`, pods are automatically restarted
- Pods always run with the latest configuration from the ConfigMap
- No manual `kubectl rollout restart` is needed after configuration changes

### Default Pod Anti-Affinity

By default, the chart configures pod anti-affinity rules to spread replicas across different nodes. This provides:

- **High Availability**: If a node fails, other replicas continue running
- **Load Distribution**: Pods are distributed across the cluster
- **Better Resource Utilization**: Avoids overloading single nodes

**Default Configuration:**
```yaml
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
```

**To Disable:** Set `affinity: {}` in your values.yaml:
```yaml
affinity: {}
```

### OpenShift Route Sticky Sessions

When deploying on OpenShift, the chart automatically adds a sticky session annotation to the Route resource:

```yaml
haproxy.router.openshift.io/balance: source
```

This ensures:
- **Session Persistence**: MCP clients (AI Agents) use stream-http and non-sticky routing would terminate connections
- **Better Performance**: Reduces overhead from routing changes

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

The application loads configuration with the following precedence (highest to lowest):

1. **Environment variables** (highest precedence) - e.g., `STACKROX_MCP__CENTRAL__URL`
2. **YAML configuration file** - `/config/config.yaml` (generated from ConfigMap)
3. **Application defaults** (lowest precedence) - Built-in fallback values

This means:
- Environment variables **override** values in the YAML config file
- YAML config file values **override** application defaults
- You can use `extraEnv` in values.yaml to override any configuration value via environment variables

**Example:**
```yaml
# Override config.central.url via environment variable
extraEnv:
  - name: STACKROX_MCP__CENTRAL__URL
    value: "custom-central.example.com:443"
```

## Authentication

The Helm chart uses **passthrough authentication** mode, which is hardcoded in the ConfigMap template. This authentication approach is specifically designed for Kubernetes deployments.

### How Passthrough Authentication Works

With passthrough authentication:

1. **No Static Credentials**: The MCP server does not store or manage API tokens directly
2. **Client-Provided Tokens**: API tokens are provided by the MCP client with each request
3. **Token Forwarding**: The server forwards authentication headers transparently to StackRox Central
4. **Per-Request Authentication**: Each request includes the necessary authentication credentials

## Server Transport

The Helm chart uses **streamable-http** transport type, which is hardcoded in the ConfigMap template. This transport is specifically optimized for Kubernetes deployments.

### How It Works

```
MCP Client (HTTPS)
  → Load Balancer/Route
    → Kubernetes Service (port 8080)
      → Pod(s) (streamable-http server on port 8443)
        → StackRox Central API
```

For more details on transport types, see the [main StackRox MCP README](../../README.md#server-configuration).

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
# For HTTP (TLS disabled)
kubectl run -i --tty --rm debug --image=curlimages/curl --restart=Never -- \
  curl http://stackrox-mcp.stackrox-mcp:8080/health

# For HTTPS (TLS enabled)
kubectl run -i --tty --rm debug --image=curlimages/curl --restart=Never -- \
  curl -k https://stackrox-mcp.stackrox-mcp.svc.cluster.local:8443/health
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
