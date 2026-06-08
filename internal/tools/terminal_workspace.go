// Package tools provides local tool implementations for the agent.
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"math"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/starclaw/starclaw/internal/agent"
)

const minTerminalWorkspaceGhosttyVersion = "1.3.0"

var terminalWorkspaceActions = []string{"status", "list_tabs", "new_tab", "new_split", "send_input"}

var (
	terminalWorkspaceAvailabilityStatusFn = terminalWorkspaceAvailabilityStatus
	terminalWorkspaceEnsureAvailableFn    = terminalWorkspaceEnsureAvailable
	terminalWorkspaceNewTabFn             = terminalWorkspaceNewTab
	terminalWorkspaceNewSplitFn           = terminalWorkspaceNewSplit
	terminalWorkspaceSendInputFn          = terminalWorkspaceSendInput
)

type terminalTabRef struct {
	windowIndex int
	tabIndex    int
}

type terminalTabRegistry struct {
	mu   sync.RWMutex
	tabs map[string]terminalTabRef
}

func newTerminalTabRegistry() *terminalTabRegistry {
	return &terminalTabRegistry{tabs: make(map[string]terminalTabRef)}
}

func (r *terminalTabRegistry) add(title string, ref terminalTabRef) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tabs[title] = ref
}

func (r *terminalTabRegistry) lookup(title string) (terminalTabRef, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ref, ok := r.tabs[title]
	return ref, ok
}

func (r *terminalTabRegistry) list() map[string]terminalTabRef {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make(map[string]terminalTabRef, len(r.tabs))
	for k, v := range r.tabs {
		out[k] = v
	}
	return out
}

// TerminalWorkspaceTool opens and controls a local visible terminal workspace.
type TerminalWorkspaceTool struct {
	tabs *terminalTabRegistry
}

type terminalWorkspaceArgs struct {
	Action      string `json:"action"`
	Description string `json:"description,omitempty"`
	Command     string `json:"command,omitempty"`
	Title       string `json:"title,omitempty"`
	Direction   string `json:"direction,omitempty"`
	Target      string `json:"target,omitempty"`
	Text        string `json:"text,omitempty"`
}

type terminalWorkspaceStatus struct {
	Platform        string   `json:"platform"`
	Backend         string   `json:"backend"`
	Available       bool     `json:"available"`
	Supported       bool     `json:"supported"`
	MinimumVersion  string   `json:"minimum_version"`
	DetectedVersion string   `json:"detected_version,omitempty"`
	Actions         []string `json:"actions"`
	TrackedTabs     int      `json:"tracked_tabs"`
	Fallback        string   `json:"fallback"`
}

type terminalWorkspaceAvailability struct {
	supported bool
	available bool
	version   string
	message   string
}

// NewTerminalWorkspaceTool creates a terminal workspace tool.
func NewTerminalWorkspaceTool() *TerminalWorkspaceTool {
	return &TerminalWorkspaceTool{tabs: newTerminalTabRegistry()}
}

func (t *TerminalWorkspaceTool) Info() agent.ToolInfo {
	return agent.ToolInfo{
		Name: "terminal_workspace",
		Description: "Open and control a local visible terminal workspace. " +
			"Use this when the user wants a terminal they can see and continue using, such as a dev server, log tail, or split-pane workspace. " +
			"Use bash when only command output is needed. Actions: status, list_tabs, new_tab, new_split, send_input.",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"action": map[string]any{
					"type":        "string",
					"description": "Action: status, list_tabs, new_tab, new_split, send_input",
				},
				"description": map[string]any{
					"type":        "string",
					"description": "Brief description of why this terminal workspace action is needed.",
				},
				"command": map[string]any{
					"type":        "string",
					"description": "Shell command to run in a new visible tab or split.",
				},
				"title": map[string]any{
					"type":        "string",
					"description": "Tracked tab title. Defaults to the command basename or terminal.",
				},
				"direction": map[string]any{
					"type":        "string",
					"description": "Split direction for new_split: right or down. Defaults to right.",
				},
				"target": map[string]any{
					"type":        "string",
					"description": "Tracked tab title to send input to.",
				},
				"text": map[string]any{
					"type":        "string",
					"description": "Text to send to a tracked terminal tab.",
				},
			},
		},
		Required: []string{"action", "description"},
	}
}

