# Browser handoff lease depth design

## Boundary

This task adds ownership tracking around the existing StarClaw browser tool. It does not change how browser actions are executed.

## Design

Add `internal/tools/browser_lease.go`:

- `BrowserUseLease`
- `WithBrowserUseLease(ctx)`
- `BrowserUseLeaseFrom(ctx)`
- `MarkBrowserUsed(ctx, owner)`
- test-only active count helpers

Add `internal/tools/browser_handoff.go`:

- A callback-based handoff helper that takes a browser owner and cleanup callback.
- It marks the owner as deprecated and runs cleanup immediately only when no lease references it.
- If a lease is active, cleanup is deferred to lease release or a later watchdog.

Since `BrowserTool` currently has no cleanup lifecycle, add minimal internal state:

- `deprecated` flag.
- test-only cleanup call count through callback wrapper or internal method.
- No user-visible behavior changes.

## Run Integration

In daemon `RunAgentWithApproval`:

- Wrap the run context with `tools.WithBrowserUseLease`.
- Defer `ReleaseAndMaybeTeardown(nil)` for this task. Future persistent browser work can pass a real cleanup callback.

`BrowserTool.Run` calls `MarkBrowserUsed(ctx, t)` before action dispatch.

## Safety

- No persistent browser process is introduced.
- No external browser automation dependency is added.
- The lease tracker is concurrency-safe and idempotent.

## Rollback

Rollback removes the lease/handoff files, BrowserTool marking, and daemon context wrapping.
