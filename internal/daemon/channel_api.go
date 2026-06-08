package daemon

import (
	"net/http"
	"strings"
)

func (s *Server) handleGetChannelRoute(w http.ResponseWriter, r *http.Request) {
	messageID := strings.TrimSpace(r.PathValue("message_id"))
	routeKey := ""
	if s.replyRoutes != nil {
		routeKey = s.replyRoutes.Get(messageID)
	}
	if routeKey == "" {
		writeError(w, http.StatusNotFound, "route not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"message_id": messageID,
		"route_key":  routeKey,
	})
}

func (s *Server) handleGetChannelState(w http.ResponseWriter, r *http.Request) {
	platform := strings.TrimSpace(r.URL.Query().Get("platform"))
	channelID := strings.TrimSpace(r.URL.Query().Get("channel_id"))
	channelLine := ""
	platformLine := ""
	preamble := []string(nil)
	if s.connectionState != nil {
		channelLine = s.connectionState.ChannelLine(platform, channelID)
		platformLine = s.connectionState.PlatformLine(platform)
		preamble = s.connectionState.Preamble()
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"platform":      platform,
		"channel_id":    channelID,
		"channel_line":  channelLine,
		"platform_line": platformLine,
		"preamble":      preamble,
	})
}

func (s *Server) handleListChannelAdapters(w http.ResponseWriter, r *http.Request) {
	adapters := []ChannelAdapterMetadata{}
	if s.channelAdapters != nil {
		adapters = s.channelAdapters.ListMetadata()
	}
	writeJSON(w, http.StatusOK, map[string]any{"adapters": adapters})
}
