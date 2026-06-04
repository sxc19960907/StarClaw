package session

import (
	"fmt"
	"html"
	"os"
	"strings"
	"time"
)

// ExportMarkdown formats a session as Markdown
func ExportMarkdown(s *Session) string {
	var b strings.Builder

	// Title
	b.WriteString("# ")
	b.WriteString(s.Title)
	b.WriteString("\n\n")

	// Metadata section
	b.WriteString("## Metadata\n\n")
	_, _ = fmt.Fprintf(&b, "- **ID**: %s\n", s.ID)
	_, _ = fmt.Fprintf(&b, "- **Created**: %s\n", s.CreatedAt.Format(time.RFC3339))
	_, _ = fmt.Fprintf(&b, "- **Updated**: %s\n", s.UpdatedAt.Format(time.RFC3339))
	_, _ = fmt.Fprintf(&b, "- **Messages**: %d\n", len(s.Messages))
	if len(s.Tags) > 0 {
		_, _ = fmt.Fprintf(&b, "- **Tags**: %s\n", strings.Join(s.Tags, ", "))
	}
	if s.Favorite {
		b.WriteString("- **Favorite**: true\n")
	}
	b.WriteString("\n")

	// Messages
	b.WriteString("## Messages\n\n")
	for _, msg := range s.Messages {
		_, _ = fmt.Fprintf(&b, "**%s**: %s\n\n", msg.Role, msg.Content)
	}

	return b.String()
}

// ExportHTML formats a session as HTML
func ExportHTML(s *Session) string {
	var b strings.Builder

	b.WriteString("<!DOCTYPE html>\n<html lang=\"en\">\n<head>\n")
	b.WriteString("<meta charset=\"UTF-8\">\n")
	_, _ = fmt.Fprintf(&b, "<title>%s</title>\n", html.EscapeString(s.Title))
	b.WriteString("<style>\n")
	b.WriteString("body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; max-width: 800px; margin: 0 auto; padding: 2em; background: #fafafa; color: #333; }\n")
	b.WriteString("h1 { border-bottom: 2px solid #eee; padding-bottom: 0.3em; }\n")
	b.WriteString(".metadata { background: #fff; border: 1px solid #ddd; border-radius: 6px; padding: 1em; margin: 1em 0; }\n")
	b.WriteString(".message { background: #fff; border: 1px solid #eee; border-radius: 6px; padding: 1em; margin: 0.5em 0; }\n")
	b.WriteString(".message .role { font-weight: bold; color: #555; margin-bottom: 0.3em; }\n")
	b.WriteString(".message .content { white-space: pre-wrap; }\n")
	b.WriteString(".message.user { border-left: 3px solid #4a9eff; }\n")
	b.WriteString(".message.assistant { border-left: 3px solid #6fcf97; }\n")
	b.WriteString(".message.system { border-left: 3px solid #f39c12; }\n")
	b.WriteString("</style>\n")
	b.WriteString("</head>\n<body>\n")

	// Title
	_, _ = fmt.Fprintf(&b, "<h1>%s</h1>\n", html.EscapeString(s.Title))

	// Metadata
	b.WriteString("<div class=\"metadata\">\n")
	_, _ = fmt.Fprintf(&b, "<p><strong>ID:</strong> %s</p>\n", html.EscapeString(s.ID))
	_, _ = fmt.Fprintf(&b, "<p><strong>Created:</strong> %s</p>\n", s.CreatedAt.Format(time.RFC3339))
	_, _ = fmt.Fprintf(&b, "<p><strong>Updated:</strong> %s</p>\n", s.UpdatedAt.Format(time.RFC3339))
	_, _ = fmt.Fprintf(&b, "<p><strong>Messages:</strong> %d</p>\n", len(s.Messages))
	if len(s.Tags) > 0 {
		_, _ = fmt.Fprintf(&b, "<p><strong>Tags:</strong> %s</p>\n", html.EscapeString(strings.Join(s.Tags, ", ")))
	}
	if s.Favorite {
		b.WriteString("<p><strong>Favorite:</strong> true</p>\n")
	}
	b.WriteString("</div>\n")

	// Messages
	for _, msg := range s.Messages {
		roleClass := html.EscapeString(msg.Role)
		b.WriteString(fmt.Sprintf("<div class=\"message %s\">\n", roleClass))
		b.WriteString(fmt.Sprintf("<div class=\"role\">%s</div>\n", html.EscapeString(msg.Role)))
		b.WriteString(fmt.Sprintf("<div class=\"content\">%s</div>\n", html.EscapeString(msg.Content)))
		b.WriteString("</div>\n")
	}

	b.WriteString("</body>\n</html>\n")

	return b.String()
}

// ExportToFile writes a session to a file in the specified format
func ExportToFile(s *Session, format string, path string) error {
	var content string

	switch format {
	case "markdown", "md":
		content = ExportMarkdown(s)
	case "html":
		content = ExportHTML(s)
	default:
		return fmt.Errorf("unsupported export format: %s", format)
	}

	return os.WriteFile(path, []byte(content), 0600)
}
