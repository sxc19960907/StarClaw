package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/starclaw/starclaw/internal/agent"
	"github.com/starclaw/starclaw/internal/daemon/desktop_rpc"
)

func fakeCalendarBroker(t *testing.T, handler func(req *desktop_rpc.RPCRequest) *desktop_rpc.RPCResult) *desktop_rpc.Broker {
	t.Helper()
	b := desktop_rpc.NewBroker()
	b.SetSendFn(func(req *desktop_rpc.RPCRequest) error {
		go func() {
			res := handler(req)
			res.RequestID = req.RequestID
			b.Resolve(req.RequestID, res)
		}()
		return nil
	})
	return b
}

func calendarSuccess(t *testing.T, body any) *desktop_rpc.RPCResult {
	t.Helper()
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return &desktop_rpc.RPCResult{OK: true, Result: raw}
}

func calendarError(code, message string) *desktop_rpc.RPCResult {
	return &desktop_rpc.RPCResult{
		OK: false,
		Error: &desktop_rpc.RPCError{
			Code:    code,
			Message: message,
		},
	}
}

func calendarToolByName(t *testing.T, broker *desktop_rpc.Broker, name string) agent.Tool {
	t.Helper()
	for _, tool := range NewCalendarTools(broker) {
		if tool.Info().Name == name {
			return tool
		}
	}
	t.Fatalf("calendar tool %q not found", name)
	return nil
}

func TestRegisterLocalToolsDoesNotExposeCalendarWithoutBroker(t *testing.T) {
	t.Parallel()
	reg := RegisterLocalTools()
	if _, ok := reg.Get("calendar_list_events"); ok {
		t.Fatal("calendar tools should not be registered without a Desktop RPC broker")
	}
}

func TestRegisterCalendarToolsWithBroker(t *testing.T) {
	t.Parallel()
	reg := agent.NewToolRegistry()
	RegisterCalendarTools(reg, desktop_rpc.NewBroker())
	for _, name := range []string{
		"calendar_check_permission",
		"calendar_request_permission",
		"calendar_list_sources",
		"calendar_list_events",
		"calendar_get_event",
		"calendar_create_event",
		"calendar_update_event",
		"calendar_delete_event",
	} {
		if _, ok := reg.Get(name); !ok {
			t.Fatalf("expected %s to be registered", name)
		}
	}
}

func TestCalendarToolNoBroker(t *testing.T) {
	t.Parallel()
	tool := calendarToolByName(t, nil, "calendar_check_permission")
	res, err := tool.Run(context.Background(), `{}`)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !res.IsError || !strings.Contains(res.Content, "Desktop") {
		t.Fatalf("result = %#v, want Desktop error", res)
	}
}

func TestCalendarCheckPermissionRoutesMethod(t *testing.T) {
	t.Parallel()
	var gotMethod string
	tool := calendarToolByName(t, fakeCalendarBroker(t, func(req *desktop_rpc.RPCRequest) *desktop_rpc.RPCResult {
		gotMethod = req.Method
		return calendarSuccess(t, desktop_rpc.CalendarCheckPermissionResult{Status: desktop_rpc.PermissionGranted})
	}), "calendar_check_permission")
	res, err := tool.Run(context.Background(), `{}`)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.IsError {
		t.Fatalf("unexpected error: %s", res.Content)
	}
	if gotMethod != desktop_rpc.MethodCalendarCheckPermission {
		t.Fatalf("method = %q, want %q", gotMethod, desktop_rpc.MethodCalendarCheckPermission)
	}
	if !strings.Contains(res.Content, desktop_rpc.PermissionGranted) {
		t.Fatalf("content = %s", res.Content)
	}
}

func TestCalendarRequestPermissionUsesExtendedTimeout(t *testing.T) {
	t.Parallel()
	var gotTimeout int
	tool := calendarToolByName(t, fakeCalendarBroker(t, func(req *desktop_rpc.RPCRequest) *desktop_rpc.RPCResult {
		gotTimeout = req.TimeoutMs
		return calendarSuccess(t, desktop_rpc.CalendarRequestPermissionResult{Status: desktop_rpc.PermissionGranted})
	}), "calendar_request_permission")
	res, err := tool.Run(context.Background(), `{"description":"ask for calendar access"}`)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.IsError {
		t.Fatalf("unexpected error: %s", res.Content)
	}
	if gotTimeout != calendarRequestPermissionTimeoutMs {
		t.Fatalf("timeout = %d, want %d", gotTimeout, calendarRequestPermissionTimeoutMs)
	}
}

