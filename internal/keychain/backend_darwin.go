//go:build darwin

package keychain

import (
	"errors"
	"os/exec"
	"strings"
)

const securityItemNotFoundExitCode = 44

type osBackend struct{}

func newOSBackend() Backend {
	return osBackend{}
}

func (osBackend) Read(service, account string) (string, error) {
	out, err := exec.Command("/usr/bin/security", "find-generic-password", "-s", service, "-a", account, "-w").Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == securityItemNotFoundExitCode {
			return "", ErrNotFound
		}
		return "", err
	}
	return strings.TrimRight(string(out), "\r\n"), nil
}

func (osBackend) Write(service, account, value string) error {
	err := exec.Command("/usr/bin/security", "add-generic-password", "-U", "-s", service, "-a", account, "-w", value).Run()
	if err != nil {
		return err
	}
	return nil
}

func (osBackend) Delete(service, account string) error {
	err := exec.Command("/usr/bin/security", "delete-generic-password", "-s", service, "-a", account).Run()
	if err == nil {
		return nil
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() == securityItemNotFoundExitCode {
		return ErrNotFound
	}
	return err
}

// NewOSStore returns a Store backed by the macOS Keychain.
func NewOSStore() (*Store, error) {
	return NewStore(newOSBackend()), nil
}
