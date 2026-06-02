# Batch GUI session sidebar polish

## Goal

Improve the session sidebar with a small batch of related usability actions.

## Requirements

- Session search should refresh results as the user types, without requiring Enter.
- The search form should include a `Clear` action that resets the query and reloads recent sessions.
- Each session row should include a `Copy ID` action that copies the session id to the clipboard.
- Copying should use the existing clipboard/toast feedback path and transient button feedback.
- Existing rename, favorite, delete, and row selection behavior must remain unchanged.
- No backend API changes.

## Acceptance Criteria

- [ ] Typing in the session search field reloads filtered sessions.
- [ ] Clicking `Clear` empties the field and reloads recent sessions.
- [ ] Clicking `Copy ID` copies the row session id and shows copy feedback.
- [ ] Existing session row actions still work.
- [ ] Browser smoke covers the new copy and clear behavior.

## Notes

- This is a batch task; keep all changes scoped to the session sidebar.
