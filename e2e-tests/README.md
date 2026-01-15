# StackRox MCP E2E Testing

This directory contains end-to-end tests for the StackRox MCP server using the [mcp-testing-framework](https://github.com/L-Qun/mcp-testing-framework).

## Prerequisites

1. **OpenAI API Key**: Required for running the AI model tests
   - Get your key from Bitwarden

2. **StackRox API Token**: Required for connecting to StackRox Central
   - Generate from StackRox Central UI: Integrations > API Token > Generate Token

## Setup

### 1. Configure Environment Variables

Create a `.env` file with your credentials:

```bash
# OpenAI API key for running tests
OPENAI_API_KEY=sk-your-openai-key-here

# StackRox API Token for accessing Central
STACKROX_API_TOKEN=your-stackrox-api-token-here
```

### 2. Update Server Configuration (Optional)

Edit `mcp-testing-framework.yaml` if you need to change the StackRox Central URL:


## Running Tests

From the `e2e-tests` directory, run:

```bash
npx mcp-testing-framework@latest evaluate
```

This will:
- Spawn the StackRox MCP server in stdio mode
- Run test cases against the configured AI models (GPT-5 and GPT-5-mini)
- Generate a test report in the `mcp-reports/` directory

## Test Configuration

The `mcp-testing-framework.yaml` file controls the test behavior:

- **testRound**: Number of times each test runs (default: 3)
- **passThreshold**: Minimum success rate (0.5 = 50%)
- **modelsToTest**: AI models to test (currently: `gpt-5`, `gpt-5-mini`)
- **testCases**: 8 test scenarios covering CVE queries and cluster listing
- **mcpServers**: Server configuration using stdio transport

## Customizing Tests

### Add More Test Cases

Add new test cases to `mcp-testing-framework.yaml`:
Use the JSON report to analyze which prompts work best with each model.
