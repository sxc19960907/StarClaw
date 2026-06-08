package share

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"time"
)

const ManifestVersion = 1

const (
	StatusActive    = "active"
	StatusRetracted = "retracted"
)

type Artifact struct {
	ID          string     `json:"id"`
	Filename    string     `json:"filename"`
	SourcePath  string     `json:"source_path,omitempty"`
	LocalPath   string     `json:"local_path"`
	URL         string     `json:"url"`
	Purpose     string     `json:"purpose,omitempty"`
	SizeBytes   int64      `json:"size_bytes"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	RetractedAt *time.Time `json:"retracted_at,omitempty"`
}

type Manifest struct {
	Version   int        `json:"version"`
	Artifacts []Artifact `json:"artifacts"`
}

type Store struct {
	starclawDir string
	now         func() time.Time
}

func NewStore(starclawDir string) *Store {
	return &Store{
		starclawDir: starclawDir,
		now:         time.Now,
	}
}

func (s *Store) ManifestPath() string {
	return filepath.Join(s.starclawDir, "web", "manifest.json")
}

func (s *Store) Record(a Artifact) (Artifact, error) {
	if a.ID == "" {
		return Artifact{}, errors.New("share: artifact id is required")
	}
	if a.Filename == "" {
		return Artifact{}, errors.New("share: artifact filename is required")
	}
	if a.LocalPath == "" {
		return Artifact{}, errors.New("share: artifact local path is required")
	}
	if a.Status == "" {
		a.Status = StatusActive
	}
	if a.CreatedAt.IsZero() {
		a.CreatedAt = s.now().UTC()
	}
	m, err := s.read()
	if err != nil {
		return Artifact{}, err
	}
	for i, existing := range m.Artifacts {
		if existing.ID == a.ID {
			m.Artifacts[i] = a
			return a, s.write(m)
		}
	}
	m.Artifacts = append(m.Artifacts, a)
	return a, s.write(m)
}

func (s *Store) List(includeRetracted bool) ([]Artifact, error) {
	m, err := s.read()
	if err != nil {
		return nil, err
	}
	out := make([]Artifact, 0, len(m.Artifacts))
	for _, a := range m.Artifacts {
		if !includeRetracted && a.Status == StatusRetracted {
			continue
		}
		out = append(out, a)
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})
	return out, nil
}

func (s *Store) Get(id string) (Artifact, bool, error) {
	m, err := s.read()
	if err != nil {
		return Artifact{}, false, err
	}
	for _, a := range m.Artifacts {
		if a.ID == id {
			return a, true, nil
		}
	}
	return Artifact{}, false, nil
}

func (s *Store) Retract(id string) (Artifact, bool, error) {
	if id == "" {
		return Artifact{}, false, errors.New("share: artifact id is required")
	}
	m, err := s.read()
	if err != nil {
		return Artifact{}, false, err
	}
	for i, a := range m.Artifacts {
		if a.ID != id {
			continue
		}
		already := a.Status == StatusRetracted
		if !already {
			now := s.now().UTC()
			a.Status = StatusRetracted
			a.RetractedAt = &now
			m.Artifacts[i] = a
			if err := os.RemoveAll(filepath.Join(s.starclawDir, "web", id)); err != nil {
				return Artifact{}, false, fmt.Errorf("remove artifact dir: %w", err)
			}
			if err := s.write(m); err != nil {
				return Artifact{}, false, err
			}
		}
		return a, already, nil
	}
	return Artifact{}, false, fs.ErrNotExist
}

func (s *Store) read() (Manifest, error) {
	path := s.ManifestPath()
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return Manifest{Version: ManifestVersion}, nil
		}
		return Manifest{}, fmt.Errorf("read share manifest: %w", err)
	}
	var m Manifest
	if err := json.Unmarshal(b, &m); err != nil {
		_ = os.WriteFile(path+".corrupt.bak", b, 0o600)
		return Manifest{Version: ManifestVersion}, nil
	}
	if m.Version != ManifestVersion {
		_ = os.WriteFile(fmt.Sprintf("%s.unknown-v%d.bak", path, m.Version), b, 0o600)
		return Manifest{Version: ManifestVersion}, nil
	}
	return m, nil
}

func (s *Store) write(m Manifest) error {
	m.Version = ManifestVersion
	dir := filepath.Dir(s.ManifestPath())
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("mkdir share manifest dir: %w", err)
	}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal share manifest: %w", err)
	}
	tmp, err := os.CreateTemp(dir, "manifest-*.json.tmp")
	if err != nil {
		return fmt.Errorf("create share manifest tmp: %w", err)
	}
	tmpPath := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
		return fmt.Errorf("write share manifest tmp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("close share manifest tmp: %w", err)
	}
	if err := os.Rename(tmpPath, s.ManifestPath()); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("rename share manifest: %w", err)
	}
	return nil
}
