# EventBus replay SSE resilience implementation plan

## Steps

1. Add `SubscribeWithReplay(id, lastID string) ([]Event, <-chan Event)` to `internal/daemon/events.go`.
2. Refactor `/events` handler to use `SubscribeWithReplay` once and then write replay events before live events.
3. Add EventBus unit tests:
   - valid cursor returns missed events and subscribes for future events,
   - invalid cursor replays all history,
   - replayed event and later live event are not duplicated.
4. Extend `/events` handler tests to cover replay followed by a live event after subscription.
5. Add minimal Astria Web UI event-stream state:
   - track last `event.lastEventId`,
   - increment reconnect count after error/open recovery,
   - expose recovered status through existing daemon pill or state.
6. Add static Web UI contract test markers.
7. Validate:
   - `python3 ./.trellis/scripts/task.py validate .trellis/tasks/06-08-eventbus-replay-sse-resilience`
   - `go test ./internal/daemon -run 'TestEventBus|TestHandleEvents|TestWebUI' -count=1 -timeout=90s`
   - `go test ./internal/daemon -count=1 -timeout=90s`
   - `go test ./...`

## Risk Areas

- Avoid changing existing EventBus publish/drop semantics.
- Avoid blocking publisher paths on slow subscribers.
- Avoid Web UI text that can overlap existing compact header layout.

## Follow-Up

If this child discovers broader lifecycle vocabulary gaps, capture them for `run-session-lifecycle-events` rather than expanding this child.
