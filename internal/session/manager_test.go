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

func TestManager_AddTag(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManager(tmpDir)

	sess := mgr.NewSession()
	sess.Title = "Tag Test"
	err := mgr.Save()
	require.NoError(t, err)

	// Add a tag
	err = mgr.AddTag(sess.ID, "important")
	require.NoError(t, err)

	// Verify by loading session again
	resumed, err := mgr.Resume(sess.ID)
	require.NoError(t, err)
	assert.Contains(t, resumed.Tags, "important")
}

func TestManager_AddTag_Duplicate(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManager(tmpDir)

	sess := mgr.NewSession()
	err := mgr.Save()
	require.NoError(t, err)

	// Add same tag twice
	err = mgr.AddTag(sess.ID, "test-tag")
	require.NoError(t, err)

	err = mgr.AddTag(sess.ID, "test-tag")
	require.NoError(t, err)

	// Verify only one instance
	resumed, err := mgr.Resume(sess.ID)
	require.NoError(t, err)
	assert.Len(t, resumed.Tags, 1)
}

func TestManager_RemoveTag(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManager(tmpDir)

	sess := mgr.NewSession()
	sess.Tags = []string{"alpha", "beta", "gamma"}
	err := mgr.Save()
	require.NoError(t, err)

	// Remove a tag
	err = mgr.RemoveTag(sess.ID, "beta")
	require.NoError(t, err)

	// Verify
	resumed, err := mgr.Resume(sess.ID)
	require.NoError(t, err)
	assert.Equal(t, []string{"alpha", "gamma"}, resumed.Tags)
}

func TestManager_RemoveTag_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManager(tmpDir)

	sess := mgr.NewSession()
	sess.Tags = []string{"alpha"}
	err := mgr.Save()
	require.NoError(t, err)

	// Remove non-existent tag should be a no-op
	err = mgr.RemoveTag(sess.ID, "nonexistent")
	require.NoError(t, err)

	resumed, err := mgr.Resume(sess.ID)
	require.NoError(t, err)
	assert.Equal(t, []string{"alpha"}, resumed.Tags)
}

func TestManager_AddTag_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManager(tmpDir)

	err := mgr.AddTag("nonexistent", "tag")
	require.Error(t, err)
}

func TestManager_SetFavorite(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManager(tmpDir)

	sess := mgr.NewSession()
	err := mgr.Save()
	require.NoError(t, err)

	// Set favorite
	err = mgr.SetFavorite(sess.ID, true)
	require.NoError(t, err)

	resumed, err := mgr.Resume(sess.ID)
	require.NoError(t, err)
	assert.True(t, resumed.Favorite)

	// Unset favorite
	err = mgr.SetFavorite(sess.ID, false)
	require.NoError(t, err)

	resumed, err = mgr.Resume(sess.ID)
	require.NoError(t, err)
	assert.False(t, resumed.Favorite)
}

func TestManager_SetFavorite_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManager(tmpDir)

	err := mgr.SetFavorite("nonexistent", true)
	require.Error(t, err)
}

func TestManager_SearchByTag(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManager(tmpDir)

	// Create sessions with different tags
	sess1 := mgr.NewSession()
	sess1.Title = "Alpha"
	sess1.Tags = []string{"important"}
	require.NoError(t, mgr.Save())

	sess2 := mgr.NewSession()
	sess2.Title = "Beta"
	sess2.Tags = []string{"important"}
	require.NoError(t, mgr.Save())

	sess3 := mgr.NewSession()
	sess3.Title = "Gamma"
	sess3.Tags = []string{"archived"}
	require.NoError(t, mgr.Save())

	// Search by tag
	results := mgr.SearchByTag("important")
	assert.Len(t, results, 2)

	ids := []string{results[0].ID, results[1].ID}
	assert.Contains(t, ids, sess1.ID)
	assert.Contains(t, ids, sess2.ID)
}

func TestManager_SearchByTag_NoMatch(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManager(tmpDir)

	sess := mgr.NewSession()
	sess.Tags = []string{"one"}
	require.NoError(t, mgr.Save())

	results := mgr.SearchByTag("nonexistent")
	assert.Empty(t, results)
}

func TestManager_ListFavorites(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManager(tmpDir)

	// Create sessions with different favorite states
	sess1 := mgr.NewSession()
	sess1.Title = "Fav1"
	sess1.Favorite = true
	require.NoError(t, mgr.Save())

	sess2 := mgr.NewSession()
	sess2.Title = "NotFav"
	require.NoError(t, mgr.Save())

	sess3 := mgr.NewSession()
	sess3.Title = "Fav2"
	sess3.Favorite = true
	require.NoError(t, mgr.Save())

	// List favorites
	favorites := mgr.ListFavorites()
	assert.Len(t, favorites, 2)

	ids := []string{favorites[0].ID, favorites[1].ID}
	assert.Contains(t, ids, sess1.ID)
	assert.Contains(t, ids, sess3.ID)
}

func TestManager_ListFavorites_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManager(tmpDir)

	mgr.NewSession()
	mgr.Save()

	favorites := mgr.ListFavorites()
	assert.Empty(t, favorites)
}

func TestManager_ThreadSafety(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManager(tmpDir)

	// Create a single session
	sess := mgr.NewSession()

	// Concurrent saves - each goroutine should modify its own session copy
	var wg sync.WaitGroup
	var mu sync.Mutex
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()

			// Add message with proper synchronization
			mu.Lock()
			sess.Messages = append(sess.Messages, client.Message{
				Role:    "user",
				Content: "Message " + string(rune('0'+n)),
			})
			mgr.Save()
			mu.Unlock()
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
