package daemon

import "testing"

func TestReplyRouteIndexPutGetUpdateEvict(t *testing.T) {
	idx := NewReplyRouteIndex(2)
	idx.Put("m1", "route-a")
	idx.Put("m2", "route-b")
	if got := idx.Get("m1"); got != "route-a" {
		t.Fatalf("m1 route = %q", got)
	}
	idx.Put("m1", "route-a2")
	if got := idx.Get("m1"); got != "route-a2" {
		t.Fatalf("updated m1 route = %q", got)
	}
	idx.Put("m3", "route-c")
	if got := idx.Get("m2"); got != "" {
		t.Fatalf("m2 should be evicted, got %q", got)
	}
	if got := idx.Get("m1"); got != "route-a2" {
		t.Fatalf("m1 should remain after update move, got %q", got)
	}
}
