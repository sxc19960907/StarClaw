# Desktop RPC launch contract design

## Decision

Extend `starclaw daemon start` with an optional paired Desktop RPC mode:

```bash
starclaw daemon start --rpc-socket <path> --rpc-pidfile <path>
```

When neither flag is set, daemon startup remains HTTP-only. When both are set,
the daemon starts `desktop_rpc.Listener` with the existing server broker and
platform metadata. A partial flag pair is invalid and must fail before daemon
startup.

## Data Flow

1. Cobra parses `--rpc-socket` and `--rpc-pidfile`.
2. `daemonStartCmd` validates the pair.
3. `daemon.NewServer(...)` creates the existing Desktop RPC broker.
4. If Desktop RPC mode is enabled, create a listener config:
   - `SockPath`
   - `PidfilePath`
   - `Broker`
   - `Platform: desktop_rpc.DefaultPlatform(Version)`
5. Register listener on the server with `SetDesktopRPCListener`.
6. Start listener under the daemon context.
7. Start HTTP server.
8. Context cancellation shuts down listener and removes scoped artifacts.

## Compatibility

- Do not require desktop flags for existing daemon/app/browser flows.
- Do not expose socket or pidfile paths through `/status`.
- Calendar tools remain registered against the broker; when no Desktop client
  is connected, existing `desktop_disconnected` behavior remains valid.

## Error Handling

- Socket only -> `--rpc-pidfile is required when --rpc-socket is set`.
- Pidfile only -> `--rpc-socket is required when --rpc-pidfile is set`.
- Listener start failure -> daemon exits before reporting readiness.
- Pidfile write failure -> listener returns error and cleanup removes socket.

## Tests

- Cobra command tests for missing pair validation.
- Daemon/desktop_rpc integration test for listener with temp socket/pidfile.
- `/status` test that `desktop_rpc.listening` is true and paths are absent.
- Existing app/doctor tests to protect fallback behavior.
