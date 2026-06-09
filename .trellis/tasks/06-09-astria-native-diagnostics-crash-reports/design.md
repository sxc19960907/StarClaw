# Astria native diagnostics export and crash reports design

## Current Shape

Astria can open the daemon `/diagnostics` URL and show native banners for daemon
and Desktop RPC session states. It does not yet have a native local export
artifact that can be attached to support, issue reports, or future crash report
flows.

## Proposed Shape

Add a Swift-side `AstriaDiagnosticsReport` boundary:

- Build a local report from `LaunchConfig`, `DaemonState`, and
  `DesktopRPCSessionState`.
- Include app version, generated timestamp, safe local URLs, daemon state,
  Desktop RPC session state, and a redacted failure summary.
- Redact API keys, bearer tokens, socket paths, pidfile paths, and raw sensitive
  values before serialization.
- Export JSON under an Astria-owned diagnostics directory in the app support
  area.

Add an `Export Diagnostics` native command that writes the report locally. Full
crash reporter upload remains out of scope.

## Test Strategy

Add `--diagnostics-export-smoke` to validate:

- report generation includes app/session/daemon metadata;
- export writes a local JSON file;
- redaction removes socket path, pidfile path, API key, bearer token, and
  secret-like values.
