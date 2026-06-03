# Release readiness diagnostics

## Goal

Provide a single CLI diagnostics entry point that helps a newly installed StarClaw user verify local readiness and find the GUI/daemon support surfaces.

## Requirements

- Add a top-level `starclaw doctor` command.
- Reuse the existing local doctor checks from `internal/tui` instead of creating duplicate check logic.
- Print version, data/config paths, app launch command, Web UI URL, diagnostics URL, and daemon status.
- When the daemon is reachable, include HTTP readiness from `/health`, `/status`, and `/diagnostics`.
- When the daemon is not reachable, keep the command successful and print actionable next commands.
- Do not expose plaintext API keys or sensitive config values.
- Keep GUI behavior unchanged unless needed to align support text.
- Cover the command with unit tests and CLI smoke coverage.

## Acceptance Criteria

- [x] `starclaw doctor` appears in CLI help.
- [x] `starclaw doctor` exits 0 without a configured daemon.
- [x] `starclaw doctor` prints local checks, daemon state, `starclaw app`, Web UI URL, diagnostics URL, data dir, and config path.
- [x] Unit tests cover daemon reachable and daemon unavailable output.
- [x] `scripts/smoke_cli.sh` validates the doctor command.
- [x] Targeted Go tests and CLI smoke pass.

## Notes

- Existing release readiness surfaces include `starclaw app --check`, `/version`, `/diagnostics`, GUI Settings -> Version, and release install smoke.
- This task should connect those surfaces through a more discoverable CLI command rather than replacing them.
