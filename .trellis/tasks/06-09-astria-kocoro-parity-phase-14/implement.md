# Astria Kocoro parity phase 14 implementation plan

## Checklist

1. `desktop-rpc-launch-contract`
   - Add paired `--rpc-socket` and `--rpc-pidfile` flags to `starclaw daemon
     start`.
   - Wire daemon start to create `desktop_rpc.Listener` when both flags are
     present.
   - Preserve HTTP-only mode when neither flag is present.
   - Add CLI/unit tests for missing pair, listener startup, pidfile after
     listen, and `/status` Desktop RPC state.
2. `desktop-rpc-capabilities-reconciliation`
   - Add Swift Desktop RPC socket client frame support for the minimal system
     methods.
   - Make Astria launch daemon with deterministic socket/pidfile paths.
   - Call `system.capabilities` and validate protocol/method/version
     compatibility.
   - Add smoke coverage for successful handshake and mismatch states.
3. `desktop-rpc-fallback-recovery`
   - Add stale pidfile/socket checks scoped to Astria runtime paths.
   - Add user-visible states and retry/fallback behavior for broken socket,
     stale pidfile, mismatch, and disconnect.
   - Update diagnostics/status docs and smoke coverage.
4. Update code-spec.
   - Record command flags, socket/pidfile contracts, cleanup rules, and
     handshake validation matrix.
5. Validate each child.
   - Trellis validation.
   - Relevant Go unit tests.
   - macOS shell smoke where Swift changes are involved.
   - `go test ./...`.
   - `git diff --check`.

## Validation Commands

- `python3 ./.trellis/scripts/task.py validate .trellis/tasks/06-09-astria-kocoro-parity-phase-14`
- `go test ./cmd ./internal/daemon ./internal/daemon/desktop_rpc -count=1`
- `scripts/smoke_macos_astria_shell.sh`
- `go test ./...`
- `git diff --check`

## Risk Points

- Do not make ordinary daemon start require a desktop socket.
- Do not silently ignore a half-configured socket/pidfile pair.
- Do not expose socket paths through `/status`.
- Do not delete arbitrary user paths during stale socket cleanup.
- Do not declare desktop-ready before `system.capabilities` succeeds.
- Do not treat cloud lifecycle or remote telemetry as part of desktop
  reconciliation.

## Review Gate

Start with `desktop-rpc-launch-contract`. Do not implement Swift capabilities
handshake until the daemon launch contract and tests are committed.
