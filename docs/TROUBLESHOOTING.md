# Troubleshooting Guide

## Common Issues

### "Configuration required"

**Cause**: No API key configured

**Solution**:
```bash
starclaw setup
# OR
export ANTHROPIC_AUTH_TOKEN="your-key"
```

### "API error (401)"

**Cause**: Invalid API key

**Solution**:
- Verify your API key is correct
- Check that the key hasn't expired
- Ensure you're using the right endpoint

### "Tool call cancelled"

**Cause**: User denied tool approval

**Solution**:
- Type 'Y' to approve, or
- Use `-y` flag for auto-approval:
  ```bash
  starclaw -y chat "your query"
  ```

### "Path outside working directory"

**Cause**: Tool tried to access files outside the project

**Solution**:
- Ensure you're running from the correct directory
- Check your configuration's `allowed_paths`

### Slow Responses

**Cause**: Large files or many tool calls

**Solution**:
- Use `offset` and `limit` for file_read
- Reduce `max_iterations` in config
- Configure `agent.token_budget` and `hard_stop` for long runs that need a firm cap
- Be more specific in your queries

### Budget exhausted

**Cause**: A configured runtime token budget was reached or projected to be reached.

**Solution**:
- Inspect the run in Astria Mission Control or `GET /runs/{id}`.
- Check `budget_status`, `routing`, and `fallback` metadata.
- Raise the relevant `agent.token_budget` limit, trim context, or rerun with a narrower prompt.

### Replay requires approval

**Cause**: `POST /runs/{id}/control` with `{"action":"replay"}` is approval-gated because replay can repeat tool calls or side effects.

**Solution**:
- Review the redacted replay plan first.
- Use `{"action":"replay","approved":true}` only when the source run and side effects are safe to repeat.
- Inspect source and replay links in run control metadata.

### Trace or metrics look incomplete

**Cause**: Metrics are aggregate-only and traces are derived from structured events, not raw prompts or provider transcripts.

**Solution**:
- Use `GET /metrics` for aggregate counts and token totals.
- Use `GET /runs/{id}/trace` for structured trace records.
- Use `GET /traces/export?path=/local/file.jsonl` for explicit local JSONL export.
- Use run detail Prompt/Result panels only for local operator review; do not treat them as shareable support bundles.

### TUI Display Issues

**Cause**: Terminal compatibility

**Solution**:
- Ensure your terminal supports Unicode
- Try a different terminal emulator
- Check terminal size (minimum 80x24)

## Error Messages

| Error | Meaning | Fix |
|-------|---------|-----|
| `EOF` | Connection closed | Check network/API status |
| `timeout` | Request took too long | Increase timeout in config |
| `permission denied` | File access denied | Check file permissions |
| `file not found` | File doesn't exist | Verify path is correct |

## Getting Help

1. Capture local readiness and diagnostics:
   ```bash
   starclaw doctor
   starclaw doctor --json
   starclaw app --check
   starclaw daemon status
   ```
2. Inspect local daemon observability:
   ```bash
   curl -s http://127.0.0.1:7533/metrics | jq .
   curl -s http://127.0.0.1:7533/runs | jq .
   ```
3. Check the [GitHub Issues](https://github.com/starclaw/starclaw/issues)
4. Run with debug logging:
   ```bash
   export STARCLAW_DEBUG=1
   starclaw chat "query"
   ```
5. Join our [Discord](https://discord.gg/starclaw)

## Report a Bug

```bash
# Include in your report:
starclaw version
go version
uname -a
```
