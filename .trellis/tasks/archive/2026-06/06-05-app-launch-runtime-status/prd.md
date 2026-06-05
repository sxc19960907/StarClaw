# Polish App Launch Runtime Status

## Problem

StarClaw already has `starclaw app`, `app --check`, `app --no-open`, `doctor`, daemon runtime APIs, a Version page, and release install smoke coverage. The remaining gap is consistency: launch readiness output, runtime JSON, GUI Version details, README, and smoke checks do not all expose the same startup/status context.

## User Value

A user who installs StarClaw should have one obvious path to launch the GUI and one consistent set of runtime facts to diagnose whether daemon + GUI are healthy.

## Confirmed Facts

- `starclaw app` starts or reuses the daemon and opens `/app/`.
- `starclaw app --no-open` starts/reuses daemon without opening a browser.
- `starclaw app --check` prints readiness without starting the daemon.
- `/version` and `/diagnostics` already include launch command, web URL, data/config paths, and diagnostic context.
- GUI Version page already shows release readiness, runtime context, update command, health/status/diagnostics URLs, data path, and config path.
- Release install smoke and app launch smoke already exercise `app --check`, `app --no-open`, `/version`, `/diagnostics`, and `doctor --json`.

## Requirements

- Make `app --check` print the same runtime endpoints users see in the GUI: Web UI, Health, Status API, Diagnostics, Data, and Config.
- Make `/version` and `/diagnostics` include any missing status endpoint fields needed for this consistency.
- Make the GUI Version page expose the readiness/status fields clearly and keep smoke coverage aligned.
- Update README launch section to describe the consistent readiness/status flow.
- Preserve existing `starclaw app` and `doctor` behavior.

## Out of Scope

- Changing daemon port selection.
- Adding an installer/updater implementation.
- Adding GUI auto-update.
- Changing CI to run targeted tool-call smoke by default.

## Acceptance Criteria

- [x] `starclaw app --check` prints Health, Status API, Diagnostics, Data, and Config.
- [x] `/version` JSON exposes `status_url`, `health_url`, `diagnostics_url`, `web_url`, `starclaw_dir`, and `config_path`.
- [x] `/diagnostics` JSON exposes matching launch/runtime context.
- [x] GUI Version smoke verifies the readiness/runtime context shown to users.
- [x] App launch and release install smoke verify the expanded readiness output.
- [x] README documents `app`, `app --no-open`, `app --check`, and `doctor` as the launch/status path.
- [x] `scripts/smoke_app_launch.sh`, `scripts/smoke_release_install.sh`, and `scripts/smoke_webui_core.sh` pass locally.
- [x] `go test ./...`, `go vet ./...`, and `git diff --check` pass locally.
