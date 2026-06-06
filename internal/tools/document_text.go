package tools

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/starclaw/starclaw/internal/agent"
)

const maxDocumentTextFileBytes = 25 * 1024 * 1024

// DocumentTextTool extracts readable text from common local document formats.
type DocumentTextTool struct{}

type documentTextArgs struct {
	Path     string `json:"path"`
	Format   string `json:"format,omitempty"`
	MaxChars int    `json:"max_chars,omitempty"`
}

func (t *DocumentTextTool) Info() agent.ToolInfo {
	return agent.ToolInfo{
		Name:        "document_text",
		Description: "Extract readable text from PDF, DOCX, XLSX, or PPTX files. This is read-only and returns bounded text for agent analysis.",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path":      map[string]any{"type": "string", "description": "Absolute or relative path to a PDF, DOCX, XLSX, or PPTX file"},
				"format":    map[string]any{"type": "string", "description": "Optional format override: pdf, docx, xlsx, or pptx"},
				"max_chars": map[string]any{"type": "integer", "description": "Maximum characters to return (default: 60000)"},
			},
		},
		Required: []string{"path"},
	}
}

func (t *DocumentTextTool) Run(ctx context.Context, argsJSON string) (agent.ToolResult, error) {
	var args documentTextArgs
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return agent.ValidationError(fmt.Sprintf("invalid arguments: %v", err)), nil
	}
	if strings.TrimSpace(args.Path) == "" {
		return agent.ValidationError("path is required"), nil
	}

	args.Path = ExpandHome(args.Path)
	if err := IsSafePath(args.Path); err != nil {
		return agent.PermissionError(err.Error()), nil
	}

	info, err := os.Stat(args.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return agent.ValidationError(fmt.Sprintf("file not found: %s", args.Path)), nil
		}
		if os.IsPermission(err) {
			return agent.PermissionError(fmt.Sprintf("permission denied: %s", args.Path)), nil
		}
		return agent.ToolResult{Content: fmt.Sprintf("error reading document metadata: %v", err), IsError: true}, nil
	}
	if info.IsDir() {
		return agent.ValidationError(fmt.Sprintf("path is a directory: %s", args.Path)), nil
	}
	if info.Size() > maxDocumentTextFileBytes {
		return agent.ValidationError(fmt.Sprintf("file too large: %d bytes exceeds %d byte limit", info.Size(), maxDocumentTextFileBytes)), nil
	}

	format := strings.ToLower(strings.TrimPrefix(strings.TrimSpace(args.Format), "."))
	if format == "" {
		format = strings.TrimPrefix(strings.ToLower(filepath.Ext(args.Path)), ".")
	}

	maxChars := args.MaxChars
	if maxChars <= 0 {
		maxChars = 60000
	}

	var text string
	switch format {
	case "docx":
		text, err = extractDOCXText(args.Path)
	case "xlsx":
		text, err = extractXLSXText(args.Path)
	case "pptx":
		text, err = extractPPTXText(args.Path)
	case "pdf":
		text, err = extractPDFText(args.Path)
	default:
		return agent.ValidationError(fmt.Sprintf("unsupported document format: %s", format)), nil
	}
	if err != nil {
		return agent.ToolResult{Content: fmt.Sprintf("error extracting document text: %v", err), IsError: true}, nil
	}

	text = normalizeDocumentText(text)
	truncated := false
	if len([]rune(text)) > maxChars {
		runes := []rune(text)
		text = string(runes[:maxChars])
		truncated = true
	}
	if strings.TrimSpace(text) == "" {
		return agent.ToolResult{Content: fmt.Sprintf("No readable text found in %s. Complex PDFs may require OCR or a dedicated PDF parser.", args.Path)}, nil
	}

	header := fmt.Sprintf("Document: %s\nFormat: %s\nCharacters: %d", args.Path, format, len([]rune(text)))
	if truncated {
		header += fmt.Sprintf("\nTruncated: true (max_chars=%d)", maxChars)
	}
	if format == "pdf" {
		header += "\nNote: PDF extraction is best-effort for embedded text streams and does not perform OCR."
	}
	return agent.ToolResult{Content: header + "\n\n" + text}, nil
}

func (t *DocumentTextTool) RequiresApproval() bool { return true }

func (t *DocumentTextTool) IsReadOnlyCall(string) bool { return true }

func (t *DocumentTextTool) IsSafeArgs(argsJSON string) bool {
	var args documentTextArgs
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return false
	}
	return IsPathUnderCWD(args.Path)
}

