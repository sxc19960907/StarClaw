# Desktop browser tool depth

## Goal

Implement the next Phase6 Kocoro parity slice by deepening StarClaw's browser tool from simple `navigate/get_title` behavior into a more inspectable local desktop/browser surface. This child establishes structured browser state contracts that future PinchTab, AX, visual verification, and Ghostty-style workspace controls can build on.

## Requirements

- Use local Kocoro evidence:
  - `/Users/timmy/PycharmProjects/Kocoro/internal/tools/browser_lease.go`
  - `/Users/timmy/PycharmProjects/Kocoro/internal/tools/browser_handoff.go`
  - `/Users/timmy/PycharmProjects/Kocoro/internal/tools/pinchtab.go`
  - `/Users/timmy/PycharmProjects/Kocoro/internal/tools/axclient.go`
- Keep this child local-first:
  - no cloud browser service
  - no external PinchTab dependency yet
  - no new bundled AX server
  - no Ghostty workspace implementation in this slice
- Add browser tool actions that expose structured state:
  - `status`: reports platform support, known supported browser apps, and action availability.
  - `snapshot`: returns structured JSON for the front/current browser when available, including app, title, URL, and window title fallback metadata.
- Preserve existing `navigate` and `get_title` behavior.
- Keep browser inspection read-only and mark it read-only for permission logic.
- Do not log or persist raw page content, screenshots, cookies, or browser storage.

## Acceptance Criteria

- [ ] `browser` tool info documents `status` and `snapshot` actions.
- [ ] `status` works on every platform and returns JSON without requiring macOS automation.
- [ ] `snapshot` is read-only; on non-macOS it returns a clear unsupported error.
- [ ] On macOS, `snapshot` attempts Safari, Chrome, Chromium, and Brave before falling back to frontmost window metadata.
- [ ] Snapshot output is structured JSON rather than prose-only text.
- [ ] Existing `navigate`, `get_title`, and read-only semantics keep working.
- [ ] Unit tests cover action metadata, read-only classification, status output, unsupported snapshot behavior, and parsing helpers.
- [ ] Focused `internal/tools` tests and full project tests pass.

