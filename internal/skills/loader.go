// Package skills provides skill loading and management.
// Skills are composable capabilities loaded from SKILL.md files.
package skills

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var skillNameRe = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,63}$`)

// Skill represents a loaded skill.
type Skill struct {
	Name          string
	Description   string
	Prompt        string
	License       string
	Compatibility string
	Metadata      map[string]string
	AllowedTools  []string
	Source        string
	Dir           string
}

// SkillMeta is the lightweight representation for listing.
type SkillMeta struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Source      string `json:"source,omitempty"`
}

// ToMeta returns API-safe metadata without the full prompt body.
func (s *Skill) ToMeta() SkillMeta {
	return SkillMeta{
		Name:        s.Name,
		Description: s.Description,
		Source:      s.Source,
	}
}

// SkillSource identifies where to load skills from.
type SkillSource struct {
	Dir    string
	Source string
}

// Source constants.
const (
	SourceGlobal  = "global"
	SourceBundled = "bundled"
	SourceAgent   = "agent"
)

// ValidateSkillName validates a skill name format.
func ValidateSkillName(name string) error {
	if !skillNameRe.MatchString(name) {
		return fmt.Errorf("invalid skill name %q: must match %s", name, skillNameRe.String())
	}
	return nil
}

// LoadSkills loads skills from multiple sources with priority (later sources override earlier).
func LoadSkills(sources ...SkillSource) ([]*Skill, error) {
	seen := make(map[string]bool)
	var result []*Skill

	for _, src := range sources {
		if _, err := os.Stat(src.Dir); os.IsNotExist(err) {
			continue
		}

		entries, err := os.ReadDir(src.Dir)
		if err != nil {
			continue
		}

		// Sort for deterministic ordering
		var names []string
		for _, e := range entries {
			if e.IsDir() {
				names = append(names, e.Name())
			}
		}
		sort.Strings(names)

		for _, name := range names {
			if seen[name] {
				continue
			}

			skillDir := filepath.Join(src.Dir, name)
			skillFile := filepath.Join(skillDir, "SKILL.md")

			if _, err := os.Stat(skillFile); os.IsNotExist(err) {
				continue
			}

			skill, err := loadSkillMD(skillFile, name, src.Source)
			if err != nil {
				return nil, fmt.Errorf("loading skill %s: %w", name, err)
			}

			skill.Dir = skillDir
			seen[name] = true
			result = append(result, skill)
		}
	}

	return result, nil
}

// loadSkillMD loads a single skill from a SKILL.md file.
func loadSkillMD(path, dirName, source string) (*Skill, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	content := string(data)

	// Parse frontmatter (simple YAML between --- markers)
	fm, body, err := parseFrontmatter(content)
	if err != nil {
		return nil, fmt.Errorf("parse frontmatter: %w", err)
	}

	// Validate name
	if fm.Name == "" {
		return nil, fmt.Errorf("skill name is required in frontmatter")
	}
	if fm.Name != dirName {
		return nil, fmt.Errorf("skill name %q must match directory name %q", fm.Name, dirName)
	}
	if err := ValidateSkillName(fm.Name); err != nil {
		return nil, err
	}

	if fm.Description == "" {
		return nil, fmt.Errorf("skill description is required")
	}

	// Parse allowed-tools
	var allowedTools []string
	if fm.AllowedTools != "" {
		allowedTools = strings.Fields(fm.AllowedTools)
	}

	return &Skill{
		Name:          fm.Name,
		Description:   fm.Description,
		Prompt:        strings.TrimSpace(body),
		License:       fm.License,
		Compatibility: fm.Compatibility,
		Metadata:      fm.Metadata,
		AllowedTools:  allowedTools,
		Source:        source,
	}, nil
}

// frontmatter represents the YAML frontmatter of a SKILL.md file.
type frontmatter struct {
	Name          string
	Description   string
	License       string
	Compatibility string
	Metadata      map[string]string
	AllowedTools  string
}

// parseFrontmatter extracts YAML frontmatter from content.
// It expects frontmatter between --- markers at the start of the file.
func parseFrontmatter(content string) (*frontmatter, string, error) {
	fm := &frontmatter{
		Metadata: make(map[string]string),
	}

	// Check if content starts with ---
	if !strings.HasPrefix(content, "---") {
		return fm, content, nil // No frontmatter
	}

	// Find the closing ---
	endIdx := strings.Index(content[3:], "\n---")
	if endIdx == -1 {
		return fm, content, nil // No closing marker, treat as no frontmatter
	}

	// Extract frontmatter YAML (between the markers)
	frontmatterYAML := content[3 : 3+endIdx]
	body := strings.TrimSpace(content[3+endIdx+4:])

	// Parse simple key: value pairs
	for _, line := range strings.Split(frontmatterYAML, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Handle metadata specially (key: value pairs under metadata:)
		if strings.HasPrefix(line, "metadata:") {
			continue // Skip the metadata header
		}
		if strings.HasPrefix(line, "  ") && strings.Contains(line, ":") {
			// This is a metadata entry
			parts := strings.SplitN(strings.TrimSpace(line), ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				value = strings.Trim(value, `"'`)
				fm.Metadata[key] = value
			}
			continue
		}

		if !strings.Contains(line, ":") {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		value = strings.Trim(value, `"'`)

		switch key {
		case "name":
			fm.Name = value
		case "description":
			fm.Description = value
		case "license":
			fm.License = value
		case "compatibility":
			fm.Compatibility = value
		case "allowed-tools":
			fm.AllowedTools = value
		}
	}

	return fm, body, nil
}

// LoadSkill loads a single skill by name (searches in priority order).
func LoadSkill(name string, sources ...SkillSource) (*Skill, error) {
	skills, err := LoadSkills(sources...)
	if err != nil {
		return nil, err
	}

	for _, s := range skills {
		if s.Name == name {
			return s, nil
		}
	}

	return nil, fmt.Errorf("skill %q not found", name)
}

// ListSkills lists all available skills.
func ListSkills(sources ...SkillSource) ([]SkillMeta, error) {
	skills, err := LoadSkills(sources...)
	if err != nil {
		return nil, err
	}

	var metas []SkillMeta
	for _, s := range skills {
		metas = append(metas, s.ToMeta())
	}

	return metas, nil
}
