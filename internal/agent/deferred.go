package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/starclaw/starclaw/internal/client"
)

// toolSearchTool is a meta-tool that loads full schemas for deferred tools on demand.
type toolSearchTool struct {
	registry *ToolRegistry
	deferred map[string]bool
}

// newToolSearchTool creates a tool_search scoped to the given deferred tool names.
func newToolSearchTool(reg *ToolRegistry, deferred map[string]bool) *toolSearchTool {
	return &toolSearchTool{registry: reg, deferred: deferred}
}

func (t *toolSearchTool) Info() ToolInfo {
	return ToolInfo{
		Name: "tool_search",
		Description: "Load deferred tool schemas so you can call them in this same request. " +
			`Use "select:name1,name2" for exact lookup or a keyword to search by name/description.`,
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"query": map[string]any{
					"type":        "string",
					"description": `Either "select:name1,name2" for exact match or a keyword to search deferred tools.`,
				},
			},
		},
		Required: []string{"query"},
	}
}

func (t *toolSearchTool) RequiresApproval() bool     { return false }
func (t *toolSearchTool) IsReadOnlyCall(string) bool { return true }

func (t *toolSearchTool) Run(_ context.Context, argsJSON string) (ToolResult, error) {
	var args struct {
		Query string `json:"query"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return ValidationError("invalid arguments: " + err.Error()), nil
	}
	if args.Query == "" {
		return ValidationError("query is required"), nil
	}

	var matched []string

	if strings.HasPrefix(args.Query, "select:") {
		names := strings.Split(strings.TrimPrefix(args.Query, "select:"), ",")
		for _, name := range names {
			name = strings.TrimSpace(name)
			if name != "" && t.deferred[name] {
				matched = append(matched, name)
			}
		}
	} else {
		query := strings.ToLower(args.Query)
		for name := range t.deferred {
			tool, ok := t.registry.Get(name)
			if !ok {
				continue
			}
			info := tool.Info()
			if strings.Contains(strings.ToLower(info.Name), query) ||
				strings.Contains(strings.ToLower(info.Description), query) {
				matched = append(matched, name)
			}
		}
		sort.Strings(matched)
	}

	var sb strings.Builder
	sb.WriteString("LOADED:")
	sb.WriteString(strings.Join(matched, ","))

	if len(matched) == 0 {
		sb.WriteString("\nNo matching deferred tools found.")
	} else {
		sb.WriteString("\nSchemas loaded. Call these tools now to continue the task.")
		schemas := t.registry.FullSchemas(matched)
		for i, s := range schemas {
			schemaJSON, _ := json.MarshalIndent(s, "", "  ")
			_, _ = fmt.Fprintf(&sb, "\n\n## %s\n%s", matched[i], string(schemaJSON))
		}
	}

	return ToolResult{Content: sb.String()}, nil
}

// DeferredToolNames returns the set of non-local tool names (MCP + gateway).
func DeferredToolNames(reg *ToolRegistry) map[string]bool {
	_, mcp, gw := reg.partitionBySource()
	names := make(map[string]bool, len(mcp)+len(gw))
	for _, n := range mcp {
		names[n] = true
	}
	for _, n := range gw {
		names[n] = true
	}
	return names
}

// DeferredToolSummaries returns sorted summaries for non-local tools.
func DeferredToolSummaries(reg *ToolRegistry) []ToolSummary {
	_, mcp, gw := reg.partitionBySource()
	all := append(mcp, gw...)
	sort.Strings(all)
	summaries := make([]ToolSummary, 0, len(all))
	for _, name := range all {
		if t, ok := reg.Get(name); ok {
			info := t.Info()
			summaries = append(summaries, ToolSummary{Name: info.Name, Description: info.Description})
		}
	}
	return summaries
}

// RebuildSchemas produces a deterministic tool schema list by iterating
// sorted names and including tools from base or loaded sets.
func RebuildSchemas(reg *ToolRegistry, baseNames map[string]bool, loaded map[string]client.Tool) []client.Tool {
	result := make([]client.Tool, 0, len(baseNames)+len(loaded))
	for _, name := range reg.SortedNames() {
		if baseNames[name] {
			if t, ok := reg.Get(name); ok {
				result = append(result, buildToolSchema(t))
			}
		} else if s, ok := loaded[name]; ok {
			result = append(result, s)
		}
	}
	return result
}

// BuildLocalOnlySchemas returns sorted schemas for local tools only.
func BuildLocalOnlySchemas(reg *ToolRegistry) []client.Tool {
	local, _, _ := reg.partitionBySource()
	sort.Strings(local)
	schemas := make([]client.Tool, 0, len(local))
	for _, name := range local {
		if t, ok := reg.Get(name); ok {
			schemas = append(schemas, buildToolSchema(t))
		}
	}
	return schemas
}
