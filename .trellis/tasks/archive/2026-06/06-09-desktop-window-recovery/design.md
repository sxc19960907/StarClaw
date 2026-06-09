# Desktop window recovery design

## Decision

Implement desktop window recovery inside the macOS Astria shell by combining:

1. Route persistence for the hosted Web UI.
2. Safe route restoration on app relaunch.
3. Daemon health monitoring after the shell is attached.
4. WebView reload after daemon recovery.

The Web UI's own Phase12 recovery remains the source of truth for run
deduplication, `/events` reconnect, and `/runs` refresh. The native shell
should not duplicate run-state logic. It should preserve and restore the
window's useful route, surface daemon health transitions, and reload the Web UI
when the daemon becomes reachable again.

## Scope

In scope:

- Persist the last useful Astria Web UI URL path/query/fragment.
- Restore that route on shell restart when it is same-origin and under `/app`.
- Fall back to `/app/` for unsafe, empty, or external URLs.
- Monitor daemon health while attached.
- Surface daemon unavailable/recovered states through the shell banner.
- Reload the WebView after daemon recovery so Phase12 browser recovery can
  refresh runs and reconnect events.
- Add smoke coverage for route persistence/sanitization.

Out of scope:

- Native reconstruction of run cards or approval state.
- Durable all-history event sourcing.
- Multiple windows/tabs.
- Kocoro pidfile/socket reconciliation.
- Packaging/signing/update policy.

## Route Contract

Persist only same-origin routes from the configured `ASTRIA_WEB_URL` origin.

Valid examples:

- `/app/`
- `/app/?view=mission`
- `/app/#runs`
- `/app/?q=a#runs`

Invalid routes must be ignored or normalized to `/app/`:

- `https://example.com/app/`
- `http://127.0.0.1:7533/diagnostics`
- `/`
- empty or malformed URLs

Storage key:

- `astria.lastWebRoute`

The value should be a relative route, not a full URL, so host/port overrides in
development do not persist stale origins.

## Daemon Health Recovery

When the shell reaches `attached`, start periodic health checks:

- If health fails, show a non-destructive banner and keep the WebView instance
  mounted when possible.
- If health recovers after a failure, reload the WebView at the restored current
  URL so the Web UI reconnect path can refresh `/events` and `/runs`.
- If a child daemon started by the shell exits, keep the existing crash state
  behavior and offer retry.

## Compatibility

- `starclaw app` browser launch remains unchanged.
- Existing Web UI recovery code remains the source of run-state recovery.
- Native shell recovery must not introduce remote telemetry or cloud sync.
- Smoke coverage must work with unsigned local builds.

## Kocoro Parity

This closes the product-shell gap where Desktop restart or daemon transition
leaves the user with a blank/initial browser-like surface. It does not yet
close Kocoro's deeper semi-bound Desktop/daemon reconciliation model; that
remains explicitly deferred.
