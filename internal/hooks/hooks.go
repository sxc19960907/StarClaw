// Package hooks provides lifecycle hook execution for the agent loop.
// Hooks are external scripts that run before/after tool calls, at session start/stop.
package hooks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

// Event names for hook lifecycle.
type Event string

const (
	PreToolUse   Event = "PreToolUse"
	PostToolUse  Event = "PostToolUse"
	SessionStart Event = "SessionStart"
	Stop         Event = "Stop"
)

const (
	defaultTimeout = 10 * time.Second
	maxOutputBytes = 10 * 1024
	exitCodeDeny   = 2
)

// Config defines hook configurations per lifecycle event.
type Config struct {
	PreToolUse   []Entry `yaml:"PreToolUse"   json:"PreToolUse"`
	PostToolUse  []Entry `yaml:"PostToolUse"  json:"PostToolUse"`
	SessionStart []Entry `yaml:"SessionStart" json:"SessionStart"`
	Stop         []Entry `yaml:"Stop"         json:"Stop"`
}

// Entry defines a single hook: a command to run, optionally filtered by tool name regex.
type Entry struct {
	Matcher string `yaml:"matcher" json:"matcher"`
	Command string `yaml:"command" json:"command"`
}

// Input is the JSON payload sent to hooks on stdin.
type Input struct {
	Event        Event           `json:"event"`
	ToolName     string          `json:"tool_name,omitempty"`
	ToolInput    json.RawMessage `json:"tool_input"`
	ToolResponse json.RawMessage `json:"tool_response"`
	SessionID    string          `json:"session_id"`
}

// Runner executes hook scripts.
type Runner struct {
	config  Config
	timeout time.Duration

	mu     sync.Mutex
	inHook bool
}

// NewRunner creates a hook runner from the given config.
func NewRunner(cfg Config) *Runner {
	return &Runner{
		config:  cfg,
		timeout: defaultTimeout,
	}
}

// RunPreToolUse runs matching PreToolUse hooks before a tool call.
// Returns ("deny", reason) if any hook rejects the call.
func (r *Runner) RunPreToolUse(ctx context.Context, toolName, toolInput, sessionID string) (string, string) {
	if r == nil {
		return "", ""
	}
	if r.enterHook() {
		log.Printf("[hooks] PreToolUse %q skipped: another hook is running", toolName)
		return "", ""
	}
	defer r.exitHook()

	entries := r.match(r.config.PreToolUse, toolName)
	if len(entries) == 0 {
		return "", ""
	}

	input := Input{
		Event:        PreToolUse,
		ToolName:     toolName,
		ToolInput:    toRawJSON(toolInput),
		ToolResponse: json.RawMessage("null"),
		SessionID:    sessionID,
	}

	for _, e := range entries {
		code, _, stderr, err := r.run(ctx, e, input)
		if err != nil {
			log.Printf("[hooks] PreToolUse %q failed: %v", e.Command, err)
			continue
		}
		if code == exitCodeDeny {
			reason := strings.TrimSpace(stderr)
			if reason == "" {
				reason = "blocked by hook"
			}
			return "deny", reason
		}
		if code != 0 {
			log.Printf("[hooks] PreToolUse %q exit %d: %s", e.Command, code, strings.TrimSpace(stderr))
		}
	}
	return "", ""
}

// RunPostToolUse runs matching PostToolUse hooks after a tool call (fire-and-forget).
func (r *Runner) RunPostToolUse(ctx context.Context, toolName, toolInput, toolResponse, sessionID string) {
	if r == nil {
		return
	}
	if r.enterHook() {
		log.Printf("[hooks] PostToolUse %q skipped: another hook is running", toolName)
		return
	}
	defer r.exitHook()

	entries := r.match(r.config.PostToolUse, toolName)
	if len(entries) == 0 {
		return
	}

	input := Input{
		Event:        PostToolUse,
		ToolName:     toolName,
		ToolInput:    toRawJSON(toolInput),
		ToolResponse: toRawJSON(toolResponse),
		SessionID:    sessionID,
	}

	for _, e := range entries {
		code, _, stderr, err := r.run(ctx, e, input)
		if err != nil {
			log.Printf("[hooks] PostToolUse %q failed: %v", e.Command, err)
			continue
		}
		if code != 0 {
			log.Printf("[hooks] PostToolUse %q exit %d: %s", e.Command, code, strings.TrimSpace(stderr))
		}
	}
}

