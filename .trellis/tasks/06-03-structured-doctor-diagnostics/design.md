# Structured Doctor Diagnostics Design

## Scope

Extend the existing `starclaw doctor` command with:

```bash
starclaw doctor --json
```

## Data Model

Add internal command-layer response structs in `cmd/doctor.go`:

- `doctorReport`
- `doctorLocalCheck`
- `doctorDaemonReport`
- `doctorDaemonStatusReport`
- `doctorDaemonDiagnosticsReport`

The plain-text output should render from the same `doctorReport` used by JSON output.

## JSON Contract

Top-level fields:

- `version`
- `launch_command`
- `web_url`
- `diagnostics_url`
- `starclaw_dir`
- `config_path`
- `local_checks`
- `daemon`

Daemon fields:

- `running`
- `status`
- `diagnostics`
- `errors`

`errors` is for non-fatal probe failures such as a malformed `/status` response.

## Error Handling

- Unreachable daemon is a normal report state, not a command failure.
- JSON encoding failures return an error.
- HTTP probe failures are captured in `daemon.errors`.

## Security

The report contains only readiness metadata already visible in existing text output or daemon diagnostics. It must not include API keys.
