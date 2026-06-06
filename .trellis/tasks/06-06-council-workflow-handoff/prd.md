# Council Workflow Handoff

## Goal

Turn Agent Council synthesis into actionable work: users should be able to convert a council result into a chat prompt, a run, or a scoped follow-up task without manually copying content.

## Requirements

- Preserve the current Council role separation.
- Add explicit user-driven handoff actions.
- Keep handoffs reviewable before execution.
- Record source council run metadata on created runs or prompts where practical.
- Avoid making Council auto-execute work.

## Acceptance Criteria

- [x] Council detail exposes clear handoff actions.
- [x] User can start a normal Astria run from a council synthesis.
- [x] Handoff preserves source council ID or goal in the run/request context.
- [x] User can still copy or send synthesis to chat.
- [x] Tests cover the handoff API/UI path.

## Non-Goals

- No autonomous swarm execution.
- No persistent task board unless planned separately.

## Dependencies

- Depends on Agent Council MVP and run history UI.
