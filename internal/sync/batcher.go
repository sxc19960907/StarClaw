package sync

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// SessionLoader returns raw session JSON bytes for a candidate.
type SessionLoader func(dir, id string) ([]byte, error)

// FileSessionLoader loads session JSON from the candidate directory.
func FileSessionLoader(dir, id string) ([]byte, error) {
	return os.ReadFile(filepath.Join(dir, filepath.Base(id)+".json"))
}

// BuildBatches packs candidates into local dry-run batch requests.
func BuildBatches(ctx context.Context, candidates []Candidate, loader SessionLoader, cfg Config, marker *Marker, now time.Time) ([]BatchRequest, error) {
	if loader == nil {
		loader = FileSessionLoader
	}
	if marker != nil && marker.Failed == nil {
		marker.Failed = map[string]FailedEntry{}
	}

	var batches []BatchRequest
	current := BatchRequest{
		SyncAt:   now,
		Sessions: []SessionEnvelope{},
	}
	currentBytes := 0
	flush := func() {
		if len(current.Sessions) > 0 {
			batches = append(batches, current)
			current = BatchRequest{SyncAt: now, Sessions: []SessionEnvelope{}}
			currentBytes = 0
		}
	}

	for _, candidate := range candidates {
		if err := ctx.Err(); err != nil {
			return batches, err
		}
		body, err := loader(candidate.Dir, candidate.SessionID)
		if err != nil {
			recordFailed(marker, candidate, "load_error", 0, now)
			continue
		}
		if stripped, err := StripThinkingFromSessionJSON(body); err == nil {
			body = stripped
		}
		size := len(body)
		if cfg.SingleSessionMaxBytes > 0 && size > cfg.SingleSessionMaxBytes {
			recordFailed(marker, candidate, "size_limit_exceeded", uint64(size), now)
			continue
		}

		env := SessionEnvelope{
			ID:        candidate.SessionID,
			AgentName: candidate.AgentName,
			Session:   json.RawMessage(body),
		}

		if cfg.BatchMaxSessions > 0 && len(current.Sessions) >= cfg.BatchMaxSessions {
			flush()
		}
		if cfg.BatchMaxBytes > 0 && currentBytes+size > cfg.BatchMaxBytes && len(current.Sessions) > 0 {
			flush()
		}
		current.Sessions = append(current.Sessions, env)
		currentBytes += size
	}
	flush()
	return batches, nil
}

func recordFailed(marker *Marker, candidate Candidate, reason string, sizeBytes uint64, now time.Time) {
	if marker == nil {
		return
	}
	if marker.Failed == nil {
		marker.Failed = map[string]FailedEntry{}
	}
	prev, ok := marker.Failed[candidate.SessionID]
	if !ok {
		prev = FailedEntry{
			FirstAttemptAt: now,
		}
	}
	prev.Reason = reason
	prev.Category = ClassifyReason(reason)
	prev.Attempts++
	prev.LastAttemptAt = now
	prev.LastObservedUpdatedAt = candidate.UpdatedAt
	prev.SizeBytes = sizeBytes
	if prev.Category == CategoryPermanent {
		prev.NextAttemptAt = nil
	}
	marker.Failed[candidate.SessionID] = prev
}

// ClassifyReason returns whether a failure is transient or permanent.
func ClassifyReason(reason string) string {
	switch reason {
	case "load_error", "size_limit_exceeded":
		return CategoryPermanent
	default:
		return CategoryTransient
	}
}
