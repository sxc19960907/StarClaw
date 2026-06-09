# Desktop RPC capabilities reconciliation

## Goal

Make Astria connect to the daemon Desktop RPC socket and reconcile
`system.capabilities` before the app considers the desktop lifecycle ready.

## Requirements

- Add a minimal Swift Desktop RPC socket client for framed `system.ping` and
  `system.capabilities` requests.
- Make Astria launch the daemon with deterministic socket and pidfile paths
  once the daemon launch contract exists.
- Validate Desktop RPC protocol version and required method set.
- Validate app/daemon version compatibility for release builds; development
  builds may warn but must not silently report full parity.
- Surface protocol mismatch, missing capabilities, connection failure, and
  version mismatch as user-visible states.
- Preserve HTTP health fallback for degraded recovery and CLI/browser paths.

## Acceptance Criteria

- [ ] Astria connects to the Desktop RPC socket after daemon launch or attach.
- [ ] Astria calls `system.capabilities` and blocks desktop-ready on mismatch.
- [ ] Required methods include at least `system.ping` and
      `system.capabilities`.
- [ ] Protocol mismatch and version mismatch produce clear shell diagnostics.
- [ ] Smoke coverage verifies success and at least one mismatch/failure path.

## Notes

This child depends on `desktop-rpc-launch-contract`.
