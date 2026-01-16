# Guide for Setting Up StackRox MCP OpenShift Lightspeed Integration

Guide tested with OpenShift Lightspeed version `1.0.8`.

### 1. Set Up OpenShift Lightspeed
- Set up your OpenShift Lightspeed integration with a large language model (LLM) service. Detailed documentation can be found in the [Red Hat OpenShift Lightspeed Configuration Guide](https://docs.redhat.com/en/documentation/red_hat_openshift_lightspeed/1.0/html/configure/ols-configuring-openshift-lightspeed).
- After OpenShift Lightspeed integration with the LLM is configured and tested, you can continue with StackRox MCP setup.

### 2. Set Up StackRox MCP
- Install StackRox MCP with Helm:
    ```bash
    # Create temp directory and checkout repository with Helm chart.
    tmp_stackrox_mcp_dir="stackrox-mcp-${RANDOM}"
    git clone --depth 1 --branch main https://github.com/stackrox/stackrox-mcp.git "${tmp_stackrox_mcp_dir}"

    # Assuming that StackRox Central is installed on the same cluster in "stackrox" namespace.
    helm install stackrox-mcp "${tmp_stackrox_mcp_dir}/charts/stackrox-mcp" --namespace stackrox-mcp --create-namespace

    # Delete temp directory.
    rm -rf "${tmp_stackrox_mcp_dir}"
    ```

    > **Note:** For advanced helm chart configuration options, see the [StackRox MCP Helm Chart README](../charts/stackrox-mcp/README.md). For OpenShift-specific deployment settings, refer to the [OpenShift Deployment](../charts/stackrox-mcp/README.md#openshift-deployment) section.

- Verify the MCP server is running:
    ```bash
    kubectl run -i --tty --rm debug --image=curlimages/curl --restart=Never -- \
      curl http://stackrox-mcp.stackrox-mcp:8080/health
    ```
    You should get `{"status":"ok"}` as a response.

### 3. Set Up Integration of StackRox MCP with OpenShift Lightspeed
- Create an API token in StackRox Central with appropriate permissions.
- Create Authorization Header Secret
  - Create a Base64 value for the authorization header secret:
    ```bash
    stackrox_api_token="<StackRox API Token>"
    echo -n "Bearer ${stackrox_api_token}" | base64
    ```
  - Create secret `stackrox-mcp-authorization-header` in the `openshift-lightspeed` namespace:
    ```yaml
    kind: Secret
    apiVersion: v1
    metadata:
      name: stackrox-mcp-authorization-header
      namespace: openshift-lightspeed
    data:
      header: "<Base64 value for authorization header>"
    type: Opaque
    ```
- Configure OpenShift Lightspeed by editing the `OLSConfig` configuration for your OpenShift Lightspeed installation and add this section to `spec`:
    ```yaml
      featureGates:
        - MCPServer
      mcpServers:
        - name: stackrox-mcp
          streamableHTTP:
            enableSSE: false
            headers:
              authorization: stackrox-mcp-authorization-header
            sseReadTimeout: 30
            timeout: 60
            url: 'http://stackrox-mcp.stackrox-mcp:8080/mcp'
    ```
- After completing the setup, test your integration with a simple prompt: "List all clusters secured by StackRox"

### Troubleshooting
If you encounter issues, refer to the [Troubleshooting](../charts/stackrox-mcp/README.md#troubleshooting) section in the Helm chart documentation.
