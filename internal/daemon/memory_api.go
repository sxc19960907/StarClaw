package daemon

import (
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	ctxmem "github.com/starclaw/starclaw/internal/context"
)

type memoryEntryView struct {
	Name     string    `json:"name"`
	Size     int64     `json:"size"`
	Modified time.Time `json:"modified"`
	Primary  bool      `json:"primary"`
}

type memoryView struct {
	MemoryDir string            `json:"memory_dir"`
	Entries   []memoryEntryView `json:"entries"`
	Content   string            `json:"content,omitempty"`
}

type memoryAppendRequest struct {
	Content string `json:"content"`
}

func (s *Server) handleGetMemory(w http.ResponseWriter, r *http.Request) {
	view, err := s.buildMemoryView()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, view)
}

func (s *Server) handleAppendMemory(w http.ResponseWriter, r *http.Request) {
	var req memoryAppendRequest
	if !decodeBody(w, r, &req) {
		return
	}
	content := strings.TrimSpace(req.Content)
	if content == "" {
		writeError(w, http.StatusBadRequest, "content is required")
		return
	}
	memoryDir := s.memoryDir()
	if memoryDir == "" {
		writeError(w, http.StatusInternalServerError, "memory directory not configured")
		return
	}
	if err := ctxmem.BoundedAppend(memoryDir, content); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	view, err := s.buildMemoryView()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, view)
}

func (s *Server) handleDeleteMemory(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimSpace(r.PathValue("name"))
	if name == "" || name != filepath.Base(name) || strings.ContainsAny(name, `/\`) || !strings.HasSuffix(name, ".md") {
		writeError(w, http.StatusBadRequest, "invalid memory entry name")
		return
	}
	memoryDir := s.memoryDir()
	if memoryDir == "" {
		writeError(w, http.StatusInternalServerError, "memory directory not configured")
		return
	}
	target := filepath.Join(memoryDir, name)
	if filepath.Clean(target) != filepath.Join(filepath.Clean(memoryDir), name) {
		writeError(w, http.StatusBadRequest, "invalid memory entry name")
		return
	}
	if err := os.Remove(target); err != nil {
		if os.IsNotExist(err) {
			writeError(w, http.StatusNotFound, "memory entry not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	view, err := s.buildMemoryView()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, view)
}

func (s *Server) buildMemoryView() (memoryView, error) {
	memoryDir := s.memoryDir()
	view := memoryView{MemoryDir: memoryDir, Entries: []memoryEntryView{}}
	if memoryDir == "" {
		return view, nil
	}
	entries, err := os.ReadDir(memoryDir)
	if err != nil {
		if os.IsNotExist(err) {
			return view, nil
		}
		return view, err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		name := entry.Name()
		if name == "MEMORY.md.lock" {
			continue
		}
		view.Entries = append(view.Entries, memoryEntryView{
			Name:     name,
			Size:     info.Size(),
			Modified: info.ModTime(),
			Primary:  name == "MEMORY.md",
		})
	}
	sort.Slice(view.Entries, func(i, j int) bool {
		if view.Entries[i].Primary != view.Entries[j].Primary {
			return view.Entries[i].Primary
		}
		return view.Entries[i].Name < view.Entries[j].Name
	})
	if data, err := os.ReadFile(filepath.Join(memoryDir, "MEMORY.md")); err == nil {
		view.Content = string(data)
	}
	return view, nil
}

func (s *Server) memoryDir() string {
	if s.deps == nil || s.deps.StarclawDir == "" {
		return ""
	}
	return filepath.Join(s.deps.StarclawDir, "memory")
}
