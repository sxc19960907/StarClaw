package tools

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func chdirForTest(t *testing.T, dir string) {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get cwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir %s: %v", dir, err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(cwd); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	})
}

func TestExpandHome(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("Cannot get home directory")
	}

	tests := []struct {
		input    string
		expected string
	}{
		{"~/test.txt", filepath.Join(home, "test.txt")},
		{"~/Documents/file.go", filepath.Join(home, "Documents", "file.go")},
		{"/absolute/path", "/absolute/path"},
		{"relative/path", "relative/path"},
		{"test", "test"},
	}

	for _, tt := range tests {
		result := ExpandHome(tt.input)
		if result != tt.expected {
			t.Errorf("ExpandHome(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestIsSafePath(t *testing.T) {
	// Create temp directory for testing
	tmpDir := t.TempDir()
	chdirForTest(t, tmpDir)

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"relative path", "file.txt", false},
		{"subdirectory", "subdir/file.go", false},
		{"current dir", ".", false},
		{"parent dir", "..", true},
		{"traversal outside home", "/etc/passwd", true}, // 系统路径
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := IsSafePath(tt.path)
			if tt.wantErr && err == nil {
				t.Errorf("IsSafePath(%q) should error", tt.path)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("IsSafePath(%q) unexpected error: %v", tt.path, err)
			}
		})
	}

	// Test system directories (Unix only)
	if runtime.GOOS != "windows" {
		tests := []struct {
			name    string
			path    string
			wantErr bool
		}{
			{"/etc", "/etc", true},
			{"/usr", "/usr", true},
			{"/bin", "/bin", true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := IsSafePath(tt.path)
				if tt.wantErr && err == nil {
					t.Errorf("IsSafePath(%q) should error", tt.path)
				}
			})
		}
	}
}

func TestIsSafePath_RejectsSiblingPrefix(t *testing.T) {
	parent := t.TempDir()
	project := filepath.Join(parent, "project")
	sibling := filepath.Join(parent, "project-other")
	if err := os.MkdirAll(project, 0755); err != nil {
		t.Fatalf("mkdir project: %v", err)
	}
	if err := os.MkdirAll(sibling, 0755); err != nil {
		t.Fatalf("mkdir sibling: %v", err)
	}
	chdirForTest(t, project)

	if err := IsSafePath(filepath.Join(sibling, "secret.txt")); err == nil {
		t.Fatal("sibling-prefix path should be rejected")
	}
}

func TestIsSafePath_RejectsSymlinkEscape(t *testing.T) {
	parent := t.TempDir()
	project := filepath.Join(parent, "project")
	outside := filepath.Join(parent, "outside")
	if err := os.MkdirAll(project, 0755); err != nil {
		t.Fatalf("mkdir project: %v", err)
	}
	if err := os.MkdirAll(outside, 0755); err != nil {
		t.Fatalf("mkdir outside: %v", err)
	}
	if err := os.WriteFile(filepath.Join(outside, "secret.txt"), []byte("secret"), 0644); err != nil {
		t.Fatalf("write outside file: %v", err)
	}
	link := filepath.Join(project, "link")
	if err := os.Symlink(outside, link); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	chdirForTest(t, project)

	if err := IsSafePath(filepath.Join(link, "secret.txt")); err == nil {
		t.Fatal("symlink escape should be rejected")
	}
}

func TestIsPathUnderCWD_RejectsSymlinkEscape(t *testing.T) {
	parent := t.TempDir()
	project := filepath.Join(parent, "project")
	outside := filepath.Join(parent, "outside")
	if err := os.MkdirAll(project, 0755); err != nil {
		t.Fatalf("mkdir project: %v", err)
	}
	if err := os.MkdirAll(outside, 0755); err != nil {
		t.Fatalf("mkdir outside: %v", err)
	}
	link := filepath.Join(project, "link")
	if err := os.Symlink(outside, link); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	chdirForTest(t, project)

	if IsPathUnderCWD(filepath.Join(link, "secret.txt")) {
		t.Fatal("symlink escape should not be considered under CWD")
	}
}

func TestIsPathUnderCWD(t *testing.T) {
	tmpDir := t.TempDir()
	chdirForTest(t, tmpDir)

	tests := []struct {
		path     string
		expected bool
	}{
		{"file.txt", true},
		{"./file.txt", true},
		{"subdir/file.go", true},
		{"../file.txt", false},
		{"/absolute/path", false},
	}

	for _, tt := range tests {
		result := IsPathUnderCWD(tt.path)
		if result != tt.expected {
			t.Errorf("IsPathUnderCWD(%q) = %v, want %v", tt.path, result, tt.expected)
		}
	}
}

func TestNormalizePath(t *testing.T) {
	tmpDir := t.TempDir()
	chdirForTest(t, tmpDir)

	// 在 macOS 上 /var 是 /private/var 的符号链接，所以我们只检查结果是否以目录名结尾
	result, err := NormalizePath("file.txt")
	if err != nil {
		t.Fatalf("NormalizePath error: %v", err)
	}
	if !strings.HasSuffix(result, "file.txt") {
		t.Errorf("NormalizePath should end with file.txt, got: %s", result)
	}

	result, err = NormalizePath("./file.txt")
	if err != nil {
		t.Fatalf("NormalizePath error: %v", err)
	}
	if !strings.HasSuffix(result, "file.txt") {
		t.Errorf("NormalizePath should end with file.txt, got: %s", result)
	}

	result, err = NormalizePath("subdir/../file.txt")
	if err != nil {
		t.Fatalf("NormalizePath error: %v", err)
	}
	if !strings.HasSuffix(result, "file.txt") {
		t.Errorf("NormalizePath should end with file.txt, got: %s", result)
	}
}
