package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/starclaw/starclaw/internal/agent"
	"github.com/starclaw/starclaw/internal/config"
	"github.com/starclaw/starclaw/internal/skills"
)

// SkillTool manages dynamic skill loading. It supports loading skills from
// the global skills directory, unloading previously loaded skills, and
// listing currently available or loaded skills.
type SkillTool struct {
	mu     sync.Mutex
	skills []*skills.Skill
	loaded map[string]*skills.Skill // name -> skill for loaded skills
}

// NewSkillTool creates a new SkillTool and pre-loads all available skills
// from the bundled and global directories.
func NewSkillTool() *SkillTool {
	t := &SkillTool{
		loaded: make(map[string]*skills.Skill),
	}

	// Pre-load bundled + global skills
	sources := t.skillSources()
	allSkills, err := skills.LoadSkills(sources...)
	if err == nil {
		t.skills = allSkills
	}

	return t
}

// SkillListEntry represents a skill entry for the list action.
type SkillListEntry struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Source      string `json:"source,omitempty"`
	Loaded      bool   `json:"loaded"`
}

type skillArgs struct {
	Action string `json:"action"`
	Name   string `json:"name,omitempty"`
}

func (t *SkillTool) Info() agent.ToolInfo {
	return agent.ToolInfo{
		Name:        "skill",
		Description: "Manage dynamically loaded skills. Actions: load (activate a skill by name), unload (deactivate a loaded skill), list (show available and loaded skills).",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"action": map[string]any{
					"type": "string",
					"enum": []string{"load", "unload", "list"},
					"description": "Action to perform: 'load' activates a skill, " +
						"'unload' deactivates it, 'list' shows available skills",
				},
				"name": map[string]any{
					"type":        "string",
					"description": "Skill name (required for load and unload)",
				},
			},
			"required": []string{"action"},
		},
	}
}

func (t *SkillTool) Run(ctx context.Context, argsJSON string) (agent.ToolResult, error) {
	var args skillArgs
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return agent.ValidationError("invalid arguments: " + err.Error()), nil
	}

	switch args.Action {
	case "list":
		return t.listSkills(ctx)
	case "load":
		return t.loadSkill(ctx, args.Name)
	case "unload":
		return t.unloadSkill(ctx, args.Name)
	default:
		return agent.ValidationError(fmt.Sprintf("unknown action %q: must be load, unload, or list", args.Action)), nil
	}
}

// skillSources returns the bundled and global skill sources in priority order.
func (t *SkillTool) skillSources() []skills.SkillSource {
	// Bundled skills are embedded in the binary (internal/skills/bundled/)
	bundledSrc := skills.SkillSource{
		Source: skills.SourceBundled,
	}
	dir := bundledSkillsDir()
	if dir != "" {
		bundledSrc.Dir = dir
	}

	// Global skills from ~/.starclaw/skills/
	globalSrc := skills.SkillSource{
		Source: skills.SourceGlobal,
	}
	starclawDir := config.StarclawDir()
	if starclawDir != "" {
		globalSrc.Dir = starclawDir + "/skills"
	}

	return []skills.SkillSource{bundledSrc, globalSrc}
}

// bundledSkillsDir returns the path to bundled skills if they exist.
func bundledSkillsDir() string {
	starclawDir := config.StarclawDir()
	if starclawDir == "" {
		return ""
	}
	return starclawDir + "/skills"
}

// refreshSkills reloads the skill list from sources.
func (t *SkillTool) refreshSkills() {
	sources := t.skillSources()
	filtered := make([]skills.SkillSource, 0, len(sources))
	for _, src := range sources {
		if src.Dir != "" {
			filtered = append(filtered, src)
		}
	}
	if len(filtered) == 0 {
		return
	}
	allSkills, err := skills.LoadSkills(filtered...)
	if err != nil {
		return
	}
	t.skills = allSkills
}

// findSkill finds a skill by name in the loaded skills list.
func (t *SkillTool) findSkill(name string) *skills.Skill {
	for _, s := range t.skills {
		if s.Name == name {
			return s
		}
	}
	return nil
}

func (t *SkillTool) listSkills(_ context.Context) (agent.ToolResult, error) {
	t.mu.Lock()
	t.refreshSkills()
	t.mu.Unlock()

	t.mu.Lock()
	defer t.mu.Unlock()

	if len(t.skills) == 0 {
		return agent.ToolResult{Content: "No skills available."}, nil
	}

	// Sort by name
	sorted := make([]*skills.Skill, len(t.skills))
	copy(sorted, t.skills)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Name < sorted[j].Name
	})

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Available skills (%d total):\n\n", len(sorted)))
	for _, s := range sorted {
		loaded := t.loaded[s.Name] != nil
		status := "available"
		if loaded {
			status = "loaded"
		}
		sb.WriteString(fmt.Sprintf("  %s [%s]\n", s.Name, status))
		if s.Description != "" {
			sb.WriteString(fmt.Sprintf("    %s\n", s.Description))
		}
		sb.WriteString(fmt.Sprintf("    source: %s\n", s.Source))
	}

	sb.WriteString(fmt.Sprintf("\nUse 'load' to activate a skill, 'unload' to deactivate."))
	return agent.ToolResult{Content: sb.String()}, nil
}

func (t *SkillTool) loadSkill(_ context.Context, name string) (agent.ToolResult, error) {
	if name == "" {
		return agent.ValidationError("name is required for action=load"), nil
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	// Check if already loaded
	if _, exists := t.loaded[name]; exists {
		return agent.ToolResult{Content: fmt.Sprintf("Skill %q is already loaded.", name)}, nil
	}

	t.refreshSkills()

	skill := t.findSkill(name)
	if skill == nil {
		return agent.ValidationError(fmt.Sprintf("skill %q not found", name)), nil
	}

	t.loaded[name] = skill
	return agent.ToolResult{Content: fmt.Sprintf("Skill %q loaded.\n\n%s\n\n%s",
		skill.Name, skill.Description, skill.Prompt)}, nil
}

func (t *SkillTool) unloadSkill(_ context.Context, name string) (agent.ToolResult, error) {
	if name == "" {
		return agent.ValidationError("name is required for action=unload"), nil
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	if _, exists := t.loaded[name]; !exists {
		return agent.ValidationError(fmt.Sprintf("skill %q is not loaded", name)), nil
	}

	delete(t.loaded, name)
	return agent.ToolResult{Content: fmt.Sprintf("Skill %q unloaded.", name)}, nil
}

func (t *SkillTool) RequiresApproval() bool { return false }