func extractDOCXText(path string) (string, error) {
	zr, err := zip.OpenReader(path)
	if err != nil {
		return "", fmt.Errorf("open docx zip: %w", err)
	}
	defer func() { _ = zr.Close() }()

	var parts []string
	for _, name := range []string{"word/document.xml", "word/footnotes.xml", "word/endnotes.xml"} {
		text, err := zipXMLText(&zr.Reader, name, map[string]bool{"p": true, "br": true, "tab": true})
		if err == nil && strings.TrimSpace(text) != "" {
			parts = append(parts, text)
		}
	}
	return strings.Join(parts, "\n\n"), nil
}

func extractPPTXText(path string) (string, error) {
	zr, err := zip.OpenReader(path)
	if err != nil {
		return "", fmt.Errorf("open pptx zip: %w", err)
	}
	defer func() { _ = zr.Close() }()

	var slideNames []string
	for _, f := range zr.File {
		if strings.HasPrefix(f.Name, "ppt/slides/slide") && strings.HasSuffix(f.Name, ".xml") {
			slideNames = append(slideNames, f.Name)
		}
	}
	sort.Strings(slideNames)

	var parts []string
	for i, name := range slideNames {
		text, err := zipXMLText(&zr.Reader, name, map[string]bool{"p": true, "br": true})
		if err != nil || strings.TrimSpace(text) == "" {
			continue
		}
		parts = append(parts, fmt.Sprintf("Slide %d:\n%s", i+1, text))
	}
	return strings.Join(parts, "\n\n"), nil
}

func extractXLSXText(path string) (string, error) {
	zr, err := zip.OpenReader(path)
	if err != nil {
		return "", fmt.Errorf("open xlsx zip: %w", err)
	}
	defer func() { _ = zr.Close() }()

	shared, err := readXLSXSharedStrings(&zr.Reader)
	if err != nil {
		return "", err
	}

	var sheetNames []string
	for _, f := range zr.File {
		if strings.HasPrefix(f.Name, "xl/worksheets/sheet") && strings.HasSuffix(f.Name, ".xml") {
			sheetNames = append(sheetNames, f.Name)
		}
	}
	sort.Strings(sheetNames)

	var sheets []string
	for i, name := range sheetNames {
		text, err := readXLSXSheet(&zr.Reader, name, shared)
		if err != nil || strings.TrimSpace(text) == "" {
			continue
		}
		sheets = append(sheets, fmt.Sprintf("Sheet %d:\n%s", i+1, text))
	}
	return strings.Join(sheets, "\n\n"), nil
}

func extractPDFText(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read pdf: %w", err)
	}
	var texts []string
	for _, raw := range extractPDFLiteralStrings(data) {
		decoded := decodePDFLiteralString(raw)
		if looksReadable(decoded) {
			texts = append(texts, decoded)
		}
	}
	return strings.Join(texts, "\n"), nil
}

func zipXMLText(zr *zip.Reader, name string, breakOn map[string]bool) (string, error) {
	f, err := findZipFile(zr, name)
	if err != nil {
		return "", err
	}
	rc, err := f.Open()
	if err != nil {
		return "", fmt.Errorf("open %s: %w", name, err)
	}
	defer func() { _ = rc.Close() }()
	return xmlText(rc, breakOn), nil
}

func xmlText(r io.Reader, breakOn map[string]bool) string {
	decoder := xml.NewDecoder(r)
	var sb strings.Builder
	for {
		tok, err := decoder.Token()
		if err != nil {
			break
		}
		switch t := tok.(type) {
		case xml.CharData:
			sb.WriteString(string(t))
		case xml.EndElement:
			if breakOn[t.Name.Local] {
				sb.WriteString("\n")
			}
		}
	}
	return sb.String()
}

func readXLSXSharedStrings(zr *zip.Reader) ([]string, error) {
	f, err := findZipFile(zr, "xl/sharedStrings.xml")
	if err != nil {
		return nil, nil
	}
	rc, err := f.Open()
	if err != nil {
		return nil, fmt.Errorf("open shared strings: %w", err)
	}
	defer func() { _ = rc.Close() }()

	decoder := xml.NewDecoder(rc)
	var values []string
	var current strings.Builder
	inSI := false
	for {
		tok, err := decoder.Token()
		if err != nil {
			break
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Local == "si" {
				inSI = true
				current.Reset()
			}
		case xml.CharData:
			if inSI {
				current.WriteString(string(t))
			}
		case xml.EndElement:
			if t.Name.Local == "si" {
				values = append(values, current.String())
				inSI = false
			}
		}
	}
	return values, nil
}

