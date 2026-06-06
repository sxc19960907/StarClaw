package tools

import (
	"archive/zip"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDocumentTextTool_DOCX(t *testing.T) {
	tmpDir := t.TempDir()
	chdirForTest(t, tmpDir)
	path := filepath.Join(tmpDir, "sample.docx")
	writeZipFixture(t, path, map[string]string{
		"word/document.xml": `<?xml version="1.0"?><w:document xmlns:w="w"><w:body><w:p><w:r><w:t>Astria mission brief</w:t></w:r></w:p><w:p><w:r><w:t>Constellation notes</w:t></w:r></w:p></w:body></w:document>`,
	})

	tool := &DocumentTextTool{}
	result, err := tool.Run(context.Background(), `{"path":"`+path+`"}`)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got %s", result.Content)
	}
	for _, want := range []string{"Format: docx", "Astria mission brief", "Constellation notes"} {
		if !strings.Contains(result.Content, want) {
			t.Fatalf("expected %q in result: %s", want, result.Content)
		}
	}
}

func TestDocumentTextTool_XLSX(t *testing.T) {
	tmpDir := t.TempDir()
	chdirForTest(t, tmpDir)
	path := filepath.Join(tmpDir, "sample.xlsx")
	writeZipFixture(t, path, map[string]string{
		"xl/sharedStrings.xml":     `<?xml version="1.0"?><sst xmlns="s"><si><t>Name</t></si><si><t>Astria</t></si></sst>`,
		"xl/worksheets/sheet1.xml": `<?xml version="1.0"?><worksheet xmlns="s"><sheetData><row><c t="s"><v>0</v></c><c t="s"><v>1</v></c></row><row><c><v>42</v></c></row></sheetData></worksheet>`,
	})

	tool := &DocumentTextTool{}
	result, err := tool.Run(context.Background(), `{"path":"`+path+`"}`)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got %s", result.Content)
	}
	for _, want := range []string{"Format: xlsx", "Sheet 1", "Name", "Astria", "42"} {
		if !strings.Contains(result.Content, want) {
			t.Fatalf("expected %q in result: %s", want, result.Content)
		}
	}
}

func TestDocumentTextTool_PPTX(t *testing.T) {
	tmpDir := t.TempDir()
	chdirForTest(t, tmpDir)
	path := filepath.Join(tmpDir, "sample.pptx")
	writeZipFixture(t, path, map[string]string{
		"ppt/slides/slide1.xml": `<?xml version="1.0"?><p:sld xmlns:p="p" xmlns:a="a"><p:cSld><p:spTree><p:sp><p:txBody><a:p><a:r><a:t>Launch panel</a:t></a:r></a:p></p:txBody></p:sp></p:spTree></p:cSld></p:sld>`,
	})

	tool := &DocumentTextTool{}
	result, err := tool.Run(context.Background(), `{"path":"`+path+`"}`)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got %s", result.Content)
	}
	for _, want := range []string{"Format: pptx", "Slide 1", "Launch panel"} {
		if !strings.Contains(result.Content, want) {
			t.Fatalf("expected %q in result: %s", want, result.Content)
		}
	}
}

func TestDocumentTextTool_PDF(t *testing.T) {
	tmpDir := t.TempDir()
	chdirForTest(t, tmpDir)
	path := filepath.Join(tmpDir, "sample.pdf")
	pdf := `%PDF-1.4
1 0 obj << /Type /Catalog /Pages 2 0 R >> endobj
2 0 obj << /Type /Pages /Kids [3 0 R] /Count 1 >> endobj
3 0 obj << /Type /Page /Parent 2 0 R /Contents 4 0 R >> endobj
4 0 obj << /Length 44 >> stream
BT /F1 12 Tf 72 720 Td (Astria PDF text) Tj ET
endstream endobj
%%EOF`
	if err := os.WriteFile(path, []byte(pdf), 0644); err != nil {
		t.Fatalf("write pdf: %v", err)
	}

	tool := &DocumentTextTool{}
	result, err := tool.Run(context.Background(), `{"path":"`+path+`"}`)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got %s", result.Content)
	}
	for _, want := range []string{"Format: pdf", "Astria PDF text"} {
		if !strings.Contains(result.Content, want) {
			t.Fatalf("expected %q in result: %s", want, result.Content)
		}
	}
}

func TestDocumentTextTool_MaxChars(t *testing.T) {
	tmpDir := t.TempDir()
	chdirForTest(t, tmpDir)
	path := filepath.Join(tmpDir, "sample.docx")
	writeZipFixture(t, path, map[string]string{
		"word/document.xml": `<?xml version="1.0"?><w:document xmlns:w="w"><w:body><w:p><w:r><w:t>abcdefghijklmnopqrstuvwxyz</w:t></w:r></w:p></w:body></w:document>`,
	})

	tool := &DocumentTextTool{}
	result, err := tool.Run(context.Background(), `{"path":"`+path+`","max_chars":5}`)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if !strings.Contains(result.Content, "Truncated: true") {
		t.Fatalf("expected truncation note: %s", result.Content)
	}
}

func writeZipFixture(t *testing.T, path string, files map[string]string) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create zip fixture: %v", err)
	}
	zw := zip.NewWriter(f)
	for name, content := range files {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatalf("create zip entry: %v", err)
		}
		if _, err := w.Write([]byte(content)); err != nil {
			t.Fatalf("write zip entry: %v", err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("close zip writer: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("close zip fixture: %v", err)
	}
}
