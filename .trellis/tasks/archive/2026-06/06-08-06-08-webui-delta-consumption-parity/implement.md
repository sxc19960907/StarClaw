# Web UI delta consumption parity implementation plan

## Steps

1. Inspect `streamMessage`, chat renderer, and agent-test renderer behavior.
2. Add duplicate-safe handling for `delta` and `assistant_text`.
3. Pass Kocoro-compatible metadata events through `renderer.onEvent`.
4. Add focused regression coverage for the event vocabulary and duplicate suppression.
5. Run validation:
   - `python3 ./.trellis/scripts/task.py validate .trellis/tasks/06-08-06-08-webui-delta-consumption-parity`
   - focused Web UI/static test
   - `go test ./internal/daemon`
   - `go test ./...`

## Rollback

Revert only the Web UI stream parser and its tests. The daemon dual-emission contract remains from the prior child.
