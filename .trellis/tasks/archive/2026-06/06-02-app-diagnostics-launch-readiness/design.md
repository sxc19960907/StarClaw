# Design

## API Shape

- Extend diagnostics response with:
  - `web_url`
  - `launch_command`
  - `executable_path`
  - `starclaw_dir`
  - `config_path`
  - `agents_dir`
  - `sessions_dir`
- Extend version response with:
  - `launch_command`

Use existing server deps for data paths and `os.Executable()` for binary path. If executable resolution fails, omit it rather than failing diagnostics.

## GUI Shape

- Diagnostics panel gets a Launch readiness card above the existing checks.
- Version card adds rows for Web UI and launch command.
- Keep cards simple data rows, matching existing `row-item` / metadata styling.

## CLI Shape

- Preserve current `starclaw app` output on success.
- On daemon ensure failure, include a hint to run `starclaw daemon status` and inspect `/diagnostics`.

## Testing

- Add/adjust daemon API tests for diagnostics/version fields.
- Extend core Web UI smoke route checks and browser assertions for launch metadata.
- Run CLI smoke to cover help/launch command presence.
