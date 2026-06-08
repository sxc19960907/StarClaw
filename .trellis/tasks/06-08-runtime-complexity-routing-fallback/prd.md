# Runtime complexity routing and fallback

## Goal

Add runtime complexity routing and model fallback so Astria can choose an execution route/model tier and downgrade or escalate based on task shape, failures, and budget posture.

## Requirements

- Define a deterministic task complexity classifier using prompt shape, requested tools, evidence needs, delivery risk, and budget settings.
- Map complexity classes to local route/model recommendations.
- Add fallback decisions for provider failure, budget pressure, missing evidence, or repeated same-class run failure.
- Keep routing transparent in run metadata or diagnostics.
- Preserve operator approval boundaries for external or risky actions.

## Acceptance Criteria

- [ ] Complexity classifier has tests for simple, evidence-heavy, council-worthy, delivery-sensitive, and budget-constrained prompts.
- [ ] Runtime can produce a route/model recommendation without executing a paid call.
- [ ] Fallback decisions are test-covered for provider error, budget exhaustion, and repeated failure.
- [ ] Run metadata exposes selected route/fallback reason where applicable.
- [ ] Existing runs and Web UI behavior remain compatible.

## Notes

- This is a backend/runtime slice; UI indicators can follow later only if needed.
