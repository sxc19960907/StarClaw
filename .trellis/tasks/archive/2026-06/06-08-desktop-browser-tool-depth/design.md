# Design

## Current State

StarClaw's `BrowserTool` currently supports:

- `navigate`: opens a URL with platform opener.
- `get_title`: macOS AppleScript probe that returns prose with browser/title/URL or frontmost app fallback.

Kocoro has deeper browser runtime infrastructure: browser leases, reload handoff, PinchTab automation, and AX client support. StarClaw already has broad macOS tools, but the browser surface lacks a structured state contract. That makes future deeper automation harder because tests and callers must parse prose.

## Proposed Shape

Extend `internal/tools/browser.go` with:

- `status`
  - Always available.
  - Returns JSON with platform, support booleans, supported browser app names, actions, and backend notes.

- `snapshot`
  - Read-only.
  - macOS only in this slice.
  - Uses AppleScript to probe Safari, Google Chrome, Chromium, and Brave Browser.
  - Falls back to frontmost application/window title.
  - Returns JSON with:
    - `supported`
    - `platform`
    - `browser`
    - `title`
    - `url`
    - `frontmost_app`
    - `window_title`
    - `source`
    - `message`

Keep `get_title` compatible by rendering the structured snapshot into the existing prose shape where possible.

## Scope Boundaries

- No PinchTab server startup or HTTP client integration yet.
- No AX server/client addition to StarClaw in this child.
- No browser screenshots or page content extraction.
- No cookies, storage, or DOM state.
- No Ghostty tool registration.

## Compatibility

- `navigate` and `get_title` action names remain unchanged.
- `RequiresApproval()` remains false at tool level because per-call read-only logic is still handled through `IsReadOnlyCall`.
- `snapshot` and `status` are read-only.

## Rollback

Revert changes to `internal/tools/browser.go` and `browser_test.go`. No persistent data migration is involved.

