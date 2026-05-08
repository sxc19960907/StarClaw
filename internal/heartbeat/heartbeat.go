// Package heartbeat provides periodic health checks for agents.
//
// The heartbeat Manager scans the agents directory for agents with a heartbeat
// config entry (every, active_hours), then runs periodic checks. Each check
// reads the agent's HEARTBEAT.md, builds a prompt, and calls a RunAgent
// callback. If the agent replies with "HEARTBEAT_OK" the check passes silently.
// Any other reply is treated as an action that the agent initiated on its own.
package heartbeat

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/starclaw/starclaw/internal/agents"
	"github.com/starclaw/starclaw/internal/watcher"
)

const maxChecklistChars = 4000

// Deps holds the external dependencies needed by the heartbeat Manager.
type Deps struct {
	// RunAgent is called to execute a heartbeat check for the given agent.
	// The prompt is the formatted heartbeat prompt. The returned string is
	// the agent's reply, which should be "HEARTBEAT_OK" if everything is
	// fine. Any other reply is treated as an action message.
	RunAgent func(ctx context.Context, agent, prompt string) (string, error)
}

// Manager runs periodic heartbeat checks for all configured agents.
type Manager struct {
	agents []*agentHeartbeat
	deps   *Deps
	cancel context.CancelFunc
	done   chan struct{}
}

// agentHeartbeat holds per-agent heartbeat state.
type agentHeartbeat struct {
	name        string
	interval    time.Duration
	activeHours string
	agentDir    string
	mu          sync.Mutex
}

// IsHeartbeatOK returns true if the agent reply is the silent acknowledgement
// token "HEARTBEAT_OK" (case-insensitive, trimmed). Any extra text outside
// whitespace causes this to return false.
func IsHeartbeatOK(reply string) bool {
	return strings.EqualFold(strings.TrimSpace(reply), "HEARTBEAT_OK")
}

// FormatPrompt builds a heartbeat prompt from a checklist body. The prompt
// instructs the agent to review the checklist items and acknowledge via
// "HEARTBEAT_OK" if everything is fine.
func FormatPrompt(checklist string) string {
	return fmt.Sprintf(`This is a periodic heartbeat check. Review the checklist below and check each item using your available tools. If everything is fine, reply with exactly "HEARTBEAT_OK" and nothing else. If something needs attention, describe the issue concisely.

Checklist:
%s`, checklist)
}

// FormatGoalPrompt builds a goal-driven heartbeat prompt. The prompt instructs
// the agent to review its goals and take action if needed, or reply with
// "HEARTBEAT_OK" if nothing requires attention.
func FormatGoalPrompt(goals string) string {
	return fmt.Sprintf(`This is a periodic check-in. Review your goals below and your current conversation context. If something needs your attention, take action using your available tools. If nothing needs doing, reply with exactly "HEARTBEAT_OK" and nothing else.

Goals:
%s`, goals)
}

// ReadChecklist reads a HEARTBEAT.md file at the given path.
// A missing file or an empty/whitespace-only file returns ("", nil) — this is
// the expected "disabled" state. Other read errors (e.g. permission denied)
// return ("", error) so callers can detect degraded monitoring.
// Content exceeding maxChecklistChars is truncated with a log warning.
func ReadChecklist(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("read HEARTBEAT.md: %w", err)
	}
	content := strings.TrimSpace(string(data))
	if content == "" {
		return "", nil
	}
	if len(content) > maxChecklistChars {
		log.Printf("heartbeat: HEARTBEAT.md at %s exceeds %d chars, truncating", path, maxChecklistChars)
		content = content[:maxChecklistChars]
	}
	return content, nil
}

