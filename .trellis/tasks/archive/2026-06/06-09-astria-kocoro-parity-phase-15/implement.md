# Astria Kocoro parity phase 15 implementation plan

## Checklist

1. `desktop-rpc-session-lifecycle`
   - Identify the current one-shot Desktop RPC client and daemon status seams.
   - Add app-side session lifecycle state after `system.capabilities`.
   - Add bounded reconnect/retry behavior for recoverable disconnects.
   - Preserve degraded HTTP fallback and mismatch behavior.
   - Add smoke/unit coverage for connected, disconnected, retry, and fallback
     states.
2. `desktop-rpc-event-monitoring`
   - Choose the smallest local event-monitoring contract compatible with the
     existing Desktop RPC frame/broker code.
   - Add daemon/app event state or subscription plumbing.
   - Test event delivery or event-state transitions without external services.
3. `native-desktop-diagnostics-recovery`
   - Surface session/event state in Astria diagnostics and visible recovery UX.
   - Keep paths redacted and unsafe cleanup boundaries intact.
   - Add smoke coverage for user-visible degraded/recovery states.
4. Update code-spec.
   - Record session state names, retry bounds, event-monitoring contract, and
     diagnostics redaction rules.
5. Final review.
   - Update Kocoro parity estimate and remaining native integration gaps.

## Validation Commands

- `python3 ./.trellis/scripts/task.py validate .trellis/tasks/06-09-astria-kocoro-parity-phase-15`
- `scripts/smoke_macos_astria_shell.sh`
- `go test ./cmd ./internal/daemon ./internal/daemon/desktop_rpc -count=1`
- `go test ./cmd -run 'Test.*App|Test.*Doctor' -count=1`
- `go test ./...`
- `git diff --check`

## Risk Points

- Do not regress Phase14's launch-time capabilities validation.
- Do not make CLI/browser flows depend on a desktop client.
- Do not introduce unbounded reconnect loops or orphaned Swift tasks.
- Do not expose runtime socket/pidfile paths in status or diagnostics.
- Do not implement cloud lifecycle, Shannon auth, or off-machine telemetry.

## Review Gate

Start with `desktop-rpc-session-lifecycle`. Do not add event-monitoring APIs
until the long-lived session lifecycle is implemented and verified.
