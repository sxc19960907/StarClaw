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
