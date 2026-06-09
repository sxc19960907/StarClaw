# Desktop RPC session lifecycle

## Goal

Turn Astria's Desktop RPC reconciliation from a launch-time one-shot
`system.capabilities` probe into a long-lived native session lifecycle. The
first slice should detect post-launch RPC disconnects, retry bounded recovery,
and preserve healthy HTTP fallback without changing CLI/browser daemon flows.

## Requirements

- Keep Phase14 `system.capabilities` validation as the initial readiness gate.
- Add app-side Desktop RPC session monitoring after successful reconciliation.
- Detect RPC disconnects or missing sockets after launch.
- Retry recoverable RPC failures with bounded attempts and delay.
- Enter degraded HTTP fallback when Desktop RPC cannot be restored while daemon
  HTTP health remains available.
- Avoid retry loops on protocol/version mismatch.
- Keep `starclaw daemon start`, `starclaw app`, and browser flows compatible.
- Do not add cloud lifecycle, Shannon auth, or off-machine telemetry.

## Acceptance Criteria

- [ ] Astria starts a long-lived Desktop RPC monitor after initial capabilities
      validation succeeds.
- [ ] The monitor can distinguish connected, reconnecting, degraded, and
      terminal mismatch outcomes.
- [ ] A recoverable missing/broken socket leads to bounded retry and then
      degraded HTTP fallback instead of process exit.
- [ ] Protocol/version mismatch remains terminal for the current daemon and
      does not spin retries.
- [ ] Health monitoring and WebView recovery continue to work when Desktop RPC
      is degraded.
- [ ] Smoke coverage exercises connected session validation and bounded
      disconnect/fallback behavior.

## Notes

- Prefer Swift-side monitoring first. Daemon protocol extensions belong in the
  later event-monitoring child task.
