package agent

import (
	"sort"

	"github.com/starclaw/starclaw/internal/client"
)

// ToolRegistry manages tool registration and lookup
type ToolRegistry struct {
	tools map[string]Tool
	order []string
}

// NewToolRegistry creates a new registry
func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools: make(map[string]Tool),
	}
}

// Register adds a tool to the registry
func (r *ToolRegistry) Register(t Tool) {
	name := t.Info().Name
	if _, exists := r.tools[name]; !exists {
		r.order = append(r.order, name)
	}
	r.tools[name] = t
}

// Get retrieves a tool by name
func (r *ToolRegistry) Get(name string) (Tool, bool) {
	t, ok := r.tools[name]
	return t, ok
}

// List returns all registered tools
func (r *ToolRegistry) List() []Tool {
	sort.Strings(r.order)
	result := make([]Tool, 0, len(r.order))
	for _, name := range r.order {
		if t, ok := r.tools[name]; ok {
			result = append(result, t)
		}
	}
	return result
}

// Names returns all registered tool names
func (r *ToolRegistry) Names() []string {
	sort.Strings(r.order)
	names := make([]string, len(r.order))
	copy(names, r.order)
	return names
}

// Count returns number of registered tools.
func (r *ToolRegistry) Count() int {
	return len(r.tools)
}

// Len returns the number of registered tools (alias for Count).
func (r *ToolRegistry) Len() int {
	return len(r.tools)
}

// All returns all registered tools in registration order.
func (r *ToolRegistry) All() []Tool {
	tools := make([]Tool, 0, len(r.order))
	for _, name := range r.order {
		tools = append(tools, r.tools[name])
	}
	return tools
}

// Schemas returns complete client.Tool schemas for all registered tools.
func (r *ToolRegistry) Schemas() []client.Tool {
	schemas := make([]client.Tool, 0, len(r.order))
	for _, name := range r.order {
		schemas = append(schemas, buildToolSchema(r.tools[name]))
	}
	return schemas
}

// SortedSchemas returns tool schemas in deterministic order:
// local tools (alpha) → MCP tools (alpha) → gateway tools (alpha).
func (r *ToolRegistry) SortedSchemas() []client.Tool {
	local, mcp, gw := r.partitionBySource()
	sort.Strings(local)
	sort.Strings(mcp)
	sort.Strings(gw)

	schemas := make([]client.Tool, 0, len(r.order))
	for _, group := range [][]string{local, mcp, gw} {
		for _, name := range group {
			schemas = append(schemas, buildToolSchema(r.tools[name]))
		}
	}
	return schemas
}

// SortedNames returns tool names in the same deterministic order as SortedSchemas.
func (r *ToolRegistry) SortedNames() []string {
	local, mcp, gw := r.partitionBySource()
	sort.Strings(local)
	sort.Strings(mcp)
	sort.Strings(gw)

	names := make([]string, 0, len(r.order))
	names = append(names, local...)
	names = append(names, mcp...)
	names = append(names, gw...)
	return names
}

// SummaryList returns name+description for all registered tools.
func (r *ToolRegistry) SummaryList() []ToolSummary {
	summaries := make([]ToolSummary, 0, len(r.order))
	for _, name := range r.order {
		info := r.tools[name].Info()
		summaries = append(summaries, ToolSummary{Name: info.Name, Description: info.Description})
	}
	return summaries
}

// FullSchemas returns complete client.Tool schemas for the named tools.
func (r *ToolRegistry) FullSchemas(names []string) []client.Tool {
	schemas := make([]client.Tool, 0, len(names))
	for _, name := range names {
		if t, ok := r.tools[name]; ok {
			schemas = append(schemas, buildToolSchema(t))
		}
	}
	return schemas
}

// partitionBySource groups tool names by their source category.
func (r *ToolRegistry) partitionBySource() (local, mcp, gw []string) {
	for _, name := range r.order {
		t := r.tools[name]
		if sourcer, ok := t.(ToolSourcer); ok {
			switch sourcer.ToolSource() {
			case SourceMCP:
				mcp = append(mcp, name)
			case SourceGateway:
				gw = append(gw, name)
			default:
				local = append(local, name)
			}
		} else {
			local = append(local, name)
		}
	}
	return
}

// buildToolSchema converts a Tool into a client.Tool schema definition.
func buildToolSchema(t Tool) client.Tool {
	info := t.Info()
	params := info.Parameters
	if params == nil {
		params = map[string]any{"type": "object", "properties": map[string]any{}}
	}
	if info.Required != nil {
		params["required"] = info.Required
	}
	return client.Tool{
		Type: "function",
		Function: client.FunctionDef{
			Name:        info.Name,
			Description: info.Description,
			Parameters:  params,
		},
	}
}

// Clone creates a copy of the registry
func (r *ToolRegistry) Clone() *ToolRegistry {
	clone := NewToolRegistry()
	for name, tool := range r.tools {
		clone.tools[name] = tool
		clone.order = append(clone.order, name)
	}
	return clone
}

// Remove removes a tool from the registry
func (r *ToolRegistry) Remove(name string) {
	delete(r.tools, name)
	for i, n := range r.order {
		if n == name {
			r.order = append(r.order[:i], r.order[i+1:]...)
			break
		}
	}
}

// FilterByAllow returns a new registry with only allowed tools
func (r *ToolRegistry) FilterByAllow(allowed []string) *ToolRegistry {
	allowedSet := make(map[string]bool)
	for _, name := range allowed {
		allowedSet[name] = true
	}

	filtered := NewToolRegistry()
	for name, tool := range r.tools {
		if allowedSet[name] {
			filtered.tools[name] = tool
			filtered.order = append(filtered.order, name)
		}
	}
	sort.Strings(filtered.order)
	return filtered
}

// FilterByDeny returns a new registry without denied tools
func (r *ToolRegistry) FilterByDeny(denied []string) *ToolRegistry {
	deniedSet := make(map[string]bool)
	for _, name := range denied {
		deniedSet[name] = true
	}

	filtered := NewToolRegistry()
	for name, tool := range r.tools {
		if !deniedSet[name] {
			filtered.tools[name] = tool
			filtered.order = append(filtered.order, name)
		}
	}
	sort.Strings(filtered.order)
	return filtered
}
