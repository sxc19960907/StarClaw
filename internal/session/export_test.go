package session

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/starclaw/starclaw/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExportMarkdown(t *testing.T) {
	sess := &Session{
		ID:        "test-123",
		Title:     "Test Session",
		CreatedAt: time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2025, 1, 1, 12, 30, 0, 0, time.UTC),
		Messages: []client.Message{
			{Role: "user", Content: "Hello"},
			{Role: "assistant", Content: "Hi there!"},
		},
		Tags:     []string{"important", "test"},
		Favorite: true,
	}

	md := ExportMarkdown(sess)

	assert.Contains(t, md, "# Test Session")
	assert.Contains(t, md, "**ID**: test-123")
	assert.Contains(t, md, "**user**: Hello")
	assert.Contains(t, md, "**assistant**: Hi there!")
	assert.Contains(t, md, "**Tags**: important, test")
	assert.Contains(t, md, "**Favorite**: true")
	assert.Contains(t, md, "**Messages**: 2")
}

func TestExportMarkdown_NoTagsOrFavorite(t *testing.T) {
	sess := &Session{
		ID:    "test-456",
		Title: "Plain Session",
		Messages: []client.Message{
			{Role: "user", Content: "Hi"},
		},
	}

	md := ExportMarkdown(sess)

	assert.Contains(t, md, "# Plain Session")
	assert.NotContains(t, md, "Tags")
	assert.NotContains(t, md, "Favorite")
	assert.Contains(t, md, "**Messages**: 1")
}

func TestExportHTML(t *testing.T) {
	sess := &Session{
		ID:        "test-789",
		Title:     "HTML Test",
		CreatedAt: time.Date(2025, 6, 1, 10, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2025, 6, 1, 11, 0, 0, 0, time.UTC),
		Messages: []client.Message{
			{Role: "user", Content: "Hello World"},
			{Role: "assistant", Content: "Test <script>alert('xss')</script>"},
		},
		Tags:     []string{"demo"},
		Favorite: false,
	}

	htmlOut := ExportHTML(sess)

	assert.Contains(t, htmlOut, "<title>HTML Test</title>")
	assert.Contains(t, htmlOut, "<h1>HTML Test</h1>")
	assert.Contains(t, htmlOut, "<strong>ID:</strong> test-789")
	assert.Contains(t, htmlOut, "Hello World")
	// HTML content should be escaped
	assert.Contains(t, htmlOut, "Test &lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;")
	assert.NotContains(t, htmlOut, "<script>alert('xss')</script>")
	assert.Contains(t, htmlOut, "<strong>Tags:</strong> demo")
}

func TestExportToFile_Markdown(t *testing.T) {
	tmpFile := t.TempDir() + "/export.md"
	sess := &Session{
		ID:    "file-test",
		Title: "File Export",
		Messages: []client.Message{
			{Role: "user", Content: "Save me"},
		},
	}

	err := ExportToFile(sess, "markdown", tmpFile)
	require.NoError(t, err)

	data, err := os.ReadFile(tmpFile)
	require.NoError(t, err)
	assert.Contains(t, string(data), "# File Export")
	assert.Contains(t, string(data), "**user**: Save me")
}

func TestExportToFile_HTML(t *testing.T) {
	tmpFile := t.TempDir() + "/export.html"
	sess := &Session{
		ID:    "file-test-html",
		Title: "HTML Export",
		Messages: []client.Message{
			{Role: "user", Content: "Hello"},
		},
	}

	err := ExportToFile(sess, "html", tmpFile)
	require.NoError(t, err)

	data, err := os.ReadFile(tmpFile)
	require.NoError(t, err)
	assert.Contains(t, string(data), "<title>HTML Export</title>")
}

func TestExportToFile_UnsupportedFormat(t *testing.T) {
	sess := &Session{ID: "test", Title: "Test"}
	err := ExportToFile(sess, "pdf", "/tmp/nonexistent.pdf")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported export format")
}

func TestExportMarkdown_EmptySession(t *testing.T) {
	sess := &Session{
		ID:    "empty",
		Title: "Empty Session",
	}

	md := ExportMarkdown(sess)
	assert.Contains(t, md, "# Empty Session")
	assert.Contains(t, md, "**Messages**: 0")
}

func TestExportHTML_EmptySession(t *testing.T) {
	sess := &Session{
		ID:    "empty-html",
		Title: "Empty HTML",
	}

	htmlOut := ExportHTML(sess)
	assert.Contains(t, htmlOut, "<h1>Empty HTML</h1>")
	assert.Contains(t, htmlOut, "<strong>Messages:</strong> 0")
}

func TestExportHTML_CharacterEscaping(t *testing.T) {
	sess := &Session{
		ID:    "escape-test",
		Title: "Test & <Title>",
		Messages: []client.Message{
			{Role: "user", Content: "A & B < C > D"},
		},
	}

	htmlOut := ExportHTML(sess)
	assert.Contains(t, htmlOut, "Test &amp; &lt;Title&gt;")
	assert.Contains(t, htmlOut, "A &amp; B &lt; C &gt; D")
}

func TestExportMarkdown_LongMessages(t *testing.T) {
	longContent := strings.Repeat("This is a long message content. ", 50)
	sess := &Session{
		ID:    "long-msg",
		Title: "Long Messages",
		Messages: []client.Message{
			{Role: "user", Content: longContent},
		},
	}

	md := ExportMarkdown(sess)
	assert.Contains(t, md, "This is a long message content")
	assert.Greater(t, len(md), 1500)
}

func TestExportHTML_SpecialRoleClass(t *testing.T) {
	sess := &Session{
		ID:    "roles",
		Title: "Roles Test",
		Messages: []client.Message{
			{Role: "user", Content: "Hello"},
			{Role: "assistant", Content: "Hi"},
			{Role: "system", Content: "System prompt"},
		},
	}

	htmlOut := ExportHTML(sess)
	assert.Contains(t, htmlOut, `class="message user"`)
	assert.Contains(t, htmlOut, `class="message assistant"`)
	assert.Contains(t, htmlOut, `class="message system"`)
}
