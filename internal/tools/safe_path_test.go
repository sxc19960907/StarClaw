package tools

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

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
	cwd, _ := os.Getwd()
	defer os.Chdir(cwd)

	// Change to temp directory
	os.Chdir(tmpDir)

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"relative path", "file.txt", false},
		{"subdirectory", "subdir/file.go", false},
		{"current dir", ".", false},
		{"parent dir", "..", false}, // .. 经过 Clean 后是合法的（当前目录的父目录）
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

func TestIsPathUnderCWD(t *testing.T) {
	tmpDir := t.TempDir()
	cwd, _ := os.Getwd()
	defer os.Chdir(cwd)
	os.Chdir(tmpDir)

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
	cwd, _ := os.Getwd()
	defer os.Chdir(cwd)
	os.Chdir(tmpDir)

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
