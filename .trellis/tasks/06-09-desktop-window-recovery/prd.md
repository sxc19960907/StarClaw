# Desktop window recovery

## Goal

Make the standalone Astria shell resilient across app restart, Web UI reload,
daemon reconnect, and daemon crash/restart so it behaves like a product shell
rather than a browser shortcut.

## Requirements

- Restore the primary Astria window to the last useful route or a safe home
  route after app restart.
- Preserve Web UI recovery behavior added in Phase12: EventSource reconnect,
  run lifecycle recovery, and refreshed run state after reconnect.
- Surface daemon health transitions inside the shell in a user-visible but
  non-destructive way.
- Handle daemon crash/restart without losing local run history or creating
  duplicate active-run UI state.
- Keep browser launch fallback behavior intact.

## Acceptance Criteria

- [ ] Shell restart restores a usable Astria window.
- [ ] Web UI reload recovers current runs through existing `/events` and
      `/runs` behavior.
- [ ] Daemon disconnect/crash produces a clear shell state and recovery path.
- [ ] Reconnect does not duplicate active run cards or stale approval states.
- [ ] Recovery behavior is covered by automated Web UI tests or documented
      native smoke steps where automation is not yet available.

## Notes

Depends on the app launcher having a shell window and health state to observe.
