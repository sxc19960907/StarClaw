# Improve app diagnostics and launch readiness

## Goal

Make `starclaw app` and the Web UI diagnostics provide enough launch/readiness context for a user to understand how to start the GUI and what local paths/config are being used.

## Requirements

- Diagnostics API must expose launch/readiness metadata:
  - Web UI URL;
  - recommended launch command;
  - StarClaw data directory;
  - config path;
  - agents path;
  - sessions path;
  - current executable path when available.
- Version API should expose the same Web UI URL and launch command context where appropriate.
- GUI Diagnostics should render a concise Launch readiness section.
- GUI Version should render startup/environment rows without requiring users to infer paths from logs.
- CLI app/open output should keep current behavior and add actionable context when startup fails.
- Existing diagnostics checks and provider setup repair flow must not regress.
- Smoke coverage must validate launch command/path metadata in API and GUI.

## Acceptance Criteria

- [x] `GET /diagnostics` includes launch metadata fields.
- [x] `GET /version` includes the recommended GUI launch command.
- [x] Diagnostics page shows Web UI URL, launch command, data/config/agents/sessions paths, and executable when available.
- [x] Version page shows launch command and Web UI URL.
- [x] Existing core Web UI smoke still passes and asserts new metadata.
- [x] Targeted Go tests, CLI smoke, Web UI core smoke, JS syntax check, and diff check pass.

## Notes

- Scope excludes installer packaging and auto-update behavior changes.
