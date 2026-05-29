//go:build windows

package filelock

import "os"

// Shared is a no-op on Windows until native file locking is implemented.
func Shared(f *os.File) error {
	return nil
}

// Exclusive is a no-op on Windows until native file locking is implemented.
func Exclusive(f *os.File) error {
	return nil
}

// Unlock is a no-op on Windows until native file locking is implemented.
func Unlock(f *os.File) error {
	return nil
}
