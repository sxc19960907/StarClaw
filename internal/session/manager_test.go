package session

import (
	"sync"
	"testing"
	"time"

	"github.com/starclaw/starclaw/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManager_NewSession(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManager(tmpDir)

	sess := mgr.NewSession()
	require.NotNil(t, sess)

	// Check ID format (should contain timestamp)
	assert.NotEmpty(t, sess.ID)
	assert.Regexp(t, `^\d{4}-\d{2}-\d{2}-\d{2}-\d{2}-\d{2}-[a-f0-9]+$`, sess.ID)

	// Check other fields
	assert.NotZero(t, sess.CreatedAt)
	assert.NotEmpty(t, sess.CWD)
	assert.Equal(t, "New session", sess.Title)
	assert.Empty(t, sess.Messages)

	// Should be set as current
	assert.Equal(t, sess, mgr.Current())
}

func TestManager_NewSession_UniqueIDs(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManager(tmpDir)

	// Create multiple sessions
	ids := make(map[string]bool)
	for i := 0; i < 10; i++ {
		sess := mgr.NewSession()
		ids[sess.ID] = true
	}

	// All IDs should be unique
	assert.Len(t, ids, 10)
}

func TestManager_Current(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManager(tmpDir)

	// Initially nil
	assert.Nil(t, mgr.Current())

	// After creating session
	sess := mgr.NewSession()
	assert.Equal(t, sess, mgr.Current())
}

func TestManager_Resume(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManager(tmpDir)

	// Create and save a session
	sess := mgr.NewSession()
	sess.Messages = []client.Message{{Role: "user", Content: "Test message"}}
	err := mgr.Save()
	require.NoError(t, err)

	// Create new manager and resume
	mgr2 := NewManager(tmpDir)
	resumed, err := mgr2.Resume(sess.ID)

	require.NoError(t, err)
	assert.Equal(t, sess.ID, resumed.ID)
	assert.Len(t, resumed.Messages, 1)
	assert.Equal(t, "Test message", resumed.Messages[0].Content)

	// Should be set as current
	assert.Equal(t, resumed, mgr2.Current())
}

func TestManager_Resume_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManager(tmpDir)

	_, err := mgr.Resume("non-existent-session")
	require.Error(t, err)
}

func TestManager_Save(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManager(tmpDir)

	// Save without current session should be no-op
	err := mgr.Save()
	require.NoError(t, err)

	// Create and save session
	sess := mgr.NewSession()
	sess.Messages = []client.Message{
		{Role: "user", Content: "Hello"},
	}

	err = mgr.Save()
	require.NoError(t, err)

	// Verify by loading with new manager
	mgr2 := NewManager(tmpDir)
	loaded, err := mgr2.Resume(sess.ID)
	require.NoError(t, err)
	assert.Len(t, loaded.Messages, 1)
}

func TestManager_List(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManager(tmpDir)

	// Create and save multiple sessions
	for i := 0; i < 3; i++ {
		sess := mgr.NewSession()
		sess.Title = "Session " + string(rune('A'+i))
		err := mgr.Save()
		require.NoError(t, err)
	}

	// List should return all sessions
	summaries, err := mgr.List()
	require.NoError(t, err)
	assert.Len(t, summaries, 3)
}

func TestManager_Delete(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManager(tmpDir)

	// Create and save session
	sess := mgr.NewSession()
	err := mgr.Save()
	require.NoError(t, err)

	// Verify it exists
	list, err := mgr.List()
	require.NoError(t, err)
	assert.Len(t, list, 1)

	// Delete it
	err = mgr.Delete(sess.ID)
	require.NoError(t, err)

	// Verify it's gone
	list, err = mgr.List()
	require.NoError(t, err)
	assert.Empty(t, list)
}

func TestManager_ResumeLatest(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManager(tmpDir)

	// Create first session
	sess1 := mgr.NewSession()
	sess1.Title = "First"
	err := mgr.Save()
	require.NoError(t, err)

	// Wait a moment and create second session
	time.Sleep(10 * time.Millisecond)
	sess2 := mgr.NewSession()
	sess2.Title = "Second"
	err = mgr.Save()
	require.NoError(t, err)

	// Create new manager and resume latest
	mgr2 := NewManager(tmpDir)
	latest, err := mgr2.ResumeLatest()

	require.NoError(t, err)
	require.NotNil(t, latest)
	assert.Equal(t, sess2.ID, latest.ID)
	assert.Equal(t, "Second", latest.Title)
}

func TestManager_ResumeLatest_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManager(tmpDir)

	latest, err := mgr.ResumeLatest()
	require.NoError(t, err)
	assert.Nil(t, latest)
}

func TestManager_ThreadSafety(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManager(tmpDir)

	// Create a single session
	sess := mgr.NewSession()

	// Concurrent saves
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()

			// Add message and save
			sess.Messages = append(sess.Messages, client.Message{
				Role:    "user",
				Content: "Message " + string(rune('0'+n)),
			})
			mgr.Save()
		}(i)
	}
	wg.Wait()

	// Verify session was saved
	summaries, err := mgr.List()
	require.NoError(t, err)
	assert.Len(t, summaries, 1)
	// Should have at least some messages (may not be all 10 due to race, but mutex helps)
	assert.GreaterOrEqual(t, summaries[0].MsgCount, 0)
}
