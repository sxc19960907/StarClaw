# Phase 5 Kocoro gap audit design

## Scope

This task produces a local-evidence gap audit after the Phase 5 hardening children. The audit is documentation only: it does not change runtime behavior, add Phase 6 code, or benchmark external products live.

## Inputs

- Completed Phase 5 child task artifacts and commits.
- Local code and tests under `internal/` and `cmd/`.
- User-facing docs under `README.md` and `docs/`.
- Prior local research in `.trellis/research/gap-analysis-starclaw-vs-shanclaw.md`.

## Output

Create `.trellis/tasks/06-08-phase5-kocoro-gap-audit/gap-audit.md` with:

- A status matrix using exactly `aligned`, `partially aligned`, `missing`, or `unknown`.
- Evidence links to local files, tests, docs, or archived task artifacts.
- A short Phase 6 recommendation ordered by user value and platform risk.

## Classification rules

- `aligned`: Phase 5 evidence shows the local StarClaw/Astria capability exists at the platform level for the intended local-first scope.
- `partially aligned`: A meaningful implementation exists, but important depth, durability, UX, or compatibility remains.
- `missing`: Local evidence shows the area is not implemented or only planned.
- `unknown`: Kocoro/Shannon behavior cannot be verified from local evidence, or parity cannot be claimed without external data.

## Constraints

- Preserve product naming: StarClaw for CLI/module/package/release/docs; Astria for product-facing Web UI.
- Keep all conclusions local-first. Do not introduce cloud sync, hosted telemetry, or external collectors as current behavior.
- Cite uncertainty explicitly. Do not upgrade a status based on assumptions about Kocoro/Shannon.
