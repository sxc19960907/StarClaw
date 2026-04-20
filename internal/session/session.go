// Package session provides session persistence for conversations
package session

import (
	"time"

	"github.com/starclaw/starclaw/internal/client"
)

// Session represents a conversation session
type Session struct {
	ID        string           `json:"id"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
	Title     string           `json:"title"`
	CWD       string           `json:"cwd"`      // Working directory
	Messages  []client.Message `json:"messages"` // Conversation history
}

// SessionSummary is a lightweight version for listing sessions
type SessionSummary struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	MsgCount  int       `json:"msg_count"`
}
