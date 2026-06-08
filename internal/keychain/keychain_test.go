package keychain

import (
	"errors"
	"fmt"
	"sync"
	"testing"
)

func newTestStore() (*Store, *MemBackend) {
	backend := NewMemBackend()
	return NewStore(backend), backend
}

func TestStoreGetAPIKeyEmpty(t *testing.T) {
	store, _ := newTestStore()
	key, err := store.GetAPIKey()
	if err != nil {
		t.Fatalf("GetAPIKey: %v", err)
	}
	if key != "" {
		t.Fatalf("key = %q, want empty", key)
	}

	userID, err := store.CurrentUserID()
	if err != nil {
		t.Fatalf("CurrentUserID: %v", err)
	}
	if userID != "" {
		t.Fatalf("userID = %q, want empty", userID)
	}
}

func TestStoreSetAPIKeyRoundTrip(t *testing.T) {
	store, _ := newTestStore()
	if err := store.SetAPIKey("user-1", "sk-test"); err != nil {
		t.Fatalf("SetAPIKey: %v", err)
	}

	userID, key, err := store.GetActiveUserAndKey()
	if err != nil {
		t.Fatalf("GetActiveUserAndKey: %v", err)
	}
	if userID != "user-1" {
		t.Fatalf("userID = %q, want user-1", userID)
	}
	if key != "sk-test" {
		t.Fatalf("key = %q, want sk-test", key)
	}
}

func TestStoreSetAPIKeyValidation(t *testing.T) {
	store, _ := newTestStore()
	if err := store.SetAPIKey("", "key"); err == nil {
		t.Fatal("SetAPIKey with empty userID should error")
	}
	if err := store.SetAPIKey("user-1", ""); err == nil {
		t.Fatal("SetAPIKey with empty key should error")
	}
}

func TestStoreDeleteAPIKeyPreservesCurrentUser(t *testing.T) {
	store, _ := newTestStore()
	if err := store.SetAPIKey("user-1", "sk-test"); err != nil {
		t.Fatalf("SetAPIKey: %v", err)
	}

	if err := store.DeleteAPIKey(); err != nil {
		t.Fatalf("DeleteAPIKey: %v", err)
	}

	userID, err := store.CurrentUserID()
	if err != nil {
		t.Fatalf("CurrentUserID: %v", err)
	}
	if userID != "user-1" {
		t.Fatalf("userID = %q, want user-1", userID)
	}

	key, err := store.GetAPIKey()
	if err != nil {
		t.Fatalf("GetAPIKey: %v", err)
	}
	if key != "" {
		t.Fatalf("key = %q, want empty", key)
	}
}

func TestStoreClearActiveUser(t *testing.T) {
	store, _ := newTestStore()
	if err := store.SetAPIKey("user-1", "sk-test"); err != nil {
		t.Fatalf("SetAPIKey: %v", err)
	}

	if err := store.ClearActiveUser(); err != nil {
		t.Fatalf("ClearActiveUser: %v", err)
	}

	userID, err := store.CurrentUserID()
	if err != nil {
		t.Fatalf("CurrentUserID: %v", err)
	}
	if userID != "" {
		t.Fatalf("userID = %q, want empty", userID)
	}

	key, err := store.GetAPIKey()
	if err != nil {
		t.Fatalf("GetAPIKey: %v", err)
	}
	if key != "" {
		t.Fatalf("key = %q, want empty", key)
	}
}

func TestStoreDeleteIdempotent(t *testing.T) {
	store, _ := newTestStore()
	if err := store.DeleteAPIKey(); err != nil {
		t.Fatalf("DeleteAPIKey on empty store: %v", err)
	}
	if err := store.ClearActiveUser(); err != nil {
		t.Fatalf("ClearActiveUser on empty store: %v", err)
	}
	if err := store.Delete(ServiceDaemonAPIKey, "missing-user"); err != nil {
		t.Fatalf("Delete missing entry: %v", err)
	}
}

