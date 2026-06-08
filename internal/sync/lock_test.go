package sync

import (
	"context"
	"errors"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestWithLockRunsCallback(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sync.lock")
	called := false
	err := WithLock(context.Background(), path, time.Second, func() error {
		called = true
		return nil
	})
	if err != nil {
		t.Fatalf("WithLock: %v", err)
	}
	if !called {
		t.Fatal("callback was not called")
	}
}

func TestWithLockContention(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("filelock is currently a no-op on Windows")
	}
	path := filepath.Join(t.TempDir(), "sync.lock")
	release, err := AcquireLock(context.Background(), path, time.Second)
	if err != nil {
		t.Fatalf("AcquireLock first: %v", err)
	}
	defer release()

	err = WithLock(context.Background(), path, 20*time.Millisecond, func() error {
		t.Fatal("callback should not run during contention")
		return nil
	})
	if !errors.Is(err, ErrLockContention) {
		t.Fatalf("err = %v, want ErrLockContention", err)
	}
}

func TestWithLockContextCancellation(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("filelock is currently a no-op on Windows")
	}
	path := filepath.Join(t.TempDir(), "sync.lock")
	release, err := AcquireLock(context.Background(), path, time.Second)
	if err != nil {
		t.Fatalf("AcquireLock first: %v", err)
	}
	defer release()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err = WithLock(ctx, path, time.Second, func() error {
		t.Fatal("callback should not run after cancellation")
		return nil
	})
	if !errors.Is(err, ErrLockContention) {
		t.Fatalf("err = %v, want ErrLockContention", err)
	}
}
