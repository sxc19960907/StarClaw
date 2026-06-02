# Add GUI run summary session action

## Goal

After a chat run completes, the GUI run summary should let the user open the session referenced by that summary without searching the sidebar.

## Requirements

- Show a session action in the run summary only when the run response includes `session_id`.
- The action opens/selects that session in the Chat panel using the same session rendering path as the sidebar session list.
- The action must not require backend API changes.
- The existing summary fields (`Session`, `Agent`, `Usage`, `Request`) must keep rendering.
- Smoke coverage must verify the new action is present in the completed-run summary.

## Acceptance Criteria

- [ ] Completed chat runs with `session_id` render an `Open session` action in the run summary.
- [ ] Clicking the action selects the corresponding session through the existing session selection flow.
- [ ] Completed chat runs without `session_id` do not render a dead session action.
- [ ] Existing Web UI smoke checks still pass.

## Notes

- Scope is frontend-only unless implementation reveals an existing backend contract problem.
