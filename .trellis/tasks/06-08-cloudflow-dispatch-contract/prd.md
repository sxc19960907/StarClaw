# Cloudflow dispatch contract

## Goal

Add the first Phase7 Kocoro parity slice: a local `internal/cloudflow` dispatch boundary that unifies slash command parsing, display metadata, and provider dispatch for `/research`, `/swarm`, and `/dag` workflows without calling external cloud services.

## Requirements

- Use Kocoro evidence:
  - `/Users/timmy/PycharmProjects/Kocoro/internal/cloudflow/parse.go`
  - `/Users/timmy/PycharmProjects/Kocoro/internal/cloudflow/display.go`
  - `/Users/timmy/PycharmProjects/Kocoro/internal/cloudflow/dispatch.go`
- Add `internal/cloudflow` with:
  - slash command parser for `/research`, `/swarm`, `/dag`
  - research strategy support: `quick`, `standard`, `deep`, `academic`
  - display status formatting helpers
  - provider interface and local/null dispatcher result
- Update daemon workflow parsing to use `internal/cloudflow`.
- Support `/dag` as `auto` workflow type mapped to a local prompt/step set.
- Keep cloud provider disabled by default; no network, credentials, or gateway calls.

## Acceptance Criteria

- [ ] `cloudflow.ParseSlash` parses `/research`, `/swarm`, and `/dag`.
- [ ] `/research deep query` captures strategy `deep`; `/research query` defaults to `standard`.
- [ ] Empty or malformed slash commands fall through or return validation errors consistently with daemon behavior.
- [ ] Daemon `parseWorkflowInvocation` supports `/dag`.
- [ ] Workflow metadata includes cloudflow type/strategy where applicable.
- [ ] Local dispatcher returns a deterministic unsupported/local result and never performs network I/O.
- [ ] Tests cover cloudflow parser/display/dispatch and daemon workflow parsing.
- [ ] Full project tests pass.

