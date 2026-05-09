package runstatus

import (
	"testing"
)

func TestRunStatus_ZeroValue(t *testing.T) {
	var s RunStatus
	if s.Code != 0 {
		t.Errorf("zero value Code = %d, want 0", s.Code)
	}
	if s.Detail != "" {
		t.Errorf("zero value Detail = %q, want empty", s.Detail)
	}
}

func TestRunStatus_WithValues(t *testing.T) {
	s := RunStatus{Code: StatusContextBloat, Detail: "context too large"}
	if s.Code != StatusContextBloat {
		t.Errorf("Code = %d, want %d", s.Code, StatusContextBloat)
	}
	if s.Detail != "context too large" {
		t.Errorf("Detail = %q, want %q", s.Detail, "context too large")
	}
}

func TestRunStatus_Constants(t *testing.T) {
	tests := []struct {
		name string
		got  int
		want int
	}{
		{"StatusOK", StatusOK, 0},
		{"StatusContextBloat", StatusContextBloat, 1},
		{"StatusLoopDetected", StatusLoopDetected, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("%s = %d, want %d", tt.name, tt.got, tt.want)
			}
		})
	}
}

func TestRunStatus_Uniqueness(t *testing.T) {
	seen := map[int]string{
		StatusOK:           "StatusOK",
		StatusContextBloat: "StatusContextBloat",
		StatusLoopDetected: "StatusLoopDetected",
	}
	if len(seen) != 3 {
		t.Error("run status constants are not unique")
	}
}
