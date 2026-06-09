# Desktop RPC fallback recovery implementation plan

## Checklist

1. Add `AstriaRuntimeArtifacts`.
   - Track runtime directory, socket path, and pidfile path.
   - Validate owned artifact paths.
   - Detect live/dead/malformed pidfiles.
   - Clean stale pidfile/socket only inside the runtime directory.
2. Integrate launch preflight.
   - Before shell-launched daemon start, run safe stale artifact cleanup.
   - Do not delete artifacts for live pidfiles.
3. Add degraded fallback state.
   - Add `.degraded(String)` to `DaemonState`.
   - Treat degraded as attached so WebView remains mounted.
   - For already-running HTTP-healthy daemons with no/bad Desktop RPC, show
     fallback banner instead of blocking the UI.
4. Add smoke coverage.
   - Extend `--desktop-rpc-smoke` or add a fallback smoke to validate stale
     artifact cleanup and unsafe cleanup refusal.
   - Keep `scripts/smoke_macos_astria_shell.sh` covering this path.
5. Update docs/spec.
   - Record runtime artifact ownership, stale cleanup, and degraded fallback.
6. Validate.

## Validation Commands

- `python3 ./.trellis/scripts/task.py validate .trellis/tasks/06-09-desktop-rpc-fallback-recovery`
- `scripts/smoke_macos_astria_shell.sh`
- `go test ./cmd ./internal/daemon ./internal/daemon/desktop_rpc -count=1`
- `go test ./cmd -run 'Test.*App|Test.*Doctor' -count=1`
- `go test ./...`
- `git diff --check`

## Risk Points

- Do not remove live daemon artifacts.
- Do not remove files outside the configured runtime directory.
- Do not fail the user out of the Web UI when HTTP health is available and the
  daemon was not launched by this shell session.
- Do not hide shell-launched Desktop RPC failure as full readiness.
