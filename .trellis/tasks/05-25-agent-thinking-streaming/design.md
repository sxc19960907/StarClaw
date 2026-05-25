# Design

## Architecture

The existing option carrier is `client.ChatOptions`. The implementation should keep using that contract instead of adding provider-specific settings to the agent loop.

`cmd` owns construction-time wiring:

- `config.Config.Agent` is converted with `agent.ThinkingConfigFromAgent`.
- The loop receives thinking config, reasoning effort, specific model, and streaming enabled.
- A small helper should apply these options in one place so `runChat` and `interactive` stay in sync.

`internal/agent` owns request-time behavior:

- `Run` builds one `client.ChatOptions` from loop fields.
- `chatWithRetry` tries streaming only when `enableStreaming` is set and the client implements `StreamingLLMClient`.
- If streaming succeeds, it returns that response without falling through to non-streaming chat.
- If streaming fails with a retryable error, retry logic remains unchanged.
- If the client does not support streaming, it uses non-streaming `Chat`.

`internal/tui` owns UI delivery:

- `TUIEventHandler.OnStreamDelta` should send `streamingMsg` into the Bubble Tea program.
- To avoid global state, the TUI model should store the program pointer after program construction.

## Compatibility

- Existing configs remain valid because fields already exist and defaults already include thinking.
- Clients that do not implement streaming continue through non-streaming chat.
- CLI final output must suppress duplicate final printing when streaming has already printed deltas.

## Trade-Offs

- Streaming is enabled at loop construction instead of adding a config flag. This gives immediate UX value and falls back safely for unsupported clients.
- Anthropic native streaming remains out of scope because no parser exists yet for Anthropic stream events in this codebase.
