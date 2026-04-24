# Specification: Skills System

## Overview

| Field | Value |
|-------|-------|
| Feature | Skills System |
| Type | Extension Feature |
| Optional | Yes (no skills = default behavior) |
| Default | Bundled skills only (if any) |

## Purpose

Enable composable, shareable capabilities that can be loaded on-demand. Skills follow the Anthropic Agent Skills specification for interoperability.

## SKILL.md Format

```markdown
---
name: github-ops
description: Advanced GitHub operations for CI/CD workflows
license: MIT
compatibility: ">=1.0.0"
metadata:
  author: starclaw-team
  version: "1.2.0"
  tags: [github, ci-cd, automation]
allowed-tools: file_read file_write bash http
---

# GitHub Operations Skill

You are a GitHub CI/CD expert. You help users with GitHub Actions, workflows, and repository management.

## Capabilities

- Set up GitHub Actions workflows
- Debug failed builds
- Optimize workflow performance
- Manage secrets and environments

## Best Practices

- Always use pinned action versions
- Cache dependencies when possible
- Use matrix builds for testing across versions
```

## Frontmatter Fields

| Field | Required | Type | Description |
|-------|----------|------|-------------|
| name | Yes | string | Skill identifier (matches directory) |
| description | Yes | string | Short description |
| license | No | string | License identifier (MIT, Apache-2.0, etc.) |
| compatibility | No | string | Semver range for compatibility |
| metadata | No | object | Arbitrary key-value pairs |
| allowed-tools | No | string | Space-separated tool names |

## Name Validation

Skill names must match regex: `^[a-z0-9][a-z0-9_-]{0,63}$`

Same rules as agent names:
- Start with alphanumeric
- Lowercase only
- Alphanumeric, underscore, hyphen allowed
- Max 64 characters

## Skill Sources

Priority order (highest wins on duplicate):

1. **Bundled** - Embedded in StarClaw binary
   - Location: Compiled-in
   - Source: `bundled`
   
2. **Global** - User's skills directory
   - Location: `~/.starclaw/skills/`
   - Source: `global`
   
3. **Agent** - Agent-specific skills
   - Location: `~/.starclaw/agents/<name>/skills/`
   - Source: `agent`

### Directory Structure

```
~/.starclaw/
├── skills/
│   ├── github-ops/
│   │   └── SKILL.md
│   ├── database/
│   │   └── SKILL.md
│   └── testing/
│       └── SKILL.md
└── agents/
    └── coder/
        ├── AGENT.md
        └── skills/
            └── code-review/
                └── SKILL.md
```

## Skill Registry API

```go
package skills

// Skill represents a loaded skill
type Skill struct {
    Name          string
    Description   string
    Prompt        string           // Body content after frontmatter
    License       string
    Compatibility string
    Metadata      map[string]string
    AllowedTools  []string         // Parsed from allowed-tools
    Source        string           // "bundled", "global", "agent"
    Dir           string           // Source directory
}

// Lightweight metadata for listing
type SkillMeta struct {
    Name        string
    Description string
    Source      string
}

// Convert to metadata (excludes prompt body)
func (s *Skill) ToMeta() SkillMeta

// Skill source for loading
type SkillSource struct {
    Dir    string
    Source string
}

// Load skills from multiple sources
func LoadSkills(sources ...SkillSource) ([]*Skill, error)

// Load single skill by name (searches in priority order)
func LoadSkill(name string) (*Skill, error)

// Validate skill name format
func ValidateSkillName(name string) error

// List available skills
func ListSkills(sources ...SkillSource) ([]SkillMeta, error)
```

## use_skill Tool

Skills are activated via the `use_skill` tool:

```go
type UseSkillArgs struct {
    Name string `json:"name"` // Skill name to activate
}

// Tool implementation
func (t *UseSkillTool) Execute(args UseSkillArgs) (string, error) {
    skill, err := skills.LoadSkill(args.Name)
    if err != nil {
        return "", fmt.Errorf("skill not found: %s", args.Name)
    }
    return skill.Prompt, nil
}
```

### Tool Schema

```json
{
  "name": "use_skill",
  "description": "Activate a skill by name. Returns the skill's prompt for injection into context.",
  "input_schema": {
    "type": "object",
    "properties": {
      "name": {
        "type": "string",
        "description": "Name of the skill to activate"
      }
    },
    "required": ["name"]
  }
}
```

## Tool Filtering

When a skill specifies `allowed-tools`, the agent should respect it:

```yaml
allowed-tools: file_read bash http
```

This means:
- Only these tools are available when skill is active
- Other tools are hidden from the LLM
- Used for safety and focus

## CLI Commands (Future)

Potential future additions:

```bash
starclaw skills list              # List available skills
starclaw skills show <name>       # Show skill details
starclaw skills install <path>    # Install skill from path
```

## Error Handling

| Error | Behavior |
|-------|----------|
| Missing frontmatter | Parse error, skill rejected |
| Missing name | Validation error |
| Name/dir mismatch | Error (name must match directory) |
| Invalid name | Validation error |
| Missing description | Validation error |
| Invalid YAML | Parse error with context |

## Testing Requirements

### Unit Tests

- Frontmatter parsing
- Name validation
- Source priority (deduplication)
- Tool filtering

### Integration Tests

- Load skills from disk
- Multiple sources
- use_skill tool execution

### Test Skills

Create test skills in temp directory:

```go
func createTestSkill(t *testing.T, dir, name string) {
    skillDir := filepath.Join(dir, name)
    os.MkdirAll(skillDir, 0755)
    content := `---
name: ` + name + `
description: Test skill
---

Test skill content.
`
    os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0644)
}
```

## Bundled Skills

Bundled skills are embedded in the binary using `//go:embed`:

```go
//go:embed bundled/*
var bundledSkills embed.FS

func ExtractBundledSkills(targetDir string) error {
    // Extract to targetDir on first run
}
```

## Compatibility

The `compatibility` field uses semver range syntax:

- `>=1.0.0` - Requires StarClaw v1.0.0 or later
- `^1.2.0` - Compatible with 1.2.0 and higher minor/patch
- `~1.2.0` - Compatible with 1.2.x

StarClaw checks compatibility before loading skills.

## Sharing Skills

Skills can be shared by:

1. Copying SKILL.md files
2. Git repositories (e.g., `starclaw-skills` repo)
3. Future: Skill registry/marketplace

## Security Considerations

- Skills are text files (markdown), not executable code
- No code execution from skills
- Prompt injection is possible (user-controlled content)
- Tool filtering provides some sandboxing

## References

- Anthropic Skills Spec: https://agentskills.io/specification
- ShanClaw Reference: `/Users/timmy/PycharmProjects/ShanClaw/internal/skills/loader.go`
