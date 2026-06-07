# Workspace session hub

## Goal

Add a compact Workspace Session Hub to Astria Home so users can see and resume the current workspace context from one place: recent chat session, recent run status, memory readiness, and local file intake.

## Requirements

- Reuse existing `/sessions`, `/runs`, `/memory`, and Home state; do not add backend endpoints.
- Place the hub on Home near the existing mission board so it feels like part of the independent workspace shell.
- Show the latest session with a direct action to open Chat when available.
- Show run health using existing run status grouping and a direct action to Mission Control.
- Show memory readiness from existing memory entries/facts/warnings and a direct action to Memory Map.
- Show File Intake as the local context gateway with a direct action to File Intake.
- Keep the layout dense, operational, and consistent with Astria celestial/workspace language.
- Do not add frontend dependencies or a build pipeline.

## Acceptance Criteria

- [x] Home renders a Workspace Session Hub with session, run, memory, and file context cards.
- [x] Hub cards update from existing loaded session/run/memory state.
- [x] Session card opens Chat; run card opens Runs; memory card opens Memory; file card opens File Intake.
- [x] Empty states are useful when sessions or runs are missing.
- [x] Core smoke verifies hub rendering and at least one hub navigation action.
- [x] JS syntax, daemon tests, Web UI smoke, and full Go tests pass.

## Non-Goals

- No new backend aggregation endpoint.
- No persisted dashboard preferences.
- No replacement of Chat, Runs, Memory, or File Intake panels.
