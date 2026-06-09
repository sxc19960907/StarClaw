# Astria native menu Dock and window shell

## Goal

Make Astria's macOS shell behave more like a native app by adding first-class
menu/window commands around the existing daemon WebView: New Window, Reload,
Open Diagnostics, and Retry Daemon.

## Requirements

- Preserve the daemon-served Web UI and existing WebView route recovery.
- Support native New Window instead of disabling the macOS new-window command.
- Add native commands for Reload, Open Diagnostics, and Retry Daemon.
- Keep commands safe when the daemon is unavailable; Diagnostics opens the
  configured local diagnostics URL and Retry restarts supervision.
- Keep CLI/browser flows unchanged.
- Add smoke coverage for command labels/availability without requiring GUI
  automation.

## Acceptance Criteria

- [ ] Astria defines a reusable native command model for expected menu actions.
- [ ] The macOS app exposes New Window, Reload, Diagnostics, and Retry Daemon
      commands.
- [ ] Reload refreshes the hosted WebView without changing the persisted route.
- [ ] Retry Daemon calls the existing `DaemonSupervisor.start()` path.
- [ ] Smoke coverage validates command labels and shortcut metadata.

## Notes

- Multi-window data isolation is limited to independent SwiftUI window
  instances around the same local daemon. Deeper window session routing can be
  expanded later.
