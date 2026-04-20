# Design: Think Tool

## Overview

The think tool is a simple pass-through tool that accepts a `thought` string and returns it unchanged. Its purpose is semantic - it signals that the model is engaging in explicit reasoning.

## Architecture

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Agent     │────▶│  ThinkTool  │────▶│   Result    │
│   Loop      │     │             │     │             │
└─────────────┘     └─────────────┘     └─────────────┘
                           │
                           ▼
                    ┌─────────────┐
                    │  thought    │
                    │  (string)   │
                    └─────────────┘
```

## Implementation

### Tool Definition

```go
// internal/tools/think.go
package tools

import (
    "context"
    "encoding/json"
    "fmt"

    "github.com/starclaw/starclaw/internal/agent"
)

type ThinkTool struct{}

type thinkArgs struct {
    Thought string `json:"thought"`
}

func (t *ThinkTool) Info() agent.ToolInfo {
    return agent.ToolInfo{
        Name:        "think",
        Description: "Use this to plan or reason through complex multi-step tasks before acting. Always use this instead of outputting plans as plain text.",
        Parameters: map[string]any{
            "type": "object",
            "properties": map[string]any{
                "thought": map[string]any{
                    "type":        "string",
                    "description": "Your reasoning or plan",
                },
            },
        },
        Required: []string{"thought"},
    }
}

func (t *ThinkTool) Run(ctx context.Context, argsJSON string) (agent.ToolResult, error) {
    var args thinkArgs
    if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
        return agent.ToolResult{
            Content: fmt.Sprintf("invalid arguments: %v", err),
            IsError: true,
        }, nil
    }
    if args.Thought == "" {
        return agent.ToolResult{
            Content: "thought is required",
            IsError: true,
        }, nil
    }
    return agent.ToolResult{Content: args.Thought}, nil
}

func (t *ThinkTool) RequiresApproval() bool { return false }
```

### Registration

Add to `internal/tools/register.go`:

```go
func RegisterLocalTools() *agent.ToolRegistry {
    reg := agent.NewToolRegistry()

    // ... existing tools ...
    reg.Register(&ThinkTool{})  // Add this line

    return reg
}
```

### System Prompt Update

Consider adding to the system prompt:

```
When facing complex multi-step tasks, use the `think` tool first to plan your approach.
This helps organize your reasoning before taking action.
```

## Testing Strategy

### Unit Tests

1. **Happy path**: Valid thought string returns successfully
2. **Empty thought**: Returns validation error
3. **Invalid JSON**: Returns parse error
4. **Large thought**: Handles long strings (up to reasonable limit)

### Integration Tests

1. Tool is registered in the registry
2. Tool info is correctly formatted
3. Tool is included in LLM tool definitions

## Error Handling

| Error Case | Behavior |
|------------|----------|
| Invalid JSON | Return validation error result |
| Empty thought | Return validation error result |
| Missing required field | Return validation error result |

## Security Considerations

- No approval required (read-only operation)
- No file system access
- No network access
- Input is only validated, not executed
