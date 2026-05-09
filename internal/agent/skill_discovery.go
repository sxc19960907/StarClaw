package agent

import (
	"os"
	"path/filepath"
	"strings"
)

// SkillInfo describes a discovered skill.
type SkillInfo struct {
	Name        string
	Description string
	Path        string
}

// DiscoverSkills scans a directory for skill files and returns info about them.
// It reads the directory listing and returns a SkillInfo for each entry.
// Directories (including symlinked directories) are included as skills with
// their name and path. Regular files are included with their name (extension
// stripped) and path.
func DiscoverSkills(skillsDir string) []SkillInfo {
	if skillsDir == "" {
		return nil
	}

	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		return nil
	}

	var skills []SkillInfo
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}

		fullPath := filepath.Join(skillsDir, name)

		// Use os.Stat to follow symlinks for correct type detection.
		fi, err := os.Stat(fullPath)
		if err != nil {
			continue
		}

		if fi.IsDir() {
			skills = append(skills, SkillInfo{
				Name: name,
				Path: fullPath,
			})
		} else if fi.Mode().IsRegular() {
			skillName := strings.TrimSuffix(name, filepath.Ext(name))
			skills = append(skills, SkillInfo{
				Name: skillName,
				Path: fullPath,
			})
		}
	}

	return skills
}
