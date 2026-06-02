# Design

## CLI Shape

- `starclaw app`: unchanged default.
- `starclaw app --no-open`: ensure daemon is running, then print a concise status plus Web UI and diagnostics URLs.
- `starclaw app --check`: print local launch metadata and current daemon state without starting the daemon or opening a browser.

## Implementation

- Extend the existing app command in `cmd/daemon.go`.
- Keep browser opening isolated behind `openURLInBrowser` so tests can assert it is not called.
- Reuse existing constants:
  - `daemonWebURL`
  - `daemonDiagnosticsURL`
  - `daemonHealthURL`
  - `Version`
  - `config.StarclawDir()`

## Compatibility

- Existing success strings for `starclaw app` and `starclaw daemon open --start` should not change.
- New flags should be additive.
