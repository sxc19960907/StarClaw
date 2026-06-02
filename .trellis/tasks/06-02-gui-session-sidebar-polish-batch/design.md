# Design

## Boundary

This task changes session sidebar Web UI behavior and smoke coverage:

- `internal/daemon/webui/index.html`
- `internal/daemon/webui/assets/app.js`
- `scripts/smoke_webui.sh`

No daemon API changes are required. Existing `/sessions`, `/sessions/search`, and `/sessions/{id}` endpoints are reused.

## Behavior

Search:

- Add an `input` listener to `#session-search`.
- Debounce calls to `loadSessions(query)` so typing does not issue one request per keystroke.
- Keep the existing submit handler as an immediate search fallback.

Clear:

- Add a `Clear` button to the search form.
- Clicking it clears the input, reloads recent sessions, and focuses the search input.

Copy ID:

- Add a `Copy ID` button to each session row.
- Reuse `copyText` and `markButtonCopied`.
- Stop row-click propagation so copying does not select/open the session.

## Compatibility

Existing row-level selection remains attached to `data-session-id`. Existing delete/rename/favorite actions already stop propagation and remain unchanged.

## Test Strategy

Extend Web UI smoke after session creation to:

- copy the session id and verify clipboard text plus transient label,
- type the renamed session title in search and verify the row remains visible,
- click `Clear` and verify the field is empty and the session row remains available.