func (t *TerminalWorkspaceTool) Run(ctx context.Context, argsJSON string) (agent.ToolResult, error) {
	var args terminalWorkspaceArgs
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return agent.ValidationError("invalid arguments: " + err.Error()), nil
	}

	switch args.Action {
	case "status":
		return t.runStatus()
	case "list_tabs":
		return t.runListTabs()
	case "new_tab":
		if err := terminalWorkspaceEnsureAvailableFn(); err != nil {
			return terminalWorkspaceUnavailableResult(err), nil
		}
		return t.runNewTab(ctx, args)
	case "new_split":
		if err := terminalWorkspaceEnsureAvailableFn(); err != nil {
			return terminalWorkspaceUnavailableResult(err), nil
		}
		return t.runNewSplit(ctx, args)
	case "send_input":
		if err := terminalWorkspaceEnsureAvailableFn(); err != nil {
			return terminalWorkspaceUnavailableResult(err), nil
		}
		return t.runSendInput(ctx, args)
	default:
		return agent.ValidationError(fmt.Sprintf("unknown action %q: use %s", args.Action, strings.Join(terminalWorkspaceActions, ", "))), nil
	}
}

func (t *TerminalWorkspaceTool) runStatus() (agent.ToolResult, error) {
	availability := terminalWorkspaceAvailabilityStatusFn()
	status := terminalWorkspaceStatus{
		Platform:        runtime.GOOS,
		Backend:         "ghostty",
		Available:       availability.available,
		Supported:       availability.supported,
		MinimumVersion:  minTerminalWorkspaceGhosttyVersion,
		DetectedVersion: availability.version,
		Actions:         append([]string(nil), terminalWorkspaceActions...),
		TrackedTabs:     len(t.tabs.list()),
		Fallback:        terminalWorkspaceFallbackMessage(),
	}
	if availability.message != "" {
		status.Fallback = availability.message + " " + terminalWorkspaceFallbackMessage()
	}
	return jsonToolResult(status)
}

func (t *TerminalWorkspaceTool) runListTabs() (agent.ToolResult, error) {
	tabs := t.tabs.list()
	if len(tabs) == 0 {
		return agent.ToolResult{Content: "No tracked terminal workspace tabs."}, nil
	}
	names := make([]string, 0, len(tabs))
	for name := range tabs {
		names = append(names, name)
	}
	sort.Strings(names)

	var sb strings.Builder
	for _, name := range names {
		ref := tabs[name]
		fmt.Fprintf(&sb, "- %s (window:%d, tab:%d)\n", name, ref.windowIndex, ref.tabIndex)
	}
	return agent.ToolResult{Content: strings.TrimRight(sb.String(), "\n")}, nil
}

func (t *TerminalWorkspaceTool) runNewTab(ctx context.Context, args terminalWorkspaceArgs) (agent.ToolResult, error) {
	title := terminalWorkspaceTitle(args.Title, args.Command)
	color := terminalWorkspaceColor(title)
	winIdx, tabIdx, err := terminalWorkspaceNewTabFn(ctx, args.Command, title, color)
	if err != nil {
		return agent.ToolResult{Content: fmt.Sprintf("terminal_workspace new_tab failed: %v", err), IsError: true}, nil
	}
	t.tabs.add(title, terminalTabRef{windowIndex: winIdx, tabIndex: tabIdx})
	return agent.ToolResult{Content: fmt.Sprintf("Opened terminal workspace tab %q (window:%d, tab:%d).", title, winIdx, tabIdx)}, nil
}

func (t *TerminalWorkspaceTool) runNewSplit(ctx context.Context, args terminalWorkspaceArgs) (agent.ToolResult, error) {
	direction := strings.TrimSpace(args.Direction)
	if direction == "" {
		direction = "right"
	}
	if direction != "right" && direction != "down" {
		return agent.ValidationError(fmt.Sprintf("invalid direction %q: use right or down", direction)), nil
	}

	title := terminalWorkspaceTitle(args.Title, args.Command)
	color := terminalWorkspaceColor(title)
	winIdx, tabIdx, err := terminalWorkspaceNewSplitFn(ctx, direction, args.Command, title, color)
	if err != nil {
		return agent.ToolResult{Content: fmt.Sprintf("terminal_workspace new_split failed: %v", err), IsError: true}, nil
	}
	t.tabs.add(title, terminalTabRef{windowIndex: winIdx, tabIndex: tabIdx})
	return agent.ToolResult{Content: fmt.Sprintf("Opened %s terminal workspace split %q (window:%d, tab:%d).", direction, title, winIdx, tabIdx)}, nil
}

