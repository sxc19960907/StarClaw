# Improve app launch install readiness

## Goal

Make the installed StarClaw app entrypoint easier to validate and use across desktop, CI, and remote/headless environments.

## Requirements

- Keep `starclaw app` default behavior unchanged: start daemon when needed and open the Web UI.
- Add a non-browser launch mode that starts/reuses the daemon and prints the Web UI URL.
- Add an install/readiness check mode that prints version, launch command, Web UI URL, diagnostics URL, data path, and daemon state without opening a browser.
- Update CLI smoke coverage for the new app modes.
- Update installation and usage docs so a user can verify an installed binary and launch the GUI.

## Acceptance Criteria

- [x] `starclaw app --check` exits successfully without starting/opening the browser and prints launch metadata.
- [x] `starclaw app --no-open` starts/reuses the daemon path and prints the Web UI URL without invoking browser open.
- [x] Existing `starclaw app` success output remains compatible.
- [x] CLI unit tests and smoke cover the new modes.
- [x] Installation docs include GUI launch verification steps.
- [x] Targeted checks pass.
