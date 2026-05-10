package session

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// SessionMeta holds searchable metadata for a session.
type SessionMeta struct {
	ID       string   `json:"id"`
	Title    string   `json:"title"`
	Tags     []string `json:"tags,omitempty"`
	Favorite bool     `json:"favorite"`
	MsgCount int      `json:"msg_count"`
}

// Index provides fast in-memory lookup of session metadata.
type Index struct {
	sessions map[string]*SessionMeta
	loaded   bool
}

// NewIndex creates an empty Index.
func NewIndex() *Index {
	return &Index{
		sessions: make(map[string]*SessionMeta),
	}
}

// Build scans sessionsDir and indexes all session JSON files.
func (idx *Index) Build(sessionsDir string) error {
	entries, err := os.ReadDir(sessionsDir)
	if err != nil {
		return err
	}

	idx.sessions = make(map[string]*SessionMeta, len(entries))

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		id := strings.TrimSuffix(e.Name(), ".json")
		data, err := os.ReadFile(filepath.Join(sessionsDir, e.Name()))
		if err != nil {
			continue
		}
		var sess Session
		if err := json.Unmarshal(data, &sess); err != nil {
			continue
		}
		idx.sessions[id] = &SessionMeta{
			ID:       sess.ID,
			Title:    sess.Title,
			Tags:     sess.Tags,
			Favorite: sess.Favorite,
			MsgCount: len(sess.Messages),
		}
	}
	idx.loaded = true
	return nil
}

// Lookup returns the metadata for a session by ID, or nil if not found.
func (idx *Index) Lookup(id string) *SessionMeta {
	if !idx.loaded {
		return nil
	}
	return idx.sessions[id]
}

// Search returns all sessions whose title matches the query (case-insensitive).
func (idx *Index) Search(query string) []SessionMeta {
	if !idx.loaded {
		return nil
	}
	q := strings.ToLower(query)
	var results []SessionMeta
	for _, meta := range idx.sessions {
		if strings.Contains(strings.ToLower(meta.Title), q) {
			results = append(results, *meta)
		}
	}
	return results
}
