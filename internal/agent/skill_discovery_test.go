package agent

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverSkills_Directory(t *testing.T) {
	dir := t.TempDir()

	// Create skill subdirectories and files.
	if err := os.MkdirAll(filepath.Join(dir, "code-review"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "git-helper"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "test-runner.yaml"), []byte("name: test-runner"), 0644); err != nil {
		t.Fatal(err)
	}

	skills := DiscoverSkills(dir)
	if len(skills) != 3 {
		t.Fatalf("expected 3 skills, got %d: %v", len(skills), skills)
	}

	// Verify each skill has a name and path.
	names := make(map[string]bool)
	for _, s := range skills {
		names[s.Name] = true
		if s.Path == "" {
			t.Errorf("skill %q has empty path", s.Name)
		}
	}
	if !names["code-review"] {
		t.Error("expected code-review skill")
	}
	if !names["git-helper"] {
		t.Error("expected git-helper skill")
	}
	if !names["test-runner"] {
		t.Error("expected test-runner skill")
	}
}

func TestDiscoverSkills_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	skills := DiscoverSkills(dir)
	if len(skills) != 0 {
		t.Errorf("expected 0 skills in empty dir, got %d", len(skills))
	}
}

func TestDiscoverSkills_NonexistentDir(t *testing.T) {
	skills := DiscoverSkills("/nonexistent/path/that/does/not/exist")
	if skills != nil {
		t.Errorf("expected nil for nonexistent dir, got %v", skills)
	}
}

func TestDiscoverSkills_EmptyString(t *testing.T) {
	skills := DiscoverSkills("")
	if skills != nil {
		t.Errorf("expected nil for empty path, got %v", skills)
	}
}

func TestDiscoverSkills_HiddenFilesIgnored(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".hidden"), []byte("secret"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "visible"), 0755); err != nil {
		t.Fatal(err)
	}

	skills := DiscoverSkills(dir)
	if len(skills) != 1 {
		t.Fatalf("expected 1 visible skill, got %d: %v", len(skills), skills)
	}
	if skills[0].Name != "visible" {
		t.Errorf("expected 'visible', got %q", skills[0].Name)
	}
}

func TestDiscoverSkills_FileExtStripped(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "my-skill.yaml"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "another.toml"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	skills := DiscoverSkills(dir)
	if len(skills) != 2 {
		t.Fatalf("expected 2 skills, got %d", len(skills))
	}

	names := make(map[string]bool)
	for _, s := range skills {
		names[s.Name] = true
	}
	if !names["my-skill"] {
		t.Errorf("expected 'my-skill' (stripped .yaml), got names %v", names)
	}
	if !names["another"] {
		t.Errorf("expected 'another' (stripped .toml), got names %v", names)
	}
}

func TestDiscoverSkills_Symlink(t *testing.T) {
	// On most platforms symlink creation works. Skip if not supported.
	dir := t.TempDir()
	realDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(realDir, "content.txt"), []byte("data"), 0644); err != nil {
		t.Fatal(err)
	}

	err := os.Symlink(realDir, filepath.Join(dir, "linked-skill"))
	if err != nil {
		t.Skip("symlink not supported on this platform")
	}

	skills := DiscoverSkills(dir)
	if len(skills) != 1 {
		t.Fatalf("expected 1 skill (symlinked dir), got %d", len(skills))
	}
	if skills[0].Name != "linked-skill" {
		t.Errorf("expected 'linked-skill', got %q", skills[0].Name)
	}
}
