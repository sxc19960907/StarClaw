# Specification: Think Tool

## Overview

| Field | Value |
|-------|-------|
| Name | `think` |
| Type | Local Tool |
| Approval Required | No |
| Read-Only | Yes |

## Purpose

Provides a dedicated mechanism for the AI to engage in explicit reasoning and planning before taking action. The think tool serves as a scratchpad for the model's internal monologue.

## Interface

### Input Parameters

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `thought` | string | Yes | The reasoning or plan content |

### Input Schema

```json
{
  "type": "object",
  "properties": {
    "thought": {
      "type": "string",
      "description": "Your reasoning or plan"
    }
  },
  "required": ["thought"]
}
```

### Output

| Field | Type | Description |
|-------|------|-------------|
| `content` | string | The thought content, returned unchanged |
| `is_error` | boolean | `false` on success, `true` on validation error |

### Success Response

Returns the `thought` string as the content.

### Error Responses

| Error | Condition |
|-------|-----------|
| Validation Error | `thought` is empty or missing |
| Parse Error | Arguments are not valid JSON |

## Examples

### Example 1: Task Planning

**Input:**
```json
{
  "thought": "I need to refactor this code. First, I'll find all usages of OldName, then rename them to NewName. Finally, I'll run tests to verify."
}
```

**Output:**
```
I need to refactor this code. First, I'll find all usages of OldName, then rename them to NewName. Finally, I'll run tests to verify.
```

### Example 2: Error Case

**Input:**
```json
{
  "thought": ""
}
```

**Output:**
```
[validation error] thought is required
```

## Behavior

1. Tool validates that `thought` is present and non-empty
2. On validation failure, returns error result
3. On success, returns thought content unchanged
4. No side effects - purely pass-through operation
5. Requires no user approval

## Design Rationale

- **Simple implementation**: No complex logic, just validation and echo
- **Explicit signal**: Model uses a tool call to indicate reasoning vs. other content
- **No approval needed**: Read-only operation with no side effects
- **Required field**: Forces model to provide actual content, not empty calls

## Future Enhancements

- Add optional `plan` field for structured planning
- Support thought categorization (reasoning, planning, analysis)
- UI rendering with collapsible sections