func TestStoreRenameLegacy(t *testing.T) {
	store, backend := newTestStore()
	if err := store.Write(ServiceDaemonAPIKey, AccountLegacy, "sk-legacy"); err != nil {
		t.Fatalf("write legacy key: %v", err)
	}
	if err := store.Write(ServiceDaemonState, AccountCurrentUser, AccountLegacy); err != nil {
		t.Fatalf("write active legacy user: %v", err)
	}

	if err := store.RenameLegacy("real-user"); err != nil {
		t.Fatalf("RenameLegacy: %v", err)
	}

	userID, key, err := store.GetActiveUserAndKey()
	if err != nil {
		t.Fatalf("GetActiveUserAndKey: %v", err)
	}
	if userID != "real-user" {
		t.Fatalf("userID = %q, want real-user", userID)
	}
	if key != "sk-legacy" {
		t.Fatalf("key = %q, want sk-legacy", key)
	}
	if _, ok := backend.Snapshot()[memKey(ServiceDaemonAPIKey, AccountLegacy)]; ok {
		t.Fatal("legacy key should be deleted after RenameLegacy")
	}
}

func TestStoreRenameLegacyNoEntry(t *testing.T) {
	store, _ := newTestStore()
	if err := store.RenameLegacy("real-user"); err != nil {
		t.Fatalf("RenameLegacy without legacy key: %v", err)
	}

	userID, err := store.CurrentUserID()
	if err != nil {
		t.Fatalf("CurrentUserID: %v", err)
	}
	if userID != "" {
		t.Fatalf("userID = %q, want empty", userID)
	}
}

func TestStoreRenameLegacyValidation(t *testing.T) {
	store, _ := newTestStore()
	if err := store.RenameLegacy(""); err == nil {
		t.Fatal("RenameLegacy with empty realUserID should error")
	}
}

func TestStoreActiveUserWithoutKey(t *testing.T) {
	store, _ := newTestStore()
	if err := store.Write(ServiceDaemonState, AccountCurrentUser, "orphan-user"); err != nil {
		t.Fatalf("write active user: %v", err)
	}

	key, err := store.GetAPIKey()
	if err != nil {
		t.Fatalf("GetAPIKey: %v", err)
	}
	if key != "" {
		t.Fatalf("key = %q, want empty", key)
	}
}

func TestStoreNilCases(t *testing.T) {
	var nilStore *Store
	if _, err := nilStore.GetAPIKey(); !errors.Is(err, ErrUnsupportedPlatform) {
		t.Fatalf("nil store GetAPIKey err = %v, want ErrUnsupportedPlatform", err)
	}
	if err := nilStore.SetAPIKey("user-1", "key"); !errors.Is(err, ErrUnsupportedPlatform) {
		t.Fatalf("nil store SetAPIKey err = %v, want ErrUnsupportedPlatform", err)
	}

	store := NewStore(nil)
	if _, err := store.Read("service", "account"); !errors.Is(err, ErrUnsupportedPlatform) {
		t.Fatalf("nil backend Read err = %v, want ErrUnsupportedPlatform", err)
	}
	if err := store.Delete("service", "account"); !errors.Is(err, ErrUnsupportedPlatform) {
		t.Fatalf("nil backend Delete err = %v, want ErrUnsupportedPlatform", err)
	}
}

func TestMemBackendConcurrentAccess(t *testing.T) {
	backend := NewMemBackend()
	var wg sync.WaitGroup
	for i := 0; i < 64; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			account := fmt.Sprintf("account-%d", i)
			value := fmt.Sprintf("value-%d", i)
			if err := backend.Write("service", account, value); err != nil {
				t.Errorf("Write(%q): %v", account, err)
				return
			}
			got, err := backend.Read("service", account)
			if err != nil {
				t.Errorf("Read(%q): %v", account, err)
				return
			}
			if got != value {
				t.Errorf("Read(%q) = %q, want %q", account, got, value)
			}
		}()
	}
	wg.Wait()

	if got := len(backend.Snapshot()); got != 64 {
		t.Fatalf("snapshot length = %d, want 64", got)
	}
}
