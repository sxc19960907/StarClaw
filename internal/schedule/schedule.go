package schedule

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/starclaw/starclaw/internal/agents"
	"github.com/starclaw/starclaw/internal/filelock"
)

// cronFieldRe matches a single cron field: wildcard, number, range, step, or list.
var cronFieldRe = regexp.MustCompile(`^(\*|[0-9]{1,2})(\/([0-9]{1,2}))?$|^[0-9]{1,2}-[0-9]{1,2}(\/[0-9]{1,2})?$|^[0-9]{1,2}(,[0-9]{1,2})*$`)

// namedCronRe matches named cron expressions like @daily, @hourly.
var namedCronRe = regexp.MustCompile(`^@(annually|yearly|monthly|weekly|daily|hourly|reboot)$`)

// Schedule represents a scheduled task.
type Schedule struct {
	ID         string    `json:"id"`
	Agent      string    `json:"agent"`
	Cron       string    `json:"cron"`
	Prompt     string    `json:"prompt"`
	Enabled    bool      `json:"enabled"`
	SyncStatus string    `json:"sync_status"`
	CreatedAt  time.Time `json:"created_at"`
}

// UpdateOpts carries optional fields for updating a schedule.
type UpdateOpts struct {
	Cron    *string
	Prompt  *string
	Enabled *bool
}

// Manager persists and manages scheduled tasks.
type Manager struct {
	indexPath string
}

// NewManager creates a new schedule manager.
func NewManager(indexPath string) *Manager {
	return &Manager{indexPath: indexPath}
}

func validateCron(expr string) error {
	if namedCronRe.MatchString(expr) {
		return nil
	}
	fields := strings.Fields(expr)
	if len(fields) != 5 {
		return fmt.Errorf("invalid cron expression: %q (need 5 fields, got %d)", expr, len(fields))
	}
	for i, f := range fields {
		if !cronFieldRe.MatchString(f) {
			return fmt.Errorf("invalid cron expression: %q (field %d %q is invalid)", expr, i+1, f)
		}
	}
	return nil
}

func validateAgent(name string) error {
	if name == "" {
		return nil
	}
	return agents.ValidateAgentName(name)
}

func validatePrompt(prompt string) error {
	if strings.TrimSpace(prompt) == "" {
		return fmt.Errorf("prompt cannot be empty")
	}
	if strings.ContainsRune(prompt, 0) {
		return fmt.Errorf("prompt contains null bytes")
	}
	return nil
}

func (m *Manager) load() ([]Schedule, error) {
	f, err := os.Open(m.indexPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()
	if err := filelock.Shared(f); err != nil {
		return nil, err
	}
	defer func() {
		_ = filelock.Unlock(f)
	}()
	var schedules []Schedule
	if err := json.NewDecoder(f).Decode(&schedules); err != nil {
		return nil, err
	}
	return schedules, nil
}

func (m *Manager) save(schedules []Schedule) error {
	dir := filepath.Dir(m.indexPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".schedules-*.tmp")
	if err != nil {
		return fmt.Errorf("create temp: %w", err)
	}
	tmpPath := tmp.Name()
	if err := filelock.Exclusive(tmp); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return err
	}
	data, err := json.MarshalIndent(schedules, "", "  ")
	if err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return err
	}
	tmp.Close()
	if err := os.Rename(tmpPath, m.indexPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("atomic rename: %w", err)
	}
	return nil
}

func (m *Manager) lockedModify(fn func([]Schedule) ([]Schedule, error)) error {
	dir := filepath.Dir(m.indexPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("create schedule dir: %w", err)
	}
	lockPath := m.indexPath + ".lock"
	lockFile, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return fmt.Errorf("open lock file: %w", err)
	}
	defer lockFile.Close()
	if err := filelock.Exclusive(lockFile); err != nil {
		return err
	}
	defer func() {
		_ = filelock.Unlock(lockFile)
	}()
	var schedules []Schedule
	if data, err := os.ReadFile(m.indexPath); err == nil {
		if err := json.Unmarshal(data, &schedules); err != nil {
			return fmt.Errorf("parse schedules: %w", err)
		}
	}
	schedules, err = fn(schedules)
	if err != nil {
		return err
	}
	return m.save(schedules)
}

// Create adds a new scheduled task.
func (m *Manager) Create(agentName, cron, prompt string) (string, error) {
	if err := validateAgent(agentName); err != nil {
		return "", err
	}
	if err := validateCron(cron); err != nil {
		return "", err
	}
	if err := validatePrompt(prompt); err != nil {
		return "", err
	}
	id := generateScheduleID()
	s := Schedule{
		ID: id, Agent: agentName, Cron: cron, Prompt: prompt,
		Enabled: true, SyncStatus: "ok", CreatedAt: time.Now(),
	}
	err := m.lockedModify(func(schedules []Schedule) ([]Schedule, error) {
		return append(schedules, s), nil
	})
	if err != nil {
		return "", err
	}
	return id, nil
}

// List returns all scheduled tasks.
func (m *Manager) List() ([]Schedule, error) {
	return m.load()
}

// Get returns a single schedule by ID.
func (m *Manager) Get(id string) (*Schedule, error) {
	schedules, err := m.load()
	if err != nil {
		return nil, err
	}
	for _, s := range schedules {
		if s.ID == id {
			return &s, nil
		}
	}
	return nil, fmt.Errorf("schedule %q not found", id)
}

// Update modifies an existing schedule.
func (m *Manager) Update(id string, opts *UpdateOpts) error {
	if opts.Cron == nil && opts.Prompt == nil && opts.Enabled == nil {
		return fmt.Errorf("no fields to update")
	}
	if opts.Cron != nil {
		if err := validateCron(*opts.Cron); err != nil {
			return err
		}
	}
	if opts.Prompt != nil {
		if err := validatePrompt(*opts.Prompt); err != nil {
			return err
		}
	}
	return m.lockedModify(func(schedules []Schedule) ([]Schedule, error) {
		for i, s := range schedules {
			if s.ID == id {
				if opts.Cron != nil {
					schedules[i].Cron = *opts.Cron
				}
				if opts.Prompt != nil {
					schedules[i].Prompt = *opts.Prompt
				}
				if opts.Enabled != nil {
					schedules[i].Enabled = *opts.Enabled
				}
				return schedules, nil
			}
		}
		return nil, fmt.Errorf("schedule %q not found", id)
	})
}

// Remove deletes a scheduled task.
func (m *Manager) Remove(id string) error {
	return m.lockedModify(func(schedules []Schedule) ([]Schedule, error) {
		filtered := make([]Schedule, 0, len(schedules))
		found := false
		for _, s := range schedules {
			if s.ID == id {
				found = true
				continue
			}
			filtered = append(filtered, s)
		}
		if !found {
			return nil, fmt.Errorf("schedule %q not found", id)
		}
		return filtered, nil
	})
}

// SetSyncStatus is a deprecated no-op kept for compatibility.
func (m *Manager) SetSyncStatus(id, status string) error {
	return nil
}

// Sync is a deprecated no-op kept for compatibility.
func (m *Manager) Sync() (int, error) {
	return 0, nil
}

func generateScheduleID() string {
	b := make([]byte, 4)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		panic(fmt.Errorf("generate schedule id randomness: %w", err))
	}
	return hex.EncodeToString(b)
}
