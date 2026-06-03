# Release Readiness Diagnostics Design

## Scope

Add a top-level CLI command:

```bash
starclaw doctor
```

The command is read-only. It reports local runtime context, existing local doctor checks, and daemon readiness when the daemon is reachable.

## Boundaries

- `cmd/doctor.go` owns Cobra wiring, output formatting, and daemon HTTP probes.
- `internal/tui/doctor.go` remains the source of local check logic.
- Existing daemon endpoints remain unchanged:
  - `GET /health`
  - `GET /status`
  - `GET /diagnostics`
- GUI Version/support panels remain unchanged for this task.

## Output Contract

The command prints:

- Version.
- Launch command: `starclaw app`.
- Web UI URL.
- Diagnostics URL.
- Data directory.
- Config path.
- Daemon state: running or not running.
- Local check rows from `tui.Doctor`.
- Daemon diagnostics summary and checks when reachable.
- Action hints when not reachable.

## Error Handling

- Unreachable daemon is not a command failure. New installs may not have the daemon running yet.
- Malformed daemon responses are warnings in the output and should not prevent local checks from printing.
- The command exits non-zero only for unexpected local command failures that prevent output generation.

## Security

- Do not print config secrets.
- Only print paths, version, status, summaries, and redacted daemon diagnostics.

## Compatibility

- Existing `starclaw app --check` remains a compact launch-readiness command.
- `starclaw doctor` becomes the broader support/debug command.
