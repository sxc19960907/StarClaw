package cloudflow

import "testing"

func TestCloudStatusLine(t *testing.T) {
	if got := CloudStatusLine("researcher", "started", "Searching"); got != "[researcher] Searching" {
		t.Fatalf("got %q", got)
	}
	if got := CloudStatusLine("orchestrator", "started", ""); got != "Agent working..." {
		t.Fatalf("got %q", got)
	}
	if got := CloudStatusLine("", "tool", ""); got != "Calling tool..." {
		t.Fatalf("got %q", got)
	}
}
