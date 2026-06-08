# External channel adapter boundaries

## Goal

Define local-first external channel adapter contracts and fake adapter registry so StarClaw has Kocoro-like channel management boundaries without enabling real off-machine transports or credentials.

This is Phase 7 child 5 under `Astria Kocoro parity phase 7: channel and cloud delivery parity`.

## Confirmed Facts

- Kocoro has Feishu/Lark Cloud passthrough management in `/Users/timmy/PycharmProjects/Kocoro/internal/daemon/feishu_handler.go`.
- StarClaw already has local inbox, queue, route index, connection state, system event, and delivery inject foundations.
- StarClaw does not have Slack/Feishu/Telegram adapter contracts or install/list/delete management boundaries.
- Real external channel transport requires credentials and off-machine delivery, which remain out of scope unless explicitly approved.

## Requirements

- Add a daemon-local channel adapter interface for management-style operations.
- Support adapter metadata: provider, display name, transport kind, configured/enabled state, capabilities, and privacy note.
- Support install/list/delete-style methods through fake/local adapters.
- Add a registry that can register, list, and retrieve adapters by provider.
- Add read-only daemon API to list channel adapter metadata.
- Add test-only or fake adapter behavior proving install/list/delete contracts without real network calls.
- Do not persist or expose secrets.
- Do not connect to Slack, Feishu/Lark, Telegram, LINE, or cloud services.

## Acceptance Criteria

- [ ] Unit tests cover registry register/list/get behavior.
- [ ] Unit tests cover fake adapter install/list/delete lifecycle.
- [ ] API tests cover listing adapters and metadata shape.
- [ ] API output makes clear real transports are disabled/local-only.
- [ ] Existing inbox/queue/channel APIs remain compatible.
- [ ] `go test ./internal/daemon` passes.
- [ ] `go test ./...` passes.
- [ ] `git diff --check` passes.

## Out of Scope

- Real external transport or Cloud Gateway calls.
- Credential storage or keychain work.
- WebSocket cloud-controller lifecycle.
- UI changes.

## Evidence

- `.trellis/research/kocoro-local-comparison-phase7-plan.md`
- `/Users/timmy/PycharmProjects/Kocoro/internal/daemon/feishu_handler.go`
