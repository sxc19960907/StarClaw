package tools

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestArchiveInspectTool_Zip(t *testing.T) {
	tmpDir := t.TempDir()
	chdirForTest(t, tmpDir)
	path := filepath.Join(tmpDir, "sample.zip")
	writeZipFixture(t, path, map[string]string{
		"notes/readme.txt": "hello",
	})

	tool := &ArchiveInspectTool{}
	result, err := tool.Run(context.Background(), `{"path":"`+path+`"}`)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got %s", result.Content)
	}
	if !strings.Contains(result.Content, `"format": "zip"`) || !strings.Contains(result.Content, `"name": "notes/readme.txt"`) {
		t.Fatalf("unexpected result: %s", result.Content)
	}
}

func TestArchiveInspectTool_TarAndTarGz(t *testing.T) {
	tmpDir := t.TempDir()
	chdirForTest(t, tmpDir)

	for _, tt := range []struct {
		name string
		gzip bool
	}{
		{name: "sample.tar"},
		{name: "sample.tar.gz", gzip: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(tmpDir, tt.name)
			writeTarFixture(t, path, tt.gzip, map[string]string{"docs/brief.txt": "astria"})

			tool := &ArchiveInspectTool{}
			result, err := tool.Run(context.Background(), `{"path":"`+path+`"}`)
			if err != nil {
				t.Fatalf("Run returned error: %v", err)
			}
			if result.IsError {
				t.Fatalf("expected success, got %s", result.Content)
			}
			if !strings.Contains(result.Content, `"name": "docs/brief.txt"`) {
				t.Fatalf("unexpected result: %s", result.Content)
			}
		})
	}
}

func TestArchiveExtractTool_Zip(t *testing.T) {
	tmpDir := t.TempDir()
	chdirForTest(t, tmpDir)
	path := filepath.Join(tmpDir, "sample.zip")
	dest := filepath.Join(tmpDir, "out")
	writeZipFixture(t, path, map[string]string{
		"notes/readme.txt": "hello astria",
	})

	tool := &ArchiveExtractTool{}
	result, err := tool.Run(context.Background(), `{"path":"`+path+`","destination":"`+dest+`"}`)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got %s", result.Content)
	}
	data, err := os.ReadFile(filepath.Join(dest, "notes", "readme.txt"))
	if err != nil {
		t.Fatalf("read extracted file: %v", err)
	}
	if string(data) != "hello astria" {
		t.Fatalf("unexpected extracted data: %s", data)
	}
}

func TestArchiveExtractTool_RejectsTraversal(t *testing.T) {
	tmpDir := t.TempDir()
	chdirForTest(t, tmpDir)
	path := filepath.Join(tmpDir, "evil.zip")
	dest := filepath.Join(tmpDir, "out")
	writeZipFixture(t, path, map[string]string{
		"../evil.txt": "bad",
	})

	tool := &ArchiveExtractTool{}
	result, err := tool.Run(context.Background(), `{"path":"`+path+`","destination":"`+dest+`"}`)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if !result.IsError || !strings.Contains(result.Content, "unsafe archive entry path") {
		t.Fatalf("expected traversal error, got: %s", result.Content)
	}
	if _, err := os.Stat(filepath.Join(tmpDir, "evil.txt")); !os.IsNotExist(err) {
		t.Fatalf("archive traversal wrote outside destination")
	}
}

func TestArchiveExtractTool_SkipsExistingWithoutOverwrite(t *testing.T) {
	tmpDir := t.TempDir()
	chdirForTest(t, tmpDir)
	path := filepath.Join(tmpDir, "sample.zip")
	dest := filepath.Join(tmpDir, "out")
	writeZipFixture(t, path, map[string]string{
		"readme.txt": "new",
	})
	if err := os.MkdirAll(dest, 0755); err != nil {
		t.Fatalf("mkdir dest: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dest, "readme.txt"), []byte("old"), 0644); err != nil {
		t.Fatalf("write existing: %v", err)
	}

	tool := &ArchiveExtractTool{}
	result, err := tool.Run(context.Background(), `{"path":"`+path+`","destination":"`+dest+`"}`)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got %s", result.Content)
	}
	data, err := os.ReadFile(filepath.Join(dest, "readme.txt"))
	if err != nil {
		t.Fatalf("read existing: %v", err)
	}
	if string(data) != "old" {
		t.Fatalf("expected existing file to remain unchanged, got %s", data)
	}
	if !strings.Contains(result.Content, `"skipped"`) {
		t.Fatalf("expected skipped entry in result: %s", result.Content)
	}
}

func TestArchiveExtractTool_TarGz(t *testing.T) {
	tmpDir := t.TempDir()
	chdirForTest(t, tmpDir)
	path := filepath.Join(tmpDir, "sample.tgz")
	dest := filepath.Join(tmpDir, "out")
	writeTarFixture(t, path, true, map[string]string{"docs/brief.txt": "tar text"})

	tool := &ArchiveExtractTool{}
	result, err := tool.Run(context.Background(), `{"path":"`+path+`","destination":"`+dest+`"}`)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got %s", result.Content)
	}
	data, err := os.ReadFile(filepath.Join(dest, "docs", "brief.txt"))
	if err != nil {
		t.Fatalf("read extracted file: %v", err)
	}
	if string(data) != "tar text" {
		t.Fatalf("unexpected extracted data: %s", data)
	}
}

func writeTarFixture(t *testing.T, path string, gzipped bool, files map[string]string) {
	t.Helper()
	var buf bytes.Buffer
	var w io.Writer = &buf
	var gz *gzip.Writer
	if gzipped {
		gz = gzip.NewWriter(&buf)
		w = gz
	}
	tw := tar.NewWriter(w)
	for name, content := range files {
		hdr := &tar.Header{
			Name: name,
			Mode: 0644,
			Size: int64(len(content)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatalf("write tar header: %v", err)
		}
		if _, err := tw.Write([]byte(content)); err != nil {
			t.Fatalf("write tar content: %v", err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("close tar writer: %v", err)
	}
	if gz != nil {
		if err := gz.Close(); err != nil {
			t.Fatalf("close gzip writer: %v", err)
		}
	}
	if err := os.WriteFile(path, buf.Bytes(), 0644); err != nil {
		t.Fatalf("write tar fixture: %v", err)
	}
}

func TestArchiveExtractTool_RejectsZipSymlinkOverwrite(t *testing.T) {
	tmpDir := t.TempDir()
	chdirForTest(t, tmpDir)
	path := filepath.Join(tmpDir, "sample.zip")
	dest := filepath.Join(tmpDir, "out")
	if err := os.MkdirAll(dest, 0755); err != nil {
		t.Fatalf("mkdir dest: %v", err)
	}
	if err := os.Symlink(filepath.Join(tmpDir, "target.txt"), filepath.Join(dest, "link.txt")); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	writeZipFixture(t, path, map[string]string{"link.txt": "content"})

	tool := &ArchiveExtractTool{}
	result, err := tool.Run(context.Background(), `{"path":"`+path+`","destination":"`+dest+`","overwrite":true}`)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if !result.IsError || !strings.Contains(result.Content, "refusing to overwrite symlink") {
		t.Fatalf("expected symlink overwrite error, got: %s", result.Content)
	}
}
