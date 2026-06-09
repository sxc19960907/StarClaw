package daemon

import (
	"sync"
	"time"

	"github.com/starclaw/starclaw/internal/daemon/desktop_rpc"
)

const defaultDesktopEventMonitorLimit = 32

type DesktopEventMonitor struct {
	mu     sync.Mutex
	limit  int
	events []desktop_rpc.DesktopEvent
}

func NewDesktopEventMonitor(limit int) *DesktopEventMonitor {
	if limit <= 0 {
		limit = defaultDesktopEventMonitorLimit
	}
	return &DesktopEventMonitor{limit: limit}
}

func (m *DesktopEventMonitor) Record(evt *desktop_rpc.DesktopEvent) {
	if m == nil || evt == nil {
		return
	}
	copyEvt := *evt
	if copyEvt.TS == "" {
		copyEvt.TS = time.Now().UTC().Format(time.RFC3339Nano)
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = append(m.events, copyEvt)
	if overflow := len(m.events) - m.limit; overflow > 0 {
		copy(m.events, m.events[overflow:])
		m.events = m.events[:m.limit]
	}
}

func (m *DesktopEventMonitor) Status() desktop_rpc.EventStatus {
	if m == nil {
		return desktop_rpc.EventStatus{}
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	status := desktop_rpc.EventStatus{Retained: len(m.events)}
	if len(m.events) == 0 {
		return status
	}
	last := m.events[len(m.events)-1]
	status.LastEvent = last.Event
	status.LastTS = last.TS
	return status
}
