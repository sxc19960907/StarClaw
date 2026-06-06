# Agent Council Swarm Workflows

## Goal

Add a multi-agent planning/review workflow inspired by Kocoro `/swarm`, adapted to Astria as an "Agent Council". It should help with planning, critique, research, and implementation review without making routine chat more complex.

## Requirements

- Provide a way to start a council workflow from Home or Chat.
- Support named roles such as planner, researcher, implementer, and reviewer where the backend can represent them.
- Show each agent's contribution, status, and final synthesis.
- Keep user approval boundaries clear before code changes or external actions.
- Allow council runs to be saved as normal sessions/runs where possible.
- Favor a narrow first workflow, likely planning/review, over a generalized multi-agent framework.

## Acceptance Criteria

- [x] User can start a council workflow from the UI or command path chosen in design.
- [x] At least two distinct agent roles contribute to one workflow.
- [x] The final synthesis is clearly separated from individual agent notes.
- [x] Errors/timeouts from one role do not lose completed contributions from others.
- [x] Tests cover orchestration state transitions.

## Non-Goals

- No autonomous background swarm with unbounded tool use.
- No requirement for distributed/cloud agents in MVP.
- No replacement of ordinary single-agent chat.

## Dependencies

- Requires design work around run/session orchestration.
- Should happen after Home Actions establishes stable launch and status UI patterns.
