# Workflow control API pause resume cancel replay

## Goal

Add workflow control APIs for pause, resume, cancel, and replay so long-running Astria work can be controlled more like Kocoro/Shannon workflows.

## Requirements

- Expose local daemon control endpoints for cancel and state inspection first; pause/resume/replay may be staged if runtime support is not yet present.
- Preserve existing stop/cancel Web UI behavior.
- Define replay semantics that do not accidentally repeat destructive/external actions.
- Ensure control actions are permission-aware and auditable.
- Record control decisions in run metadata or event history.

## Acceptance Criteria

- [ ] Control API contract is documented in PRD/design before implementation.
- [ ] Cancel endpoint is test-covered and remains compatible with current stop behavior.
- [ ] Replay requires explicit approval boundary for tool calls or external effects.
- [ ] Run/event metadata reflects cancel/replay decisions.
- [ ] Existing run smoke tests continue to pass.

## Notes

- This is a backend/workflow-control slice.
