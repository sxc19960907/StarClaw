package agent

import (
	"sync"
	"testing"
)

func TestNewWarmSet(t *testing.T) {
	reg := NewToolRegistry()
	ws := NewWarmSet(reg)
	if ws == nil {
		t.Fatal("NewWarmSet returned nil")
	}
	if ws.Len() != 0 {
		t.Errorf("expected empty warm set, got %d", ws.Len())
	}
}

func TestWarmSet_WarmAndGet(t *testing.T) {
	reg := NewToolRegistry()
	tool1 := TestTool("tool_one")
	tool2 := TestTool("tool_two")
	reg.Register(tool1)
	reg.Register(tool2)

	ws := NewWarmSet(reg)
	ws.Warm("tool_one")

	if tool := ws.Get("tool_one"); tool == nil {
		t.Error("expected tool_one to be warmed")
	} else if tool.Info().Name != "tool_one" {
		t.Errorf("got tool %q", tool.Info().Name)
	}

	if tool := ws.Get("tool_two"); tool != nil {
		t.Error("expected tool_two NOT to be warmed")
	}
}

func TestWarmSet_WarmMultiple(t *testing.T) {
	reg := NewToolRegistry()
	names := []string{"a", "b", "c"}
	for _, n := range names {
		reg.Register(TestTool(n))
	}
	ws := NewWarmSet(reg)
	ws.Warm(names...)

	if ws.Len() != 3 {
		t.Errorf("expected 3 warmed tools, got %d", ws.Len())
	}
	for _, n := range names {
		if !ws.Contains(n) {
			t.Errorf("expected %s to be contained", n)
		}
	}
}

func TestWarmSet_WarmAll(t *testing.T) {
	reg := NewToolRegistry()
	reg.Register(TestTool("a"))
	reg.Register(TestTool("b"))

	ws := NewWarmSet(reg)
	ws.WarmAll()

	if ws.Len() != 2 {
		t.Errorf("expected 2 warmed tools, got %d", ws.Len())
	}
}

func TestWarmSet_WarmUnknown(t *testing.T) {
	reg := NewToolRegistry()
	reg.Register(TestTool("known"))

	ws := NewWarmSet(reg)
	ws.Warm("known", "unknown")
	if ws.Len() != 1 {
		t.Errorf("expected 1 warmed tool (unknown skipped), got %d", ws.Len())
	}
}

func TestWarmSet_GetNil(t *testing.T) {
	reg := NewToolRegistry()
	ws := NewWarmSet(reg)
	if tool := ws.Get("nothing"); tool != nil {
		t.Errorf("expected nil, got %v", tool)
	}
}

func TestWarmSet_Contains(t *testing.T) {
	reg := NewToolRegistry()
	reg.Register(TestTool("exists"))
	ws := NewWarmSet(reg)
	ws.Warm("exists")

	if !ws.Contains("exists") {
		t.Error("expected Contains to return true")
	}
	if ws.Contains("nonexistent") {
		t.Error("expected Contains to return false for unknown tool")
	}
}

func TestWarmSet_Clear(t *testing.T) {
	reg := NewToolRegistry()
	reg.Register(TestTool("a"))
	ws := NewWarmSet(reg)
	ws.Warm("a")

	ws.Clear()
	if ws.Len() != 0 {
		t.Errorf("expected 0 after clear, got %d", ws.Len())
	}
	if ws.Contains("a") {
		t.Error("expected not to contain 'a' after clear")
	}
}

func TestWarmSet_NilReceiver(t *testing.T) {
	var ws *WarmSet
	if tool := ws.Get("x"); tool != nil {
		t.Errorf("expected nil from nil receiver")
	}
	if ws.Contains("x") {
		t.Error("expected false from nil receiver")
	}
	if ws.Len() != 0 {
		t.Errorf("expected 0 from nil receiver")
	}
	ws.Warm("x") // should not panic
	ws.Clear()   // should not panic
}

func TestWarmSet_Concurrency(t *testing.T) {
	reg := NewToolRegistry()
	for i := 0; i < 20; i++ {
		reg.Register(TestTool(string(rune('a' + i))))
	}
	ws := NewWarmSet(reg)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ws.Get("a")
			ws.Contains("b")
			ws.Len()
		}()
		wg.Add(1)
		go func() {
			defer wg.Done()
			ws.Warm("a", "b", "c")
		}()
		wg.Add(1)
		go func() {
			defer wg.Done()
			ws.Clear()
		}()
	}
	wg.Wait()
}
