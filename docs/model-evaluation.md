# LLM Model Evaluation Results

## Overview

This document tracks evaluation results of LLM models used with the StackRox MCP server. Evaluations measure how well a model selects the correct MCP tools, passes appropriate parameters, stays within expected tool call bounds, and produces accurate responses.

All evaluations use the [mcpchecker](https://github.com/mcpchecker/mcpchecker) framework against a deterministic WireMock-based mock backend, ensuring reproducible results across runs.

## Evaluation Methodology

### Test Framework

Evaluations are run using **mcpchecker**, configured in [`e2e-tests/mcpchecker/eval.yaml`](../e2e-tests/mcpchecker/eval.yaml). The framework:

1. Sends a natural language prompt to the model under test
2. The model interacts with the MCP server (tool calls, parameter selection)
3. Assertions validate tool usage against expected behavior
4. An LLM judge evaluates response quality against reference answers

### Test Environment

- **Backend**: WireMock mock server with deterministic fixtures (no live StackRox Central required)
- **MCP Config**: [`e2e-tests/mcpchecker/mcp-config-mock.yaml`](../e2e-tests/mcpchecker/mcp-config-mock.yaml)
- **Task definitions**: [`e2e-tests/mcpchecker/tasks/`](../e2e-tests/mcpchecker/tasks/)

### Assertions

Each task defines assertions from the following set:

| Assertion | Description |
|-----------|-------------|
| `toolsUsed` | Required tool(s) must be called, optionally with matching arguments (`argumentsMatch`) |
| `minToolCalls` | Minimum total tool calls across all tools |
| `maxToolCalls` | Maximum total tool calls (prevents runaway tool usage) |

A task passes when **all** its assertions pass **and** the LLM judge approves the response.

## Evaluation Results

<!-- model:gpt-5-mini start -->

### gpt-5-mini — 2026-03-31

**Overall: 10/11 tasks passed (90%)**

#### Task Results

| # | Task | Result | toolsUsed | minCalls | maxCalls | Input Tokens | Output Tokens |
|---|------|--------|-----------|----------|----------|--------------|---------------|
| 1 | list-clusters | Pass | Pass | Pass | Pass | 1728 | 962 |
| 2 | cve-detected-workloads | Pass | Pass | Pass | Pass | 565 | 1187 |
| 3 | cve-detected-clusters | Pass | **Fail** | Pass | Pass | 640 | 1998 |
| 4 | cve-nonexistent | Pass | Pass | Pass | Pass | 1077 | 2605 |
| 5 | cve-cluster-does-exist | **Fail** | Pass | Pass | Pass | 539 | 1285 |
| 6 | cve-cluster-does-not-exist | Pass | **Fail** | Pass | Pass | 1528 | 1324 |
| 7 | cve-clusters-general | Pass | Pass | Pass | Pass | 796 | 2304 |
| 8 | cve-cluster-list | Pass | Pass | Pass | Pass | 488 | 1917 |
| 9 | cve-log4shell | Pass | Pass | Pass | Pass | 1008 | 2936 |
| 10 | cve-multiple | Pass | Pass | Pass | Pass | 1142 | 2493 |
| 11 | rhsa-not-supported | Pass | — | Pass | Pass | 650 | 2488 |

**Total input tokens**: 10161 | **Total output tokens**: 21499

<!-- model:gpt-5-mini end -->

<!-- model:gpt-5 start -->

### gpt-5 — 2026-03-31

**Overall: 9/11 tasks passed (81%)**

#### Task Results

| # | Task | Result | toolsUsed | minCalls | maxCalls | Input Tokens | Output Tokens |
|---|------|--------|-----------|----------|----------|--------------|---------------|
| 1 | list-clusters | Pass | Pass | Pass | Pass | 1720 | 552 |
| 2 | cve-detected-workloads | Pass | Pass | Pass | Pass | 1589 | 1003 |
| 3 | cve-detected-clusters | Pass | Pass | Pass | Pass | 521 | 1702 |
| 4 | cve-nonexistent | **Fail** | Pass | Pass | Pass | 2406 | 2085 |
| 5 | cve-cluster-does-exist | Pass | Pass | Pass | Pass | 1563 | 1682 |
| 6 | cve-cluster-does-not-exist | **Fail** | **Fail** | Pass | Pass | 504 | 1868 |
| 7 | cve-clusters-general | Pass | Pass | Pass | Pass | 516 | 1477 |
| 8 | cve-cluster-list | Pass | Pass | Pass | Pass | 706 | 1964 |
| 9 | cve-log4shell | Pass | Pass | Pass | Pass | 1008 | 2304 |
| 10 | cve-multiple | Pass | Pass | Pass | Pass | 2166 | 2492 |
| 11 | rhsa-not-supported | Pass | — | Pass | Pass | 818 | 2187 |

**Total input tokens**: 13517 | **Total output tokens**: 19316

<!-- model:gpt-5 end -->

## How to Run Evaluations

### Prerequisites

- Go 1.25+
- LLM judge credentials configured via environment variables (see below)

### Running an Evaluation

1. **Configure the agent model** via environment variable or in `e2e-tests/mcpchecker/eval.yaml`:

   ```bash
   export MODEL_NAME=gpt-5-nano
   ```

2. **Set judge environment variables**:

   ```bash
   export JUDGE_TYPE=openai
   export JUDGE_API_KEY=<your-key>
   export JUDGE_MODEL_NAME=<judge-model>
   ```

3. **Run the evaluation**:

   ```bash
   make e2e-test
   ```

4. **Update this document** with the results:

   ```bash
   ./scripts/update-model-evaluation.sh \
     --model-id <model-id> \
     --results e2e-tests/mcpchecker/mcpchecker-stackrox-mcp-e2e-out.json
   ```

   The script generates a markdown section with the task results table and
   inserts or updates it in this document using HTML comment markers.

   If results for the given `--model-id` already exist, the script replaces
   the existing section. Otherwise, it appends a new section.
