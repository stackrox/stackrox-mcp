# Guide for Setting Up StackRox MCP OpenShift Lightspeed Integration

Guide tested with OpenShift Lightspeed version `1.1.3`.

### 1. Set Up OpenShift Lightspeed
- Set up your OpenShift Lightspeed integration with a large language model (LLM) service. Detailed documentation can be found in the [Red Hat OpenShift Lightspeed Configuration Guide](https://docs.redhat.com/en/documentation/red_hat_openshift_lightspeed/1.0/html/configure/ols-configuring-openshift-lightspeed).
- After OpenShift Lightspeed integration with the LLM is configured and tested, you can continue with StackRox MCP setup.

### 2. Set Up StackRox MCP
- Install StackRox MCP with Helm:
    ```bash
    # Create temp directory and checkout repository with Helm chart.
    tmp_stackrox_mcp_dir="stackrox-mcp-${RANDOM}"
    git clone --depth 1 --branch main https://github.com/stackrox/stackrox-mcp.git "${tmp_stackrox_mcp_dir}"

    # Assuming StackRox Central is installed on the same cluster in the "stackrox" namespace.
    # This installs MCP for a local Central: served over HTTP in-cluster and trusting
    # Central's self-signed certificate. For production, use TLS (see the chart README).
    helm upgrade stackrox-mcp "${tmp_stackrox_mcp_dir}/charts/stackrox-mcp" \
      --install \
      --namespace stackrox-mcp --create-namespace \
      --set config.central.insecureSkipTLSVerify=true \
      --set config.server.TLSEnabled=false \
      --set config.server.port=8080 \
      --set service.type=ClusterIP \
      --set service.port=8080 \
      --set openshift.route.tls.termination=edge \
      --set replicaCount=1

    # Delete temp directory.
    rm -rf "${tmp_stackrox_mcp_dir}"
    ```

    The flags above adapt the chart (which defaults to TLS) for a local Central:
    - `config.central.insecureSkipTLSVerify=true` — trust Central's self-signed certificate.
    - `config.server.TLSEnabled=false` + `config.server.port=8080` — serve MCP over HTTP on 8080 (no server certificate needed).
    - `service.type=ClusterIP` + `service.port=8080` — OpenShift Lightspeed reaches MCP via the in-cluster Service.
    - `openshift.route.tls.termination=edge` — required because the pod now serves HTTP.

    > **Note:** For advanced helm chart configuration options, see the [StackRox MCP Helm Chart README](../charts/stackrox-mcp/README.md). For OpenShift-specific deployment settings, refer to the [OpenShift Deployment](../charts/stackrox-mcp/README.md#openshift-deployment) section.

- Verify the MCP server is running:
    ```bash
    kubectl run -i --tty --rm debug --image=quay.io/curl/curl:latest --restart=Never -- \
      curl http://stackrox-mcp.stackrox-mcp:8080/health
    ```
    You should get `{"status":"ok"}` as a response.

### 3. Set Up Integration of StackRox MCP with OpenShift Lightspeed
- Create an API token in StackRox Central with appropriate permissions.
- Create the `stackrox-mcp-authorization-header` secret in the `openshift-lightspeed` namespace (kubectl encodes the value for you):
    ```bash
    stackrox_api_token="<StackRox API Token>"
    kubectl create secret generic stackrox-mcp-authorization-header \
      --namespace openshift-lightspeed \
      --from-literal=header="Bearer ${stackrox_api_token}"
    ```
- Configure OpenShift Lightspeed by editing the `OLSConfig` configuration for your OpenShift Lightspeed installation and add this section to `spec`:
    ```yaml
      featureGates:
        - MCPServer
      mcpServers:
        - name: stackrox-mcp
          headers:
            - name: authorization
              valueFrom:
                type: secret
                secretRef:
                  name: stackrox-mcp-authorization-header
          timeout: 120
          url: 'http://stackrox-mcp.stackrox-mcp:8080/mcp'
    ```
- After completing the setup, test your integration with a simple prompt: "List all clusters secured by StackRox"

### Troubleshooting
If you encounter issues, refer to the [Troubleshooting](../charts/stackrox-mcp/README.md#troubleshooting) section in the Helm chart documentation.
