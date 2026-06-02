# Improve run detail timeline and rerun actions

## Goal

Make Run detail useful as an execution review surface instead of a raw event dump.

## Requirements

- Run detail must present top-level actions:
  - copy run summary;
  - copy prompt;
  - open session when a session is available;
  - re-run with the same agent and prompt.
- Run detail timeline must group related tool call and tool result events into one tool card where possible.
- Timeline must render text, preamble, usage, approval, and generic events in readable event cards.
- Tool cards must show tool name, status, args, result, and error category when present.
- Long args/results must remain scrollable and not break layout.
- Re-run must move the user to Chat, prefill agent/prompt, and start a new session by default; it should not immediately execute without user action.
- Existing Run list, Agent Test Runner auto-open behavior, and Chat run summary actions must continue to work.
- Browser smoke must cover copy actions, open session, rerun prefill, and grouped timeline rendering.

## Acceptance Criteria

- [x] Run detail includes Copy summary, Copy prompt, Open session, and Re-run actions when applicable.
- [x] Tool call and tool result events for the same tool are displayed as one grouped card.
- [x] Text/preamble/usage/generic events still render in the timeline.
- [x] Re-run preselects the original agent, prompt, and new-session mode in Chat.
- [x] Web UI runs smoke validates the new actions and grouped timeline.
- [x] Targeted Go tests, Web UI smoke, JS syntax check, and diff check pass.

## Notes

- Scope is Web UI only unless existing run data is insufficient.
