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
- Be more specific in your queries

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

1. Check the [GitHub Issues](https://github.com/starclaw/starclaw/issues)
2. Run with debug logging:
   ```bash
   export STARCLAW_DEBUG=1
   starclaw chat "query"
   ```
3. Join our [Discord](https://discord.gg/starclaw)

## Report a Bug

```bash
# Include in your report:
starclaw version
go version
uname -a
```
