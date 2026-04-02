# mcpchecker v0.0.14 Bug Analysis

## Summary

E2E tests are failing after upgrading from mcpchecker v0.0.12 to v0.0.14 due to an issue where the OpenAI mock agent makes tool calls but doesn't send a final AgentMessageChunk update. This causes llmJudge verification to fail with:

```
cannot run llmJudge step before agent (must be in verification)
```

## Affected Tests

- `cve-cluster-does-exist` - ❌ Makes 1 tool call, no final message
- `cve-cluster-does-not-exist` - ❌ Makes 1 tool call, no final message
- `cve-log4shell` - ❌ Makes 3 tool calls (including failing node call), no final message
- `cve-nonexistent` - ⚠️ Makes 4 tool calls, has final message but judge fails for other reasons

Working tests like `cve-detected-clusters`, `cve-multiple`, etc. all have a final message step.

## Root Cause

### Normal Flow (Working)
1. Agent calls tool via ACP ToolCall update
2. Tool executes and returns result via ToolCallUpdate
3. **OpenAI mock sends follow-up chat completion with "Evaluation complete."**
4. llmagent converts response to AgentMessageChunk via OnStepFinish callback
5. ExtractOutputSteps produces steps including final "message" type
6. llmJudge can evaluate the final message

### Broken Flow (Failing Tests)
1. Agent calls tool via ACP ToolCall update
2. Tool executes and returns result via ToolCallUpdate
3. **No follow-up message is sent** (or fantasy doesn't call OnStepFinish)
4. ExtractOutputSteps produces only "tool_call" type steps
5. FinalMessageFromSteps returns empty string
6. llmJudge fails because `input.Agent.Output == ""`

## Technical Details

### OpenAI Mock Server Behavior

The mock in `functional/servers/openai/server.go` is supposed to send a follow-up message:

```go
// If request contains tool result messages, this is a follow-up after a tool call.
// Return a simple text response to end the agentic loop.
for _, msg := range req.Messages {
    if msg.Role == "tool" {
        followUp := &ChatCompletionResponse{
            // ...
            Message: Message{
                Role:    "assistant",
                Content: "Evaluation complete.",
            },
            FinishReason: "stop",
        }
        // ...
    }
}
```

### llmagent ACP Agent

The `acp_agent.go` processes OpenAI responses via fantasy's OnStepFinish:

```go
OnStepFinish: func(step fantasy.StepResult) error {
    text := step.Response.Content.Text()
    if text == "" {
        return nil  // ← Early return if no text!
    }

    return a.conn.SessionUpdate(promptCtx, acp.SessionNotification{
        SessionId: params.SessionId,
        Update:    acp.UpdateAgentMessageText(text),
    })
},
```

If `step.Response.Content.Text()` is empty, no AgentMessageText update is sent.

### llmJudge Validation

The llmJudge step validates agent output in `pkg/steps/llm_judge.go:88-90`:

```go
if input.Agent == nil || input.Agent.Prompt == "" || input.Agent.Output == "" {
    return nil, fmt.Errorf("cannot run llmJudge step before agent (must be in verification)")
}
```

## Reproduction Test

Added test in mcpchecker repo:

```go
// TestAgentWithOnlyToolCallsNoFinalMessage reproduces issue #268
func TestAgentWithOnlyToolCallsNoFinalMessage(t *testing.T) {
    updates := []acp.SessionUpdate{
        {
            ToolCall: &acp.SessionUpdateToolCall{
                ToolCallId: "call-1",
                Title:      "get_clusters_with_orchestrator_cve",
                // ...
            },
        },
        {
            ToolCallUpdate: &acp.SessionToolCallUpdate{
                ToolCallId: "call-1",
                Status:     ptr(acp.ToolCallStatusCompleted),
                // ...
            },
        },
        // BUG: No AgentMessageChunk update here!
    }

    steps := agent.ExtractOutputSteps(updates)
    assert.Len(t, steps, 1, "Only has tool_call, no message")

    finalMessage := agent.FinalMessageFromSteps(steps)
    assert.Empty(t, finalMessage)  // ← Fails llmJudge validation
}
```

To run: `cd /tmp/mcpchecker && go test -v -run TestAgentWithOnlyToolCallsNoFinalMessage ./pkg/agent/`

## Hypothesis

The issue may be related to:

1. **fantasy library behavior change** - The charm.land/fantasy package was updated in v0.0.13 (bump from 0.16.0 to 0.17.1). The OnStepFinish callback might not be called in all scenarios.

2. **OpenAI streaming response handling** - The mock server's streaming implementation might not be properly triggering OnStepFinish for the follow-up message after tool results.

3. **ACP protocol handling** - The conversion from OpenAI chat completion responses to ACP SessionUpdate messages might have edge cases.

## Investigation Steps

1. ✅ Reproduced issue with unit test
2. ✅ Identified that FinalMessageFromSteps returns empty for failing tests
3. ✅ Traced to missing AgentMessageChunk updates in SessionUpdate stream
4. ⏭️ **TODO**: Check if fantasy v0.17.1 has breaking changes in OnStepFinish behavior
5. ⏭️ **TODO**: Add debug logging to llmagent acp_agent.go to see if OnStepFinish is called
6. ⏭️ **TODO**: Test with real OpenAI API instead of mock to confirm it's a mock issue
7. ⏭️ **TODO**: Review PR #268 discussion on mcpchecker repo for context

## Workaround

For now, we've:
1. Migrated all tasks to v1alpha2 format (required by v0.0.14)
2. Fixed wiremock ExportNodeResponse fixture format
3. Waiting to see if these fixes resolve the remaining failures

If failures persist, we may need to:
- Downgrade to mcpchecker v0.0.12 temporarily
- Report bug upstream to mcpchecker with reproduction test
- Investigate fantasy library update as potential cause

## Related Links

- mcpchecker PR #268: https://github.com/mcpchecker/mcpchecker/pull/268
- Our PR #102: https://github.com/stackrox/stackrox-mcp/pull/102
- Test run: https://github.com/stackrox/stackrox-mcp/actions/runs/23899405760