func TestCalendarListEventsValidationAndLimitClamp(t *testing.T) {
	t.Parallel()
	tool := calendarToolByName(t, fakeCalendarBroker(t, func(req *desktop_rpc.RPCRequest) *desktop_rpc.RPCResult {
		var params desktop_rpc.CalendarListEventsParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			t.Fatalf("params: %v", err)
		}
		if params.Limit != 2000 {
			t.Fatalf("limit = %d, want 2000", params.Limit)
		}
		return calendarSuccess(t, desktop_rpc.CalendarListEventsResult{Events: []desktop_rpc.CalendarEvent{}, Truncated: false})
	}), "calendar_list_events")

	bad, err := tool.Run(context.Background(), `{"start":"2026-06-08T09:00:00","end":"2026-06-08T10:00:00+08:00"}`)
	if err != nil {
		t.Fatalf("bad Run: %v", err)
	}
	if !bad.IsError || !strings.Contains(bad.Content, "start") {
		t.Fatalf("bad result = %#v, want start validation", bad)
	}

	res, err := tool.Run(context.Background(), `{"start":"2026-06-08T09:00:00+08:00","end":"2026-06-08T10:00:00+08:00","limit":9999}`)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.IsError {
		t.Fatalf("unexpected error: %s", res.Content)
	}
}

func TestCalendarCreateStripsDescription(t *testing.T) {
	t.Parallel()
	var gotParams string
	tool := calendarToolByName(t, fakeCalendarBroker(t, func(req *desktop_rpc.RPCRequest) *desktop_rpc.RPCResult {
		gotParams = string(req.Params)
		return calendarSuccess(t, desktop_rpc.CalendarMutationResult{ID: "evt_1", PendingRemoteSync: true})
	}), "calendar_create_event")
	res, err := tool.Run(context.Background(), `{"title":"Planning","start":"2026-06-08T09:00:00+08:00","end":"2026-06-08T10:00:00+08:00","description":"create event"}`)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.IsError {
		t.Fatalf("unexpected error: %s", res.Content)
	}
	if strings.Contains(gotParams, "description") {
		t.Fatalf("description leaked to Desktop RPC params: %s", gotParams)
	}
	if !strings.Contains(gotParams, "Planning") {
		t.Fatalf("event fields missing from params: %s", gotParams)
	}
}

func TestCalendarUpdateRejectsScopeAll(t *testing.T) {
	t.Parallel()
	tool := calendarToolByName(t, desktop_rpc.NewBroker(), "calendar_update_event")
	res, err := tool.Run(context.Background(), `{"id":"evt_1","scope":"all","description":"update"}`)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !res.IsError || !strings.Contains(res.Content, "scope") {
		t.Fatalf("result = %#v, want scope validation", res)
	}
}

func TestCalendarDeleteAcceptsScopeAllAndStripsDescription(t *testing.T) {
	t.Parallel()
	var gotParams string
	tool := calendarToolByName(t, fakeCalendarBroker(t, func(req *desktop_rpc.RPCRequest) *desktop_rpc.RPCResult {
		gotParams = string(req.Params)
		if req.Method != desktop_rpc.MethodCalendarDeleteEvent {
			t.Fatalf("method = %q", req.Method)
		}
		return calendarSuccess(t, desktop_rpc.CalendarDeleteEventResult{ID: "evt_1", Deleted: true})
	}), "calendar_delete_event")
	res, err := tool.Run(context.Background(), `{"id":"evt_1","scope":"all","description":"delete series"}`)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.IsError {
		t.Fatalf("unexpected error: %s", res.Content)
	}
	if strings.Contains(gotParams, "description") {
		t.Fatalf("description leaked to Desktop RPC params: %s", gotParams)
	}
	if !strings.Contains(gotParams, `"scope":"all"`) {
		t.Fatalf("scope missing from params: %s", gotParams)
	}
}

func TestCalendarRPCErrorMapping(t *testing.T) {
	t.Parallel()
	tool := calendarToolByName(t, fakeCalendarBroker(t, func(req *desktop_rpc.RPCRequest) *desktop_rpc.RPCResult {
		return calendarError(desktop_rpc.ErrCodePermissionDenied, "denied")
	}), "calendar_list_sources")
	res, err := tool.Run(context.Background(), `{}`)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !res.IsError || res.ErrorCategory != agent.ErrCategoryPermission {
		t.Fatalf("result = %#v, want permission error", res)
	}
}

func TestCalendarToolApprovalAndReadOnly(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		wantApprove bool
		wantRead    bool
	}{
		{"calendar_check_permission", false, true},
		{"calendar_request_permission", true, false},
		{"calendar_list_sources", false, true},
		{"calendar_list_events", false, true},
		{"calendar_get_event", false, true},
		{"calendar_create_event", true, false},
		{"calendar_update_event", true, false},
		{"calendar_delete_event", true, false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			tool := calendarToolByName(t, desktop_rpc.NewBroker(), tt.name)
			if tool.RequiresApproval() != tt.wantApprove {
				t.Fatalf("RequiresApproval = %v, want %v", tool.RequiresApproval(), tt.wantApprove)
			}
			checker, ok := tool.(agent.ReadOnlyChecker)
			if !ok {
				t.Fatalf("%s does not implement ReadOnlyChecker", tt.name)
			}
			if checker.IsReadOnlyCall(`{}`) != tt.wantRead {
				t.Fatalf("IsReadOnlyCall = %v, want %v", checker.IsReadOnlyCall(`{}`), tt.wantRead)
			}
		})
	}
}
