# Implementation Plan

## Steps

1. Add delivery result payload and class constants in `internal/daemon/types.go`.
2. Add `delivery_inject.go` with label, formatting, consumer, and exported single-call helper.
3. Add focused tests for formatting, success silence, known-route enqueue, route miss drop, and trusted route-scoped events.
4. Run:
   - `go test ./internal/daemon`
   - `go test ./...`
   - `git diff --check`
   - `python3 ./.trellis/scripts/task.py validate .trellis/tasks/06-08-delivery-inject-lifecycle-depth`
5. Commit and archive the child task.

## Review Gates

- Do not add real external channel transport.
- Do not enqueue delivery failures without a route index hit.
- Do not use broad observability stores for model-facing delivery events.
- Do not claim transient failures are permanent.

## Completion Notes

TBD.
