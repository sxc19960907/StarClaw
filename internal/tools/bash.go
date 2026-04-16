package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/starclaw/starclaw/internal/agent"
)

// BashTool executes shell commands
type BashTool struct {
	ExtraSafeCommands []string
}

type bashArgs struct {
	Command string `json:"command"`
	Timeout int    `json:"timeout,omitempty"`
}

var safeCommands = []string{
	"ls", "pwd", "which", "echo", "cat", "head", "tail", "wc",
	"git status", "git diff", "git log", "git branch", "git show",
	"go build", "go test", "go vet", "go fmt", "go mod",
	"make", "cargo build", "cargo test", "npm test", "npm run",
	"python -m pytest", "python -m py_compile",
}

var shellOperators = []string{"&&", "||", ";", "|", ">", "<", "`", "$(", "${", "&"}

func isSafeCommand(cmd string, extraSafe []string) bool {
	trimmed := strings.TrimSpace(cmd)

	// Reject commands containing shell operators
	for _, op := range shellOperators {
		if strings.Contains(trimmed, op) {
			return false
		}
	}

	for _, safe := range safeCommands {
		if trimmed == safe || strings.HasPrefix(trimmed, safe+" ") {
			return true
		}
	}

	for _, safe := range extraSafe {
		if trimmed == safe || strings.HasPrefix(trimmed, safe+" ") {
			return true
		}
	}

	return false
}

func (t *BashTool) Info() agent.ToolInfo {
	return agent.ToolInfo{
		Name:        "bash",
		Description: "Execute a shell command. Use for running scripts, data processing, file management, automation, and system operations.",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"command": map[string]any{"type": "string", "description": "Shell command to execute"},
				"timeout": map[string]any{"type": "integer", "description": "Timeout in seconds (default: 120)"},
			},
		},
		Required: []string{"command"},
	}
}

func (t *BashTool) Run(ctx context.Context, argsJSON string) (agent.ToolResult, error) {
	var args bashArgs
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return agent.ValidationError(fmt.Sprintf("invalid arguments: %v", err)), nil
	}

	timeout := 120 * time.Second
	if args.Timeout > 0 {
		timeout = time.Duration(args.Timeout) * time.Second
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sh", "-c", args.Command)
	output, err := cmd.CombinedOutput()

	result := string(output)
	maxOut := 30000
	if len(result) > maxOut {
		keepHead := maxOut * 3 / 4
		keepTail := maxOut / 4
		result = result[:keepHead] + "\n\n[... truncated " +
			strconv.Itoa(len(result)-maxOut) + " chars ...]\n\n" +
			result[len(result)-keepTail:]
	}

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			timeoutSecs := int(timeout.Seconds())
			return agent.TransientError(fmt.Sprintf("command timed out after %ds\n%s", timeoutSecs, result)), nil
		}
		return agent.ToolResult{
			Content: fmt.Sprintf("exit code: %v\n%s", err, result),
			IsError: true,
		}, nil
	}

	return agent.ToolResult{Content: result}, nil
}

func (t *BashTool) RequiresApproval() bool { return true }

func (t *BashTool) IsReadOnlyCall(string) bool { return false }

func (t *BashTool) IsSafe(command string) bool {
	return isSafeCommand(command, t.ExtraSafeCommands)
}

func (t *BashTool) IsSafeArgs(argsJSON string) bool {
	var args bashArgs
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return false
	}
	return isSafeCommand(args.Command, t.ExtraSafeCommands)
}
