package daemon

import (
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	ChannelAxisTransport  = "transport"
	ChannelAxisBinding    = "binding"
	ChannelAxisMembership = "membership"

	ChannelChangeJoin           = "join"
	ChannelChangeLeave          = "leave"
	ChannelChangeKicked         = "kicked"
	ChannelChangeBan            = "ban"
	ChannelChangeInstallRevoked = "install_revoked"
	ChannelChangeTokenRevoked   = "token_revoked"
	ChannelChangeDisconnected   = "disconnected"
)

type ChannelStateEvent struct {
	Axis      string `json:"axis"`
	Platform  string `json:"platform"`
	ChannelID string `json:"channel_id,omitempty"`
	Change    string `json:"change"`
}

type channelState struct {
	Change string    `json:"change"`
	At     time.Time `json:"at"`
}

type ConnectionStateCache struct {
	mu        sync.RWMutex
	channels  map[string]channelState
	binding   map[string]channelState
	transport map[string]channelState
}

func NewConnectionStateCache() *ConnectionStateCache {
	return &ConnectionStateCache{
		channels:  map[string]channelState{},
		binding:   map[string]channelState{},
		transport: map[string]channelState{},
	}
}

func (c *ConnectionStateCache) Apply(evt ChannelStateEvent, now time.Time) {
	if c == nil || evt.Platform == "" || evt.Change == "" {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	state := channelState{Change: evt.Change, At: now.UTC()}
	switch evt.Axis {
	case ChannelAxisMembership:
		if evt.ChannelID != "" {
			c.channels[channelStateKey(evt.Platform, evt.ChannelID)] = state
		}
	case ChannelAxisBinding:
		c.binding[evt.Platform] = state
	case ChannelAxisTransport:
		c.transport[evt.Platform] = state
	}
}

func (c *ConnectionStateCache) MarkChannelHealthy(platform, channelID string) {
	if c == nil || platform == "" || channelID == "" {
		return
	}
	c.mu.Lock()
	delete(c.channels, channelStateKey(platform, channelID))
	c.mu.Unlock()
}

func (c *ConnectionStateCache) MarkPlatformHealthy(platform string) {
	if c == nil || platform == "" {
		return
	}
	c.mu.Lock()
	delete(c.binding, platform)
	delete(c.transport, platform)
	c.mu.Unlock()
}

func (c *ConnectionStateCache) ChannelLine(platform, channelID string) string {
	if c == nil || platform == "" || channelID == "" {
		return ""
	}
	c.mu.RLock()
	st, ok := c.channels[channelStateKey(platform, channelID)]
	c.mu.RUnlock()
	if !ok {
		return ""
	}
	return changePhrase(platform, st.Change)
}

func (c *ConnectionStateCache) PlatformLine(platform string) string {
	if c == nil || platform == "" {
		return ""
	}
	c.mu.RLock()
	st, ok := c.binding[platform]
	if !ok {
		st, ok = c.transport[platform]
	}
	c.mu.RUnlock()
	if !ok {
		return ""
	}
	return changePhrase(platform, st.Change)
}

func (c *ConnectionStateCache) Preamble() []string {
	if c == nil {
		return nil
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	states := make(map[string]channelState, len(c.binding)+len(c.transport))
	for platform, st := range c.transport {
		states[platform] = st
	}
	for platform, st := range c.binding {
		states[platform] = st
	}
	platforms := make([]string, 0, len(states))
	for platform := range states {
		if phrase := changePhrase(platform, states[platform].Change); phrase != "" {
			platforms = append(platforms, platform)
		}
	}
	sort.Strings(platforms)
	out := make([]string, 0, len(platforms))
	for _, platform := range platforms {
		out = append(out, title(platform)+": "+changePhrase(platform, states[platform].Change))
	}
	return out
}

func changePhrase(platform, change string) string {
	switch change {
	case ChannelChangeKicked, ChannelChangeLeave, ChannelChangeBan:
		return "the bot was removed from this channel and can no longer read or post here until re-added"
	case ChannelChangeInstallRevoked:
		return title(platform) + " app authorization was revoked; the bot cannot send or receive until re-installed"
	case ChannelChangeTokenRevoked:
		return title(platform) + " authorization token was revoked; re-authorize to restore"
	case ChannelChangeDisconnected:
		return title(platform) + " connection dropped; reconnecting"
	default:
		return ""
	}
}

func channelStateKey(platform, channelID string) string {
	return platform + ":" + channelID
}

func title(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
