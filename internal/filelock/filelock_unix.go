//go:build !windows

package filelock

import (
	"fmt"
	"os"
	"syscall"
)

// Shared takes a shared advisory lock on f.
func Shared(f *os.File) error {
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_SH); err != nil {
		return fmt.Errorf("flock shared: %w", err)
	}
	return nil
}

// Exclusive takes an exclusive advisory lock on f.
func Exclusive(f *os.File) error {
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		return fmt.Errorf("flock exclusive: %w", err)
	}
	return nil
}

// Unlock releases an advisory lock on f.
func Unlock(f *os.File) error {
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_UN); err != nil {
		return fmt.Errorf("flock unlock: %w", err)
	}
	return nil
}
