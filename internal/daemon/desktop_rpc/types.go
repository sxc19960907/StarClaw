package desktop_rpc

import "encoding/json"

const ProtocolVersion = "1.0.0"

var ProtocolMethods = []string{
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

const (
	MethodSystemPing                = "system.ping"
	MethodSystemCapabilities        = "system.capabilities"
	MethodCalendarListSources       = "calendar.list_sources"
	MethodCalendarListEvents        = "calendar.list_events"
	MethodCalendarGetEvent          = "calendar.get_event"
	MethodCalendarCreateEvent       = "calendar.create_event"
	MethodCalendarUpdateEvent       = "calendar.update_event"
	MethodCalendarDeleteEvent       = "calendar.delete_event"
	MethodCalendarCheckPermission   = "calendar.check_permission"
	MethodCalendarRequestPermission = "calendar.request_permission"
)

const (
	FrameDesktopRPCRequest = "desktop_rpc_request"
	FrameDesktopRPCResult  = "desktop_rpc_result"
	FrameDesktopEvent      = "desktop_event"
	FrameDesktopRPCCancel  = "desktop_rpc_cancel"
)

const (
	ErrCodePermissionDenied        = "calendar_permission_denied"
	ErrCodePermissionNotDetermined = "calendar_permission_not_determined"
	ErrCodeNotFound                = "not_found"
	ErrCodeInvalidArgument         = "invalid_argument"
	ErrCodeReadOnlyCalendar        = "read_only_calendar"
	ErrCodeInternal                = "internal_error"
	ErrCodeTimeout                 = "timeout"
	ErrCodeDesktopDisconnected     = "desktop_disconnected"
)

const (
	PermissionNotDetermined = "not_determined"
	PermissionRestricted    = "restricted"
	PermissionDenied        = "denied"
	PermissionGranted       = "granted"
	PermissionWriteOnly     = "write_only"
)

const (
	AccountTypeICloud       = "icloud"
	AccountTypeGoogle       = "google"
	AccountTypeExchange     = "exchange"
	AccountTypeOutlook      = "outlook"
	AccountTypeLocal        = "local"
	AccountTypeSubscription = "subscription"
	AccountTypeOther        = "other"
)

const (
	AttendeeAccepted    = "accepted"
	AttendeeTentative   = "tentative"
	AttendeeDeclined    = "declined"
	AttendeeNeedsAction = "needs_action"
)

const (
	ScopeThis          = "this"
	ScopeThisAndFuture = "this_and_future"
	ScopeAll           = "all"
)

const (
	EventDesktopOnline             = "desktop_online"
	EventDesktopOffline            = "desktop_offline"
	EventCalendarPermissionChanged = "calendar_permission_changed"
	EventCalendarDataChanged       = "calendar_data_changed"
)

const (
	FrequencyDaily   = "daily"
	FrequencyWeekly  = "weekly"
	FrequencyMonthly = "monthly"
	FrequencyYearly  = "yearly"
)

const MaxFrameBodyBytes = 4 * 1024 * 1024

type Frame struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type RPCRequest struct {
	RequestID string          `json:"request_id"`
	Method    string          `json:"method"`
	Params    json.RawMessage `json:"params,omitempty"`
	TimeoutMs int             `json:"timeout_ms,omitempty"`
	SessionID string          `json:"session_id,omitempty"`
	Agent     string          `json:"agent,omitempty"`
	Source    string          `json:"source,omitempty"`
	TS        string          `json:"ts,omitempty"`
}

type RPCResult struct {
	RequestID string          `json:"request_id"`
	OK        bool            `json:"ok"`
	Result    json.RawMessage `json:"result,omitempty"`
	Error     *RPCError       `json:"error,omitempty"`
}

type RPCError struct {
	Code      string          `json:"code"`
	Message   string          `json:"message"`
	Retriable bool            `json:"retriable"`
	Details   json.RawMessage `json:"details,omitempty"`
}

type DesktopEvent struct {
	Event string          `json:"event"`
	Data  json.RawMessage `json:"data,omitempty"`
	TS    string          `json:"ts,omitempty"`
}

type SystemCapabilitiesResult struct {
	Version  string   `json:"version"`
	Methods  []string `json:"methods"`
	Platform Platform `json:"platform"`
}

type Platform struct {
	OS         string `json:"os"`
	OSVersion  string `json:"os_version,omitempty"`
	AppVersion string `json:"app_version"`
}

type SystemPingParams struct {
	Echo string `json:"echo,omitempty"`
}

type SystemPingResult struct {
	Pong       string `json:"pong"`
	ServerTime string `json:"server_time"`
}

type CalendarCheckPermissionResult struct {
	Status string `json:"status"`
}

type CalendarRequestPermissionResult struct {
	Status string `json:"status"`
}

type CalendarListSourcesResult struct {
	Sources []CalendarSource `json:"sources"`
}

type CalendarSource struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	AccountType string `json:"account_type"`
	ColorHex    string `json:"color_hex,omitempty"`
	Writable    bool   `json:"writable"`
}

