# Provider stream watchdog parity design

## Baseline

Kocoro exposes `agent.stream_idle_timeout_secs` with a default of `90`. Its provider stream watchdog closes the response body when no SSE chunk arrives within the configured gap and returns `client.ErrStreamIdleTimeout`. Agent retry logic treats that error as terminal for the turn so a silent upstream stream does not hang and then re-hang in non-streaming fallback.

StarClaw currently has SSE client idle-watchdog support for local event streams, but provider streaming uses blocking `bufio.Scanner` loops in `ParseOpenAIStream` and `ParseAnthropicStream`. A silent upstream can keep the HTTP body read blocked until transport timeout or manual cancellation.

## Runtime Contract

- `agent.stream_idle_timeout_secs > 0`: provider stream readers use a watchdog that measures time between scanner lines.
- `agent.stream_idle_timeout_secs == 0`: provider stream readers use the existing scanner path.
- Timeout returns `ErrStreamIdleTimeout`, wrapped by callers where useful but still detectable via `errors.Is`.
- If partial content/tool state exists, parser returns the partial `Response` alongside the timeout error.
- Agent retry logic treats `ErrStreamIdleTimeout` as non-retryable and does not fall back to non-streaming for the same attempt.

## Implementation Shape

- Add `StreamIdleTimeoutSecs` to `config.AgentConfig`, defaults, validation, and tests.
- Add a sentinel `ErrStreamIdleTimeout` plus stream parser options in `internal/client`.
- Keep existing `ParseOpenAIStream` and `ParseAnthropicStream` signatures by delegating to option-enabled helpers with zero timeout.
- Add timeout-enabled helpers used by provider clients.
- Add optional `SetStreamIdleTimeout` / `StreamIdleTimeout` methods to Anthropic/OpenAI/Ollama clients to keep construction localized and testable.
- Configure clients from `cmd/root.go` and daemon client setup where provider instances are created.
- Update agent `chatWithRetry` to terminally return stream-idle timeout without retry/fallback.

## Compatibility

The default enables the watchdog, which is a behavior change only for silent stalled streams. Healthy streams continue through the same parser logic. Users can set `agent.stream_idle_timeout_secs: 0` to restore legacy blocking scanner behavior.
