// Package prompt assembles the system prompt from layered components.
package prompt

import (
	"fmt"
	"strings"
	"time"
)

const (
	maxMemoryChars       = 2000
	maxInstructionsChars = 16000
)

// Options configures system prompt assembly.
type Options struct {
	BasePrompt    string   // persona + core operational rules
	Memory        string   // agent memory content
	Instructions  string   // agent-specific instructions
	ToolNames     []string // registered tool names
	MCPContext    string   // context strings from MCP servers
	SkillNames    []string // available skill names
	CWD           string   // current working directory
	ModelName     string   // model identifier
	ContextWindow int      // context window size in tokens (0 = omit)
	MemoryDir     string   // directory for persistent memory writes
}

// Parts separates the system prompt into cacheable and volatile sections.
type Parts struct {
	System          string // static: persona + tools + skills
	VolatileContext string // per-turn: memory, instructions, date, CWD, MCP
}

// Build assembles the system prompt from the given options.
func Build(opts Options) Parts {
	return Parts{
		System:          buildSystem(opts),
		VolatileContext: buildVolatile(opts),
	}
}

func buildSystem(opts Options) string {
	var sb strings.Builder

	// 1. Base prompt
	sb.WriteString(opts.BasePrompt)

	// 2. Available tools
	sb.WriteString("\n\n## Available Tools\n")
	if len(opts.ToolNames) > 0 {
		sb.WriteString("You have these tools: ")
		sb.WriteString(strings.Join(opts.ToolNames, ", "))
		sb.WriteString(".")
	} else {
		sb.WriteString("No tools are available.")
	}

	// 3. Available skills
	if len(opts.SkillNames) > 0 {
		sb.WriteString("\n\n## Available Skills\n")
		sb.WriteString("You can activate a skill by calling `use_skill` with the skill name.\n")
		sb.WriteString("Only activate a skill when relevant to the user's request.\n\n")
		for _, name := range opts.SkillNames {
			sb.WriteString(fmt.Sprintf("- %s\n", name))
		}
	}

	// 4. Memory persistence guidance
	if opts.MemoryDir != "" {
		sb.WriteString("\n\n## Memory Persistence\n")
		sb.WriteString("Your memory is shown in the context below.\n")
		sb.WriteString("When you discover something worth remembering across sessions,\n")
		sb.WriteString("use `memory_append` to add it to persistent memory.\n")
		sb.WriteString("Do NOT use file_write on MEMORY.md directly.\n")
	}

	return sb.String()
}

func buildVolatile(opts Options) string {
	var sb strings.Builder

	// Date/time + CWD + model
	sb.WriteString("## Context\n")
	sb.WriteString("Current date: " + time.Now().Format("2006-01-02 15:04 MST"))
	if opts.CWD != "" {
		sb.WriteString("\nWorking directory: " + opts.CWD)
	}
	if opts.ModelName != "" {
		sb.WriteString("\nModel: " + opts.ModelName)
	}
	if opts.ContextWindow > 0 {
		sb.WriteString(fmt.Sprintf("\nContext window: %d tokens", opts.ContextWindow))
	}

	// Output format
	sb.WriteString("\n\n## Output Format\n")
	sb.WriteString("Format responses using GitHub-flavored markdown (GFM).")

	// Memory
	if mem := strings.TrimSpace(opts.Memory); mem != "" {
		sb.WriteString("\n\n## Memory\n")
		sb.WriteString(truncate(mem, maxMemoryChars))
	}

	// Instructions
	if inst := strings.TrimSpace(opts.Instructions); inst != "" {
		sb.WriteString("\n\n## Instructions\n")
		sb.WriteString(truncate(inst, maxInstructionsChars))
	}

	// MCP context
	if mcp := strings.TrimSpace(opts.MCPContext); mcp != "" {
		sb.WriteString("\n\n## MCP Server Context\n")
		sb.WriteString(mcp)
	}

	return sb.String()
}

func truncate(s string, maxChars int) string {
	r := []rune(s)
	if len(r) <= maxChars {
		return s
	}
	return string(r[:maxChars]) + "\n[truncated]"
}
