package agent

import (
	"testing"
	"time"
)

func TestSuggestionStateSetGetClearAccept(t *testing.T) {
	state := NewSuggestionState()
	at := time.Date(2026, 6, 8, 12, 0, 0, 0, time.UTC)

	if _, ok := state.Get("s1"); ok {
		t.Fatal("unexpected suggestion before set")
	}
	state.Set("s1", "continue implementation", at)
	got, ok := state.Get("s1")
	if !ok {
		t.Fatal("suggestion missing after set")
	}
	if got.Text != "continue implementation" || !got.SuggestedAt.Equal(at) {
		t.Fatalf("suggestion = %#v", got)
	}
	got.Text = "mutated"
	gotAgain, _ := state.Get("s1")
	if gotAgain.Text != "continue implementation" {
		t.Fatalf("Get returned mutable state: %#v", gotAgain)
	}
	if !state.MarkAccepted("s1") {
		t.Fatal("MarkAccepted returned false")
	}
	accepted, _ := state.Get("s1")
	if accepted.AcceptedAt == nil {
		t.Fatal("accepted timestamp missing")
	}
	state.Clear("s1")
	if _, ok := state.Get("s1"); ok {
		t.Fatal("suggestion still present after clear")
	}
	if state.MarkAccepted("s1") {
		t.Fatal("accepted missing suggestion")
	}
}

func TestSuggestionStateNilAndEmptyNoop(t *testing.T) {
	var state *SuggestionState
	state.Set("s1", "ignored", time.Now())
	if _, ok := state.Get("s1"); ok {
		t.Fatal("nil state returned suggestion")
	}
	state.Clear("s1")
	if state.MarkAccepted("s1") {
		t.Fatal("nil state accepted suggestion")
	}

	realState := NewSuggestionState()
	realState.Set("", "ignored", time.Now())
	if _, ok := realState.Get(""); ok {
		t.Fatal("empty session suggestion stored")
	}
}
