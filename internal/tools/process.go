// Package tools provides local tool implementations for the agent.
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/starclaw/starclaw/internal/agent"
)

// ProcessTool manages processes: list, kill, start, signal, and status.
type ProcessTool struct {
	mu      sync.Mutex
	spawned map[int]*spawnedProcess
	timeout time.Duration
}

type spawnedProcess struct {
	cmd    *exec.Cmd
	stdout *safeProcessBuffer
	stderr *safeProcessBuffer
	done   chan struct{}
}

type safeProcessBuffer struct {
	mu  sync.Mutex
	buf []byte
}

func (b *safeProcessBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.buf = append(b.buf, p...)
	return len(p), nil
}

func (b *safeProcessBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return string(b.buf)
}

func (b *safeProcessBuffer) Len() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.buf)
}

func (b *safeProcessBuffer) Tail(max int) string {
	b.mu.Lock()
	defer b.mu.Unlock()
	if len(b.buf) <= max {
		return string(b.buf)
	}
	return string(b.buf[len(b.buf)-max:])
}

type processArgs struct {
	Action  string   `json:"action"`
	PID     int      `json:"pid,omitempty"`
	Name    string   `json:"name,omitempty"`
	Command string   `json:"command,omitempty"`
	Args    []string `json:"args,omitempty"`
	Signal  string   `json:"signal,omitempty"`
	Timeout int      `json:"timeout,omitempty"`
}

// NewProcessTool creates a process tool with a default spawn timeout.
func NewProcessTool(defaultTimeout time.Duration) *ProcessTool {
	if defaultTimeout == 0 {
		defaultTimeout = 300 * time.Second
	}
	return &ProcessTool{
		spawned: make(map[int]*spawnedProcess),
		timeout: defaultTimeout,
	}
}

func (t *ProcessTool) Info() agent.ToolInfo {
	return agent.ToolInfo{
		Name:        "process",
		Description: "Manage processes: list, kill, start a background process, send signals, or check status.",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"action": map[string]any{
					"type":        "string",
					"description": "Action: 'list', 'kill', 'start', 'signal', or 'status'",
				},
				"pid": map[string]any{
					"type":        "integer",
					"description": "Process ID (for kill/signal/status)",
				},
				"name": map[string]any{
					"type":        "string",
					"description": "Process name (alternative to pid for kill)",
				},
				"command": map[string]any{
					"type":        "string",
					"description": "Command to run (for start action)",
				},
				"args": map[string]any{
					"type":        "array",
					"items":       map[string]any{"type": "string"},
					"description": "Arguments for the command (for start action)",
				},
				"signal": map[string]any{
					"type":        "string",
					"description": "Signal name: SIGTERM, SIGKILL, SIGINT, SIGHUP (for signal action)",
				},
				"timeout": map[string]any{
					"type":        "integer",
					"description": "Timeout in seconds for spawned process (for start action, default 300)",
				},
			},
		},
		Required: []string{"action"},
	}
}

func (t *ProcessTool) Run(ctx context.Context, argsJSON string) (agent.ToolResult, error) {
	var args processArgs
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return agent.ValidationError("invalid arguments: " + err.Error()), nil
	}

	switch args.Action {
	case "list":
		return t.listProcesses(ctx)
	case "kill":
		return t.killProcess(ctx, args)
	case "start":
		return t.startProcess(ctx, args)
	case "signal":
		return t.signalProcess(args)
	case "status":
		return t.statusProcess(args)
	default:
		return agent.ValidationError(fmt.Sprintf("unknown action: %q (use list/kill/start/signal/status)", args.Action)), nil
	}
}

func (t *ProcessTool) listProcesses(ctx context.Context) (agent.ToolResult, error) {
	cmd := exec.CommandContext(ctx, "ps", "aux")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return agent.ToolResult{Content: fmt.Sprintf("ps error: %v\n%s", err, string(output)), IsError: true}, nil
	}
	result := string(output)
	if len(result) > 30000 {
		result = result[:30000] + "\n... (truncated)"
	}
	return agent.ToolResult{Content: result}, nil
}

func (t *ProcessTool) killProcess(ctx context.Context, args processArgs) (agent.ToolResult, error) {
	if args.PID > 0 {
		cmd := exec.CommandContext(ctx, "kill", fmt.Sprintf("%d", args.PID))
		output, err := cmd.CombinedOutput()
		if err != nil {
			return agent.ToolResult{Content: fmt.Sprintf("kill error: %v\n%s", err, string(output)), IsError: true}, nil
		}
		t.mu.Lock()
		delete(t.spawned, args.PID)
		t.mu.Unlock()
		return agent.ToolResult{Content: fmt.Sprintf("sent SIGTERM to PID %d", args.PID)}, nil
	}
	if strings.TrimSpace(args.Name) != "" {
		cmd := exec.CommandContext(ctx, "pkill", "-x", "--", args.Name)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return agent.ToolResult{Content: fmt.Sprintf("pkill error: %v\n%s", err, string(output)), IsError: true}, nil
		}
		return agent.ToolResult{Content: fmt.Sprintf("sent SIGTERM to process '%s'", args.Name)}, nil
	}
	return agent.ValidationError("pid or name is required for kill action"), nil
}

