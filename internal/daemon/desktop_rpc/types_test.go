package desktop_rpc

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestProtocolMethodsIncludeCalendarV1(t *testing.T) {
	t.Parallel()
	want := []string{
		MethodSystemPing,
		MethodSystemCapabilities,
		MethodCalendarListSources,
		MethodCalendarListEvents,
		MethodCalendarGetEvent,
		MethodCalendarCreateEvent,
		MethodCalendarUpdateEvent,
		MethodCalendarDeleteEvent,
		MethodCalendarCheckPermission,
		MethodCalendarRequestPermission,
	}
	if !reflect.DeepEqual(ProtocolMethods, want) {
		t.Fatalf("ProtocolMethods = %#v, want %#v", ProtocolMethods, want)
	}
}

func TestCalendarProtocolConstants(t *testing.T) {
	t.Parallel()
	tests := map[string]string{
		"method list sources":       MethodCalendarListSources,
		"method list events":        MethodCalendarListEvents,
		"method get event":          MethodCalendarGetEvent,
		"method create event":       MethodCalendarCreateEvent,
		"method update event":       MethodCalendarUpdateEvent,
		"method delete event":       MethodCalendarDeleteEvent,
		"method check permission":   MethodCalendarCheckPermission,
		"method request permission": MethodCalendarRequestPermission,
		"cancel frame":              FrameDesktopRPCCancel,
		"permission denied error":   ErrCodePermissionDenied,
		"permission unknown error":  ErrCodePermissionNotDetermined,
		"not found error":           ErrCodeNotFound,
		"read only error":           ErrCodeReadOnlyCalendar,
		"permission granted":        PermissionGranted,
		"permission write only":     PermissionWriteOnly,
		"account subscription":      AccountTypeSubscription,
		"attendee needs action":     AttendeeNeedsAction,
		"scope this and future":     ScopeThisAndFuture,
		"desktop offline":           EventDesktopOffline,
		"calendar changed":          EventCalendarDataChanged,
		"frequency yearly":          FrequencyYearly,
	}
	want := map[string]string{
		"method list sources":       "calendar.list_sources",
		"method list events":        "calendar.list_events",
		"method get event":          "calendar.get_event",
		"method create event":       "calendar.create_event",
		"method update event":       "calendar.update_event",
		"method delete event":       "calendar.delete_event",
		"method check permission":   "calendar.check_permission",
		"method request permission": "calendar.request_permission",
		"cancel frame":              "desktop_rpc_cancel",
		"permission denied error":   "calendar_permission_denied",
		"permission unknown error":  "calendar_permission_not_determined",
		"not found error":           "not_found",
		"read only error":           "read_only_calendar",
		"permission granted":        "granted",
		"permission write only":     "write_only",
		"account subscription":      "subscription",
		"attendee needs action":     "needs_action",
		"scope this and future":     "this_and_future",
		"desktop offline":           "desktop_offline",
		"calendar changed":          "calendar_data_changed",
		"frequency yearly":          "yearly",
	}
	for name, got := range tests {
		if got != want[name] {
			t.Fatalf("%s = %q, want %q", name, got, want[name])
		}
	}
}

func TestCalendarPayloadJSONShape(t *testing.T) {
	t.Parallel()
	calendarID := "cal_work"
	params := CalendarCreateEventParams{
		CalendarID: &calendarID,
		Title:      "Planning",
		Start:      "2026-06-08T09:00:00+08:00",
		End:        "2026-06-08T09:30:00+08:00",
		Attendees: []CalendarAttendee{{
			Email:  "a@example.com",
			Name:   "A",
			Status: AttendeeNeedsAction,
		}},
		Alarms: []CalendarAlarm{{MinutesBefore: 10}},
		RecurrenceRule: &CalendarRecurrence{
			Frequency: FrequencyWeekly,
			Interval:  1,
			ByDay:     []string{"MO"},
		},
	}
	raw, err := json.Marshal(params)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("Unmarshal object: %v", err)
	}
	for _, key := range []string{"calendar_id", "title", "start", "end", "attendees", "alarms", "recurrence_rule"} {
		if _, ok := got[key]; !ok {
			t.Fatalf("encoded payload missing key %q: %s", key, raw)
		}
	}
	if _, ok := got["description"]; ok {
		t.Fatalf("protocol payload must not encode approval description: %s", raw)
	}
}

func TestCalendarUpdatePayloadRoundTrip(t *testing.T) {
	t.Parallel()
	raw := []byte(`{
		"id":"evt_1",
		"scope":"this_and_future",
		"patch":{"title":"New title","attendees":[]},
		"clear_recurrence":true
	}`)
	var params CalendarUpdateEventParams
	if err := json.Unmarshal(raw, &params); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if params.ID != "evt_1" || params.Scope != ScopeThisAndFuture || params.Patch == nil {
		t.Fatalf("params = %#v", params)
	}
	if params.Patch.Title == nil || *params.Patch.Title != "New title" {
		t.Fatalf("patch title = %#v", params.Patch.Title)
	}
	if params.Patch.Attendees == nil || len(*params.Patch.Attendees) != 0 {
		t.Fatalf("empty attendees array should round-trip as explicit empty slice: %#v", params.Patch.Attendees)
	}
	if !params.ClearRecurrence {
		t.Fatal("clear_recurrence did not decode")
	}
}

func TestCalendarPatchPayloadPreservesExplicitClears(t *testing.T) {
	t.Parallel()
	empty := ""
	attendees := []CalendarAttendee{}
	payload := CalendarEventPatchPayload{
		Location:  &empty,
		Attendees: &attendees,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("Unmarshal object: %v", err)
	}
	if v, ok := got["location"]; !ok || v != "" {
		t.Fatalf("location explicit clear not preserved: %s", raw)
	}
	if v, ok := got["attendees"].([]any); !ok || len(v) != 0 {
		t.Fatalf("attendees explicit clear not preserved: %s", raw)
	}
	if _, ok := got["title"]; ok {
		t.Fatalf("nil patch fields should stay omitted: %s", raw)
	}
}
