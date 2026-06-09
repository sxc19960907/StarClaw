# Daemon supervision app launcher implementation plan

## Checklist

1. Update the macOS shell source.
   - Add `DaemonSupervisor` as an `ObservableObject`.
   - Resolve `starclaw` binary from `ASTRIA_STARCLAW_BIN`, bundled resources,
     or `/usr/bin/env starclaw`.
   - Probe `/health` before spawning.
   - Spawn `starclaw daemon start` only when no healthy daemon is found.
   - Poll health until ready or timeout.
   - Track child process exit after launch.
2. Update the shell UI.
   - Show starting/attached/failure/crash state without hiding the Web UI when
     attach succeeds.
   - Keep diagnostics action available on failure.
   - Load Web UI only after the supervisor reaches attached state.
3. Update build/smoke scripts.
   - Keep unsigned local app build working.
   - Add a supervision smoke that builds a temporary `starclaw` binary, points
     `ASTRIA_STARCLAW_BIN` at it, starts the app or exercises a testable
     supervisor path, and verifies daemon health.
   - Do not require private signing credentials.
4. Update docs.
   - `desktop/macos/Astria/README.md`
   - `docs/INSTALL.md`
   - `docs/RELEASE.md`
5. Validate.
   - Trellis task validation.
   - macOS shell smoke.
   - targeted app/doctor Go tests.
   - full Go tests.

## Validation Commands

- `python3 ./.trellis/scripts/task.py validate .trellis/tasks/06-09-daemon-supervision-app-launcher`
- `scripts/smoke_macos_astria_shell.sh`
- `go test ./cmd -run 'Test.*App|Test.*Doctor' -count=1`
- `go test ./...`
- `git diff --check`

If implementation adds a dedicated supervisor smoke script, run it before the
existing shell smoke.

## Risk Points

- `Process` launch must not block the Swift main thread.
- The shell must not spawn a duplicate daemon when `/health` already reports
  healthy.
- Startup failures must be visible and actionable, not blank-window failures.
- Do not terminate an already-running daemon that the shell did not start.
- Do not add broad App Transport Security exceptions beyond local networking.
- Avoid depending on generated `build/` artifacts in committed code.
