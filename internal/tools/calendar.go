package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/starclaw/starclaw/internal/agent"
	"github.com/starclaw/starclaw/internal/daemon/desktop_rpc"
)

const calendarRequestPermissionTimeoutMs = 5 * 60 * 1000

type calendarTool struct {
	broker *desktop_rpc.Broker
	name   string
}

func NewCalendarTools(broker *desktop_rpc.Broker) []agent.Tool {
	return []agent.Tool{
		&calendarTool{broker: broker, name: "calendar_check_permission"},
		&calendarTool{broker: broker, name: "calendar_request_permission"},
		&calendarTool{broker: broker, name: "calendar_list_sources"},
		&calendarTool{broker: broker, name: "calendar_list_events"},
		&calendarTool{broker: broker, name: "calendar_get_event"},
		&calendarTool{broker: broker, name: "calendar_create_event"},
		&calendarTool{broker: broker, name: "calendar_update_event"},
		&calendarTool{broker: broker, name: "calendar_delete_event"},
	}
}

func (t *calendarTool) Info() agent.ToolInfo {
	descriptionField := map[string]any{
		"type":        "string",
		"description": "User-facing reason shown in the approval request.",
	}
	switch t.name {
	case "calendar_check_permission":
		return agent.ToolInfo{
			Name:        t.name,
			Description: "Check whether the user has granted StarClaw Desktop calendar access. Returns one of: not_determined, restricted, denied, granted, write_only.",
			Parameters:  emptyObjectSchema(),
		}
	case "calendar_request_permission":
		return agent.ToolInfo{
			Name:        t.name,
			Description: "Request macOS calendar access from the user through StarClaw Desktop. Use only after calendar_check_permission returns not_determined.",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"description": descriptionField,
				},
			},
			Required: []string{"description"},
		}
	case "calendar_list_sources":
		return agent.ToolInfo{
			Name:        t.name,
			Description: "List calendar sources configured in the local Desktop calendar account provider. Read-only.",
			Parameters:  emptyObjectSchema(),
		}
	case "calendar_list_events":
		return agent.ToolInfo{
			Name:        t.name,
			Description: "List calendar events in a time window. Times must be RFC3339 with timezone offset. Optional limit defaults to 500 and is capped at 2000.",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"start":        map[string]any{"type": "string", "description": "Window start, RFC3339 with timezone offset"},
					"end":          map[string]any{"type": "string", "description": "Window end, RFC3339 with timezone offset"},
					"calendar_ids": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
					"query":        map[string]any{"type": "string"},
					"limit":        map[string]any{"type": "integer"},
				},
			},
			Required: []string{"start", "end"},
		}
	case "calendar_get_event":
		return agent.ToolInfo{
			Name:        t.name,
			Description: "Fetch full detail for one calendar event by id. Read-only.",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"id": map[string]any{"type": "string"},
				},
			},
			Required: []string{"id"},
		}
	case "calendar_create_event":
		return agent.ToolInfo{
			Name:        t.name,
			Description: "Create a calendar event through StarClaw Desktop. Requires approval. Invitations are not sent by v1 Desktop RPC.",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"calendar_id":     map[string]any{"type": "string"},
					"title":           map[string]any{"type": "string"},
					"start":           map[string]any{"type": "string"},
					"end":             map[string]any{"type": "string"},
					"all_day":         map[string]any{"type": "boolean"},
					"location":        map[string]any{"type": "string"},
					"notes":           map[string]any{"type": "string"},
					"url":             map[string]any{"type": "string"},
					"attendees":       map[string]any{"type": "array", "items": map[string]any{"type": "object"}},
					"alarms":          map[string]any{"type": "array", "items": map[string]any{"type": "object"}},
					"recurrence_rule": map[string]any{"type": "object"},
					"description":     descriptionField,
				},
			},
			Required: []string{"title", "start", "end", "description"},
		}
	case "calendar_update_event":
		return agent.ToolInfo{
			Name:        t.name,
			Description: "Update a calendar event through StarClaw Desktop. Requires approval. scope must be this or this_and_future.",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"id":               map[string]any{"type": "string"},
					"scope":            map[string]any{"type": "string", "enum": []string{desktop_rpc.ScopeThis, desktop_rpc.ScopeThisAndFuture}},
					"patch":            map[string]any{"type": "object"},
					"clear_recurrence": map[string]any{"type": "boolean"},
					"description":      descriptionField,
				},
			},
			Required: []string{"id", "scope", "description"},
		}
	case "calendar_delete_event":
		return agent.ToolInfo{
			Name:        t.name,
			Description: "Delete a calendar event through StarClaw Desktop. Requires approval. scope may be this, this_and_future, or all.",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"id":          map[string]any{"type": "string"},
					"scope":       map[string]any{"type": "string", "enum": []string{desktop_rpc.ScopeThis, desktop_rpc.ScopeThisAndFuture, desktop_rpc.ScopeAll}},
					"description": descriptionField,
				},
			},
			Required: []string{"id", "scope", "description"},
		}
	default:
		return agent.ToolInfo{Name: t.name, Parameters: emptyObjectSchema()}
	}
}

