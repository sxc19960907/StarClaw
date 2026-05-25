# Implementation Plan

1. Add Anthropic stream state and `ParseAnthropicStream` tests.
2. Refactor Anthropic request body construction into a shared helper.
3. Implement `AnthropicClient.StreamChat`.
4. Add client-level HTTP test for request body, headers, delta callback, and final response.
5. Run focused tests.
6. Run full test suite.

## Validation Commands

```bash
go test ./internal/client ./internal/agent ./cmd ./internal/tui
go test ./...
```
