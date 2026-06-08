# Phase 5 Web UI bug bash

## Goal

Fix integration UX issues discovered while validating Phase 3/4 runtime features across Astria Web UI surfaces.

## Requirements

- Focus on Mission Control, Run detail, Budget, Quality, Reuse, Share, Memory, Recovery, and Trace navigation.
- Fix defects that block validation or confuse the runtime workflow.
- Keep UI changes scoped and consistent with existing embedded static assets.
- Avoid new panels unless a missing state prevents validation.

## Acceptance Criteria

- [x] Runtime/recovery/trace/budget states are readable and do not overlap on common viewport sizes.
- [x] Navigation from quality/reuse/share/memory-related cards to runs or chat remains coherent.
- [x] Empty/error states appear for failed trace/API calls.
- [x] Web UI smoke/static tests cover any added hooks.
- [x] `go test ./internal/daemon ./cmd`, `go test ./...`, and `git diff --check` pass after fixes.

## Non-Goals

- No broad visual redesign.
- No frontend build pipeline.
- No new backend runtime behavior unless required to fix an integration bug.
