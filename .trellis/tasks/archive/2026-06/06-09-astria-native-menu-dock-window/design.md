# Astria native menu Dock and window shell design

## Current Shape

Astria has a single `WindowGroup("Astria")` and explicitly replaces `.newItem`
with an empty command group. The root view owns `reloadToken`, `webURL`, and
`DaemonSupervisor`, but there is no shared native command bridge.

## Proposed Shape

Add an `AstriaAppActions` environment object:

- `newWindow`: uses SwiftUI `openWindow` for the main window scene.
- `reload`: increments a published reload token through the root view.
- `openDiagnostics`: opens the configured diagnostics URL.
- `retryDaemon`: calls `DaemonSupervisor.start()`.

Add a small `AstriaNativeCommandSpec` smoke-testable model for labels and
keyboard shortcuts. SwiftUI commands render from this contract.

## Scene Model

Give the main `WindowGroup` a stable id such as `astria-main` so New Window can
open another independent root view around the same local daemon URL.

## Test Strategy

Add `--native-command-smoke` to validate the command spec labels/shortcuts and
basic action identifiers without UI automation.
