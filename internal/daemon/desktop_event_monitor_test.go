package daemon

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/starclaw/starclaw/internal/daemon/desktop_rpc"
)

func TestDesktopEventMonitorStatusAndRetention(t *testing.T) {
	monitor := NewDesktopEventMonitor(2)
	monitor.Record(&desktop_rpc.DesktopEvent{Event: desktop_rpc.EventDesktopOnline, Data: json.RawMessage(`{"secret":"hidden"}`), TS: "2026-06-09T01:00:00Z"})
	monitor.Record(&desktop_rpc.DesktopEvent{Event: desktop_rpc.EventCalendarDataChanged, TS: "2026-06-09T01:01:00Z"})
	monitor.Record(&desktop_rpc.DesktopEvent{Event: desktop_rpc.EventDesktopOffline, TS: "2026-06-09T01:02:00Z"})

	status := monitor.Status()
	if status.Retained != 2 {
		t.Fatalf("retained = %d, want 2", status.Retained)
	}
	if status.LastEvent != desktop_rpc.EventDesktopOffline {
		t.Fatalf("last event = %q", status.LastEvent)
	}
	if status.LastTS != "2026-06-09T01:02:00Z" {
		t.Fatalf("last ts = %q", status.LastTS)
	}
	encoded, err := json.Marshal(status)
	if err != nil {
		t.Fatalf("marshal status: %v", err)
	}
	if string(encoded) == "" || json.Valid(encoded) != true {
		t.Fatalf("invalid status json: %s", encoded)
	}
	if got := string(encoded); strings.Contains(got, "secret") || strings.Contains(got, "hidden") {
		t.Fatalf("status exposed raw event payload: %s", got)
	}
}

func TestDesktopEventMonitorNormalizesMissingTimestamp(t *testing.T) {
	monitor := NewDesktopEventMonitor(1)
	monitor.Record(&desktop_rpc.DesktopEvent{Event: desktop_rpc.EventDesktopOnline})
	status := monitor.Status()
	if status.LastTS == "" {
		t.Fatal("missing normalized timestamp")
	}
}