type CalendarListEventsParams struct {
	Start       string    `json:"start"`
	End         string    `json:"end"`
	CalendarIDs *[]string `json:"calendar_ids,omitempty"`
	Query       *string   `json:"query,omitempty"`
	Limit       int       `json:"limit,omitempty"`
}

type CalendarListEventsResult struct {
	Events    []CalendarEvent `json:"events"`
	Truncated bool            `json:"truncated"`
}

type CalendarGetEventParams struct {
	ID string `json:"id"`
}

type CalendarEvent struct {
	ID             string              `json:"id"`
	CalendarID     string              `json:"calendar_id"`
	Title          string              `json:"title"`
	Start          string              `json:"start"`
	End            string              `json:"end"`
	AllDay         bool                `json:"all_day"`
	Location       string              `json:"location,omitempty"`
	Notes          string              `json:"notes,omitempty"`
	URL            string              `json:"url,omitempty"`
	Attendees      []CalendarAttendee  `json:"attendees,omitempty"`
	Alarms         []CalendarAlarm     `json:"alarms,omitempty"`
	RecurrenceRule *CalendarRecurrence `json:"recurrence_rule,omitempty"`
	SeriesMasterID string              `json:"series_master_id,omitempty"`
}

type CalendarAttendee struct {
	Email  string `json:"email,omitempty"`
	Name   string `json:"name,omitempty"`
	Status string `json:"status,omitempty"`
}

type CalendarAlarm struct {
	MinutesBefore int `json:"minutes_before"`
}

type CalendarRecurrence struct {
	Frequency       string   `json:"frequency,omitempty"`
	Interval        int      `json:"interval,omitempty"`
	ByDay           []string `json:"by_day,omitempty"`
	EndDate         string   `json:"end_date,omitempty"`
	OccurrenceCount int      `json:"occurrence_count,omitempty"`
	RawRRule        string   `json:"raw_rrule,omitempty"`
}

type CalendarEventPayload struct {
	CalendarID     *string             `json:"calendar_id,omitempty"`
	Title          string              `json:"title,omitempty"`
	Start          string              `json:"start,omitempty"`
	End            string              `json:"end,omitempty"`
	AllDay         bool                `json:"all_day,omitempty"`
	Location       string              `json:"location,omitempty"`
	Notes          string              `json:"notes,omitempty"`
	URL            string              `json:"url,omitempty"`
	Attendees      []CalendarAttendee  `json:"attendees,omitempty"`
	Alarms         []CalendarAlarm     `json:"alarms,omitempty"`
	RecurrenceRule *CalendarRecurrence `json:"recurrence_rule,omitempty"`
}

type CalendarCreateEventParams = CalendarEventPayload

type CalendarMutationResult struct {
	ID                string `json:"id"`
	PendingRemoteSync bool   `json:"pending_remote_sync"`
	InvitationsSent   bool   `json:"invitations_sent,omitempty"`
}

type CalendarUpdateEventParams struct {
	ID              string                     `json:"id"`
	Scope           string                     `json:"scope"`
	Patch           *CalendarEventPatchPayload `json:"patch,omitempty"`
	ClearRecurrence bool                       `json:"clear_recurrence,omitempty"`
}

type CalendarEventPatchPayload struct {
	CalendarID     *string             `json:"calendar_id,omitempty"`
	Title          *string             `json:"title,omitempty"`
	Start          *string             `json:"start,omitempty"`
	End            *string             `json:"end,omitempty"`
	AllDay         *bool               `json:"all_day,omitempty"`
	Location       *string             `json:"location,omitempty"`
	Notes          *string             `json:"notes,omitempty"`
	URL            *string             `json:"url,omitempty"`
	Attendees      *[]CalendarAttendee `json:"attendees,omitempty"`
	Alarms         *[]CalendarAlarm    `json:"alarms,omitempty"`
	RecurrenceRule *CalendarRecurrence `json:"recurrence_rule,omitempty"`
}

type CalendarDeleteEventParams struct {
	ID    string `json:"id"`
	Scope string `json:"scope"`
}

type CalendarDeleteEventResult struct {
	ID                string `json:"id"`
	Deleted           bool   `json:"deleted"`
	PendingRemoteSync bool   `json:"pending_remote_sync"`
}
