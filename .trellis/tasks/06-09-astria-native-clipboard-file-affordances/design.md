# Astria native clipboard file affordances design

## Current Shape

Astria already has native commands for New Window, Reload, Open Diagnostics,
Export Diagnostics, and Retry Daemon. The shell can export a redacted local
diagnostics JSON report, but it does not provide quick native copy/reveal
actions.

## Proposed Shape

Extend `AstriaAppActions` and command specs with:

- Copy Current Route: writes the safe relative `/app` route or `/app/`.
- Copy Support Summary: writes a redacted local support text generated from
  `LaunchConfig`, `DaemonState`, and `DesktopRPCSessionState`.
- Reveal Diagnostics Folder: opens the local diagnostics export directory in
  Finder.

Use `NSPasteboard.general` for clipboard writes and `NSWorkspace` for Finder
reveal.

## Test Strategy

Add smoke assertions for:

- command metadata;
- route normalization;
- support summary redaction for token/socket/pidfile values.
