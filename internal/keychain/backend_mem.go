package keychain

import "sync"

// MemBackend is an in-memory Backend for tests. It is safe for concurrent use.
type MemBackend struct {
	mu   sync.Mutex
	data map[string]string
}

// NewMemBackend returns an empty in-memory backend.
func NewMemBackend() *MemBackend {
	return &MemBackend{data: map[string]string{}}
}

func memKey(service, account string) string {
	return service + "\x00" + account
}

// Read returns the value for service/account, or ErrNotFound when absent.
func (m *MemBackend) Read(service, account string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	value, ok := m.data[memKey(service, account)]
	if !ok {
		return "", ErrNotFound
	}
	return value, nil
}

// Write stores value for service/account.
func (m *MemBackend) Write(service, account, value string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[memKey(service, account)] = value
	return nil
}

// Delete removes service/account, or ErrNotFound when absent.
func (m *MemBackend) Delete(service, account string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := memKey(service, account)
	if _, ok := m.data[key]; !ok {
		return ErrNotFound
	}
	delete(m.data, key)
	return nil
}

// Snapshot returns a defensive copy of the backend map for tests.
func (m *MemBackend) Snapshot() map[string]string {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make(map[string]string, len(m.data))
	for key, value := range m.data {
		out[key] = value
	}
	return out
}
