# Design

## CLI Shape

- `starclaw app`
  - user-facing primary GUI launcher
  - ensures daemon health before opening the browser
- `starclaw daemon open --start`
  - compatible advanced form under existing daemon namespace
  - same ensure-and-open implementation
- `starclaw daemon open`
  - unchanged: only opens the configured Web UI URL

## Implementation

Add helper functions in `cmd/daemon.go` near existing daemon commands:

- `daemonHealthURL()`
- `isDaemonHealthy(ctx)`
- `startDaemonBackground()`
- `waitForDaemonHealth(ctx)`
- `ensureDaemonRunning(ctx)`
- `openDaemonWebUI(cmd, ensure bool)`

Use package-level function variables for test seams:

- health check HTTP client or function
- executable path resolver
- command starter
- browser opener already exists

Background daemon startup uses the current executable:

```bash
starclaw daemon start
```

The command is detached enough for CLI return: stdout/stderr go to `os.DevNull`; the process is started and not waited on. Health polling determines success.

## Error Handling

- If health is already OK, do not start another daemon.
- If the daemon process cannot start, return `daemon: start background daemon: ...`.
- If health does not become ready before timeout, return an actionable timeout error.
- Browser open failures still return the existing `daemon: open web UI: ...` style error.

## Docs

Update README and docs examples to show:

```bash
starclaw app
```

and keep the manual daemon commands as advanced/lifecycle alternatives.
