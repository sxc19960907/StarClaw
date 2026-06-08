# Astria Phase 5 integrated hardening and E2E validation design

## Scope

Phase 5 acts as an integration acceptance layer over already implemented runtime and product surfaces. The parent owns sequencing and success criteria; child tasks own concrete tests, fixes, docs, and the final gap audit.

## Validation Surfaces

- Browser/Web UI:
  - Home mission composer
  - Chat stream
  - Runs / Mission Control
  - Run detail timeline
  - Runtime recovery, workflow steps, control history, and trace sections
  - Budget, Quality, Reuse, Share, Memory, and Result routes when linked from run state
- Local daemon API:
  - `POST /message`
  - `GET /runs`, `GET /runs/{id}`, `POST /runs/{id}/control`
  - `GET /runs/{id}/trace`, `GET /traces/export`
  - `GET /metrics`
  - `POST /v1/chat/completions`
- Runtime records:
  - token `budget_status`
  - routing/fallback records
  - structured events
  - durable run store
  - workflow steps and control decisions

## Child Sequencing

1. Runtime E2E smoke establishes the main user/product path.
2. API/observability smoke validates local daemon contracts as integration interfaces.
3. Secret-leakage regression hardens safety across all export-like surfaces.
4. Web UI bug bash fixes issues discovered by the first three children.
5. Docs current capabilities updates user-facing guidance after validation.
6. Kocoro gap audit uses the validated product state to decide Phase 6.

## Boundaries

Phase 5 can fix defects discovered during validation. It should not add new broad feature areas, new remote services, or new build tooling.

## Rollback

Each child should be independently revertible. Parent-level artifacts can be archived without code changes if child tasks complete separately.
