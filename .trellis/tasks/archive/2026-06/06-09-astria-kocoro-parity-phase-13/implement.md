# Astria Kocoro parity phase 13 implementation plan

## Checklist

1. Plan and skeleton the standalone desktop shell.
   - Choose first-shell technology and repository layout.
   - Define app-to-daemon launch/attach contract.
   - Keep existing CLI/browser launch unchanged.
2. Implement daemon supervision in the app launcher.
   - Start or attach to daemon.
   - Monitor `/health`, `/status`, and diagnostics.
   - Add explicit states for port conflict, stale daemon, version mismatch, and
     startup timeout.
3. Implement desktop window recovery.
   - Restore last window URL/state.
   - Recover after Web UI reload and EventSource reconnect.
   - Surface daemon crash/restart states without data loss.
4. Define packaging, signing, and update boundary.
   - Document local dev build.
   - Document release artifact shape.
   - Add smoke checks that can run without private signing credentials.
5. Close the phase.
   - Archive all children.
   - Record final Kocoro gap review.

## Validation Commands

- `python3 ./.trellis/scripts/task.py validate .trellis/tasks/06-09-astria-kocoro-parity-phase-13`
- `go test ./...`
- Existing smoke checks that remain relevant:
  - `scripts/smoke_app_launch.sh`
  - `scripts/smoke_release_local.sh`
  - `scripts/smoke_webui_core.sh`

Native app build and signing validation will be specified by the child that
introduces the app project.

## Risk Points

- Do not break current CLI/npm/binary installation paths while adding a
  desktop app.
- Do not hide daemon startup failures behind silent retry loops.
- Avoid coupling pidfile and socket paths by implicit filename derivation.
- Keep the local-first boundary explicit: no cloud telemetry or remote
  lifecycle transport by default.
- Keep private signing/notarization credentials out of the repository.
