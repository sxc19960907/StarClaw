# Tool Reference

## Overview

StarClaw provides 7 built-in tools that the AI can use to interact with your local system.

## Tool Index

| Tool | Purpose | Approval Required |
|------|---------|-------------------|
| [file_read](#file_read) | Read file contents | Yes |
| [file_write](#file_write) | Write files | Yes |
| [file_edit](#file_edit) | Edit files | Yes |
| [glob](#glob) | Find files by pattern | No |
| [directory_list](#directory_list) | List directories | No |
| [grep](#grep) | Search file contents | No |
| [bash](#bash) | Execute commands | Yes |

---

## file_read

Reads a file with line numbers.

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `path` | string | Yes | File path (absolute or relative) |
| `offset` | integer | No | Start line (0-based, default 0) |
| `limit` | integer | No | Max lines to read |

### Example

```json
{
  "path": "main.go",
  "offset": 0,
  "limit": 50
}
```

### Security

- Path must be under current working directory
- Cannot access files outside project root

---

## file_write

Writes content to a file.

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `path` | string | Yes | File path |
| `content` | string | Yes | File content |

### Example

```json
{
  "path": "output.txt",
  "content": "Hello, World!"
}
```

### Security

- Cannot overwrite existing files without confirmation
- Path restricted to working directory

---

## file_edit

Find and replace text in a file.

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `path` | string | Yes | File path |
| `old_string` | string | Yes | Text to find |
| `new_string` | string | Yes | Replacement text |

### Example

```json
{
  "path": "config.go",
  "old_string": "const Version = \"1.0.0\"",
  "new_string": "const Version = \"1.1.0\""
}
```

---

## glob

Find files matching a pattern.

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `pattern` | string | Yes | Glob pattern (e.g., "*.go", "**/*.md") |

### Example

```json
{
  "pattern": "**/*_test.go"
}
```

---

## directory_list

List directory contents.

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `path` | string | Yes | Directory path (use "." for current) |

### Example

```json
{
  "path": "./cmd"
}
```

---

## grep

Search file contents with regex.

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `pattern` | string | Yes | Regex pattern |
| `path` | string | Yes | Directory or file to search |
| `output_mode` | string | No | "content", "files", or "count" |

### Example

```json
{
  "pattern": "func main",
  "path": ".",
  "output_mode": "files"
}
```

---

## bash

Execute shell commands.

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `command` | string | Yes | Shell command to execute |
| `timeout` | integer | No | Timeout in seconds (default: 120) |

### Example

```json
{
  "command": "go test ./...",
  "timeout": 60
}
```

### Security

- Requires explicit approval
- Respects system PATH
- Output limited by `bash_max_output` config

---

## Tool Approval

Tools marked as "Requires Approval" will prompt you before execution:

```
⚠️  Tool Approval Required

Tool: file_write
Args: {"path":"test.txt","content":"Hello"}

Approve? [Y/n]
```

Use `-y` flag to auto-approve all tools:

```bash
starclaw -y chat "Command that uses tools"
```