// RunSessionStart runs all SessionStart hooks.
func (r *Runner) RunSessionStart(ctx context.Context, sessionID string) {
	if r == nil || r.enterHook() {
		return
	}
	defer r.exitHook()

	if len(r.config.SessionStart) == 0 {
		return
	}

	input := Input{
		Event:        SessionStart,
		ToolInput:    json.RawMessage("null"),
		ToolResponse: json.RawMessage("null"),
		SessionID:    sessionID,
	}

	for _, e := range r.config.SessionStart {
		code, _, stderr, err := r.run(ctx, e, input)
		if err != nil {
			log.Printf("[hooks] SessionStart %q failed: %v", e.Command, err)
			continue
		}
		if code != 0 {
			log.Printf("[hooks] SessionStart %q exit %d: %s", e.Command, code, strings.TrimSpace(stderr))
		}
	}
}

// RunStop runs all Stop hooks.
func (r *Runner) RunStop(ctx context.Context, sessionID string) {
	if r == nil || r.enterHook() {
		return
	}
	defer r.exitHook()

	if len(r.config.Stop) == 0 {
		return
	}

	input := Input{
		Event:        Stop,
		ToolInput:    json.RawMessage("null"),
		ToolResponse: json.RawMessage("null"),
		SessionID:    sessionID,
	}

	for _, e := range r.config.Stop {
		code, _, stderr, err := r.run(ctx, e, input)
		if err != nil {
			log.Printf("[hooks] Stop %q failed: %v", e.Command, err)
			continue
		}
		if code != 0 {
			log.Printf("[hooks] Stop %q exit %d: %s", e.Command, code, strings.TrimSpace(stderr))
		}
	}
}

func (r *Runner) run(ctx context.Context, entry Entry, input Input) (int, string, string, error) {
	cmdPath, err := resolveCommand(entry.Command)
	if err != nil {
		return -1, "", "", err
	}

	inputJSON, err := json.Marshal(input)
	if err != nil {
		return -1, "", "", fmt.Errorf("marshal hook input: %w", err)
	}

	hookCtx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	cmd := exec.CommandContext(hookCtx, cmdPath)
	cmd.Stdin = bytes.NewReader(inputJSON)

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &limitedWriter{buf: &stdoutBuf, limit: maxOutputBytes}
	cmd.Stderr = &limitedWriter{buf: &stderrBuf, limit: maxOutputBytes}

	if runErr := cmd.Run(); runErr != nil {
		if hookCtx.Err() == context.DeadlineExceeded {
			return -1, stdoutBuf.String(), stderrBuf.String(), fmt.Errorf("hook timed out")
		}
		if exitErr, ok := runErr.(*exec.ExitError); ok {
			return exitErr.ExitCode(), stdoutBuf.String(), stderrBuf.String(), nil
		}
		return -1, stdoutBuf.String(), stderrBuf.String(), runErr
	}

	return 0, stdoutBuf.String(), stderrBuf.String(), nil
}

func (r *Runner) match(entries []Entry, toolName string) []Entry {
	var matched []Entry
	for _, e := range entries {
		if e.Matcher == "" {
			matched = append(matched, e)
			continue
		}
		re, err := regexp.Compile(e.Matcher)
		if err != nil {
			log.Printf("[hooks] invalid matcher %q: %v", e.Matcher, err)
			continue
		}
		if re.MatchString(toolName) {
			matched = append(matched, e)
		}
	}
	return matched
}

func (r *Runner) enterHook() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.inHook {
		return true
	}
	r.inHook = true
	return false
}

func (r *Runner) exitHook() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.inHook = false
}

func resolveCommand(command string) (string, error) {
	if command == "" {
		return "", fmt.Errorf("hook command must not be empty")
	}

	if strings.HasPrefix(command, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolve home: %w", err)
		}
		command = filepath.Join(home, command[2:])
	}

	if !filepath.IsAbs(command) && !strings.HasPrefix(command, "./") &&
		!strings.Contains(command, string(filepath.Separator)) {
		return "", fmt.Errorf("bare command %q rejected: use absolute path or ./ prefix", command)
	}

	// Only allow scripts under ~/.starclaw/
	if filepath.IsAbs(command) {
		starclawDir, err := starclawDir()
		cleaned := filepath.Clean(command)
		if err != nil || !strings.HasPrefix(cleaned, starclawDir+string(filepath.Separator)) {
			return "", fmt.Errorf("absolute path %q rejected: must be under ~/.starclaw", command)
		}
	}

	return command, nil
}

func starclawDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".starclaw"), nil
}

func toRawJSON(s string) json.RawMessage {
	if s == "" {
		return json.RawMessage("null")
	}
	if json.Valid([]byte(s)) {
		return json.RawMessage(s)
	}
	data, _ := json.Marshal(s)
	return data
}

type limitedWriter struct {
	buf   *bytes.Buffer
	limit int
}

func (w *limitedWriter) Write(p []byte) (int, error) {
	remaining := w.limit - w.buf.Len()
	if remaining <= 0 {
		return len(p), nil
	}
	if len(p) > remaining {
		w.buf.Write(p[:remaining])
	} else {
		w.buf.Write(p)
	}
	return len(p), nil
}
