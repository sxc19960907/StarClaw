# Permissions System

> Tool-level security: 4-layer model that evaluates every tool call before execution.

---

## Architecture

```
LLM tool call → executeTool() → permissions.CheckToolCall(name, args, cfg)
                                    │
                          ┌─────────┼─────────┐
                     bash │  file_* │  http   │ others
                          │         │         │
                   CheckCommand  CheckFilePath  CheckNetwork  (pass)
```

## 4-Layer Security Model

| Layer | Priority | Source | Effect |
|-------|----------|--------|--------|
| **Hard-block** | 1 (highest) | Built-in | Always deny: `rm -rf /`, `curl \| sh` |
| **User denied** | 2 | Config `denied_commands` | Patterns to always reject |
| **User allowed** | 3 | Config `allowed_commands` | Explicit whitelist |
| **Default safe** | 4 (lowest) | Built-in | ~120 common safe commands |
| **Ask** | fallback | — | Prompt user for approval |

## Decision Type

```go
type Decision string
const (
    Allow Decision = "allow"  // execute silently
    Deny  Decision = "deny"   // block with PermissionError
    Ask   Decision = "ask"    // prompt user (interactive mode)
)
```

## Integration

In `internal/agent/loop.go` `executeTool()`:

```go
if a.permsConfig != nil {
    decision, reason := permissions.CheckToolCall(toolUse.Name, string(toolUse.Input), a.permsConfig)
    if decision == permissions.Deny {
        return PermissionError(fmt.Sprintf("%s: blocked (%s)", toolUse.Name, reason))
    }
}
```

## Scenario: Daemon Tool Approval

### 1. Scope / Trigger

- Trigger: daemon or Web UI runs need to execute a tool call whose permission decision is `Ask`, or a tool declares `RequiresApproval()` and no permission rule explicitly allows it.
- Scope: local daemon approval only; CLI/TUI approval UX is separate.

### 2. Signatures

- Agent loop setter: `SetApprovalRequester(requester ApprovalRequester)`.
- Agent approval request:
  ```go
  type ApprovalRequest struct {
      Tool string
      Args string
      Reason string
  }
  ```
- Daemon API: `POST /approval` with `{"request_id":"apr_x","decision":"allow|deny"}`.
- Daemon events: `approval_needed`, `approval_resolved` on `GET /events`.

### 3. Contracts

- `permissions.Deny` always returns a permission error and does not ask.
- `permissions.Allow` bypasses explicit approval for the matching call.
- `permissions.Ask` must ask through the configured approval requester.
- If no approval requester is configured and a permission rule returns `Ask`, the tool must be denied, not silently executed.
- `RequiresApproval()` is enforced for daemon runs because the daemon injects an approval requester; non-daemon surfaces keep their existing behavior until they wire their own requester.

### 4. Validation & Error Matrix

- Unknown approval id -> `/approval` returns OK but no waiting call is resolved.
- Approval timeout -> deny.
- Request context cancellation -> deny and return `ctx.Err()` to the requester.
- User denies -> return `PermissionError`, do not execute the tool.

### 5. Good/Base/Bad Cases

- Good: Web UI receives `approval_needed`, user clicks Allow, tool executes.
- Base: Web UI is closed; request times out and denies.
- Bad: approval-needed tools run without a requester or explicit allow rule.

### 6. Tests Required

- Agent loop tests for approval allow and deny.
- Daemon requester tests for `approval_needed` and `approval_resolved` event payloads.
- API tests for `/approval` resolving a pending broker request.

### 7. Wrong vs Correct

#### Wrong

```go
if decision == permissions.Ask {
    // Treat as allow in daemon mode.
}
```

#### Correct

```go
if decision == permissions.Ask {
    decision, err := approver.RequestApproval(ctx, req)
    if err != nil || decision != ApprovalAllow {
        return PermissionError("denied by user")
    }
}
```

## Configuration

```yaml
permissions:
  allowed_dirs:
    - ~
    - .
    - /tmp
  allowed_commands:
    - "python3 myscript.py"
  denied_commands:
    - "shutdown*"
    - "reboot"
  network_allowlist:
    - "api.github.com"
    - "*.mycompany.com"
  sensitive_patterns:
    - "*.secret"
```

## Hard-Block Patterns (never overridable)

```
rm -rf /, rm -rf ~, rm -rf /System
> /dev/sd*, > /dev/disk*
mkfs.*, dd if=* of=/dev/*
curl * | sh, wget * | sh
```

## Default Safe Commands (no config needed)

Read-only system tools, build tools, linters: `ls`, `cat`, `grep`, `go build`, `go test`, `git status`, `golangci-lint`, `docker ps`, `gh pr list`, etc.

## Path Containment

File-oriented tools must validate paths before filesystem access. Expand `~`, convert to an absolute clean path, resolve symlinks or the nearest existing ancestor for new files, then check containment with relative-path logic such as `filepath.Rel`. Do not use raw string prefixes for containment; sibling paths like `/project-other` and symlink escapes must not pass as being under `/project`.

## Adding a New Permissions Handler

In `CheckToolCall()`, add a case for the tool name:

```go
case "my_new_tool":
    return checkMyNewThing(extractField(argsJSON, "key_field"), cfg)
```
