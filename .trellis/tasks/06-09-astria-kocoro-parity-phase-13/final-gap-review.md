# Phase13 final Kocoro gap review

## Baseline

- Kocoro reference: `/Users/timmy/PycharmProjects/Kocoro` at
  `74cdb3cfa010bc07a5ec3a5f9bda1854da41b245`.
- Phase13 scope: standalone desktop app shell, daemon supervision, window
  recovery, and packaging/signing/update boundary.
- Out of scope: full native UI replacement, cloud sync, off-machine telemetry,
  private signing credentials, and deeper pidfile/socket reconciliation.

## Closed in Phase13

- Added `desktop/macos/Astria` as a thin SwiftUI/WKWebView desktop shell that
  hosts the existing daemon-served Web UI.
- Preserved `starclaw app`, browser launch, and daemon-only workflows.
- Implemented app-side daemon attach/start behavior through the existing
  `/health` readiness contract.
- Added user-visible daemon checking, starting, failed, crashed, unavailable,
  and recovered states.
- Added route persistence for same-origin `/app` paths only, with unsafe-route
  fallback to `/app/`.
- Kept Web UI run/event recovery in the daemon/Web UI layer instead of
  duplicating run-state reconstruction in Swift.
- Added WebView reload after daemon health recovery.
- Added unsigned macOS shell smoke coverage for bundle structure, route
  recovery, bundled daemon resolution, supervision, attach, and launch failure.
- Added build-time support for optional bundled daemon placement at
  `Astria.app/Contents/Resources/starclaw`.
- Added build-time app version metadata via `ASTRIA_APP_VERSION` and
  `ASTRIA_APP_BUILD`.
- Documented signing, notarization, updater, and app/daemon compatibility
  boundaries without committing private credentials or update keys.

## Remaining Kocoro gaps

- StarClaw's Astria shell still uses HTTP health/readiness for supervision. It
  does not yet launch the daemon with Desktop RPC `--rpc-socket` /
  `--rpc-pidfile` and perform `system.capabilities` reconciliation before
  declaring the desktop online.
- Version mismatch is documented at the packaging boundary but not yet enforced
  as a fail-fast runtime check between the app bundle and daemon.
- Production macOS distribution remains unsigned/local-first. There is no
  committed DMG/pkg build, Developer ID signing, notarization, stapling, or
  external secret-backed macOS release workflow.
- Astria does not yet implement a desktop auto-updater. Update behavior remains
  CLI/release driven, with future checksum/signature requirements documented.
- Native Desktop OS integrations such as Kocoro's calendar/TCC sidecar model
  remain deferred.
- Native app affordances remain minimal: no polished menu model, multi-window
  management, Dock integration, crash reporter, or signed helper applications.

## Updated parity estimate

Phase13 moves StarClaw from browser-opened Web UI toward a real standalone app
shell and closes the public Kocoro packaging-boundary gap that is visible in
the open repository. Against Kocoro's broader Desktop lifecycle model, Astria is
now roughly 80-85% aligned for local-first standalone shell behavior.

The next parity step should focus on Desktop RPC handshake and daemon
reconciliation:

1. Launch daemon from the app with explicit socket and pidfile paths.
2. Call `system.capabilities` before declaring desktop-online readiness.
3. Enforce app/daemon version compatibility at runtime.
4. Preserve the current HTTP fallback for CLI/browser launch.
