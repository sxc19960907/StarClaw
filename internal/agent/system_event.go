package agent

import (
	"strings"
	"time"
)

// SystemEvent is a daemon-authored runtime signal that can be scoped to a
// route and surfaced to the model on a later turn.
type SystemEvent struct {
	Text       string
	ContextKey string
	Trusted    bool
	TS         time.Time
}

var systemEventReplacer = strings.NewReplacer(
	"\r\n", " ",
	"\r", " ",
	"\n", " ",
	"[", "(",
	"]", ")",
	"<", "(",
	">", ")",
)

// SanitizeSystemEventText neutralizes framing-sensitive characters and
// collapses whitespace before an event is rendered.
func SanitizeSystemEventText(s string) string {
	s = systemEventReplacer.Replace(s)
	return strings.Join(strings.Fields(s), " ")
}

// FormatSystemEventBlock renders events as a single reminder block. Empty or
// fully sanitized-away events are skipped.
func FormatSystemEventBlock(events []SystemEvent) string {
	if len(events) == 0 {
		return ""
	}
	lines := make([]string, 0, len(events))
	for _, ev := range events {
		text := SanitizeSystemEventText(ev.Text)
		if text == "" {
			continue
		}
		prefix := "System"
		if !ev.Trusted {
			prefix = "System (untrusted)"
		}
		ts := ev.TS
		if ts.IsZero() {
			ts = time.Unix(0, 0).UTC()
		}
		lines = append(lines, prefix+": ["+ts.Format("15:04:05")+"] "+text)
	}
	if len(lines) == 0 {
		return ""
	}
	return "<system-reminder>\n" + strings.Join(lines, "\n") + "\n</system-reminder>"
}
