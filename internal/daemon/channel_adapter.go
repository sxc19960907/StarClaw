package daemon

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

type ChannelAdapter interface {
	Metadata() ChannelAdapterMetadata
	Install(context.Context, ChannelInstallRequest) (ChannelInstall, error)
	ListInstalls(context.Context) ([]ChannelInstall, error)
	DeleteInstall(context.Context, string) error
}

type ChannelAdapterMetadata struct {
	Provider     string   `json:"provider"`
	DisplayName  string   `json:"display_name"`
	Kind         string   `json:"kind"`
	Configured   bool     `json:"configured"`
	Enabled      bool     `json:"enabled"`
	Capabilities []string `json:"capabilities"`
	PrivacyNote  string   `json:"privacy_note"`
}

type ChannelInstallRequest struct {
	Provider    string            `json:"provider,omitempty"`
	Agent       string            `json:"agent,omitempty"`
	DisplayName string            `json:"display_name,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

type ChannelInstall struct {
	ID          string            `json:"id"`
	Provider    string            `json:"provider"`
	Agent       string            `json:"agent,omitempty"`
	DisplayName string            `json:"display_name,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
}

type ChannelAdapterRegistry struct {
	mu       sync.RWMutex
	adapters map[string]ChannelAdapter
}

func NewChannelAdapterRegistry() *ChannelAdapterRegistry {
	return &ChannelAdapterRegistry{adapters: make(map[string]ChannelAdapter)}
}

func (r *ChannelAdapterRegistry) Register(adapter ChannelAdapter) error {
	if r == nil {
		return fmt.Errorf("channel adapter registry unavailable")
	}
	if adapter == nil {
		return fmt.Errorf("channel adapter is nil")
	}
	meta := adapter.Metadata()
	provider := strings.TrimSpace(meta.Provider)
	if provider == "" {
		return fmt.Errorf("channel adapter provider is required")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.adapters[provider] = adapter
	return nil
}

func (r *ChannelAdapterRegistry) Get(provider string) (ChannelAdapter, bool) {
	if r == nil {
		return nil, false
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	adapter, ok := r.adapters[strings.TrimSpace(provider)]
	return adapter, ok
}

func (r *ChannelAdapterRegistry) ListMetadata() []ChannelAdapterMetadata {
	if r == nil {
		return nil
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	providers := make([]string, 0, len(r.adapters))
	for provider := range r.adapters {
		providers = append(providers, provider)
	}
	sort.Strings(providers)
	out := make([]ChannelAdapterMetadata, 0, len(providers))
	for _, provider := range providers {
		out = append(out, cloneChannelAdapterMetadata(r.adapters[provider].Metadata()))
	}
	return out
}

type fakeChannelAdapter struct {
	mu       sync.RWMutex
	meta     ChannelAdapterMetadata
	installs map[string]ChannelInstall
	next     int
}

func NewFakeChannelAdapter(meta ChannelAdapterMetadata) *fakeChannelAdapter {
	meta.Provider = strings.TrimSpace(meta.Provider)
	meta.Capabilities = cloneStringSlice(meta.Capabilities)
	return &fakeChannelAdapter{
		meta:     meta,
		installs: make(map[string]ChannelInstall),
	}
}

func (a *fakeChannelAdapter) Metadata() ChannelAdapterMetadata {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return cloneChannelAdapterMetadata(a.meta)
}

func (a *fakeChannelAdapter) Install(_ context.Context, req ChannelInstallRequest) (ChannelInstall, error) {
	if a == nil {
		return ChannelInstall{}, fmt.Errorf("channel adapter unavailable")
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.next++
	provider := a.meta.Provider
	if strings.TrimSpace(req.Provider) != "" {
		provider = strings.TrimSpace(req.Provider)
	}
	install := ChannelInstall{
		ID:          fmt.Sprintf("%s-install-%d", provider, a.next),
		Provider:    provider,
		Agent:       strings.TrimSpace(req.Agent),
		DisplayName: strings.TrimSpace(req.DisplayName),
		Metadata:    cloneStringMap(req.Metadata),
		CreatedAt:   time.Now().UTC(),
	}
	a.installs[install.ID] = install
	return cloneChannelInstall(install), nil
}

func (a *fakeChannelAdapter) ListInstalls(context.Context) ([]ChannelInstall, error) {
	if a == nil {
		return nil, fmt.Errorf("channel adapter unavailable")
	}
	a.mu.RLock()
	defer a.mu.RUnlock()
	ids := make([]string, 0, len(a.installs))
	for id := range a.installs {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	out := make([]ChannelInstall, 0, len(ids))
	for _, id := range ids {
		out = append(out, cloneChannelInstall(a.installs[id]))
	}
	return out, nil
}

func (a *fakeChannelAdapter) DeleteInstall(_ context.Context, id string) error {
	if a == nil {
		return fmt.Errorf("channel adapter unavailable")
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return fmt.Errorf("install id is required")
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if _, ok := a.installs[id]; !ok {
		return fmt.Errorf("channel install %q not found", id)
	}
	delete(a.installs, id)
	return nil
}

func newDefaultChannelAdapterRegistry() *ChannelAdapterRegistry {
	registry := NewChannelAdapterRegistry()
	for _, meta := range []ChannelAdapterMetadata{
		{
			Provider:     "feishu",
			DisplayName:  "Feishu/Lark",
			Kind:         "external",
			Configured:   false,
			Enabled:      false,
			Capabilities: []string{"install", "list", "delete"},
			PrivacyNote:  "Disabled local contract; no external Feishu/Lark transport is active.",
		},
		{
			Provider:     "slack",
			DisplayName:  "Slack",
			Kind:         "external",
			Configured:   false,
			Enabled:      false,
			Capabilities: []string{"install", "list", "delete"},
			PrivacyNote:  "Disabled local contract; no external Slack transport is active.",
		},
		{
			Provider:     "telegram",
			DisplayName:  "Telegram",
			Kind:         "external",
			Configured:   false,
			Enabled:      false,
			Capabilities: []string{"install", "list", "delete"},
			PrivacyNote:  "Disabled local contract; no external Telegram transport is active.",
		},
		{
			Provider:     "webhook",
			DisplayName:  "Local webhook",
			Kind:         "local",
			Configured:   true,
			Enabled:      true,
			Capabilities: []string{"install", "list", "delete"},
			PrivacyNote:  "Local fake adapter only; no off-machine transport is active.",
		},
	} {
		_ = registry.Register(NewFakeChannelAdapter(meta))
	}
	return registry
}

func cloneChannelAdapterMetadata(meta ChannelAdapterMetadata) ChannelAdapterMetadata {
	meta.Capabilities = cloneStringSlice(meta.Capabilities)
	return meta
}

func cloneChannelInstall(install ChannelInstall) ChannelInstall {
	install.Metadata = cloneStringMap(install.Metadata)
	return install
}

func cloneStringSlice(values []string) []string {
	if values == nil {
		return nil
	}
	out := make([]string, len(values))
	copy(out, values)
	return out
}

func cloneStringMap(values map[string]string) map[string]string {
	if values == nil {
		return nil
	}
	out := make(map[string]string, len(values))
	for key, value := range values {
		out[key] = value
	}
	return out
}
