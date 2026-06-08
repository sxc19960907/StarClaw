# Council stage workflow

## Goal

Make Agent Council feel like a staged multi-agent workflow instead of a one-shot role list. The Council detail should expose how a goal moves through planning, research, review, synthesis, and handoff, with direct operator actions at each stage.

## Requirements

- Keep the embedded static Web UI architecture; do not add a frontend build pipeline.
- Preserve the existing Council daemon API shape and run handoff behavior.
- Render a Council stage rail for Planner, Researcher, Reviewer, Synthesis, and Handoff.
- Each stage must show status, role/output summary, and an operational next action.
- Role stages must support copying the role notes and drafting the role output into Chat.
- Synthesis and Handoff stages must keep existing copy, send-to-chat, and start-run actions.
- Styling should stay dense and operational with subtle Astria/celestial identity.

## Acceptance Criteria

- [x] Council detail renders a stage rail for a completed council run.
- [x] Planner, Researcher, and Reviewer stages show role status, summary, notes preview, and copy/draft actions.
- [x] Synthesis stage shows the final synthesis and keeps copy/send-to-chat actions.
- [x] Handoff stage exposes Start run and preserves the existing `council_handoff` run path.
- [x] Web UI smoke verifies stage rail rendering, role copy/draft actions, synthesis copy, and run handoff.

## Notes

- Keep `prd.md` focused on requirements, constraints, and acceptance criteria.
- Lightweight tasks can remain PRD-only.
- For complex tasks, add `design.md` for technical design and `implement.md` for execution planning before `task.py start`.
