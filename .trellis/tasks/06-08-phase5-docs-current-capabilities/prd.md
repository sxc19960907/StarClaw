# Phase 5 docs current capabilities

## Goal

Update user-facing documentation so StarClaw/Astria describes the current local runtime, API, workflow-control, observability, recovery, and safety capabilities accurately.

## Requirements

- Preserve StarClaw naming for CLI/module/package/release docs and Astria naming for product-facing Web UI.
- Document local-only safety boundaries and explicit non-goals.
- Reflect Phase 3/4 capabilities only after validation children confirm behavior.
- Keep docs concise enough for users to operate the app.

## Acceptance Criteria

- [ ] README or relevant docs describe Astria Web UI current workflow surfaces.
- [ ] API docs mention local OpenAI-compatible gateway and workflow-control endpoints.
- [ ] Runtime docs mention budget enforcement, routing/fallback, observability, trace export, durable recovery, and replay approval boundaries.
- [ ] Safety docs state prompt/secret redaction guarantees and limitations.
- [ ] Documentation checks or relevant tests pass.

## Non-Goals

- No marketing rewrite.
- No hosted/cloud documentation.
- No docs for unimplemented Phase 6 ideas.

## Goal

TBD.

## Requirements

- TBD

## Acceptance Criteria

- [ ] TBD

## Notes

- Keep `prd.md` focused on requirements, constraints, and acceptance criteria.
- Lightweight tasks can remain PRD-only.
- For complex tasks, add `design.md` for technical design and `implement.md` for execution planning before `task.py start`.
