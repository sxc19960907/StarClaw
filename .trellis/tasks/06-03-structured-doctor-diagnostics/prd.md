# Structured doctor diagnostics

## Goal

Make `starclaw doctor` usable by automation and release smoke scripts through a stable JSON output mode.

## Requirements

- Add `starclaw doctor --json`.
- JSON output must include version, launch URLs, data/config paths, local checks, daemon state, and daemon diagnostics when reachable.
- Plain-text `starclaw doctor` behavior must remain readable and backward compatible.
- Daemon unavailable must still exit 0 and serialize as `daemon.running=false`.
- Do not serialize config secrets or API keys.
- Update app launch and release install smoke scripts to validate doctor JSON.
- Keep implementation in the CLI layer and reuse existing local doctor checks.

## Acceptance Criteria

- [x] `starclaw doctor --json` exits 0 with no daemon running.
- [x] JSON contains `version`, `launch_command`, `web_url`, `diagnostics_url`, `starclaw_dir`, `config_path`, `local_checks`, and `daemon.running`.
- [x] When daemon is reachable, JSON includes daemon status and diagnostics summary/checks.
- [x] Existing plain-text `starclaw doctor` tests continue to pass.
- [x] `scripts/smoke_cli.sh`, `scripts/smoke_app_launch.sh`, and `scripts/smoke_release_install.sh` validate doctor output.
- [x] Targeted tests, available smoke scripts, and release install smoke syntax check pass.

## Notes

- This builds on the prior `starclaw doctor` command.
