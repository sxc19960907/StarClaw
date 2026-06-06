# Astria Home Task Launchers and Activity Center

## Goal

Turn the Astria Home surface from a polished prompt entry page into a functional product hub. Home chips, cards, counters, and recommendations should launch real workflows, show live state, and keep the Kocoro-like directness the user wants.

## Requirements

- Home shortcut chips must either navigate to a working panel, seed a useful task with mode hints, or show a clearly disabled future state.
- Activity counters must reflect real daemon/run/approval state where APIs exist.
- Home cards must map to product capabilities:
  - Mission Orbit: active/running/recent work.
  - MCP Starport: MCP servers and tool health.
  - Memory Map: reusable memory and session context.
- The large home composer must continue submitting through the existing chat/run stream.
- UI should stay minimal and native-app-like, not become a dashboard grid.
- Astria visual language should use subtle constellation/orbit cues around actual state and actions.

## Acceptance Criteria

- [x] Each visible Home chip has a deterministic action.
- [x] Live counters match the same data shown in existing Runs/Approvals surfaces.
- [x] Clicking core Home cards routes to real panels or shows a non-broken disabled state.
- [x] Home composer still passes existing Web UI smoke submission flow.
- [x] Narrow and desktop layouts remain usable without text overlap.
- [x] Web UI smoke tests pass.

## Non-Goals

- No new backend capability unless required to expose already-existing state.
- No large visual redesign beyond making the current Astria Home operational.
- No external frontend framework or build pipeline.

## Dependencies

- Depends on completion of `06-06-astria-webui-home`.
- Should remain compatible with existing Web UI selectors and smoke scripts.