// New creates a heartbeat Manager by scanning the agents directory for agents
// that have a heartbeat config entry. Returns an empty (but valid) Manager if
// no agents have heartbeat configured. Only returns an error if listing agents
// fails at the filesystem level.
func New(agentsDir string, deps *Deps) (*Manager, error) {
	if deps == nil {
		return nil, fmt.Errorf("heartbeat: deps is required")
	}

	infos, err := agents.ListAgents(agentsDir)
	if err != nil {
		return nil, fmt.Errorf("heartbeat: list agents: %w", err)
	}

	var entries []*agentHeartbeat
	for _, info := range infos {
		ag, lErr := agents.LoadAgent(agentsDir, info.Name)
		if lErr != nil {
			log.Printf("heartbeat: skip agent %q: %v", info.Name, lErr)
			continue
		}
		if ag.Config == nil || ag.Config.Heartbeat == nil || ag.Config.Heartbeat.Every == "" {
			continue
		}
		hb := ag.Config.Heartbeat

		interval, pErr := time.ParseDuration(hb.Every)
		if pErr != nil {
			log.Printf("heartbeat: skip agent %q: invalid interval %q: %v", info.Name, hb.Every, pErr)
			continue
		}
		if interval < 1*time.Minute {
			log.Printf("heartbeat: skip agent %q: interval %s too short (min 1m)", info.Name, interval)
			continue
		}

		entries = append(entries, &agentHeartbeat{
			name:        info.Name,
			interval:    interval,
			activeHours: hb.ActiveHours,
			agentDir:    filepath.Join(agentsDir, info.Name),
		})
	}

	return &Manager{
		agents: entries,
		deps:   deps,
		done:   make(chan struct{}),
	}, nil
}

// Start launches per-agent ticker goroutines. It creates a derived cancellable
// context so that Close() can stop all tickers. Start returns immediately after
// launching the goroutines. The caller should call Close() to clean up.
func (m *Manager) Start(ctx context.Context) {
	ctx, m.cancel = context.WithCancel(ctx)

	var wg sync.WaitGroup
	for _, ah := range m.agents {
		wg.Add(1)
		go func(ah *agentHeartbeat) {
			defer wg.Done()
			m.runTicker(ctx, ah)
		}(ah)
	}

	go func() {
		wg.Wait()
		close(m.done)
	}()
}

// runTicker runs the heartbeat ticker loop for a single agent.
func (m *Manager) runTicker(ctx context.Context, ah *agentHeartbeat) {
	ticker := time.NewTicker(ah.interval)
	defer ticker.Stop()

	log.Printf("heartbeat: started for agent %q every %s", ah.name, ah.interval)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.tick(ctx, ah)
		}
	}
}

// tick executes a single heartbeat check for an agent. It uses TryLock to
// prevent overlapping ticks, checks active hours, reads the HEARTBEAT.md,
// builds a prompt, and calls RunAgent.
func (m *Manager) tick(ctx context.Context, ah *agentHeartbeat) {
	if !ah.mu.TryLock() {
		log.Printf("heartbeat: skip %q (previous tick still running)", ah.name)
		return
	}
	defer ah.mu.Unlock()

	if !watcher.InActiveHours(ah.activeHours, time.Now()) {
		log.Printf("heartbeat: skip %q (outside active hours)", ah.name)
		return
	}

	checklistPath := filepath.Join(ah.agentDir, "HEARTBEAT.md")
	checklist, rErr := ReadChecklist(checklistPath)
	if rErr != nil || checklist == "" {
		log.Printf("heartbeat: skip %q (no checklist)", ah.name)
		return
	}

	prompt := FormatPrompt(checklist)
	reply, rErr := m.deps.RunAgent(ctx, ah.name, prompt)
	if rErr != nil {
		if ctx.Err() != nil {
			log.Printf("heartbeat: %q canceled", ah.name)
			return
		}
		log.Printf("heartbeat: %q error: %v", ah.name, rErr)
		return
	}

	if IsHeartbeatOK(reply) {
		log.Printf("heartbeat: %q ok", ah.name)
		return
	}

	log.Printf("heartbeat: %q action: %s", ah.name, reply)
}

// Close cancels all running ticker goroutines and waits for them to finish.
// It is safe to call multiple times; subsequent calls are no-ops.
func (m *Manager) Close() {
	if m.cancel != nil {
		m.cancel()
	}
	<-m.done
}
