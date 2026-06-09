# Native desktop diagnostics and recovery UX

## Goal

Use Phase15 session and event-monitoring state in the macOS Astria shell so
Desktop RPC degradation is visible and recoverable to the user. The UX should
remain a thin native shell around the daemon Web UI, not a new native
replacement for the Web UI.

## Requirements

- Surface Desktop RPC session states in user-visible native diagnostics:
  connecting, connected, reconnecting, degraded, and mismatch.
- Preserve existing daemon/WebView launch, route recovery, and HTTP fallback.
- Keep recovery actions local: retry supervision, open diagnostics, and keep
  the WebView usable when HTTP health is available.
- Avoid exposing socket paths, pidfile paths, raw Desktop RPC event payloads,
  API keys, provider headers, or user content.
- Add smoke coverage for session diagnostics copy/state classification.

## Acceptance Criteria

- [ ] Astria has a native diagnostic summary for Desktop RPC session states.
- [ ] Reconnecting/degraded/mismatch states produce user-visible banner text.
- [ ] HTTP fallback remains mounted for degraded Desktop RPC states.
- [ ] Retry is available where the existing daemon state can safely retry.
- [ ] Smoke coverage validates diagnostic text for key session states.

## Notes

- Full native menus, Dock actions, crash reporter, and production release UX
  remain future phases.
