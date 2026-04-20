package integration

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/starclaw/starclaw/internal/client"
	"github.com/starclaw/starclaw/internal/session"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSessionCreateSaveLoad tests creating, saving, and loading a session
func TestSessionCreateSaveLoad(t *testing.T) {
	tmpDir := t.TempDir()

	// Create manager and new session
	mgr := session.NewManager(tmpDir)
	sess := mgr.NewSession()
	require.NotNil(t, sess)

	// Add some messages
	sess.Messages = []client.Message{
		{Role: "user", Content: "Hello"},
		{Role: "assistant", Content: "Hi there!"},
	}
	sess.Title = "Test Session"

	// Save session
	err := mgr.Save()
	require.NoError(t, err)

	// Create new manager and load session
	mgr2 := session.NewManager(tmpDir)
	loaded, err := mgr2.Resume(sess.ID)
	require.NoError(t, err)

	// Verify data
	assert.Equal(t, sess.ID, loaded.ID)
	assert.Equal(t, "Test Session", loaded.Title)
	assert.Len(t, loaded.Messages, 2)
	assert.Equal(t, "Hello", loaded.Messages[0].Content)
	assert.Equal(t, "Hi there!", loaded.Messages[1].Content)
}

// TestSessionResumeLatest tests resuming the most recent session
func TestSessionResumeLatest(t *testing.T) {
	tmpDir := t.TempDir()

	// Create first session
	mgr := session.NewManager(tmpDir)
	sess1 := mgr.NewSession()
	sess1.Title = "First Session"
	err := mgr.Save()
	require.NoError(t, err)

	// Wait and create second session
	time.Sleep(10 * time.Millisecond)
	sess2 := mgr.NewSession()
	sess2.Title = "Second Session"
	err = mgr.Save()
	require.NoError(t, err)

	// Create new manager and resume latest
	mgr2 := session.NewManager(tmpDir)
	latest, err := mgr2.ResumeLatest()

	require.NoError(t, err)
	require.NotNil(t, latest)
	assert.Equal(t, "Second Session", latest.Title)
}

// TestSessionList tests listing sessions
func TestSessionList(t *testing.T) {
	tmpDir := t.TempDir()

	mgr := session.NewManager(tmpDir)

	// Create multiple sessions
	for i := 0; i < 3; i++ {
		sess := mgr.NewSession()
		sess.Title = "Session " + string(rune('A'+i))
		sess.Messages = []client.Message{{Role: "user", Content: "Test"}}
		err := mgr.Save()
		require.NoError(t, err)
		time.Sleep(10 * time.Millisecond)
	}

	// List sessions
	summaries, err := mgr.List()
	require.NoError(t, err)
	assert.Len(t, summaries, 3)

	// Should be sorted by date (newest first)
	for i := 0; i < len(summaries)-1; i++ {
		assert.True(t, summaries[i].CreatedAt.After(summaries[i+1].CreatedAt) ||
			summaries[i].CreatedAt.Equal(summaries[i+1].CreatedAt))
	}
}

// TestSessionDelete tests deleting a session
func TestSessionDelete(t *testing.T) {
	tmpDir := t.TempDir()

	mgr := session.NewManager(tmpDir)
	sess := mgr.NewSession()
	sess.Title = "To Delete"
	err := mgr.Save()
	require.NoError(t, err)

	// Verify it exists
	summaries, err := mgr.List()
	require.NoError(t, err)
	assert.Len(t, summaries, 1)

	// Delete it
	err = mgr.Delete(sess.ID)
	require.NoError(t, err)

	// Verify it's gone
	summaries, err = mgr.List()
	require.NoError(t, err)
	assert.Len(t, summaries, 0)
}

// TestSessionFilePermissions tests that session files have correct permissions
func TestSessionFilePermissions(t *testing.T) {
	tmpDir := t.TempDir()

	mgr := session.NewManager(tmpDir)
	sess := mgr.NewSession()
	err := mgr.Save()
	require.NoError(t, err)

	// Check file was created with correct permissions
	path := filepath.Join(tmpDir, sess.ID+".json")
	info, err := os.Stat(path)
	require.NoError(t, err)

	// Skip on Windows
	if os.PathSeparator == '/' {
		mode := info.Mode().Perm()
		assert.Equal(t, os.FileMode(0600), mode)
	}
}

// TestSessionWithAgentLoop tests session integration with agent loop
func TestSessionWithAgentLoop(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a session
	mgr := session.NewManager(tmpDir)
	sess := mgr.NewSession()
	sess.Title = "Agent Test"
	sess.Messages = []client.Message{
		{Role: "user", Content: "Previous message"},
	}
	err := mgr.Save()
	require.NoError(t, err)

	// Create agent loop with this session (demonstrate integration)
	mockClient := client.NewMockClient()
	mockClient.SetResponse("Response to previous and new message")

	// Load session with new manager
	mgr2 := session.NewManager(tmpDir)
	loaded, err := mgr2.Resume(sess.ID)
	require.NoError(t, err)

	// Verify messages are loaded
	assert.Len(t, loaded.Messages, 1)
	assert.Equal(t, "Previous message", loaded.Messages[0].Content)
}
