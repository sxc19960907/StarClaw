package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/starclaw/starclaw/internal/agent"
	"github.com/starclaw/starclaw/internal/config"
	ctxwin "github.com/starclaw/starclaw/internal/context"
)

// MemoryTool manages agent memory files. It supports listing, searching, and
// deleting memory entries. This is different from MemoryAppendTool which only
// appends to MEMORY.md. The memory directory is resolved from context (set by
// AgentLoop.Run), falling back to ~/.starclaw/memory/.
type MemoryTool struct{}

type memoryArgs struct {
	Action string `json:"action"`
	Query  string `json:"query,omitempty"`
	Name   string `json:"name,omitempty"`
}

func (t *MemoryTool) Info() agent.ToolInfo {
	return agent.ToolInfo{
		Name:        "memory",
		Description: "Manage agent memory entries. Supports: list (list all memory files), search (search memory content for a query), delete (delete a memory entry by name). Operates on the agent's memory directory.",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"action": map[string]any{
					"type": "string",
					"enum": []string{"list", "search", "delete"},
					"description": "Action to perform: 'list' enumerates memory files, " +
						"'search' finds content matching a query, 'delete' removes a named entry",
				},
				"query": map[string]any{
					"type":        "string",
					"description": "Search term (required for action=search)",
				},
				"name": map[string]any{
					"type":        "string",
					"description": "Entry name to delete (required for action=delete, e.g. 'MEMORY.md')",
				},
			},
			"required": []string{"action"},
		},
	}
}

func (t *MemoryTool) Run(ctx context.Context, argsJSON string) (agent.ToolResult, error) {
	var args memoryArgs
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return agent.ValidationError("invalid arguments: " + err.Error()), nil
	}

	switch args.Action {
	case "list":
		return t.listMemory(ctx)
	case "search":
		return t.searchMemory(ctx, args.Query)
	case "delete":
		return t.deleteMemory(ctx, args.Name)
	default:
		return agent.ValidationError(fmt.Sprintf("unknown action %q: must be list, search, or delete", args.Action)), nil
	}
}

// memoryDir resolves the memory directory from context or falls back to
// ~/.starclaw/memory/.
func (t *MemoryTool) memoryDir(ctx context.Context) (string, error) {
	// Try context first (set by AgentLoop)
	dir := ctxwin.MemoryDirFromContext(ctx)
	if dir != "" {
		return dir, nil
	}

	// Fall back to ~/.starclaw/memory/
	starclawDir := config.StarclawDir()
	if starclawDir == "" {
		return "", fmt.Errorf("cannot resolve memory directory: starclaw home not found")
	}
	dir = filepath.Join(starclawDir, "memory")
	return dir, nil
}

// listMemory returns all files in the memory directory.
func (t *MemoryTool) listMemory(ctx context.Context) (agent.ToolResult, error) {
	dir, err := t.memoryDir(ctx)
	if err != nil {
		return agent.ToolResult{Content: "memory: " + err.Error(), IsError: true}, nil
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return agent.ToolResult{Content: "No memory entries found (directory does not exist)."}, nil
		}
		return agent.ToolResult{Content: fmt.Sprintf("memory: failed to read directory: %v", err), IsError: true}, nil
	}

	var fileNames []string
	totalSize := int64(0)
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		fileNames = append(fileNames, fmt.Sprintf("  %s (%d bytes)", e.Name(), info.Size()))
		totalSize += info.Size()
	}

	if len(fileNames) == 0 {
		return agent.ToolResult{Content: "Memory directory is empty."}, nil
	}

	sort.Strings(fileNames)
	var sb strings.Builder
	fmt.Fprintf(&sb, "Memory directory: %s\n\n", dir)
	sb.WriteString(strings.Join(fileNames, "\n"))
	fmt.Fprintf(&sb, "\n\n%d files, %d bytes total\n", len(fileNames), totalSize)

	return agent.ToolResult{Content: sb.String()}, nil
}

// searchMemory searches memory file contents for the given query string.
func (t *MemoryTool) searchMemory(ctx context.Context, query string) (agent.ToolResult, error) {
	if query == "" {
		return agent.ValidationError("query is required for action=search"), nil
	}

	dir, err := t.memoryDir(ctx)
	if err != nil {
		return agent.ToolResult{Content: "memory: " + err.Error(), IsError: true}, nil
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return agent.ToolResult{Content: fmt.Sprintf("No memory entries found matching %q.", query)}, nil
		}
		return agent.ToolResult{Content: fmt.Sprintf("memory: failed to read directory: %v", err), IsError: true}, nil
	}

	q := strings.ToLower(query)
	var results []string

	for _, e := range entries {
		if e.IsDir() {
			continue
		}

		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}

		content := string(data)
		if strings.Contains(strings.ToLower(content), q) {
			// Show matching lines
			lines := strings.Split(content, "\n")
			var matchLines []string
			for i, line := range lines {
				if strings.Contains(strings.ToLower(line), q) {
					matchLines = append(matchLines, fmt.Sprintf("    %s:%d: %s", e.Name(), i+1, strings.TrimSpace(line)))
				}
			}
			if len(matchLines) > 0 {
				results = append(results, strings.Join(matchLines, "\n"))
			}
		}
	}

	if len(results) == 0 {
		return agent.ToolResult{Content: fmt.Sprintf("No memory entries found matching %q.", query)}, nil
	}

	return agent.ToolResult{Content: fmt.Sprintf("Found matches in %d file(s):\n\n%s", len(results), strings.Join(results, "\n\n"))}, nil
}

// deleteMemory deletes a named memory entry file.
func (t *MemoryTool) deleteMemory(ctx context.Context, name string) (agent.ToolResult, error) {
	if name == "" {
		return agent.ValidationError("name is required for action=delete"), nil
	}

	dir, err := t.memoryDir(ctx)
	if err != nil {
		return agent.ToolResult{Content: "memory: " + err.Error(), IsError: true}, nil
	}

	target := filepath.Join(dir, name)
	// Prevent path traversal
	if !strings.HasPrefix(filepath.Clean(target), filepath.Clean(dir)+string(filepath.Separator)) &&
		filepath.Clean(target) != filepath.Clean(dir) {
		return agent.ValidationError("invalid memory entry name"), nil
	}

	if err := os.Remove(target); err != nil {
		if os.IsNotExist(err) {
			return agent.ValidationError(fmt.Sprintf("memory entry %q not found", name)), nil
		}
		return agent.ToolResult{Content: fmt.Sprintf("memory: failed to delete: %v", err), IsError: true}, nil
	}

	return agent.ToolResult{Content: fmt.Sprintf("Deleted memory entry: %s", name)}, nil
}

func (t *MemoryTool) RequiresApproval() bool { return false }
