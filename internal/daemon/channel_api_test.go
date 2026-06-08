package daemon

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestChannelRouteAPIFromQueueCreate(t *testing.T) {
	s := newTestServer(t, newTestServerDeps(t))
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	postJSON(t, ts.URL+"/queue", `{"route_key":"route-a","text":"hello","source":"slack","external_id":"msg-1"}`, http.StatusAccepted, &map[string]any{})

	var route struct {
		MessageID string `json:"message_id"`
		RouteKey  string `json:"route_key"`
	}
	getJSON(t, ts.URL+"/channel/routes/msg-1", http.StatusOK, &route)
	if route.MessageID != "msg-1" || route.RouteKey != "route-a" {
		t.Fatalf("route = %#v", route)
	}
	getJSON(t, ts.URL+"/channel/routes/missing", http.StatusNotFound, &map[string]string{})
}

func TestChannelStateAPI(t *testing.T) {
	s := newTestServer(t, newTestServerDeps(t))
	s.connectionState.Apply(ChannelStateEvent{Axis: ChannelAxisBinding, Platform: "slack", Change: ChannelChangeTokenRevoked}, nowForChannelAPITest())
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	var body struct {
		Platform     string   `json:"platform"`
		PlatformLine string   `json:"platform_line"`
		Preamble     []string `json:"preamble"`
	}
	getJSON(t, ts.URL+"/channel/state?platform=slack", http.StatusOK, &body)
	if body.Platform != "slack" {
		t.Fatalf("platform = %q", body.Platform)
	}
	if !strings.Contains(body.PlatformLine, "authorization token was revoked") {
		t.Fatalf("platform line = %q", body.PlatformLine)
	}
	if len(body.Preamble) != 1 {
		t.Fatalf("preamble = %#v", body.Preamble)
	}
}

func TestChannelAdaptersAPI(t *testing.T) {
	s := newTestServer(t, newTestServerDeps(t))
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	var body struct {
		Adapters []ChannelAdapterMetadata `json:"adapters"`
	}
	getJSON(t, ts.URL+"/channel/adapters", http.StatusOK, &body)
	if len(body.Adapters) == 0 {
		t.Fatal("expected adapters")
	}
	var feishu *ChannelAdapterMetadata
	for i := range body.Adapters {
		if body.Adapters[i].Provider == "feishu" {
			feishu = &body.Adapters[i]
			break
		}
	}
	if feishu == nil {
		t.Fatalf("feishu adapter missing: %#v", body.Adapters)
	}
	if feishu.Enabled || feishu.Configured {
		t.Fatalf("feishu should be disabled local contract: %#v", *feishu)
	}
	if !strings.Contains(feishu.PrivacyNote, "no external") {
		t.Fatalf("privacy note = %q", feishu.PrivacyNote)
	}
}

func nowForChannelAPITest() time.Time {
	return time.Date(2026, 6, 8, 10, 0, 0, 0, time.UTC)
}
