# Usage Examples

## Basic Queries

### Simple Question
```bash
starclaw chat "What is the capital of France?"
```

### Code Explanation
```bash
starclaw chat "Explain what this project does"
```

### File Analysis
```bash
starclaw chat "Read main.go and explain the entry point"
```

## File Operations

### Read a File
```bash
starclaw chat "Read the README.md file"
```

### Find Files
```bash
starclaw chat "Find all Go test files"
```

### Search Code
```bash
starclaw chat "Search for 'TODO' comments in all Go files"
```

## Code Generation

### Create a Script
```bash
starclaw chat "Create a Python script to sort CSV files by column"
```

### Generate Config
```bash
starclaw chat "Create a sample docker-compose.yml for a web app"
```

## Refactoring

### Rename Variables
```bash
starclaw chat "Rename 'userId' to 'userID' in all files"
```

### Extract Functions
```bash
starclaw chat "Refactor main.go to extract the config loading into a separate function"
```

## Testing

### Run Tests
```bash
starclaw -y chat "Run go test ./... and report any failures"
```

### Analyze Failures
```bash
starclaw chat "The tests are failing. Look at the output and suggest fixes"
```

## Interactive Mode

### Launch TUI
```bash
starclaw interactive
```

**Keyboard Shortcuts:**
- `Ctrl+Enter` - Send message
- `Ctrl+Q` - Quit
- `Ctrl+L` - Clear screen
- `Ctrl+Y` - Auto-approve all tools

## Daemon Web UI

### Open the Web UI
```bash
starclaw app
```

This starts the daemon when needed and opens the embedded Web UI at:

```text
http://127.0.0.1:7533/app/
```

The Web UI supports chat, streaming output, cancellable runs, expandable tool-call details, named agents, skills, session history, and schedule management.

### Manual Daemon Lifecycle
```bash
starclaw daemon start
starclaw daemon open
```

You can also start the daemon on demand while opening the UI:

```bash
starclaw daemon open --start
```

### Check Daemon Status
```bash
starclaw daemon status
```

The status output includes the Web UI URL when the daemon is reachable.

### Smoke Test the Web UI
```bash
scripts/smoke_webui.sh
```

The smoke script builds a temporary local binary, starts the daemon with an isolated home directory, opens the embedded Web UI in a browser, checks schedule controls and approval UI behavior, and writes a screenshot to `output/playwright/daemon-webui-smoke.png`.

## Advanced Usage

### Pipe Input
```bash
cat error.log | starclaw chat "Analyze these errors"
```

### Chain Commands
```bash
starclaw chat "List all Go files" | grep "_test.go"
```

### With Auto-Approval
```bash
starclaw -y chat "Clean up all temp files in /tmp"
```

## Tips

1. **Be Specific**: "Find unused imports in cmd/" vs "Find unused imports"
2. **Use Context**: Reference specific files when possible
3. **Iterative**: Start simple, then refine based on results
4. **Review**: Always review tool calls before approving
