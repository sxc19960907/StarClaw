package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/starclaw/starclaw/internal/agent"
	"github.com/starclaw/starclaw/internal/skills"
)

// UseSkillTool allows the agent to activate a skill by name.
// When activated, the skill's prompt is returned to be injected into the system prompt.
type UseSkillTool struct {
	skillsDir string
}

// NewUseSkillTool creates a new use_skill tool.
func NewUseSkillTool(skillsDir string) *UseSkillTool {
	return &UseSkillTool{
		skillsDir: skillsDir,
	}
}

// UseSkillArgs defines the arguments for the use_skill tool.
type UseSkillArgs struct {
	Name string `json:"name"` // Name of the skill to activate
}

// Info returns the tool's metadata.
func (t *UseSkillTool) Info() agent.ToolInfo {
	return agent.ToolInfo{
		Name:        "use_skill",
		Description: "Activate a skill by name. Returns the skill's prompt content for injection into the conversation context. Use this to access specialized capabilities and knowledge.",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"name": map[string]interface{}{
					"type":        "string",
					"description": "Name of the skill to activate (e.g., 'github-ops', 'database-queries')",
				},
			},
			"required": []string{"name"},
		},
	}
}

// Run activates the skill and returns its prompt.
func (t *UseSkillTool) Run(ctx context.Context, args string) (agent.ToolResult, error) {
	var skillArgs UseSkillArgs
	if err := json.Unmarshal([]byte(args), &skillArgs); err != nil {
		return agent.ToolResult{
			Content: fmt.Sprintf("Error: invalid arguments: %v", err),
			IsError: true,
		}, nil
	}

	// Validate skill name
	if err := skills.ValidateSkillName(skillArgs.Name); err != nil {
		return agent.ToolResult{
			Content: fmt.Sprintf("Error: %v", err),
			IsError: true,
		}, nil
	}

	// Load skill from global directory
	source := skills.SkillSource{
		Dir:    t.skillsDir,
		Source: skills.SourceGlobal,
	}

	skill, err := skills.LoadSkill(skillArgs.Name, source)
	if err != nil {
		return agent.ToolResult{
			Content: fmt.Sprintf("Error: skill %q not found: %v", skillArgs.Name, err),
			IsError: true,
		}, nil
	}

	// Return the skill prompt for injection
	result := fmt.Sprintf("Skill '%s' activated.\n\nDescription: %s\n\n--- Skill Instructions ---\n\n%s",
		skill.Name,
		skill.Description,
		skill.Prompt,
	)

	return agent.ToolResult{
		Content: result,
		IsError: false,
	}, nil
}

// RequiresApproval returns false - use_skill is safe to auto-approve.
func (t *UseSkillTool) RequiresApproval() bool {
	return false
}

// SetSkillsDir updates the skills directory.
func (t *UseSkillTool) SetSkillsDir(dir string) {
	t.skillsDir = dir
}
