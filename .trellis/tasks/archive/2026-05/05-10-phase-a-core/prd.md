# Phase A: Client Streaming, Retry, Agent Loop Depth

## Goal

Implement full SSE streaming for OpenAI and Ollama clients, add a retry-aware gateway client layer, and deepen the agent loop with web query normalization and context partition retry — closing the critical "core usability" gap with ShanClaw.

## Requirements

### 1. OpenAI Client StreamChat (SSE)

- Implement `StreamChat` on `OpenAIClient` satisfying the `StreamingLLMClient` interface
- Use `"stream": true` in the request body
- Parse SSE `data:` lines incrementally
- Accumulate `tool_calls` across multiple deltas (index-based merging)
- Call `onDelta(text)` for each content chunk
- Handle `[DONE]` sentinel
- Return final assembled `*Response` identical to non-streaming path
- Handle partial JSON in function arguments (buffer until complete)

### 2. Ollama Client StreamChat (SSE)

- Implement `StreamChat` on `OllamaClient`
- Ollama's `/v1/chat/completions` supports `"stream": true` with same SSE format as OpenAI
- Reuse the SSE parsing logic from OpenAI (extract shared helper)
- Longer timeout (600s already set)

### 3. Shared SSE Parser

- Extract `internal/client/sse.go` with reusable SSE line parser
- Handle: `data:`, `event:`, empty lines (event boundary), `[DONE]`
- Provide `parseSSEStream(reader io.Reader, onEvent func(data []byte) error) error`
- Handle connection drops mid-stream gracefully (return partial + error)

### 4. Retry Layer Enhancement

- Current `chatWithRetry` in agent/loop.go already handles basic retry
- Add: rate-limit aware backoff (parse `Retry-After` header when 429)
- Add: jitter to exponential backoff to avoid thundering herd
- Move retry classification to `internal/client/retry.go` so it's reusable outside agent loop

### 5. Multi-level Config Merge

- Support three config layers: global (`~/.starclaw/config.yaml`), project (`.starclaw/config.yaml`), local (`.starclaw/config.local.yaml`)
- Project and local configs override global using pointer-based overlay merge
- Add `ConfigSource` tracking: for each field, record which layer set it
- Implement `MergeConfigs(global, project, local *Config) *Config`

### 6. Agent Loop: Web Query Normalize

- Add `internal/agent/normalize.go`
- Normalize user input before sending to LLM: extract URLs, detect search intent
- Strip markdown artifacts from pasted content
- Normalize whitespace and control characters

### 7. Agent Loop: Context Partition Retry

- When LLM returns context-too-large error, automatically partition messages
- Split conversation into chunks, summarize older chunks, retry with reduced context
- Integrate with existing `contextWindow` field in AgentLoop

## Acceptance Criteria

- [ ] `OpenAIClient` implements `StreamingLLMClient` — streaming works with real OpenAI API
- [ ] `OllamaClient` implements `StreamingLLMClient` — streaming works with local Ollama
- [ ] SSE parser handles partial reads, connection drops, and `[DONE]`
- [ ] Tool call deltas are correctly merged by index across multiple SSE events
- [ ] Retry logic respects `Retry-After` header and adds jitter
- [ ] Config loads from 3 levels with correct override precedence
- [ ] `ConfigSource` can report which layer set each field
- [ ] Web query normalize strips control chars and extracts URLs
- [ ] Context partition retry triggers on oversize error and succeeds with reduced context
- [ ] All new code has unit tests
- [ ] `go build ./...` passes
- [ ] `go test ./...` passes

## Technical Notes

- The `StreamingLLMClient` interface already exists in `internal/agent/loop.go:49`
- `chatWithRetry` already exists at `loop.go:567` with basic exponential backoff
- Config currently loads from single file via viper (`config.go:87`)
- OpenAI SSE format: each line is `data: {json}\n\n`, final line is `data: [DONE]\n\n`
- Ollama uses identical SSE format via its OpenAI-compatible endpoint
- `parseOpenAIResponse` in `openai.go` can be reused for final assembly after streaming
