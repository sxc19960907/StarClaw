# Guard watcher debounce timer

## Goal

Prevent stale watcher debounce timer callbacks from flushing a newer batch after the debounce timer has been reset or stopped.

## Requirements

- Add a generation or cancellation guard around watcher debounce timer callbacks.
- Preserve debounce behavior: bursts for the same agent should still coalesce into a single run after the quiet period.
- Preserve watcher close behavior: closing the watcher should prevent pending callbacks from dispatching runs.
- Keep the change local to watcher timer lifecycle handling.

## Acceptance Criteria

- [ ] A stale debounce callback does not flush a batch created by a newer timer generation.
- [ ] Pending debounce callbacks are invalidated when the watcher closes.
- [ ] Watcher tests pass.
- [ ] Full repository tests pass.

## Notes

- Relevant guideline: `.trellis/spec/backend/quality-guidelines.md` requires generation/cancellation guards for resettable and stoppable timers.
