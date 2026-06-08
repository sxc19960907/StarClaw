# Design

## Current StarClaw State

StarClaw has a daemon HTTP API and embedded Web UI, but no native Desktop RPC channel. Future native Astria/Desktop integration needs a separate local boundary so native-only capabilities can be handled outside the daemon while preserving daemon ownership of run state, tools, approvals, and observability.

## Kocoro Evidence

Kocoro implements Desktop RPC as:

- a broker that correlates daemon requests to Desktop results
- length-prefixed JSON frames over Unix domain socket
- listener that installs a transport send function on connect
- versioned `system.ping` and `system.capabilities`
- calendar methods layered on top of the same boundary

This StarClaw child mirrors the foundation without calendar/EventKit.

## Proposed Architecture

Add `internal/daemon/desktop_rpc/`:

- `types.go`
  - protocol version
  - frame/request/result/error types
  - method and error constants
  - capability payloads

- `codec.go`
  - 4-byte big-endian length prefix
  - max frame body cap
  - JSON encode/decode

- `broker.go`
  - pending request map
  - request id generation
  - request timeout
  - result resolution
  - disconnect cancellation

- `listener.go`
  - Unix socket listener
  - accept loop
  - frame read loop
  - send callback installation
  - local handling for Desktop-originated `system.ping` / `system.capabilities`

Add daemon wiring:

- `Server` owns `desktopRPC *desktop_rpc.Broker`.
- `Server` can report Desktop RPC status in `GET /status`.
- Start/stop listener support can be added with a method and tests. If full daemon startup wiring is too invasive, this child can expose construction/listening APIs and unit tests as the verified boundary.

## Wire Contract

Frame:

```json
{"type":"desktop_rpc_request","payload":{...}}
```

Request:

```json
{"request_id":"drpc_...","method":"system.ping","params":{"echo":"hi"},"timeout_ms":30000,"ts":"..."}
```

Result:

```json
{"request_id":"drpc_...","ok":true,"result":{"pong":"hi","server_time":"..."}}
```

## Scope Boundaries

- No calendar methods in this child.
- No native Desktop app launch.
- No cloud or external transport.
- No payload-rich structured telemetry.

## Rollback

Remove `internal/daemon/desktop_rpc` and status wiring. Existing daemon HTTP/API behavior remains unchanged.
