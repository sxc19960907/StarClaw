package daemon

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type TraceExportRecord struct {
	SchemaVersion string         `json:"schema_version"`
	TraceID       string         `json:"trace_id"`
	SpanID        string         `json:"span_id"`
	ParentSpanID  string         `json:"parent_span_id,omitempty"`
	RunID         string         `json:"run_id"`
	EventID       string         `json:"event_id"`
	Name          string         `json:"name"`
	Phase         string         `json:"phase"`
	Timestamp     time.Time      `json:"timestamp"`
	Attributes    map[string]any `json:"attributes,omitempty"`
}

func (s *RunStore) TraceEvents(runID string) ([]TraceExportRecord, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	record := s.records[runID]
	if record == nil {
		return nil, false
	}
	return traceRecordsForRun(record), true
}

func (s *RunStore) AllTraceEvents() []TraceExportRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	records := []TraceExportRecord{}
	for _, id := range s.order {
		record := s.records[id]
		if record == nil {
			continue
		}
		records = append(records, traceRecordsForRun(record)...)
	}
	return records
}

func (s *RunStore) ExportTracesJSONL(path string) error {
	return writeTraceJSONL(path, s.AllTraceEvents())
}

func (s *RunStore) ExportRunTraceJSONL(runID, path string) error {
	records, ok := s.TraceEvents(runID)
	if !ok {
		return fmt.Errorf("run trace %q: run not found", runID)
	}
	return writeTraceJSONL(path, records)
}

func traceRecordsForRun(record *RunRecord) []TraceExportRecord {
	if record == nil {
		return nil
	}
	out := make([]TraceExportRecord, 0, len(record.StructuredEvents))
	for _, evt := range record.StructuredEvents {
		out = append(out, TraceExportRecord{
			SchemaVersion: evt.SchemaVersion,
			TraceID:       evt.RunID,
			SpanID:        evt.ID,
			RunID:         evt.RunID,
			EventID:       evt.ID,
			Name:          evt.Type,
			Phase:         evt.Phase,
			Timestamp:     evt.At,
			Attributes:    sanitizeTraceAttributes(evt.Data),
		})
	}
	return out
}

func writeTraceJSONL(path string, records []TraceExportRecord) error {
	if path == "" {
		return fmt.Errorf("trace export path is required")
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create trace export directory: %w", err)
	}
	tmp, err := os.CreateTemp(dir, filepath.Base(path)+".tmp-*")
	if err != nil {
		return fmt.Errorf("create trace export temp file: %w", err)
	}
	tmpPath := tmp.Name()
	encoder := json.NewEncoder(tmp)
	for _, record := range records {
		if err := encoder.Encode(record); err != nil {
			_ = tmp.Close()
			_ = os.Remove(tmpPath)
			return fmt.Errorf("encode trace export: %w", err)
		}
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("close trace export temp file: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("replace trace export: %w", err)
	}
	return nil
}

func sanitizeTraceAttributes(attrs map[string]any) map[string]any {
	if attrs == nil {
		return nil
	}
	out := make(map[string]any, len(attrs))
	for key, value := range attrs {
		if traceKeyRedacted(key) {
			out[key+"_redacted"] = true
			continue
		}
		out[key] = sanitizeTraceValue(value)
	}
	return out
}

func sanitizeTraceValue(value any) any {
	switch v := value.(type) {
	case map[string]any:
		return sanitizeTraceAttributes(v)
	case []any:
		out := make([]any, 0, len(v))
		for _, item := range v {
			out = append(out, sanitizeTraceValue(item))
		}
		return out
	default:
		return redactScalar(v)
	}
}

func traceKeyRedacted(key string) bool {
	switch key {
	case "args", "content", "text", "delta", "preamble", "prompt", "request", "response":
		return true
	default:
		return looksSensitive(key)
	}
}
