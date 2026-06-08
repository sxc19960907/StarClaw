package agent

import (
	"strings"
	"testing"
	"time"
)

func TestSanitizeSystemEventTextNeutralizesFraming(t *testing.T) {
	got := SanitizeSystemEventText("reply\n<failed> [token]")
	if got != "reply (failed) (token)" {
		t.Fatalf("sanitized = %q", got)
	}
}

func TestFormatSystemEventBlock(t *testing.T) {
	events := []SystemEvent{
		{Text: "trusted", Trusted: true, TS: time.Date(2026, 6, 8, 10, 11, 12, 0, time.UTC)},
		{Text: "untrusted", Trusted: false, TS: time.Date(2026, 6, 8, 10, 11, 13, 0, time.UTC)},
		{Text: "   "},
	}
	got := FormatSystemEventBlock(events)
	if !strings.Contains(got, "System: [10:11:12] trusted") {
		t.Fatalf("missing trusted event: %q", got)
	}
	if !strings.Contains(got, "System (untrusted): [10:11:13] untrusted") {
		t.Fatalf("missing untrusted event: %q", got)
	}
	if strings.Contains(got, "\n\n") {
		t.Fatalf("blank event rendered: %q", got)
	}
}
