package cloudflow

import "testing"

func TestParseSlash(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		wantType string
		wantCmd  string
		wantStr  string
		wantQ    string
	}{
		{name: "ordinary text", text: "research this"},
		{name: "no args", text: "/research"},
		{name: "research default", text: "/research compare Kocoro", wantType: TypeResearch, wantCmd: "/research", wantStr: "standard", wantQ: "compare Kocoro"},
		{name: "research strategy", text: "/research deep compare Kocoro", wantType: TypeResearch, wantCmd: "/research", wantStr: "deep", wantQ: "compare Kocoro"},
		{name: "strategy without query", text: "/research deep"},
		{name: "swarm", text: "/swarm ship phase seven", wantType: TypeSwarm, wantCmd: "/swarm", wantQ: "ship phase seven"},
		{name: "dag", text: "/dag coordinate this", wantType: TypeAuto, wantCmd: "/dag", wantQ: "coordinate this"},
		{name: "case insensitive", text: "/DAG coordinate this", wantType: TypeAuto, wantCmd: "/dag", wantQ: "coordinate this"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got := ParseSlash(tt.text)
			if tt.wantType == "" {
				if got != nil {
					t.Fatalf("got %#v, want nil", got)
				}
				return
			}
			if got == nil {
				t.Fatal("got nil")
			}
			if got.Type != tt.wantType || got.Command != tt.wantCmd || got.Strategy != tt.wantStr || got.Query != tt.wantQ {
				t.Fatalf("got %#v", got)
			}
		})
	}
}
