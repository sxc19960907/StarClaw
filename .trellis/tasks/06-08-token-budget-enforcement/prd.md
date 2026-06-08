# Token budget enforcement

## Goal

Add real local token budget tracking and hard-stop enforcement to Astria/StarClaw runs so Budget Guard becomes an executable runtime constraint instead of only a planning surface.

## Requirements

- Add a runtime budget model that can represent max input tokens, max output tokens, max total tokens, and stop behavior.
- Track token usage from provider responses when usage metadata is available.
- Enforce hard-stop behavior before follow-up model calls when a configured budget is exhausted or projected to exceed limits.
- Surface budget status in run records or daemon responses without leaking secrets.
- Keep behavior local and provider-agnostic; support missing usage metadata with conservative unknown/estimated status rather than false precision.
- Provide CLI/config/API touchpoints only where they already fit existing runtime patterns.

## Acceptance Criteria

- [x] Runtime has a test-covered budget data structure and enforcement decision function.
- [x] Runs can record budget status and usage when provider usage is available.
- [x] A run that exceeds hard budget stops with a clear budget-exhausted result instead of continuing model/tool loops.
- [x] Missing provider usage metadata is handled explicitly.
- [x] Targeted unit tests cover under-budget, at-budget, over-budget, and unknown-usage cases.
- [x] Relevant daemon/agent tests pass.

## Notes

- This is a backend/runtime slice, not a UI style task.
