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
