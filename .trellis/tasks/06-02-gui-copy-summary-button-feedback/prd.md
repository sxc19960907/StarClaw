# Add GUI copy summary button feedback

## Goal

Make the run summary copy action give immediate in-card feedback after a successful copy.

## Requirements

- After `Copy summary` succeeds, the clicked button should temporarily show `Copied`.
- The button should return to `Copy summary` automatically.
- Existing toast feedback should remain.
- Copy behavior and copied text format must not change.
- No backend API changes.

## Acceptance Criteria

- [ ] Clicking `Copy summary` copies the existing summary text.
- [ ] The clicked button changes to `Copied` after success.
- [ ] The button returns to `Copy summary` without requiring navigation or reload.
- [ ] Existing `Open session` behavior still works.
- [ ] Web UI smoke verifies the button feedback.

## Notes

- Scope is frontend-only.