func (t *TerminalWorkspaceTool) runSendInput(ctx context.Context, args terminalWorkspaceArgs) (agent.ToolResult, error) {
	target := strings.TrimSpace(args.Target)
	if target == "" {
		return agent.ValidationError("target is required for send_input"), nil
	}
	if args.Text == "" {
		return agent.ValidationError("text is required for send_input"), nil
	}
	ref, ok := t.tabs.lookup(target)
	if !ok {
		known := make([]string, 0, len(t.tabs.list()))
		for name := range t.tabs.list() {
			known = append(known, name)
		}
		sort.Strings(known)
		if len(known) == 0 {
			return agent.ValidationError(fmt.Sprintf("tracked terminal tab %q not found; no tabs are currently tracked", target)), nil
		}
		return agent.ValidationError(fmt.Sprintf("tracked terminal tab %q not found; known tabs: %s", target, strings.Join(known, ", "))), nil
	}
	if err := terminalWorkspaceSendInputFn(ctx, ref.windowIndex, ref.tabIndex, args.Text); err != nil {
		return agent.ToolResult{Content: fmt.Sprintf("terminal_workspace send_input failed: %v", err), IsError: true}, nil
	}
	return agent.ToolResult{Content: fmt.Sprintf("Sent input to terminal workspace tab %q.", target)}, nil
}

func (t *TerminalWorkspaceTool) RequiresApproval() bool { return true }

func (t *TerminalWorkspaceTool) IsReadOnlyCall(argsJSON string) bool {
	var args terminalWorkspaceArgs
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return false
	}
	return args.Action == "status" || args.Action == "list_tabs"
}

func terminalWorkspaceTitle(title, command string) string {
	if trimmed := strings.TrimSpace(title); trimmed != "" {
		return trimmed
	}
	fields := strings.Fields(command)
	if len(fields) > 0 {
		base := filepath.Base(fields[0])
		if base != "." && base != string(filepath.Separator) {
			return base
		}
	}
	return "terminal"
}

func terminalWorkspaceColor(name string) string {
	h := fnv.New32a()
	_, _ = h.Write([]byte(name))
	hue := float64(h.Sum32() % 360)
	r, g, b := terminalWorkspaceHSLToRGB(hue, 0.65, 0.45)
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

func terminalWorkspaceHSLToRGB(h, s, l float64) (uint8, uint8, uint8) {
	c := (1 - math.Abs(2*l-1)) * s
	hPrime := h / 60
	x := c * (1 - math.Abs(math.Mod(hPrime, 2)-1))
	var r1, g1, b1 float64
	switch {
	case hPrime < 1:
		r1, g1, b1 = c, x, 0
	case hPrime < 2:
		r1, g1, b1 = x, c, 0
	case hPrime < 3:
		r1, g1, b1 = 0, c, x
	case hPrime < 4:
		r1, g1, b1 = 0, x, c
	case hPrime < 5:
		r1, g1, b1 = x, 0, c
	default:
		r1, g1, b1 = c, 0, x
	}
	m := l - c/2
	return uint8(math.Round((r1 + m) * 255)),
		uint8(math.Round((g1 + m) * 255)),
		uint8(math.Round((b1 + m) * 255))
}

func compareSemverLike(a, b string) int {
	aParts := strings.Split(strings.TrimPrefix(strings.TrimSpace(a), "v"), ".")
	bParts := strings.Split(strings.TrimPrefix(strings.TrimSpace(b), "v"), ".")
	for i := 0; i < len(aParts) || i < len(bParts); i++ {
		var av, bv int
		if i < len(aParts) {
			av, _ = strconv.Atoi(numericVersionPrefix(aParts[i]))
		}
		if i < len(bParts) {
			bv, _ = strconv.Atoi(numericVersionPrefix(bParts[i]))
		}
		if av < bv {
			return -1
		}
		if av > bv {
			return 1
		}
	}
	return 0
}

func numericVersionPrefix(s string) string {
	for i, r := range s {
		if r < '0' || r > '9' {
			if i == 0 {
				return "0"
			}
			return s[:i]
		}
	}
	if s == "" {
		return "0"
	}
	return s
}

func terminalWorkspaceFallbackMessage() string {
	return "Fallback: use bash for non-visible command output, or applescript with Terminal.app for a visible macOS terminal."
}

func terminalWorkspaceUnavailableResult(err error) agent.ToolResult {
	return agent.BusinessError(fmt.Sprintf("%v. %s", err, terminalWorkspaceFallbackMessage()))
}