func (t *ProcessTool) startProcess(_ context.Context, args processArgs) (agent.ToolResult, error) {
	if args.Command == "" {
		return agent.ValidationError("command is required for start action"), nil
	}

	stdout := &safeProcessBuffer{}
	stderr := &safeProcessBuffer{}
	cmd := exec.Command(args.Command, args.Args...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	if err := cmd.Start(); err != nil {
		return agent.ToolResult{Content: fmt.Sprintf("failed to start process: %v", err), IsError: true}, nil
	}

	pid := cmd.Process.Pid
	sp := &spawnedProcess{cmd: cmd, stdout: stdout, stderr: stderr, done: make(chan struct{})}

	t.mu.Lock()
	t.spawned[pid] = sp
	t.mu.Unlock()

	timeout := t.timeout
	if args.Timeout > 0 {
		timeout = time.Duration(args.Timeout) * time.Second
	}

	go func() {
		defer close(sp.done)
		timer := time.AfterFunc(timeout, func() {
			cmd.Process.Kill()
		})
		cmd.Wait()
		timer.Stop()
		t.mu.Lock()
		delete(t.spawned, pid)
		t.mu.Unlock()
	}()

	// Wait briefly for initial output
	time.Sleep(100 * time.Millisecond)

	result := fmt.Sprintf("Started PID %d: %s %s", pid, args.Command, strings.Join(args.Args, " "))
	if stdout.Len() > 0 {
		result += fmt.Sprintf("\n\nstdout:\n%s", stdout.String())
	}
	if stderr.Len() > 0 {
		result += fmt.Sprintf("\n\nstderr:\n%s", stderr.String())
	}

	return agent.ToolResult{Content: result}, nil
}

func (t *ProcessTool) signalProcess(args processArgs) (agent.ToolResult, error) {
	if args.PID <= 0 {
		return agent.ValidationError("pid is required for signal action"), nil
	}

	sig := parseSignal(args.Signal)
	if sig == nil {
		return agent.ValidationError(fmt.Sprintf("unknown signal: %q (use SIGTERM, SIGKILL, SIGINT, SIGHUP)", args.Signal)), nil
	}

	proc, err := os.FindProcess(args.PID)
	if err != nil {
		return agent.ToolResult{Content: fmt.Sprintf("process not found: %v", err), IsError: true}, nil
	}

	if err := proc.Signal(sig); err != nil {
		return agent.ToolResult{Content: fmt.Sprintf("signal error: %v", err), IsError: true}, nil
	}

	return agent.ToolResult{Content: fmt.Sprintf("sent %s to PID %d", args.Signal, args.PID)}, nil
}

func (t *ProcessTool) statusProcess(args processArgs) (agent.ToolResult, error) {
	if args.PID <= 0 {
		return agent.ValidationError("pid is required for status action"), nil
	}

	t.mu.Lock()
	sp, tracked := t.spawned[args.PID]
	t.mu.Unlock()

	if tracked {
		select {
		case <-sp.done:
			return agent.ToolResult{Content: fmt.Sprintf("PID %d: exited", args.PID)}, nil
		default:
			result := fmt.Sprintf("PID %d: running", args.PID)
			if sp.stdout.Len() > 0 {
				result += fmt.Sprintf("\n\nstdout (tail):\n%s", sp.stdout.Tail(2000))
			}
			return agent.ToolResult{Content: result}, nil
		}
	}

	// Check untracked process via signal 0
	proc, err := os.FindProcess(args.PID)
	if err != nil {
		return agent.ToolResult{Content: fmt.Sprintf("PID %d: not found", args.PID)}, nil
	}
	if err := proc.Signal(syscall.Signal(0)); err != nil {
		return agent.ToolResult{Content: fmt.Sprintf("PID %d: not running (%v)", args.PID, err)}, nil
	}
	return agent.ToolResult{Content: fmt.Sprintf("PID %d: running (not tracked by this tool)", args.PID)}, nil
}

func parseSignal(name string) os.Signal {
	switch strings.ToUpper(strings.TrimSpace(name)) {
	case "SIGTERM", "TERM":
		return syscall.SIGTERM
	case "SIGKILL", "KILL":
		return syscall.SIGKILL
	case "SIGINT", "INT":
		return syscall.SIGINT
	case "SIGHUP", "HUP":
		return syscall.SIGHUP
	default:
		return nil
	}
}

func (t *ProcessTool) RequiresApproval() bool { return false }

func (t *ProcessTool) IsSafeArgs(argsJSON string) bool {
	var args processArgs
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return false
	}
	return args.Action == "list" || args.Action == "status"
}

func (t *ProcessTool) IsReadOnlyCall(argsJSON string) bool {
	return t.IsSafeArgs(argsJSON)
}