func readXLSXSheet(zr *zip.Reader, name string, shared []string) (string, error) {
	f, err := findZipFile(zr, name)
	if err != nil {
		return "", err
	}
	rc, err := f.Open()
	if err != nil {
		return "", fmt.Errorf("open %s: %w", name, err)
	}
	defer func() { _ = rc.Close() }()

	decoder := xml.NewDecoder(rc)
	var rows []string
	var cells []string
	var cellType string
	var inValue bool
	var value strings.Builder
	for {
		tok, err := decoder.Token()
		if err != nil {
			break
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "row":
				cells = nil
			case "c":
				cellType = ""
				for _, attr := range t.Attr {
					if attr.Name.Local == "t" {
						cellType = attr.Value
					}
				}
			case "v", "t":
				inValue = true
				value.Reset()
			}
		case xml.CharData:
			if inValue {
				value.WriteString(string(t))
			}
		case xml.EndElement:
			switch t.Name.Local {
			case "v", "t":
				inValue = false
				cells = append(cells, resolveXLSXCell(value.String(), cellType, shared))
			case "row":
				row := strings.TrimSpace(strings.Join(cells, "\t"))
				if row != "" {
					rows = append(rows, row)
				}
			}
		}
	}
	return strings.Join(rows, "\n"), nil
}

func resolveXLSXCell(value, cellType string, shared []string) string {
	value = strings.TrimSpace(value)
	if cellType == "s" {
		idx, err := strconv.Atoi(value)
		if err == nil && idx >= 0 && idx < len(shared) {
			return shared[idx]
		}
	}
	return value
}

func findZipFile(zr *zip.Reader, name string) (*zip.File, error) {
	for _, f := range zr.File {
		if f.Name == name {
			return f, nil
		}
	}
	return nil, fmt.Errorf("zip part not found: %s", name)
}

func normalizeDocumentText(text string) string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	spaceRE := regexp.MustCompile(`[ \t]+`)
	text = spaceRE.ReplaceAllString(text, " ")
	lineRE := regexp.MustCompile(`\n{3,}`)
	text = lineRE.ReplaceAllString(text, "\n\n")
	return strings.TrimSpace(text)
}

func extractPDFLiteralStrings(data []byte) [][]byte {
	var out [][]byte
	for i := 0; i < len(data); i++ {
		if data[i] != '(' {
			continue
		}
		start := i + 1
		depth := 1
		escaped := false
		for j := start; j < len(data); j++ {
			if escaped {
				escaped = false
				continue
			}
			switch data[j] {
			case '\\':
				escaped = true
			case '(':
				depth++
			case ')':
				depth--
				if depth == 0 {
					out = append(out, data[start:j])
					i = j
					goto next
				}
			}
		}
	next:
	}
	return out
}

func decodePDFLiteralString(raw []byte) string {
	var buf bytes.Buffer
	for i := 0; i < len(raw); i++ {
		if raw[i] != '\\' || i+1 >= len(raw) {
			buf.WriteByte(raw[i])
			continue
		}
		i++
		switch raw[i] {
		case 'n':
			buf.WriteByte('\n')
		case 'r':
			buf.WriteByte('\r')
		case 't':
			buf.WriteByte('\t')
		case 'b':
			buf.WriteByte('\b')
		case 'f':
			buf.WriteByte('\f')
		case '\\', '(', ')':
			buf.WriteByte(raw[i])
		default:
			if raw[i] >= '0' && raw[i] <= '7' {
				end := i + 1
				for end < len(raw) && end < i+3 && raw[end] >= '0' && raw[end] <= '7' {
					end++
				}
				v, err := strconv.ParseInt(string(raw[i:end]), 8, 32)
				if err == nil {
					buf.WriteRune(rune(v))
				}
				i = end - 1
			} else {
				buf.WriteByte(raw[i])
			}
		}
	}
	return buf.String()
}

func looksReadable(s string) bool {
	s = strings.TrimSpace(s)
	if len(s) < 2 {
		return false
	}
	printable := 0
	lettersOrDigits := 0
	for _, r := range s {
		if unicode.IsPrint(r) || unicode.IsSpace(r) {
			printable++
		}
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			lettersOrDigits++
		}
	}
	return printable >= len([]rune(s))*8/10 && lettersOrDigits > 0
}
