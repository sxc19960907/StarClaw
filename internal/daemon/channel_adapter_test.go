package daemon

import (
	"context"
	"testing"
)

func TestChannelAdapterRegistryRegisterListGet(t *testing.T) {
	registry := NewChannelAdapterRegistry()
	if err := registry.Register(NewFakeChannelAdapter(ChannelAdapterMetadata{
		Provider:     "slack",
		DisplayName:  "Slack",
		Capabilities: []string{"install"},
	})); err != nil {
		t.Fatalf("register slack: %v", err)
	}
	if err := registry.Register(NewFakeChannelAdapter(ChannelAdapterMetadata{Provider: "feishu", DisplayName: "Feishu"})); err != nil {
		t.Fatalf("register feishu: %v", err)
	}
	list := registry.ListMetadata()
	if len(list) != 2 {
		t.Fatalf("metadata len = %d, want 2", len(list))
	}
	if list[0].Provider != "feishu" || list[1].Provider != "slack" {
		t.Fatalf("metadata not sorted: %#v", list)
	}
	list[1].Capabilities[0] = "mutated"
	again := registry.ListMetadata()
	if again[1].Capabilities[0] != "install" {
		t.Fatalf("metadata was not copied defensively: %#v", again[1])
	}
	if _, ok := registry.Get("slack"); !ok {
		t.Fatal("registered adapter missing")
	}
	if _, ok := registry.Get("missing"); ok {
		t.Fatal("unexpected missing adapter")
	}
}

func TestChannelAdapterRegistryRejectsInvalid(t *testing.T) {
	registry := NewChannelAdapterRegistry()
	if err := registry.Register(nil); err == nil {
		t.Fatal("expected nil adapter error")
	}
	if err := registry.Register(NewFakeChannelAdapter(ChannelAdapterMetadata{})); err == nil {
		t.Fatal("expected empty provider error")
	}
}

func TestFakeChannelAdapterInstallListDelete(t *testing.T) {
	adapter := NewFakeChannelAdapter(ChannelAdapterMetadata{Provider: "webhook", DisplayName: "Local webhook"})
	install, err := adapter.Install(context.Background(), ChannelInstallRequest{
		Agent:       "assistant",
		DisplayName: "Local bridge",
		Metadata:    map[string]string{"secret": "redacted"},
	})
	if err != nil {
		t.Fatalf("install: %v", err)
	}
	if install.Provider != "webhook" || install.ID == "" || install.CreatedAt.IsZero() {
		t.Fatalf("install = %#v", install)
	}
	install.Metadata["secret"] = "mutated"
	installs, err := adapter.ListInstalls(context.Background())
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(installs) != 1 || installs[0].Metadata["secret"] != "redacted" {
		t.Fatalf("installs = %#v", installs)
	}
	if err := adapter.DeleteInstall(context.Background(), installs[0].ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	installs, err = adapter.ListInstalls(context.Background())
	if err != nil {
		t.Fatalf("list after delete: %v", err)
	}
	if len(installs) != 0 {
		t.Fatalf("installs after delete = %#v", installs)
	}
	if err := adapter.DeleteInstall(context.Background(), "missing"); err == nil {
		t.Fatal("expected missing delete error")
	}
}
