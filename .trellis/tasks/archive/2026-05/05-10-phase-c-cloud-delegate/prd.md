# Phase C: Cloud Agent Delegation

## Goal

Enable StarClaw to delegate sub-tasks to remote cloud agents, allowing multi-agent collaboration. A local agent can offload work (research, code generation, review) to a cloud endpoint, receive streaming progress updates, and incorporate results back into its context.

## Requirements

### 1. Cloud Delegate Tool (`internal/tools/cloud_delegate.go`)

- New tool `cloud_delegate` registered in the tool registry
- Accepts: `task` (string description), `agent` (optional remote agent name), `timeout` (optional seconds)
- Sends task to configured cloud endpoint via HTTP POST
- Streams progress events back via SSE
- Returns final result as tool output
- Respects cancellation (context.Done)

### 2. Cloud Client (`internal/client/cloud.go`)

- `CloudClient` struct wrapping HTTP + SSE for cloud agent communication
- `Delegate(ctx, req CloudDelegateRequest) (*CloudDelegateResponse, error)` — blocking call
- `DelegateStream(ctx, req, onProgress func(CloudProgress)) (*CloudDelegateResponse, error)` — streaming
- Handles auth (API key header), timeout, retry on transient errors
- Progress events: `thinking`, `tool_call`, `text`, `done`, `error`

### 3. Cloud Event Bus Integration

- Cloud delegate progress events are forwarded to the daemon EventBus
- New event types: `EventCloudDelegateStart`, `EventCloudDelegateProgress`, `EventCloudDelegateComplete`
- SSE subscribers (TUI, web clients) can observe cloud delegation in real-time

### 4. Config: Cloud Section

- Add `Cloud` field to `Config` struct:
  ```yaml
  cloud:
    enabled: true
    endpoint: "https://api.starclaw.cloud/v1"
    api_key: ""
    timeout: 3600
    max_concurrent: 3
  ```
- Add to `defaultConfig()` in multilevel.go
- Add to overlay merge logic

### 5. Agent Loop Integration

- When cloud is enabled and the agent loop detects a delegatable sub-task, it can use the `cloud_delegate` tool
- The tool is registered like any other tool — the LLM decides when to use it
- Progress callbacks update the EventHandler (OnStreamDelta for intermediate text)

## Acceptance Criteria

- [ ] `cloud_delegate` tool is registered and callable by the LLM
- [ ] `CloudClient.Delegate` sends POST with task description and returns result
- [ ] `CloudClient.DelegateStream` streams progress events via SSE
- [ ] Progress events are published to daemon EventBus
- [ ] Cloud config section loads from all 3 config levels
- [ ] Tool respects timeout and context cancellation
- [ ] Graceful handling when cloud is disabled or endpoint unreachable
- [ ] Unit tests for CloudClient (mock HTTP server), tool, and config
- [ ] `go build ./...` and `go test ./...` pass

## Technical Notes

- Cloud endpoint API contract (what we POST to):
  ```
  POST /v1/delegate
  Headers: X-API-Key: <key>, Content-Type: application/json
  Body: {"task": "...", "agent": "...", "context": "..."}
  Response: SSE stream with progress events, final event has type "done"
  ```
- Reuse `internal/client/gateway.go` patterns for HTTP
- Reuse `internal/client/stream.go` SSE parsing patterns (adapt for non-OpenAI format)
- The `cloud_delegate` tool should be in `microcompact.go`'s exclusion list (already is)
- Event types follow existing daemon/events.go pattern
