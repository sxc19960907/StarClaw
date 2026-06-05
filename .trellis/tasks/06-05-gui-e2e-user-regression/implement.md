# Implementation Plan

## Checklist

1. Start from existing deterministic Web UI smoke to verify the scripted broad path.
2. Run targeted tool-call smoke to verify real tool call UI path.
3. Start GUI locally and inspect with Playwright CLI:
   - Diagnostics
   - Config
   - Agents / Agent Test
   - Chat
   - Sessions
   - Runs / Run Detail
   - Permissions
   - Version
4. Capture screenshots.
5. Fix any discovered blocking UI regression.
6. Rerun affected smoke/tests.
7. Record outcome and archive task.

## Validation Commands

- `scripts/smoke_webui_core.sh`
- `scripts/smoke_webui_tool_call.sh`
- targeted Playwright CLI screenshots/snapshots
- `git diff --check` if code changes occur

## Risk Points

- Browser smoke uses port 7533; do not run overlapping daemon smoke scripts.
- Local npm commits are ahead of remote until network push succeeds; do not lose them.
