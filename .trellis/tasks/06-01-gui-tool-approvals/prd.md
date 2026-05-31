# Add GUI tool approvals

## Goal

Make the daemon-hosted Web UI capable of handling human approval for tool calls that require explicit confirmation, so a browser-based StarClaw run can continue or deny risky operations without switching to the terminal.

## Confirmed Facts

- The daemon already exposes `POST /approval` and an `ApprovalBroker`.
- The daemon already exposes `GET /events` as an SSE event stream and defines `approval_needed` / `approval_resolved` event types.
- The Web UI already supports `/message` streaming, tool-call detail rendering, and cancellation.
- The agent loop currently checks configured permissions for `deny`, but `ask` is not yet wired to a blocking approval decision.
- CLI and TUI approval UX are not complete blocking approval flows; this task targets daemon/Web UI only.

## Requirements

- Add a daemon-run approval path that blocks a tool call when permissions return `ask` or when a tool requires approval and no explicit allow rule bypasses it.
- Publish `approval_needed` events with enough data for the GUI to render the request: approval request id, run request id/thread id, tool name, args, agent, channel, and reason when available.
- Resolve approval decisions through existing `POST /approval` with `allow` or `deny`.
- Publish `approval_resolved` events after a decision is accepted.
- Update the Web UI to subscribe to `/events`, render pending approvals in the chat workspace, and provide Allow/Deny controls.
- Show approval state transitions clearly: pending, allowed, denied, expired/cancelled where detectable.
- Keep the embedded GUI dependency-free and same-origin.
- Preserve existing `/message` SSE behavior and existing CLI/TUI behavior unless directly needed for shared agent-loop interfaces.
- Leave unrelated `.agents/skills/obsidian-cli/SKILL.md` untouched.

## Acceptance Criteria

- [ ] A daemon `/message` run that reaches an `ask` permission decision emits an approval request and waits for a GUI/API decision before executing the tool.
- [ ] `POST /approval` with `allow` lets the waiting tool execute and the run continue.
- [ ] `POST /approval` with `deny` returns a permission-style tool result instead of executing the tool.
- [ ] The Web UI subscribes to `/events` and renders pending approval cards with tool args.
- [ ] The Web UI Allow/Deny buttons call `/approval` and update card state.
- [ ] Backend tests cover approval wait/allow/deny behavior and event payloads.
- [ ] Web UI route/static tests remain green.
- [ ] `node --check internal/daemon/webui/assets/app.js`, `go test ./internal/daemon ./cmd`, `go test ./...`, and `go vet ./...` pass.

## Out Of Scope

- Full interactive CLI/TUI approval implementation.
- Persistent approval rules or editing permissions config from the GUI.
- Multi-user authentication or remote access control for the local daemon.
