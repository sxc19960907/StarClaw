# Phase A Research: Existing Patterns

## 1. Client Package Structure

- `client.go` — Anthropic client + shared types (Message, ToolUse, Response, LLMClient interface)
- `openai.go` — OpenAI-compatible client (non-streaming)
- `ollama.go` — Ollama client (uses OpenAI-compatible endpoint, non-streaming)
- `mock.go` — Mock client for tests (no StreamChat)
- `gateway.go` — Generic HTTP client for daemon API (Post/Get with APIError)
- `sse.go` — SSE consumer for daemon event streams (reconnect with backoff)

## 2. StreamingLLMClient Interface (agent/loop.go:48-51)

```go
type StreamingLLMClient interface {
    StreamChat(ctx context.Context, systemPrompt string, messages []client.Message, tools []client.ToolDef, maxTokens int, opts *client.ChatOptions, onDelta func(delta string)) (*client.Response, error)
}
```

Used in `chatWithRetry` at loop.go:575-585 — tries streaming on first attempt if enabled.

## 3. Retry Logic (agent/loop.go:566-643)

Already has:
- 3 retries with exponential backoff (1s, 2s, 4s)
- `isRetryableLLMError` checks: 429, 500-504, timeout, connection errors, EOF
- Context cancellation short-circuits

Missing:
- `Retry-After` header parsing (can't access headers from error string alone)
- Jitter
- Reusable outside agent loop

## 4. SSE Parsing (sse.go)

The existing `SSEClient` is for daemon event consumption (GET /events). It:
- Uses `bufio.Scanner` line-by-line
- Handles `id:`, `event:`, `data:` fields
- Multi-line data via newline concatenation
- Reconnects with exponential backoff

For LLM streaming we need a simpler inline parser (POST response body, no reconnect).

## 5. Config Loading (config/config.go:87-120)

Currently single-file via viper:
- Reads `~/.starclaw/config.yaml`
- Sets defaults for all fields
- No project-level or local override

## 6. Context Window (context/window.go)

- `EstimateTokens` — chars/3.5 + 4 overhead per message
- `ShouldCompact` — triggers at 85% of context window
- `ShapeHistory` — sliding window with summary injection

## 7. Test Patterns (client/*_test.go)

- `sse_test.go` (7293 lines) — tests SSE event parsing
- `gateway_test.go` (4663 lines) — tests GatewayClient
- `client_test.go` (3056 lines) — tests AnthropicClient
- `ollama_test.go` (10178 lines) — tests OllamaClient
- Pattern: httptest.NewServer for mock HTTP, table-driven tests

## 8. Key Decisions

- SSE for LLM streaming should NOT reuse the daemon SSEClient (different lifecycle: POST body vs GET reconnect)
- Extract a `parseSSELines(reader, callback)` helper that both can share
- OpenAI streaming format: `data: {"choices":[{"delta":{"content":"..."}}]}\n\n` then `data: [DONE]\n\n`
- Tool call deltas come as `{"choices":[{"delta":{"tool_calls":[{"index":0,"function":{"arguments":"..."}}]}}]}`
- Retry logic should move to `client/retry.go` for reuse by daemon/gateway
