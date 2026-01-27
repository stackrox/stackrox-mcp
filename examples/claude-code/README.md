# Claude Code Usage Examples

This directory contains configuration examples and usage prompts for integrating StackRox MCP with Claude Code.

## Prerequirements

- **Build binary**:
    ```bash
    make build
    ```

## Setup

1. **Start the StackRox MCP server**:
   ```bash
   STACKROX_MCP__CENTRAL__URL=<StackRox host with port> ./stackrox-mcp --config=examples/config-read-only.yaml
   ```

2. **Configure Claude Code**:

   Copy the `mcp.json` file to your workspace directory. This will scope Claude Code to use StackRox MCP only when executed in that directory:
   ```bash
   mkdir -p ~/stackrox-workspace

   cp examples/claude-code/mcp.json ~/stackrox-workspace/.mcp.json

   cd ~/stackrox-workspace
   ```

3. **Update the API token**:

   Edit `~/stackrox-workspace/.mcp.json` and replace `<YOUR_STACKROX_API_TOKEN>` with your actual StackRox API token.

4. **Verify the connection**:

    Start Claude Code
   ```bash
   claude
   ```

   You can get prompt to allow using MCP server. After that you can run `/mcp` to see if `stackrox-mcp` is in the list.

## Example Prompts

Once configured, you can use natural language prompts with Claude Code:

### List all clusters
```
Can you list all the clusters secured by StackRox?
```

### Check for a specific CVE
```
Is CVE-2021-44228 detected in any of my clusters?
```

### CVE analysis in specific namespace
```
Check if CVE-2021-44228 is present in deployments in namespace "backend"
```

### Filter by cluster
```
Show me all deployments affected by CVE-2021-44228 in the dev-cluster
```

## Notes

- The server must be running before using Claude Code
- Make sure your API token has appropriate permissions for the operations you want to perform
- The default configuration uses `http://localhost:8080/mcp` - adjust if your server runs on a different port or host
