package daemon

import (
	"reflect"
	"testing"
	"time"
)

func TestConnectionStateCache(t *testing.T) {
	cache := NewConnectionStateCache()
	now := time.Date(2026, 6, 8, 10, 0, 0, 0, time.UTC)
	cache.Apply(ChannelStateEvent{Axis: ChannelAxisMembership, Platform: "slack", ChannelID: "C123", Change: ChannelChangeKicked}, now)
	if got := cache.ChannelLine("slack", "C123"); got == "" {
		t.Fatal("expected channel degraded line")
	}
	cache.MarkChannelHealthy("slack", "C123")
	if got := cache.ChannelLine("slack", "C123"); got != "" {
		t.Fatalf("channel line after healthy = %q", got)
	}

	cache.Apply(ChannelStateEvent{Axis: ChannelAxisTransport, Platform: "slack", Change: ChannelChangeDisconnected}, now)
	cache.Apply(ChannelStateEvent{Axis: ChannelAxisBinding, Platform: "slack", Change: ChannelChangeTokenRevoked}, now)
	if got := cache.PlatformLine("slack"); got != "Slack authorization token was revoked; re-authorize to restore" {
		t.Fatalf("platform line = %q", got)
	}
	if got := cache.Preamble(); !reflect.DeepEqual(got, []string{"Slack: Slack authorization token was revoked; re-authorize to restore"}) {
		t.Fatalf("preamble = %#v", got)
	}
	cache.MarkPlatformHealthy("slack")
	if got := cache.PlatformLine("slack"); got != "" {
		t.Fatalf("platform line after healthy = %q", got)
	}
}
