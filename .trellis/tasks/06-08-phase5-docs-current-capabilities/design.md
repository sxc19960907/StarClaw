# Phase 5 docs current capabilities design

## Scope

Update concise user-facing documentation to match validated Phase 3/4 platform capabilities and Phase 5 hardening evidence. Documentation must distinguish StarClaw CLI/runtime naming from Astria Web UI product naming.

## Target Content

- Astria Web UI workflow surfaces and runtime validation areas.
- Local OpenAI-compatible API gateway.
- Workflow-control endpoints for cancel, pause, resume, and replay approval boundaries.
- Runtime token budget enforcement, complexity routing, and fallback metadata.
- Structured events, metrics, trace read/export, durable recovery, and replay metadata.
- Local-only safety boundaries, redaction guarantees, and limitations.

## Approach

1. Inspect existing README/docs structure and avoid duplicating outdated sections.
2. Update the smallest user-facing docs set that covers operational use.
3. Keep claims grounded in tests and implemented routes.
4. Avoid Phase 6 speculation, cloud claims, or hosted docs.

## Rollback

Revert documentation/task artifact changes. No runtime code changes are expected.
