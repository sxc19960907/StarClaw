# Design

## Architecture

Add approval as an optional agent-loop dependency:

- `internal/agent/loop.go` defines a small optional approval interface and calls it only when a tool needs human confirmation.
- `internal/daemon/approval_handler.go` adapts `ApprovalBroker` and `EventBus` to that interface.
- `internal/daemon/runner.go` injects the daemon approval handler for daemon-originated runs.
- `internal/daemon/webui/assets/app.js` subscribes to `/events`, renders approval cards, and posts decisions to `/approval`.

## Approval Decision Flow

1. LLM requests a tool call.
2. Agent loop evaluates `permissions.CheckToolCall`.
3. `deny` returns `PermissionError` immediately.
4. `allow` executes the tool.
5. `ask` calls the optional approval requester.
6. Daemon approval requester:
   - creates an approval request id
   - publishes `approval_needed`
   - waits on `ApprovalBroker.WaitForApproval`
   - publishes `approval_resolved`
7. `allow` executes the tool.
8. `deny`, timeout, or context cancellation returns a permission-style tool result without executing the tool.

For tools outside the permissions engine, `tool.RequiresApproval()` still triggers approval unless the permission decision already returned `allow`.

## Event Contracts

`approval_needed` payload:

```json
{
  "request_id": "apr_x",
  "thread_id": "web-run-request-id",
  "channel": "http",
  "tool": "bash",
  "args": "{\"command\":\"...\"}",
  "agent": "helper",
  "reason": "requires approval"
}
```

`approval_resolved` payload:

```json
{
  "request_id": "apr_x",
  "decision": "allow",
  "resolved_by": "web"
}
```

`POST /approval` remains:

```json
{"request_id":"apr_x","decision":"allow"}
```

## Compatibility

- Existing handlers that do not configure an approval requester continue to treat `ask` as denied, not silently allowed.
- Existing daemon approval endpoint remains backward-compatible.
- Web UI listens to global daemon events; cards are displayed in the current browser workspace even when the related run is active.

## Trade-offs

- The first pass uses daemon-wide `/events` rather than per-run message SSE approval events. This reuses existing infrastructure and keeps `/message` streaming stable.
- CLI/TUI approval remains separate. Extending the shared approval requester into those surfaces should be a later task with dedicated UX decisions.

## Risks

- If a browser is not open, approval requests may time out and deny.
- If multiple browser tabs are open, all tabs may see the same approval. The broker resolves only once; later decisions are no-ops.
