package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// SessionEnvelope carries one already-prepared session payload for dry-run
// outbox writing. Privacy transformations happen before this type is built.
type SessionEnvelope struct {
	ID        string          `json:"id"`
	AgentName string          `json:"agent_name,omitempty"`
	Session   json.RawMessage `json:"session"`
}

// BatchRequest is the local dry-run batch shape.
type BatchRequest struct {
	ClientVersion string            `json:"client_version,omitempty"`
	SyncAt        time.Time         `json:"sync_at"`
	Sessions      []SessionEnvelope `json:"sessions"`
}

// BatchResponse reports accepted and rejected session ids.
type BatchResponse struct {
	Accepted []string        `json:"accepted"`
	Rejected []RejectedEntry `json:"rejected,omitempty"`
}

// RejectedEntry reserves response shape for later batching/upload tasks.
type RejectedEntry struct {
	ID     string `json:"id"`
	Reason string `json:"reason"`
}

// DryRunUploader writes batches to local JSON outbox files and accepts every id.
type DryRunUploader struct {
	OutboxDir string
	Now       func() time.Time
}

// Send writes batch to OutboxDir and returns all non-empty session ids accepted.
func (u *DryRunUploader) Send(ctx context.Context, batch BatchRequest) (BatchResponse, error) {
	if err := ctx.Err(); err != nil {
		return BatchResponse{}, err
	}
	if u.OutboxDir == "" {
		return BatchResponse{}, fmt.Errorf("dry-run uploader: empty outbox dir")
	}
	if err := os.MkdirAll(u.OutboxDir, 0700); err != nil {
		return BatchResponse{}, fmt.Errorf("mkdir dry-run outbox: %w", err)
	}
	now := time.Now()
	if u.Now != nil {
		now = u.Now()
	}
	body, err := json.MarshalIndent(batch, "", "  ")
	if err != nil {
		return BatchResponse{}, fmt.Errorf("marshal dry-run batch: %w", err)
	}
	name := fmt.Sprintf("%s-%03d.json", now.UTC().Format("20060102T150405Z"), now.UnixNano()%1000)
	path := filepath.Join(u.OutboxDir, name)
	if err := os.WriteFile(path, body, 0600); err != nil {
		return BatchResponse{}, fmt.Errorf("write dry-run outbox: %w", err)
	}

	accepted := make([]string, 0, len(batch.Sessions))
	for _, session := range batch.Sessions {
		if session.ID != "" {
			accepted = append(accepted, session.ID)
		}
	}
	return BatchResponse{Accepted: accepted}, nil
}
