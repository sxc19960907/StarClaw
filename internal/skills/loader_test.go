package skills

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestValidateSkillName(t *testing.T) {
	tests := []struct {
		name  string
		valid bool
	}{
		{"skill1", true},
		{"my-skill", true},
		{"my_skill", true},
		{"a", true},
		{"skill-123", true},
		{"Skill", false},       // uppercase
		{"-skill", false},      // starts with hyphen
		{"", false},            // empty
		{"skill name", false},  // space
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSkillName(tt.name)
			if tt.valid && err != nil {
				t.Errorf("ValidateSkillName(%q) returned error: %v", tt.name, err)
			}
			if !tt.valid && err == nil {
				t.Errorf("ValidateSkillName(%q) should have returned error", tt.name)
			}
		})
	}
}

func TestParseFrontmatter(t *testing.T) {
	tests := []struct {
		name         string
		content      string
		wantName     string
		wantDesc     string
		wantLicense  string
		wantBody     string
		wantMetadata map[string]string
	}{
		{
			name: "full frontmatter",
			content: `---
name: test-skill
description: A test skill
license: MIT
compatibility: ">=1.0.0"
allowed-tools: file_read bash
---

# Test Skill

This is the skill body.
`,
			wantName:    "test-skill",
			wantDesc:    "A test skill",
			wantLicense: "MIT",
			wantBody:    "# Test Skill\n\nThis is the skill body.",
			wantMetadata: map[string]string{},
		},
		{
			name: "minimal frontmatter",
			content: `---
name: simple-skill
description: A simple skill
---

Simple body.`,
			wantName: "simple-skill",
			wantDesc: "A simple skill",
			wantBody: "Simple body.",
		},
		{
			name:         "no frontmatter",
			content:      "# Just body\n\nNo frontmatter here.",
			wantName:     "",
			wantDesc:     "",
			wantBody:     "# Just body\n\nNo frontmatter here.",
			wantMetadata: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm, body, err := parseFrontmatter(tt.content)
			if err != nil {
				t.Fatalf("parseFrontmatter failed: %v", err)
			}

			if fm.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", fm.Name, tt.wantName)
			}
			if fm.Description != tt.wantDesc {
				t.Errorf("Description = %q, want %q", fm.Description, tt.wantDesc)
			}
			if fm.License != tt.wantLicense {
				t.Errorf("License = %q, want %q", fm.License, tt.wantLicense)
			}
			if body != tt.wantBody {
				t.Errorf("Body = %q, want %q", body, tt.wantBody)
			}
			if tt.wantMetadata != nil {
				for k, v := range tt.wantMetadata {
					if fm.Metadata[k] != v {
						t.Errorf("Metadata[%q] = %q, want %q", k, fm.Metadata[k], v)
					}
				}
			}
		})
	}
}

func TestLoadSkills(t *testing.T) {
	// Create temp skills directory
	tmpDir := t.TempDir()

	// Create skill source
	source := SkillSource{
		Dir:    tmpDir,
		Source: SourceGlobal,
	}

	// Create test skill
	skillDir := filepath.Join(tmpDir, "test-skill")
	os.MkdirAll(skillDir, 0755)

	skillContent := `---
name: test-skill
description: A test skill for unit testing
license: MIT
allowed-tools: file_read file_write
---

# Test Skill

You are a test skill.
`
	os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(skillContent), 0644)

	// Load skills
	skills, err := LoadSkills(source)
	if err != nil {
		t.Fatalf("LoadSkills failed: %v", err)
	}

	if len(skills) != 1 {
		t.Fatalf("Expected 1 skill, got %d", len(skills))
	}

	skill := skills[0]
	if skill.Name != "test-skill" {
		t.Errorf("Name = %q, want %q", skill.Name, "test-skill")
	}
	if skill.Description != "A test skill for unit testing" {
		t.Errorf("Description = %q, want %q", skill.Description, "A test skill for unit testing")
	}
	if skill.License != "MIT" {
		t.Errorf("License = %q, want %q", skill.License, "MIT")
	}
	if len(skill.AllowedTools) != 2 {
		t.Errorf("AllowedTools length = %d, want 2", len(skill.AllowedTools))
	}
	if skill.Source != SourceGlobal {
		t.Errorf("Source = %q, want %q", skill.Source, SourceGlobal)
	}
}