func (t *calendarTool) Run(ctx context.Context, argsJSON string) (agent.ToolResult, error) {
	switch t.name {
	case "calendar_check_permission":
		return callCalendarRPC(ctx, t.broker, desktop_rpc.MethodCalendarCheckPermission, struct{}{}, 0)
	case "calendar_request_permission":
		var args struct {
			Description string `json:"description"`
		}
		if res := decodeCalendarArgs(argsJSON, &args); res != nil {
			return *res, nil
		}
		if args.Description == "" {
			return agent.ValidationError("calendar_request_permission: missing required `description` parameter"), nil
		}
		return callCalendarRPC(ctx, t.broker, desktop_rpc.MethodCalendarRequestPermission, struct{}{}, calendarRequestPermissionTimeoutMs)
	case "calendar_list_sources":
		return callCalendarRPC(ctx, t.broker, desktop_rpc.MethodCalendarListSources, struct{}{}, 0)
	case "calendar_list_events":
		var args desktop_rpc.CalendarListEventsParams
		if res := decodeCalendarArgs(argsJSON, &args); res != nil {
			return *res, nil
		}
		if args.Start == "" {
			return agent.ValidationError("calendar_list_events: missing required `start` parameter"), nil
		}
		if args.End == "" {
			return agent.ValidationError("calendar_list_events: missing required `end` parameter"), nil
		}
		if err := validateCalendarRFC3339(args.Start); err != nil {
			return agent.ValidationError(fmt.Sprintf("calendar_list_events: invalid `start`: %v", err)), nil
		}
		if err := validateCalendarRFC3339(args.End); err != nil {
			return agent.ValidationError(fmt.Sprintf("calendar_list_events: invalid `end`: %v", err)), nil
		}
		args.Limit = clampCalendarLimit(args.Limit)
		return callCalendarRPC(ctx, t.broker, desktop_rpc.MethodCalendarListEvents, args, 0)
	case "calendar_get_event":
		var args desktop_rpc.CalendarGetEventParams
		if res := decodeCalendarArgs(argsJSON, &args); res != nil {
			return *res, nil
		}
		if args.ID == "" {
			return agent.ValidationError("calendar_get_event: missing required `id` parameter"), nil
		}
		return callCalendarRPC(ctx, t.broker, desktop_rpc.MethodCalendarGetEvent, args, 0)
	case "calendar_create_event":
		var args struct {
			Title       string `json:"title"`
			Start       string `json:"start"`
			End         string `json:"end"`
			Description string `json:"description"`
		}
		if res := decodeCalendarArgs(argsJSON, &args); res != nil {
			return *res, nil
		}
		if args.Title == "" {
			return agent.ValidationError("calendar_create_event: missing required `title` parameter"), nil
		}
		if args.Start == "" {
			return agent.ValidationError("calendar_create_event: missing required `start` parameter"), nil
		}
		if args.End == "" {
			return agent.ValidationError("calendar_create_event: missing required `end` parameter"), nil
		}
		if args.Description == "" {
			return agent.ValidationError("calendar_create_event: missing required `description` parameter"), nil
		}
		if err := validateCalendarRFC3339(args.Start); err != nil {
			return agent.ValidationError(fmt.Sprintf("calendar_create_event: invalid `start`: %v", err)), nil
		}
		if err := validateCalendarRFC3339(args.End); err != nil {
			return agent.ValidationError(fmt.Sprintf("calendar_create_event: invalid `end`: %v", err)), nil
		}
		return callCalendarRPCRaw(ctx, t.broker, desktop_rpc.MethodCalendarCreateEvent, stripCalendarDescription([]byte(argsJSON)), 0)
	case "calendar_update_event":
		var args struct {
			ID          string `json:"id"`
			Scope       string `json:"scope"`
			Description string `json:"description"`
		}
		if res := decodeCalendarArgs(argsJSON, &args); res != nil {
			return *res, nil
		}
		if args.ID == "" {
			return agent.ValidationError("calendar_update_event: missing required `id` parameter"), nil
		}
		if args.Description == "" {
			return agent.ValidationError("calendar_update_event: missing required `description` parameter"), nil
		}
		switch args.Scope {
		case desktop_rpc.ScopeThis, desktop_rpc.ScopeThisAndFuture:
		case "":
			return agent.ValidationError("calendar_update_event: missing required `scope` parameter"), nil
		case desktop_rpc.ScopeAll:
			return agent.ValidationError("calendar_update_event: `scope` cannot be `all`; use delete + create instead"), nil
		default:
			return agent.ValidationError("calendar_update_event: `scope` must be `this` or `this_and_future`"), nil
		}
		return callCalendarRPCRaw(ctx, t.broker, desktop_rpc.MethodCalendarUpdateEvent, stripCalendarDescription([]byte(argsJSON)), 0)
	case "calendar_delete_event":
		var args struct {
			ID          string `json:"id"`
			Scope       string `json:"scope"`
			Description string `json:"description"`
		}
		if res := decodeCalendarArgs(argsJSON, &args); res != nil {
			return *res, nil
		}
		if args.ID == "" {
			return agent.ValidationError("calendar_delete_event: missing required `id` parameter"), nil
		}
		if args.Description == "" {
			return agent.ValidationError("calendar_delete_event: missing required `description` parameter"), nil
		}
		switch args.Scope {
		case desktop_rpc.ScopeThis, desktop_rpc.ScopeThisAndFuture, desktop_rpc.ScopeAll:
		case "":
			return agent.ValidationError("calendar_delete_event: missing required `scope` parameter"), nil
		default:
			return agent.ValidationError("calendar_delete_event: `scope` must be `this`, `this_and_future`, or `all`"), nil
		}
		return callCalendarRPCRaw(ctx, t.broker, desktop_rpc.MethodCalendarDeleteEvent, stripCalendarDescription([]byte(argsJSON)), 0)
	default:
		return agent.ValidationError(fmt.Sprintf("unknown calendar tool %q", t.name)), nil
	}
}

