//go:build !darwin

package keychain

// NewOSStore is unsupported outside macOS. Callers should fall back to existing
// config/env credential sources unless the user explicitly opts into another
// future backend.
func NewOSStore() (*Store, error) {
	return nil, ErrUnsupportedPlatform
}
