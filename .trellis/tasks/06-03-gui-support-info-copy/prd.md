# Add GUI support info copy

## Goal

Let users copy a safe support summary from the Web UI Version page for troubleshooting daemon startup, update, and runtime context issues.

## Requirements

- Version page should include a `Copy support info` action.
- Copied text should include version/build/update status plus runtime URLs and local paths already exposed by `/version`.
- Copied text should include diagnostics status/summary when diagnostics are loaded.
- Copied text must not include API keys or raw config contents.
- Web UI smoke should verify the action and copied text.

## Acceptance Criteria

- [x] Version page has a `Copy support info` button.
- [x] Clipboard text includes version, platform, Web UI, diagnostics URL, data dir, config path, and diagnostics status.
- [x] Clipboard text omits secrets and raw config content.
- [x] Browser smoke covers the copy action.
- [x] Diff check passes.
