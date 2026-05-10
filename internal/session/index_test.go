package session

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIndex_Build(t *testing.T) {
	tmpDir := t.TempDir()

	sessions := []*Session{
		{ID: "sess-1", Title: "First Session", Tags: []string{"important"}},
		{ID: "sess-2", Title: "Research on Go", Favorite: true},
		{ID: "sess-3", Title: "Debug session"},
	}
	store := NewStore(tmpDir)
	for _, s := range sessions {
		require.NoError(t, store.Save(s))
	}

	idx := NewIndex()
	err := idx.Build(tmpDir)
	require.NoError(t, err)

	// Lookup by ID
	meta := idx.Lookup("sess-1")
	require.NotNil(t, meta)
	assert.Equal(t, "First Session", meta.Title)
	assert.Equal(t, []string{"important"}, meta.Tags)
	assert.Equal(t, 0, meta.MsgCount)

	// Search by title (substring, case-insensitive)
	results := idx.Search("research")
	assert.Len(t, results, 1)
	assert.Equal(t, "Research on Go", results[0].Title)
	assert.True(t, results[0].Favorite)

	results = idx.Search("session")
	assert.Len(t, results, 2) // "First Session" and "Debug session", not "Research on Go"
}

func TestIndex_Lookup_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewStore(tmpDir)
	require.NoError(t, store.Save(&Session{ID: "sess-1", Title: "Test"}))

	idx := NewIndex()
	require.NoError(t, idx.Build(tmpDir))

	meta := idx.Lookup("nonexistent")
	assert.Nil(t, meta)
}

func TestIndex_Search_NoMatch(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewStore(tmpDir)
	require.NoError(t, store.Save(&Session{ID: "sess-1", Title: "Alpha"}))

	idx := NewIndex()
	require.NoError(t, idx.Build(tmpDir))

	results := idx.Search("omega")
	assert.Empty(t, results)
}

func TestIndex_Build_EmptyDir(t *testing.T) {
	tmpDir := t.TempDir()

	idx := NewIndex()
	err := idx.Build(tmpDir)
	require.NoError(t, err)

	assert.Nil(t, idx.Lookup("anything"))
	assert.Empty(t, idx.Search("test"))
}

func TestIndex_Build_CorruptedFile(t *testing.T) {
	tmpDir := t.TempDir()

	store := NewStore(tmpDir)
	require.NoError(t, store.Save(&Session{ID: "valid", Title: "Valid Session"}))

	// Create a corrupted session file
	err := os.WriteFile(filepath.Join(tmpDir, "corrupt.json"), []byte("invalid json"), 0600)
	require.NoError(t, err)

	idx := NewIndex()
	err = idx.Build(tmpDir)
	require.NoError(t, err)

	// Should only contain the valid session; corrupted file is skipped
	assert.NotNil(t, idx.Lookup("valid"))
	assert.Nil(t, idx.Lookup("corrupt"))
}

func TestIndex_NotBuilt(t *testing.T) {
	idx := NewIndex()
	assert.Nil(t, idx.Lookup("test"))
	assert.Empty(t, idx.Search("test"))
}

func TestIndex_ReBuild(t *testing.T) {
	tmpDir := t.TempDir()

	store := NewStore(tmpDir)
	require.NoError(t, store.Save(&Session{ID: "sess-1", Title: "Alpha"}))

	idx := NewIndex()
	require.NoError(t, idx.Build(tmpDir))
	assert.NotNil(t, idx.Lookup("sess-1"))

	// Add another session and rebuild
	require.NoError(t, store.Save(&Session{ID: "sess-2", Title: "Beta"}))

	idx2 := NewIndex()
	require.NoError(t, idx2.Build(tmpDir))
	assert.NotNil(t, idx2.Lookup("sess-1"))
	assert.NotNil(t, idx2.Lookup("sess-2"))
	assert.Len(t, idx2.Search(""), 2)
}

func TestIndex_Search_CaseInsensitive(t *testing.T) {
	tmpDir := t.TempDir()

	store := NewStore(tmpDir)
	require.NoError(t, store.Save(&Session{ID: "s1", Title: "Go Programming"}))
	require.NoError(t, store.Save(&Session{ID: "s2", Title: "GOLANG TUTORIAL"}))
	require.NoError(t, store.Save(&Session{ID: "s3", Title: "Python 101"}))

	idx := NewIndex()
	require.NoError(t, idx.Build(tmpDir))

	// Should match both "Go Programming" and "GOLANG TUTORIAL"
	results := idx.Search("go")
	assert.Len(t, results, 2)

	results = idx.Search("python")
	assert.Len(t, results, 1)
}
