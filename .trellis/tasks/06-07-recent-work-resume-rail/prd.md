# Recent work resume rail

## Goal

Make recent work resumable from Astria's app shell by adding recent session/run commands and direct resume behavior to the Home Workspace Hub.

## Requirements

- Reuse existing loaded `state.sessions` and `state.runs`; do not add backend endpoints.
- Add recent session commands to Command Center when sessions are available.
- Add recent run commands to Command Center when runs are available.
- Workspace Hub Session card should resume the latest session when one exists.
- Workspace Hub Runs card should open the latest run when one exists.
- Empty Workspace Hub cards should keep their existing panel navigation behavior.
- Keep command filtering simple and consistent with the existing Command Center.
- Do not add frontend dependencies or a build pipeline.

## Acceptance Criteria

- [x] Command Center includes recent session commands after sessions load.
- [x] Command Center includes recent run commands after runs load.
- [x] Running a recent session command opens Chat with that session.
- [x] Running a recent run command opens Runs detail for that run.
- [x] Workspace Hub Session/Run cards resume latest work when available.
- [x] Core smoke covers at least one recent work command or resume action.
- [x] JS syntax, daemon tests, Web UI smoke, and full Go tests pass.

## Non-Goals

- No persisted command history.
- No backend recent-work aggregation.
- No new run/session detail UI.