func (t *calendarTool) RequiresApproval() bool {
	switch t.name {
	case "calendar_request_permission", "calendar_create_event", "calendar_update_event", "calendar_delete_event":
		return true
	default:
		return false
	}
}

func (t *calendarTool) IsReadOnlyCall(string) bool {
	switch t.name {
	case "calendar_check_permission", "calendar_list_sources", "calendar_list_events", "calendar_get_event":
		return true
	default:
		return false
	}
}

func emptyObjectSchema() map[string]any {
	return map[string]any{"type": "object", "properties": map[string]any{}}
}

func decodeCalendarArgs(argsJSON string, target any) *agent.ToolResult {
	if err := json.Unmarshal([]byte(argsJSON), target); err != nil {
		res := agent.ValidationError("invalid arguments: " + err.Error())
		return &res
	}
	return nil
}

func callCalendarRPC(ctx context.Context, broker *desktop_rpc.Broker, method string, params any, timeoutMs int) (agent.ToolResult, error) {
	raw, err := json.Marshal(params)
	if err != nil {
		return agent.ValidationError(fmt.Sprintf("calendar: invalid RPC params: %v", err)), nil
	}
	return callCalendarRPCRaw(ctx, broker, method, raw, timeoutMs)
}

func callCalendarRPCRaw(ctx context.Context, broker *desktop_rpc.Broker, method string, raw json.RawMessage, timeoutMs int) (agent.ToolResult, error) {
	if broker == nil {
		return agent.ToolResult{Content: "Calendar tools require StarClaw Desktop to be running and connected.", IsError: true}, nil
	}
	res, err := broker.Request(ctx, &desktop_rpc.RPCRequest{
		Method:    method,
		Params:    raw,
		TimeoutMs: timeoutMs,
		TS:        time.Now().UTC().Format(time.RFC3339),
	})
	if err != nil {
		if errors.Is(err, desktop_rpc.ErrNotConnected) {
			return agent.ToolResult{Content: "Calendar tools require StarClaw Desktop to be running and connected.", IsError: true}, nil
		}
		return agent.ToolResult{Content: fmt.Sprintf("calendar: Desktop RPC error: %v", err), IsError: true}, nil
	}
	if !res.OK {
		return mapCalendarRPCError(res.Error), nil
	}
	return agent.ToolResult{Content: string(res.Result)}, nil
}

