# Tasks: Add Think Tool

## Task 1: Create ThinkTool Implementation

**ID**: T1  
**Status**: completed  
**Owner**:  
**Blocked By**: -

### Description
Create the `ThinkTool` struct and its methods in `internal/tools/think.go`.

### Acceptance Criteria
- [x] `think.go` file created with `ThinkTool` struct
- [x] `Info()` method returns correct tool definition
- [x] `Run()` method parses arguments and returns thought content
- [x] `RequiresApproval()` returns `false`
- [x] File includes package documentation comment

### Notes
Reference: ShanClaw `internal/tools/think.go`

---

## Task 2: Write Unit Tests for ThinkTool

**ID**: T2  
**Status**: completed  
**Owner**:  
**Blocked By**: T1

### Description
Create comprehensive unit tests in `internal/tools/think_test.go`.

### Acceptance Criteria
- [x] Test valid thought returns success
- [x] Test empty thought returns error
- [x] Test invalid JSON returns error
- [x] Test large thought (1000+ chars) handled correctly
- [x] All tests pass with `go test ./internal/tools/...`

### Test Cases
```go
// Happy path
{"thought": "I need to plan this task"}

// Empty thought
{"thought": ""} → error

// Invalid JSON
"not json" → error

// Missing field
{} → error
```

---

## Task 3: Register ThinkTool

**ID**: T3  
**Status**: completed  
**Owner**:  
**Blocked By**: T1

### Description
Add `ThinkTool` registration to `internal/tools/register.go`.

### Acceptance Criteria
- [x] `&ThinkTool{}` added to `RegisterLocalTools()`
- [x] Registration order is logical (after file tools, before bash)
- [x] Import statement added if needed

### Notes
Place after file tools (file_read, file_write, file_edit) and before bash.

---

## Task 4: Write Integration Tests

**ID**: T4  
**Status**: completed  
**Owner**:  
**Blocked By**: T3

### Description
Add integration test to verify tool registration.

### Acceptance Criteria
- [x] Test that `think` tool is in the registry
- [x] Test that tool info is correctly formatted
- [x] Test passes with `go test ./tests/...`

### Notes
Can add to existing `tests/integration_test.go` or create new test.

---

## Task 5: Update System Prompt (Optional)

**ID**: T5  
**Status**: completed  
**Owner**:  
**Blocked By**: T3

### Description
Optionally update the system prompt to encourage think tool usage.

### Acceptance Criteria
- [x] System prompt includes guidance on using think tool
- [x] Guidance is concise and actionable

### Notes
This is optional and can be done if there's a system prompt configuration.

---

## Dependencies

```
T1 (Implement)
  └── T2 (Unit Tests)
  └── T3 (Register)
       └── T4 (Integration Tests)
       └── T5 (System Prompt - optional)
```

## Estimated Effort

- T1: 30 minutes
- T2: 45 minutes
- T3: 15 minutes
- T4: 30 minutes
- T5: 15 minutes (optional)

**Total**: ~2 hours
