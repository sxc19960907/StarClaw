package tools

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/starclaw/starclaw/internal/agent"
	"github.com/starclaw/starclaw/internal/schedule"
)

func setupScheduleMgr(t *testing.T) *schedule.Manager {
	t.Helper()
	return schedule.NewManager(filepath.Join(t.TempDir(), "schedules.json"))
}

func TestScheduleTool_Create(t *testing.T) {
	mgr := setupScheduleMgr(t)
	tools := NewScheduleTools(mgr)
	tool := tools[0] // create

	result, err := tool.Run(context.Background(), `{"cron":"0 9 * * *","prompt":"daily report"}`)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected error: %s", result.Content)
	}
	if result.Content == "" {
		t.Error("expected non-empty result")
	}

	// Verify it's in the list
	list, _ := mgr.List()
	if len(list) != 1 {
		t.Fatalf("got %d schedules, want 1", len(list))
	}
}

func TestScheduleTool_CreateWithAgent(t *testing.T) {
	mgr := setupScheduleMgr(t)
	tools := NewScheduleTools(mgr)
	tool := tools[0] // create

	result, err := tool.Run(context.Background(), `{"agent":"ops-bot","cron":"*/5 * * * *","prompt":"check"}`)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected error: %s", result.Content)
	}

	list, _ := mgr.List()
	if list[0].Agent != "ops-bot" {
		t.Errorf("agent = %q, want ops-bot", list[0].Agent)
	}
}

func TestScheduleTool_CreateMissingRequired(t *testing.T) {
	mgr := setupScheduleMgr(t)
	tools := NewScheduleTools(mgr)
	tool := tools[0] // create

	result, err := tool.Run(context.Background(), `{"cron":"0 9 * * *"}`)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error for missing prompt")
	}
}

func TestScheduleTool_List(t *testing.T) {
	mgr := setupScheduleMgr(t)
	tools := NewScheduleTools(mgr)
	createTool := tools[0]
	listTool := tools[1]

	if result, err := createTool.Run(context.Background(), `{"cron":"0 9 * * *","prompt":"task one"}`); err != nil {
		t.Fatalf("Create task one: %v", err)
	} else if result.IsError {
		t.Fatalf("Create task one returned error: %s", result.Content)
	}
	if result, err := createTool.Run(context.Background(), `{"cron":"30 */2 * * *","prompt":"task two"}`); err != nil {
		t.Fatalf("Create task two: %v", err)
	} else if result.IsError {
		t.Fatalf("Create task two returned error: %s", result.Content)
	}

	result, err := listTool.Run(context.Background(), `{}`)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected error: %s", result.Content)
	}
	if result.Content == "" {
		t.Error("expected non-empty list")
	}
}

func TestScheduleTool_ListEmpty(t *testing.T) {
	mgr := setupScheduleMgr(t)
	tools := NewScheduleTools(mgr)
	listTool := tools[1]

	result, err := listTool.Run(context.Background(), `{}`)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if result.IsError {
		t.Fatal("expected no error for empty list")
	}
}

func TestScheduleTool_Update(t *testing.T) {
	mgr := setupScheduleMgr(t)
	tools := NewScheduleTools(mgr)
	createTool := tools[0]
	updateTool := tools[2]

	cr, _ := createTool.Run(context.Background(), `{"cron":"0 9 * * *","prompt":"old"}`)
	id := cr.Content[len("Schedule created: "):]

	updateArgs := map[string]any{"id": id, "prompt": "new prompt"}
	updateJSON, _ := json.Marshal(updateArgs)
	result, err := updateTool.Run(context.Background(), string(updateJSON))
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected error: %s", result.Content)
	}

	list, _ := mgr.List()
	if list[0].Prompt != "new prompt" {
		t.Errorf("prompt = %q, want %q", list[0].Prompt, "new prompt")
	}
}

func TestScheduleTool_Remove(t *testing.T) {
	mgr := setupScheduleMgr(t)
	tools := NewScheduleTools(mgr)
	createTool := tools[0]
	removeTool := tools[3]

	cr, _ := createTool.Run(context.Background(), `{"cron":"0 9 * * *","prompt":"temp"}`)
	id := cr.Content[len("Schedule created: "):]

	removeArgs := map[string]any{"id": id}
	removeJSON, _ := json.Marshal(removeArgs)
	result, err := removeTool.Run(context.Background(), string(removeJSON))
	if err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected error: %s", result.Content)
	}

	list, _ := mgr.List()
	if len(list) != 0 {
		t.Fatalf("got %d schedules after remove, want 0", len(list))
	}
}

func TestScheduleTool_RequiresApproval(t *testing.T) {
	mgr := setupScheduleMgr(t)
	tools := NewScheduleTools(mgr)

	// create, update, remove require approval; list does not
	if !tools[0].RequiresApproval() {
		t.Error("create should require approval")
	}
	if tools[1].RequiresApproval() {
		t.Error("list should not require approval")
	}
	if !tools[2].RequiresApproval() {
		t.Error("update should require approval")
	}
	if !tools[3].RequiresApproval() {
		t.Error("remove should require approval")
	}
}

func TestScheduleTool_IsReadOnlyCall(t *testing.T) {
	mgr := setupScheduleMgr(t)
	tools := NewScheduleTools(mgr)

	// create, update, remove are not read-only; list is
	for i, tool := range tools {
		rc, ok := tool.(agent.ReadOnlyChecker)
		if !ok {
			t.Fatalf("tool %d does not implement ReadOnlyChecker", i)
		}
		switch i {
		case 0: // create
			if rc.IsReadOnlyCall("") {
				t.Error("create should not be read-only")
			}
		case 1: // list
			if !rc.IsReadOnlyCall("") {
				t.Error("list should be read-only")
			}
		case 2: // update
			if rc.IsReadOnlyCall("") {
				t.Error("update should not be read-only")
			}
		case 3: // remove
			if rc.IsReadOnlyCall("") {
				t.Error("remove should not be read-only")
			}
		}
	}
}
