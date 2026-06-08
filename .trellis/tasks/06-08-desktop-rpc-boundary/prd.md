# Desktop RPC boundary

## Goal

Implement the next Phase6 Kocoro parity slice: add a local Unix-socket Desktop RPC boundary so a future native Astria/Kocoro-style desktop app can attach to the StarClaw daemon, exchange versioned JSON frames, answer system capability probes, and provide a tested broker/listener foundation for later calendar/browser/native desktop tools.

## Requirements

- Use local Kocoro evidence:
  - `/Users/timmy/PycharmProjects/Kocoro/internal/daemon/desktop_rpc/broker.go`
  - `/Users/timmy/PycharmProjects/Kocoro/internal/daemon/desktop_rpc/codec.go`
  - `/Users/timmy/PycharmProjects/Kocoro/internal/daemon/desktop_rpc/listener.go`
  - `/Users/timmy/PycharmProjects/Kocoro/internal/daemon/desktop_rpc/types.go`
- Keep this child local-first and daemon-internal:
  - no cloud
  - no external Desktop app dependency
  - no calendar/EventKit implementation yet
  - no auto-launching native apps
- Add a StarClaw desktop RPC package or daemon module with:
  - versioned protocol constants
  - length-prefixed JSON frame codec
  - broker for pending request/response correlation
  - listener over Unix domain socket
  - system methods: `system.ping` and `system.capabilities`
- Add a fake desktop smoke harness in tests that connects to the socket, reads/writes frames, and proves request/result flow.
- Expose daemon status/capability information enough for diagnostics or tests to know whether Desktop RPC is listening and connected.
- Ensure disconnects cancel pending requests without hanging agent/runtime goroutines.
- Keep secrets and request payloads out of structured telemetry and metrics.

## Acceptance Criteria

- [ ] Desktop RPC codec round-trips frames and rejects oversized/invalid frames.
- [ ] Desktop RPC broker returns not-connected when no client is attached.
- [ ] Broker correlates requests/results by request id and times out pending calls.
- [ ] Listener accepts a Unix socket connection, installs broker send function, and cancels pending calls on disconnect.
- [ ] `system.ping` returns a local pong/server time response.
- [ ] `system.capabilities` returns protocol version, supported methods, platform, and StarClaw daemon version.
- [ ] A fake desktop test proves daemon-to-desktop request flow and desktop-to-daemon system method flow.
- [ ] `GET /status` or diagnostics expose Desktop RPC listening/connected state without leaking socket secrets or payloads.
- [ ] Full project tests pass.

## Notes

- Calendar tools and macOS-specific native actions are intentionally deferred. This child builds the boundary they will use.
