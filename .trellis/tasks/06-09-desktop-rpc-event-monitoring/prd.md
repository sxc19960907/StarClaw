# Desktop RPC event monitoring

## Goal

Add the first local Desktop RPC event-monitoring path for Phase15 by making
daemon-side `desktop_event` frames observable and testable. The goal is not a
full native event subscription system yet; it is to preserve local lifecycle /
tool events sent by the desktop client and expose safe aggregate state for
diagnostics and follow-up UX.

## Requirements

- Reuse the existing Desktop RPC `FrameDesktopEvent` and `DesktopEvent` wire
  shape.
- Preserve existing request/response Desktop RPC behavior and Phase14/15
  session smoke coverage.
- Record recent desktop events locally in daemon memory with bounded retention.
- Expose redacted event-monitoring state through daemon status or a similarly
  local diagnostics surface.
- Do not expose socket paths, pidfile paths, or raw sensitive event payloads in
  `/status`.
- Keep browser/CLI flows valid without a desktop client.
- Do not add cloud routing, Shannon auth, or off-machine telemetry.

## Acceptance Criteria

- [ ] A connected desktop client can send a `desktop_event` frame and the daemon
      records it through the existing listener `EventSink`.
- [ ] Event retention is bounded and concurrency-safe.
- [ ] `/status` or local diagnostics reports safe event-monitoring metadata
      such as last event type/time and retained event count.
- [ ] Existing Desktop RPC request/response tests continue to pass.
- [ ] Tests cover event ingestion, bounded retention, and redacted status.

## Notes

- This child should not implement full native macOS event producers. Native UX
  consumption belongs in `native-desktop-diagnostics-recovery`.
