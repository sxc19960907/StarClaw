# Daemon SSE event vocabulary parity design

## Boundary

Only the per-request `POST /message` SSE path changes. The global `/events` endpoint, OpenAI-compatible gateway, provider clients, Web UI, and cloud integration surfaces are out of scope.

## Event Mapping

| StarClaw callback | Existing event | New Kocoro-compatible event |
|---|---|---|
| `SetSessionID(id)` | none | `session_started` |
| `OnToolCall(name,args)` | `tool_call` | `tool` with `status=running` |
| `OnToolResult(name,result)` | `tool_result` | `tool` with `status=completed|error` |
| `OnStreamDelta(delta)` | `text` | `delta` |
| `OnPreamble(text)` | `preamble` | `assistant_text` |
| `OnUsage(usage)` | none | `usage` |
| run success/failure | `done` / `error` | unchanged |

## Compatibility

The task uses additive dual-emission. Legacy StarClaw event names remain available, while Kocoro-compatible clients can subscribe to the new names. This avoids a breaking API change and lets the later Web UI child choose which vocabulary to consume.

## Payload Contracts

`session_started`:

```json
{"session_id":"..."}
```

`delta`:

```json
{"text":"chunk"}
```

`assistant_text`:

```json
{"text":"mid-turn narration"}
```

`usage`:

```json
{"input_tokens":10,"output_tokens":20,"total_tokens":30}
```

`tool` running:

```json
{"tool":"file_read","status":"running","args":"..."}
```

`tool` completed/error:

```json
{"tool":"file_read","status":"completed","is_error":false,"preview":"...","error_category":""}
```

StarClaw cannot yet provide Kocoro's `tool_use_id`, elapsed seconds, cost, or model on this path without widening `agent.EventHandler`; those are explicit follow-ups.

## Privacy

Use the existing tool argument/result redaction and truncation helpers where available. Do not include raw provider payloads, API keys, or request bodies in new events.
