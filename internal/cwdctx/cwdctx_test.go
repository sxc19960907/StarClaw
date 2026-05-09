package cwdctx

import (
	"testing"
)

func TestNew(t *testing.T) {
	c := New()
	if c == nil {
		t.Fatal("New() returned nil")
	}
	if c.Get() != "" {
		t.Errorf("initial Get() = %q, want empty string", c.Get())
	}
}

func TestSetAndGet(t *testing.T) {
	c := New()
	c.Set("/home/user/project")
	if got := c.Get(); got != "/home/user/project" {
		t.Errorf("Get() = %q, want %q", got, "/home/user/project")
	}
}

func TestSet_Override(t *testing.T) {
	c := New()
	c.Set("/first/path")
	c.Set("/second/path")
	if got := c.Get(); got != "/second/path" {
		t.Errorf("Get() after override = %q, want %q", got, "/second/path")
	}
}

func TestSet_Empty(t *testing.T) {
	c := New()
	c.Set("/some/path")
	c.Set("")
	if got := c.Get(); got != "" {
		t.Errorf("Get() after empty Set = %q, want empty", got)
	}
}

func TestMultipleInstances(t *testing.T) {
	c1 := New()
	c2 := New()

	c1.Set("/path/one")
	c2.Set("/path/two")

	if c1.Get() != "/path/one" {
		t.Errorf("c1.Get() = %q, want %q", c1.Get(), "/path/one")
	}
	if c2.Get() != "/path/two" {
		t.Errorf("c2.Get() = %q, want %q", c2.Get(), "/path/two")
	}
}
