# Implementation Plan

## Scope

Deepen the browser tool with structured local status/snapshot contracts while preserving existing behavior.

## Steps

1. Read backend and tool quality guidelines.
2. Update `BrowserTool.Info()` to include `status` and `snapshot`.
3. Add browser status/snapshot response structs.
4. Implement `status` action with platform/action/backend metadata.
5. Implement `snapshot` action:
   - non-macOS clear unsupported error
   - macOS AppleScript browser probing
   - structured JSON output
6. Adapt `get_title` to use the same browser snapshot helper while preserving prose output.
7. Update read-only classification for `status` and `snapshot`.
8. Add/adjust tests in `internal/tools/browser_test.go`.
9. Run:
   - `go test ./internal/tools`
   - `go test ./...`
   - `git diff --check`
10. Commit and archive child task.

## Review Gates

- Browser inspection must stay content-free: no page body, cookies, storage, screenshots, or request payloads.
- Snapshot must be structured JSON for machine parsing.
- Non-macOS behavior must be deterministic in tests.
- Existing actions remain compatible.

## Completion Notes

- Added `browser status` with platform, backend, supported browser apps, and action metadata.
- Added `browser snapshot` with structured JSON output for current browser/window metadata.
- Preserved existing `navigate` and `get_title`; `get_title` now renders from the structured snapshot helper on macOS.
- Updated read-only classification for `get_title`, `status`, and `snapshot`.
- Added tests for metadata, read-only semantics, status JSON, non-macOS snapshot behavior, and snapshot parser cases.

## Validation

- `go test ./internal/tools` — passed.
- `go test ./...` — passed.
- `git diff --check` — passed.
- `python3 ./.trellis/scripts/task.py validate .trellis/tasks/06-08-desktop-browser-tool-depth` — passed.
