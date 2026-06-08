package sync

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/starclaw/starclaw/internal/filelock"
)

var ErrLockContention = errors.New("sync: lock contention")

// WithLock acquires an exclusive lock for the duration of fn. Timeout or
// context cancellation returns ErrLockContention so callers can treat another
// sync run as a noop instead of a data failure.
func WithLock(ctx context.Context, lockPath string, timeout time.Duration, fn func() error) error {
	release, err := AcquireLock(ctx, lockPath, timeout)
	if err != nil {
		return err
	}
	defer release()
	return fn()
}

// AcquireLock acquires an exclusive lock and returns a release function.
func AcquireLock(ctx context.Context, lockPath string, timeout time.Duration) (func(), error) {
	if err := os.MkdirAll(filepath.Dir(lockPath), 0700); err != nil {
		return nil, fmt.Errorf("mkdir sync lock dir: %w", err)
	}
	file, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, fmt.Errorf("open sync lock: %w", err)
	}

	deadline := time.Time{}
	if timeout > 0 {
		deadline = time.Now().Add(timeout)
	}

	for {
		if err := filelock.TryExclusive(file); err == nil {
			return func() {
				_ = filelock.Unlock(file)
				_ = file.Close()
			}, nil
		} else if !errors.Is(err, filelock.ErrWouldBlock) {
			_ = file.Close()
			return nil, err
		}

		if timeout <= 0 || (!deadline.IsZero() && !time.Now().Before(deadline)) {
			_ = file.Close()
			return nil, ErrLockContention
		}

		select {
		case <-ctx.Done():
			_ = file.Close()
			return nil, fmt.Errorf("%w: %v", ErrLockContention, ctx.Err())
		case <-time.After(lockPollInterval()):
		}
	}
}

func lockPollInterval() time.Duration {
	if runtime.GOOS == "windows" {
		return time.Millisecond
	}
	return 10 * time.Millisecond
}