func mapCalendarRPCError(e *desktop_rpc.RPCError) agent.ToolResult {
	if e == nil {
		return agent.BusinessError("calendar: unspecified Desktop RPC error")
	}
	switch e.Code {
	case desktop_rpc.ErrCodePermissionDenied:
		return agent.PermissionError("Calendar access was denied. Open StarClaw Desktop settings and grant Calendar permission.")
	case desktop_rpc.ErrCodePermissionNotDetermined:
		return agent.PermissionError("Calendar permission has not been requested yet. Call calendar_request_permission first.")
	case desktop_rpc.ErrCodeNotFound:
		return agent.BusinessError("Calendar event or source not found.")
	case desktop_rpc.ErrCodeInvalidArgument:
		return agent.ValidationError("calendar: invalid argument: " + e.Message)
	case desktop_rpc.ErrCodeReadOnlyCalendar:
		return agent.PermissionError("The target calendar is read-only. Choose a writable calendar.")
	case desktop_rpc.ErrCodeTimeout:
		return agent.TransientError("Calendar RPC timed out: " + e.Message)
	case desktop_rpc.ErrCodeDesktopDisconnected:
		return agent.TransientError("StarClaw Desktop is not connected.")
	case desktop_rpc.ErrCodeInternal:
		return agent.BusinessError("StarClaw Desktop reported an internal calendar error: " + e.Message)
	default:
		return agent.BusinessError(fmt.Sprintf("calendar: %s: %s", e.Code, e.Message))
	}
}

func validateCalendarRFC3339(value string) error {
	if _, err := time.Parse(time.RFC3339Nano, value); err != nil {
		return fmt.Errorf("not a valid RFC3339 timestamp with timezone: %q", value)
	}
	return nil
}

func clampCalendarLimit(limit int) int {
	switch {
	case limit <= 0:
		return 500
	case limit > 2000:
		return 2000
	default:
		return limit
	}
}

func stripCalendarDescription(raw []byte) json.RawMessage {
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(raw, &obj); err != nil {
		return raw
	}
	delete(obj, "description")
	out, err := json.Marshal(obj)
	if err != nil {
		return raw
	}
	return out
}
