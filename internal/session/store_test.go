package session

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/starclaw/starclaw/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStore_Save(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewStore(tmpDir)

	sess := &Session{
		ID:    "test-session",
		Title: "Test Session",
		CWD:   "/tmp",
		Messages: []client.Message{
			{Role: "user", Content: "Hello"},
			{Role: "assistant", Content: "Hi!"},
		},
	}

	err := store.Save(sess)
	require.NoError(t, err)

	// Check file was created with correct permissions
	path := filepath.Join(tmpDir, "test-session.json")
	_, err = os.Stat(path)
	require.NoError(t, err, "Session file should exist")

	// Verify permissions (Unix only)
	if os.PathSeparator == '/' {
		info, err := os.Stat(path)
		require.NoError(t, err)
		mode := info.Mode().Perm()
		assert.Equal(t, os.FileMode(0600), mode, "File should have 0600 permissions")
	}
}

func TestStore_Save_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	sessionsDir := filepath.Join(tmpDir, "nested", "sessions")

	store := NewStore(sessionsDir)

	sess := &Session{
		ID:    "test",
		Title: "Test",
	}

	err := store.Save(sess)
	require.NoError(t, err)

	// Verify directory was created
	_, err = os.Stat(sessionsDir)
	require.NoError(t, err, "Directory should be created")
}

func TestStore_Load(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewStore(tmpDir)

	// Create and save a session
	sess := &Session{
		ID:       "test-session",
		Title:    "Test Session",
		CWD:      "/home/user",
		Messages: []client.Message{{Role: "user", Content: "Hello"}},
	}

	err := store.Save(sess)
	require.NoError(t, err)

	// Load it back
	loaded, err := store.Load("test-session")
	require.NoError(t, err)

	assert.Equal(t, sess.ID, loaded.ID)
	assert.Equal(t, sess.Title, loaded.Title)
	assert.Equal(t, sess.CWD, loaded.CWD)
	assert.Len(t, loaded.Messages, 1)
	assert.Equal(t, "user", loaded.Messages[0].Role)
	assert.Equal(t, "Hello", loaded.Messages[0].Content)
	assert.NotZero(t, loaded.CreatedAt)
	assert.NotZero(t, loaded.UpdatedAt)
}

func TestStore_Load_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewStore(tmpDir)

	_, err := store.Load("non-existent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "read session")
}

func TestStore_Load_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewStore(tmpDir)

	// Write invalid JSON
	path := filepath.Join(tmpDir, "invalid.json")
	err := os.WriteFile(path, []byte("not valid json"), 0600)
	require.NoError(t, err)

	_, err = store.Load("invalid")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parse session")
}

func TestStore_List(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewStore(tmpDir)

	// Create multiple sessions
	sessions := []*Session{
		{ID: "session-1", Title: "First Session"},
		{ID: "session-2", Title: "Second Session"},
		{ID: "session-3", Title: "Third Session"},
	}

	for _, sess := range sessions {
		sess.Messages = []client.Message{{Role: "user", Content: "Test"}}
		err := store.Save(sess)
		require.NoError(t, err)
	}

	// List sessions
	summaries, err := store.List()
	require.NoError(t, err)
	assert.Len(t, summaries, 3)

	// Should be sorted by creation time (newest first)
	// Since we just created them, order should be 3, 2, 1
	ids := []string{summaries[0].ID, summaries[1].ID, summaries[2].ID}
	assert.Contains(t, ids, "session-1")
	assert.Contains(t, ids, "session-2")
	assert.Contains(t, ids, "session-3")

	// Check summaries have correct data
	for _, s := range summaries {
		assert.NotEmpty(t, s.ID)
		assert.NotEmpty(t, s.Title)
		assert.NotZero(t, s.CreatedAt)
		assert.Equal(t, 1, s.MsgCount)
	}
}

func TestStore_List_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewStore(tmpDir)

	summaries, err := store.List()
	require.NoError(t, err)
	assert.Empty(t, summaries)
}

func TestStore_List_SkipsCorrupted(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewStore(tmpDir)

	// Create a valid session
	sess := &Session{ID: "valid", Title: "Valid"}
	err := store.Save(sess)
	require.NoError(t, err)

	// Create a corrupted file
	corruptPath := filepath.Join(tmpDir, "corrupt.json")
	err = os.WriteFile(corruptPath, []byte("not valid"), 0600)
	require.NoError(t, err)

	// Should only return the valid session
	summaries, err := store.List()
	require.NoError(t, err)
	assert.Len(t, summaries, 1)
	assert.Equal(t, "valid", summaries[0].ID)
}

func TestStore_Delete(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewStore(tmpDir)

	// Create a session
	sess := &Session{ID: "to-delete", Title: "Delete Me"}
	err := store.Save(sess)
	require.NoError(t, err)

	// Verify it exists
	path := filepath.Join(tmpDir, "to-delete.json")
	_, err = os.Stat(path)
	require.NoError(t, err)

	// Delete it
	err = store.Delete("to-delete")
	require.NoError(t, err)

	// Verify it's gone
	_, err = os.Stat(path)
	assert.True(t, os.IsNotExist(err))
}

func TestStore_Delete_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewStore(tmpDir)

	// Deleting non-existent should error
	err := store.Delete("non-existent")
	require.Error(t, err)
}

func TestStore_SaveUpdatesTimestamps(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewStore(tmpDir)

	sess := &Session{
		ID:    "timestamp-test",
		Title: "Timestamp Test",
	}

	// First save
	err := store.Save(sess)
	require.NoError(t, err)

	createdAt := sess.CreatedAt
	updatedAt := sess.UpdatedAt

	assert.NotZero(t, createdAt)
	assert.NotZero(t, updatedAt)
	assert.Equal(t, createdAt, updatedAt)

	// Wait a bit and save again
	time.Sleep(10 * time.Millisecond)
	err = store.Save(sess)
	require.NoError(t, err)

	// CreatedAt should stay the same, UpdatedAt should change
	assert.Equal(t, createdAt, sess.CreatedAt)
	assert.True(t, sess.UpdatedAt.After(updatedAt))
}
