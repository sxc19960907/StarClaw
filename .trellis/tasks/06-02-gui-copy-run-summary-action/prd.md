# Add GUI copy run summary action

## Goal

Let users copy the completed chat run summary from the GUI so session metadata can be pasted into notes, issue reports, or debugging context.

## Requirements

- Show a `Copy summary` action in each successful run summary card.
- Copy the same fields shown in the card: session id, agent, usage, and request id.
- Use the browser clipboard API and existing toast feedback.
- Keep the existing `Open session` action behavior unchanged.
- Do not add backend API changes.

## Acceptance Criteria

- [ ] Successful run summary cards include a `Copy summary` action.
- [ ] Clicking `Copy summary` writes a readable plain-text summary to the clipboard.
- [ ] The UI shows success feedback after copying.
- [ ] Existing run summary and session actions still render.
- [ ] Web UI smoke covers the copy action and clipboard content.

## Notes

- Scope is frontend-only.
