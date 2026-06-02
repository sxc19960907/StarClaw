# Design

## Current State

- Run detail renders metadata, prompt, result, and an event timeline.
- `renderRunEvents()` prints each raw event independently.
- Run summaries already expose copy/open-run/open-session actions in Chat and Agent Test Runner result cards.

## UI Design

- Add a `run-detail-actions` toolbar below the metadata section:
  - `data-run-copy-summary`
  - `data-run-copy-prompt`
  - `data-run-open-session`
  - `data-run-rerun`
- Generate summary text from run ID, status, agent, session, usage, and prompt.
- Re-run behavior:
  - switch to Chat;
  - set `chat-agent` to the run agent if available;
  - set `chat-input` to run prompt/request text;
  - set `chat-new-session` checked;
  - clear active session and update labels;
  - focus the composer.

## Timeline Grouping

- Build display entries from `run.events`.
- For `tool_call`, create or update a grouped tool entry keyed by tool name.
- For `tool_result`, merge into the latest matching tool entry when available; otherwise create a standalone tool entry.
- Preserve chronological order by the first event timestamp for grouped tool entries.
- Render event kinds:
  - `tool`: tool card with status, args, result, error category;
  - `text` / `preamble`: compact transcript card;
  - `usage`: token card;
  - `approval_*`: approval card;
  - fallback generic card.

## Testing

- Extend `smoke_webui_runs.sh` data with tool call/result and usage events.
- Validate grouped tool card shows one tool event with args/result.
- Validate Copy summary, Copy prompt, Open session, and Re-run.
