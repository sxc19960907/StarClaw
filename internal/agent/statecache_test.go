package agent

import (
	"sync"
	"testing"
)

func TestNewStateCache(t *testing.T) {
	sc := NewStateCache()
	if sc == nil {
		t.Fatal("NewStateCache returned nil")
	}
}

func TestStateCache_SetGet(t *testing.T) {
	sc := NewStateCache()
	sc.Set("key1", "value1")

	val, ok := sc.Get("key1")
	if !ok {
		t.Error("expected key1 to be found")
	}
	if val != "value1" {
		t.Errorf("expected 'value1', got %v", val)
	}
}

func TestStateCache_GetMissing(t *testing.T) {
	sc := NewStateCache()
	_, ok := sc.Get("nonexistent")
	if ok {
		t.Error("expected nonexistent key to not be found")
	}
}

func TestStateCache_Set_NilValue(t *testing.T) {
	sc := NewStateCache()
	sc.Set("nilkey", nil)

	val, ok := sc.Get("nilkey")
	if !ok {
		t.Error("expected nilkey to be found")
	}
	if val != nil {
		t.Errorf("expected nil value, got %v", val)
	}
}

func TestStateCache_Overwrite(t *testing.T) {
	sc := NewStateCache()
	sc.Set("key", "original")
	sc.Set("key", "updated")

	val, ok := sc.Get("key")
	if !ok {
		t.Error("expected key to be found")
	}
	if val != "updated" {
		t.Errorf("expected 'updated', got %v", val)
	}
}

func TestStateCache_Clear(t *testing.T) {
	sc := NewStateCache()
	sc.Set("key1", "value1")
	sc.Set("key2", "value2")

	if sc.Len() != 2 {
		t.Errorf("expected len 2, got %d", sc.Len())
	}

	sc.Clear()
	if sc.Len() != 0 {
		t.Errorf("expected len 0 after clear, got %d", sc.Len())
	}

	_, ok := sc.Get("key1")
	if ok {
		t.Error("expected key1 to be removed after clear")
	}
}

func TestStateCache_ClearEmpty(t *testing.T) {
	sc := NewStateCache()
	sc.Clear() // should not panic
}

func TestStateCache_IntValue(t *testing.T) {
	sc := NewStateCache()
	sc.Set("count", 42)

	val, ok := sc.Get("count")
	if !ok {
		t.Fatal("expected count to be found")
	}
	v, ok := val.(int)
	if !ok {
		t.Fatalf("expected int, got %T", val)
	}
	if v != 42 {
		t.Errorf("expected 42, got %d", v)
	}
}

func TestStateCache_MapValue(t *testing.T) {
	sc := NewStateCache()
	data := map[string]string{"foo": "bar"}
	sc.Set("data", data)

	val, ok := sc.Get("data")
	if !ok {
		t.Fatal("expected data to be found")
	}
	m, ok := val.(map[string]string)
	if !ok {
		t.Fatalf("expected map[string]string, got %T", val)
	}
	if m["foo"] != "bar" {
		t.Errorf("expected 'bar', got %q", m["foo"])
	}
}

func TestStateCache_Len(t *testing.T) {
	sc := NewStateCache()
	if sc.Len() != 0 {
		t.Errorf("initial len should be 0, got %d", sc.Len())
	}
	sc.Set("a", 1)
	if sc.Len() != 1 {
		t.Errorf("len should be 1, got %d", sc.Len())
	}
	sc.Set("b", 2)
	if sc.Len() != 2 {
		t.Errorf("len should be 2, got %d", sc.Len())
	}
	sc.Set("a", 3) // overwrite
	if sc.Len() != 2 {
		t.Errorf("len should still be 2 after overwrite, got %d", sc.Len())
	}
}

func TestStateCache_Concurrent(t *testing.T) {
	sc := NewStateCache()
	var wg sync.WaitGroup

	// Concurrent writes.
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			sc.Set("key", i)
		}(i)
	}

	// Concurrent reads.
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sc.Get("key")
		}()
	}

	// Concurrent clear.
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sc.Clear()
		}()
	}

	wg.Wait()
}

func TestStateCache_EmptyGetAfterClear(t *testing.T) {
	sc := NewStateCache()
	sc.Set("key", "value")
	sc.Clear()
	_, ok := sc.Get("key")
	if ok {
		t.Error("key should not exist after clear")
	}
}
