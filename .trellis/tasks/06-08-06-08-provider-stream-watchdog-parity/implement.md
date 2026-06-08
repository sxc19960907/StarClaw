# Provider stream watchdog parity implementation plan

## Steps

1. Update config:
   - add `Agent.StreamIdleTimeoutSecs`
   - default to `90`
   - validate `>= 0`
   - add config tests for default, YAML override, and negative rejection

2. Update client stream parsing:
   - add `ErrStreamIdleTimeout`
   - add timeout-enabled parser options for OpenAI and Anthropic stream readers
   - keep legacy parser signatures intact
   - add tests with short timeouts and hanging readers

3. Wire providers:
   - add stream idle timeout fields/setters to Anthropic, OpenAI-compatible, and Ollama clients
   - have `StreamChat` use timeout-enabled parsing when configured
   - wire configured timeout in CLI and daemon client setup

4. Update agent retry behavior:
   - ensure `ErrStreamIdleTimeout` exits immediately from streaming path
   - do not retry and do not fall back to non-streaming for this error
   - add regression test

5. Verify:
   - `python3 ./.trellis/scripts/task.py validate .trellis/tasks/06-08-06-08-provider-stream-watchdog-parity`
   - focused tests in `internal/client`, `internal/config`, `internal/agent`, `cmd`
   - `go test ./...`

## Rollback

Revert client parser timeout options, provider setters, config field/default/validation, and agent retry special-case together. Leaving only the config or only the parser would create misleading knobs.
