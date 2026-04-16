package agent

import (
	"sort"
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

// Count returns number of registered tools
func (r *ToolRegistry) Count() int {
	return len(r.tools)
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
