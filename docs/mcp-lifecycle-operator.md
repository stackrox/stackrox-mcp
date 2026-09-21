# Guide for Deploying StackRox MCP with the MCP Lifecycle Operator

This guide describes how to deploy the StackRox MCP server using the [MCP Lifecycle Operator](https://catalog.redhat.com/en/software/containers/mcp-lifecycle-operator-beta/mcp-lifecycle-rhel9-operator/69de44471b613610210c8c01), which manages MCP servers through an `MCPServer` custom resource.

### 1. Prerequisites
- The [MCP Lifecycle Operator](https://catalog.redhat.com/en/software/containers/mcp-lifecycle-operator-beta/mcp-lifecycle-rhel9-operator/69de44471b613610210c8c01) is installed on the cluster.
- StackRox Central is reachable from the cluster. This guide assumes Central is installed on the **same cluster** (for example, in the `stackrox` namespace).

### 2. Create the namespace
```bash
kubectl create namespace acs-mcp
```

### 3. Deploy the MCP server
Create an `MCPServer` resource. The operator provisions the deployment and an in-cluster Service from this specification:
```yaml
apiVersion: mcp.x-k8s.io/v1alpha1
kind: MCPServer
metadata:
  name: acs-mcp
  namespace: acs-mcp
spec:
  config:
    env:
      - name: STACKROX_MCP__TOOLS__CONFIG_MANAGER__ENABLED
        value: 'true'
      - name: STACKROX_MCP__TOOLS__VULNERABILITY__ENABLED
        value: 'true'
      - name: STACKROX_MCP__CENTRAL__INSECURE_SKIP_TLS_VERIFY
        value: 'true'
    port: 8080
  source:
    containerImage:
      ref: 'registry.redhat.io/agentic-cluster-security-suite-tech-preview/acs-mcp-server-rhel9:0.2'
    type: ContainerImage
```

Apply it:
```bash
kubectl apply -f mcpserver.yaml
```

The `spec.config.env` entries configure the MCP server:
- `STACKROX_MCP__TOOLS__CONFIG_MANAGER__ENABLED=true` — enable the config management tools (disabled by default).
- `STACKROX_MCP__TOOLS__VULNERABILITY__ENABLED=true` — enable the vulnerability management tools (disabled by default).

### 4. Verify the deployment
- Check that the `MCPServer` resource and its pod are running:
    ```bash
    kubectl get mcpserver -n acs-mcp
    kubectl get pods -n acs-mcp
    ```

- Verify the MCP server responds:
    ```bash
    kubectl run -i --tty --rm debug --image=quay.io/curl/curl:latest --restart=Never -- curl http://acs-mcp.acs-mcp:8080/health
    ```
    You should get `{"status":"ok"}` as a response.

### Appendix: Integrating with OpenShift Lightspeed
The operator exposes the MCP server through an in-cluster Service, so you can integrate it with OpenShift Lightspeed the same way as a Helm-based deployment. Follow [Step 3 of the OpenShift Lightspeed Integration Guide](lightspeed-integration.md) to create the authorization-header secret and update the `OLSConfig`.

When configuring `mcpServers` in the `OLSConfig`, set the `url` to the Service created by the operator:
```yaml
      url: 'http://acs-mcp.acs-mcp:8080/mcp'
```
