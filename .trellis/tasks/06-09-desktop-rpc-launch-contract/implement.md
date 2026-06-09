# Desktop RPC launch contract implementation plan

## Checklist

1. Inspect current daemon start command and Desktop RPC listener lifecycle.
2. Add package-level variables for daemon start RPC socket/pidfile flags, wired
   in `init`.
3. Add validation helper for paired flags.
4. In `daemonStartCmd`, when desktop mode is enabled:
   - create listener config with server broker and default platform metadata;
   - attach listener to server;
   - run listener under daemon context before HTTP server readiness.
5. Add tests.
   - Missing socket/pidfile pair.
   - Listener writes pidfile and status redacts paths.
   - Existing app launch tests still pass.
6. Update docs/spec.
7. Validate.

## Validation Commands

- `python3 ./.trellis/scripts/task.py validate .trellis/tasks/06-09-desktop-rpc-launch-contract`
- `go test ./cmd ./internal/daemon ./internal/daemon/desktop_rpc -count=1`
- `go test ./cmd -run 'Test.*App|Test.*Doctor' -count=1`
- `go test ./...`
- `git diff --check`

## Risk Points

- Avoid global flag state leaking between tests.
- Do not start HTTP server if Desktop RPC listener cannot start in desktop
  mode.
- Do not expose socket/pidfile paths in status or diagnostics.
- Do not change default `starclaw daemon start` behavior.
