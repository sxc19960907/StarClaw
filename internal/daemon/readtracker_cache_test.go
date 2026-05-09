package daemon

import (
	"sync"
	"testing"
	"time"
)

func TestNewReadTrackerCache(t *testing.T) {
	c := NewReadTrackerCache()
	if c == nil {
		t.Fatal("NewReadTrackerCache returned nil")
	}
	if c.items == nil {
		t.Error("expected non-nil items map")
	}
}

func TestMarkAndIsRead(t *testing.T) {
	c := NewReadTrackerCache()
	path := "/home/user/file.txt"

	if c.IsRead(path) {
		t.Error("expected path to not be read initially")
	}

	c.MarkRead(path)

	if !c.IsRead(path) {
		t.Error("expected path to be read after MarkRead")
	}
}

func TestIsRead_MultiplePaths(t *testing.T) {
	c := NewReadTrackerCache()
	paths := []string{
		"/home/user/a.txt",
		"/home/user/b.txt",
		"/home/user/c.txt",
	}

	for _, p := range paths {
		if c.IsRead(p) {
			t.Errorf("expected %q to not be read initially", p)
		}
		c.MarkRead(p)
	}

	for _, p := range paths {
		if !c.IsRead(p) {
			t.Errorf("expected %q to be read after MarkRead", p)
		}
	}
}

func TestIsRead_UnknownPath(t *testing.T) {
	c := NewReadTrackerCache()
	c.MarkRead("/known/path")
	if c.IsRead("/unknown/path") {
		t.Error("expected unknown path to not be read")
	}
}

func TestClear(t *testing.T) {
	c := NewReadTrackerCache()
	c.MarkRead("/some/file")
	c.MarkRead("/other/file")

	if !c.IsRead("/some/file") {
		t.Error("expected path to be read before Clear")
	}

	c.Clear()

	if c.IsRead("/some/file") {
		t.Error("expected path to not be read after Clear")
	}
	if c.IsRead("/other/file") {
		t.Error("expected path to not be read after Clear")
	}
}

func TestClear_EmptyCache(t *testing.T) {
	c := NewReadTrackerCache()
	c.Clear() // Should not panic.
	if c.IsRead("/anything") {
		t.Error("expected path to not be read")
	}
}

func TestMarkRead_Duplicate(t *testing.T) {
	c := NewReadTrackerCache()
	path := "/dup/file"

	c.MarkRead(path)
	c.MarkRead(path) // Should not panic or error.

	if !c.IsRead(path) {
		t.Error("expected path to be read")
	}
}

func TestMarkRead_EmptyPath(t *testing.T) {
	c := NewReadTrackerCache()
	c.MarkRead("")
	if !c.IsRead("") {
		t.Error("expected empty path to be read after MarkRead")
	}
}

func TestReadTrackerCache_ConcurrentSafety(t *testing.T) {
	c := NewReadTrackerCache()
	var wg sync.WaitGroup

	// Concurrent marks and reads.
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			path := "/concurrent/file"
			c.MarkRead(path)
			_ = c.IsRead(path)
			_ = c.IsRead("/other/file")
		}(i)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// All goroutines completed without race.
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for concurrent operations")
	}
}

func TestReadTrackerCache_MarkAndClearConcurrent(t *testing.T) {
	c := NewReadTrackerCache()
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.MarkRead("/path/a")
			c.Clear()
			c.MarkRead("/path/b")
			_ = c.IsRead("/path/a")
			_ = c.IsRead("/path/b")
		}()
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for concurrent operations")
	}
}