func TestLoadSkills_MultipleSources(t *testing.T) {
	// Create two temp directories
	tmpDir1 := t.TempDir()
	tmpDir2 := t.TempDir()

	sources := []SkillSource{
		{Dir: tmpDir1, Source: SourceBundled},
		{Dir: tmpDir2, Source: SourceGlobal},
	}

	// Create skill in first source
	skillDir1 := filepath.Join(tmpDir1, "skill-a")
	os.MkdirAll(skillDir1, 0755)
	os.WriteFile(filepath.Join(skillDir1, "SKILL.md"), []byte(`---
name: skill-a
description: Skill A
---
`), 0644)

	// Create skills in second source (including duplicate)
	skillDir2a := filepath.Join(tmpDir2, "skill-a") // Duplicate
	skillDir2b := filepath.Join(tmpDir2, "skill-b") // New
	os.MkdirAll(skillDir2a, 0755)
	os.MkdirAll(skillDir2b, 0755)
	os.WriteFile(filepath.Join(skillDir2a, "SKILL.md"), []byte(`---
name: skill-a
description: Skill A Override
---
`), 0644)
	os.WriteFile(filepath.Join(skillDir2b, "SKILL.md"), []byte(`---
name: skill-b
description: Skill B
---
`), 0644)

	skills, err := LoadSkills(sources...)
	if err != nil {
		t.Fatalf("LoadSkills failed: %v", err)
	}

	// Should have 2 skills (skill-a from first source, skill-b from second)
	if len(skills) != 2 {
		t.Errorf("Expected 2 skills, got %d", len(skills))
	}

	// Verify skill-a came from bundled (first source)
	for _, s := range skills {
		if s.Name == "skill-a" {
			if s.Source != SourceBundled {
				t.Error("skill-a should be from bundled source (first wins)")
			}
			if s.Description != "Skill A" {
				t.Error("skill-a should have description from bundled")
			}
		}
	}
}

func TestLoadSkill(t *testing.T) {
	tmpDir := t.TempDir()
	source := SkillSource{
		Dir:    tmpDir,
		Source: SourceGlobal,
	}

	// Create skill
	skillDir := filepath.Join(tmpDir, "find-me")
	os.MkdirAll(skillDir, 0755)
	os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(`---
name: find-me
description: Found me!
---
`), 0644)

	// Load specific skill
	skill, err := LoadSkill("find-me", source)
	if err != nil {
		t.Fatalf("LoadSkill failed: %v", err)
	}

	if skill.Name != "find-me" {
		t.Errorf("Name = %q, want %q", skill.Name, "find-me")
	}

	// Try loading non-existent skill
	_, err = LoadSkill("not-exist", source)
	if err == nil {
		t.Error("LoadSkill should return error for missing skill")
	}
}

func TestListSkills(t *testing.T) {
	tmpDir := t.TempDir()
	source := SkillSource{
		Dir:    tmpDir,
		Source: SourceGlobal,
	}

	// Create skills
	for _, name := range []string{"zebra", "alpha", "beta"} {
		skillDir := filepath.Join(tmpDir, name)
		os.MkdirAll(skillDir, 0755)
		content := fmt.Sprintf(`---
name: %s
description: Skill %s
---
`, name, name)
		os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0644)
	}

	metas, err := ListSkills(source)
	if err != nil {
		t.Fatalf("ListSkills failed: %v", err)
	}

	if len(metas) != 3 {
		t.Errorf("Expected 3 skills, got %d", len(metas))
	}

	// Should be sorted by name
	if len(metas) >= 3 && metas[0].Name != "alpha" {
		t.Error("Skills should be sorted alphabetically")
	}
}

func TestSkill_ToMeta(t *testing.T) {
	skill := &Skill{
		Name:        "test-skill",
		Description: "Test description",
		Source:      SourceGlobal,
		Prompt:      "This should not appear in meta",
	}

	meta := skill.ToMeta()

	if meta.Name != "test-skill" {
		t.Errorf("Meta.Name = %q, want %q", meta.Name, "test-skill")
	}
	if meta.Description != "Test description" {
		t.Errorf("Meta.Description = %q, want %q", meta.Description, "Test description")
	}
	if meta.Source != SourceGlobal {
		t.Errorf("Meta.Source = %q, want %q", meta.Source, SourceGlobal)
	}
}
