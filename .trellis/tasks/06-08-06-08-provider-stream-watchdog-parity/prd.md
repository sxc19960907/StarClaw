# Provider stream watchdog parity

## Goal

Close the provider streaming liveness gap with the local Kocoro baseline by adding a per-chunk idle watchdog for StarClaw provider `StreamChat` paths.

Kocoro baseline: `/Users/timmy/PycharmProjects/Kocoro` at `74cdb3cfa010bc07a5ec3a5f9bda1854da41b245`.

## Requirements

- Add a local provider stream idle timeout setting under `agent.stream_idle_timeout_secs`.
- Default the setting to `90` seconds, matching Kocoro's provider stream watchdog default.
- Allow `0` to disable the watchdog and preserve the legacy blocking scanner path.
- Reject negative `agent.stream_idle_timeout_secs` during config validation.
- Apply the watchdog to provider streaming readers used by Anthropic, OpenAI-compatible, and Ollama streaming clients.
- When no stream line arrives before the configured idle timeout, abort the body read and return a sentinel stream-idle timeout error.
- Preserve partial response accumulation where available.
- Prevent agent retry/fallback from re-hanging on the same silent stream failure; the first stream-idle timeout should surface deterministically.
- Do not add cloud transport, off-machine telemetry, or Shannon-specific dependencies.

## Acceptance Criteria

- [ ] Config has `Agent.StreamIdleTimeoutSecs`, default `90`, YAML override support, and validation for negative values.
- [ ] Provider stream parsing supports a timeout-enabled path and returns a recognizable `ErrStreamIdleTimeout`.
- [ ] Anthropic, OpenAI-compatible, and Ollama `StreamChat` use the configured watchdog when enabled.
- [ ] Agent streaming retry logic does not retry or fall back to non-streaming after `ErrStreamIdleTimeout`.
- [ ] Unit tests cover timeout-enabled OpenAI and Anthropic stream parsing, disabled legacy parsing, config defaults/overrides/validation, and agent no-retry behavior.
- [ ] Existing streaming behavior and non-streaming providers remain compatible.

## Notes

- Kocoro also has richer gateway/cloud stream watchdog behavior. This child is scoped to StarClaw local provider streaming.
