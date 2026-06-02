# Add GUI version and update status

## Goal

Improve release/install readiness by exposing StarClaw version, platform, daemon web URL, and update-check status from the daemon Web UI.

## Confirmed Facts

- The CLI already has `starclaw version` and `starclaw update --check`.
- The daemon already receives the build version through `daemon.NewServer(..., Version)`.
- The Web UI already displays a compact version metric from `/status`, but there is no full version/about page.
- The update package skips update checks for non-semver development builds.
- The Web UI is embedded static HTML/CSS/JS with no frontend build step.

## Requirements

- Add daemon API support for version/install metadata:
  - current version
  - semver/update support flag
  - current platform
  - daemon Web UI URL
  - update command hint
- Add a daemon update-check endpoint that checks GitHub releases only on user action.
- Development builds must not attempt network update checks and must clearly report that updates require a release build.
- Add a Version/About panel in the Web UI:
  - navigation entry
  - current version/platform/runtime data
  - update status and manual check button
  - command hint for CLI update
- Add backend tests and Web UI smoke coverage.

## Acceptance Criteria

- [x] `GET /version` returns structured version/install metadata.
- [x] `GET /update/check` returns a clear no-update response for `dev`.
- [x] The Web UI has a Version panel reachable from navigation.
- [x] The Version panel displays current version, platform, Web UI URL, and update command.
- [x] The Version panel can manually check updates and shows development-build status during smoke.
- [x] `node --check internal/daemon/webui/assets/app.js`, targeted Go tests, full Go tests, vet, and Web UI core smoke pass.

## Out of Scope

- Installing updates from the GUI.
- Changing release publishing or GoReleaser configuration.
- Adding a desktop app wrapper or system service installer.

## Goal

TBD.

## Requirements

- TBD

## Acceptance Criteria

- [ ] TBD

## Notes

- Keep `prd.md` focused on requirements, constraints, and acceptance criteria.
- Lightweight tasks can remain PRD-only.
- For complex tasks, add `design.md` for technical design and `implement.md` for execution planning before `task.py start`.
