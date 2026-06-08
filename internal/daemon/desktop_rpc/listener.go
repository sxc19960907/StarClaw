package desktop_rpc

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type MethodHandler func(ctx context.Context, params json.RawMessage) (any, *RPCError)

type EventSink func(evt *DesktopEvent)

type ListenerConfig struct {
	SockPath    string
	PidfilePath string
	Platform    Platform
	Broker      *Broker
	EventSink   EventSink
	ReadyCh     chan<- struct{}
}

type Listener struct {
	cfg        ListenerConfig
	methods    map[string]MethodHandler
	running    atomic.Bool
	activeConn atomic.Pointer[net.Conn]
	writeMu    sync.Mutex
}

func NewListener(cfg ListenerConfig) *Listener {
	l := &Listener{
		cfg:     cfg,
		methods: make(map[string]MethodHandler),
	}
	l.methods[MethodSystemPing] = l.handleSystemPing
	l.methods[MethodSystemCapabilities] = l.handleSystemCapabilities
	return l
}

func (l *Listener) RegisterMethod(method string, h MethodHandler) {
	if l.running.Load() {
		panic(fmt.Sprintf("desktop_rpc: RegisterMethod(%q) called after Listener.Run started", method))
	}
	l.methods[method] = h
}

func (l *Listener) IsListening() bool {
	return l.running.Load()
}

func (l *Listener) IsConnected() bool {
	return l.activeConn.Load() != nil
}

func (l *Listener) Status() Status {
	pending := 0
	if l.cfg.Broker != nil {
		pending = l.cfg.Broker.PendingCount()
	}
	return Status{
		Listening: l.IsListening(),
		Connected: l.IsConnected(),
		Pending:   pending,
	}
}

type Status struct {
	Listening bool `json:"listening"`
	Connected bool `json:"connected"`
	Pending   int  `json:"pending"`
}

func (l *Listener) Run(ctx context.Context) (retErr error) {
	defer l.cleanupArtifacts()
	l.running.Store(true)
	defer l.running.Store(false)

	if l.cfg.SockPath == "" {
		return errors.New("desktop_rpc: empty sock path")
	}
	if l.cfg.PidfilePath == "" {
		return errors.New("desktop_rpc: empty pidfile path")
	}
	if l.cfg.Broker == nil {
		return errors.New("desktop_rpc: nil broker")
	}
	if err := os.MkdirAll(filepath.Dir(l.cfg.SockPath), 0o700); err != nil {
		return fmt.Errorf("create parent dir: %w", err)
	}
	if err := os.Remove(l.cfg.SockPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("remove stale sock %s: %w", l.cfg.SockPath, err)
	}

	ln, err := net.Listen("unix", l.cfg.SockPath)
	if err != nil {
		return fmt.Errorf("listen unix %s: %w", l.cfg.SockPath, err)
	}
	defer ln.Close()
	if err := os.Chmod(l.cfg.SockPath, 0o600); err != nil {
		return fmt.Errorf("chmod sock %s: %w", l.cfg.SockPath, err)
	}
	if err := writePidfileAtomic(l.cfg.PidfilePath, os.Getpid()); err != nil {
		return fmt.Errorf("write pidfile: %w", err)
	}
	if l.cfg.ReadyCh != nil {
		close(l.cfg.ReadyCh)
	}

	go func() {
		<-ctx.Done()
		_ = ln.Close()
	}()

	const maxAcceptRetries = 10
	consecutiveErrors := 0
	for {
		conn, err := ln.Accept()
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			consecutiveErrors++
			if consecutiveErrors >= maxAcceptRetries {
				return fmt.Errorf("desktop_rpc: %d consecutive accept errors, last: %w", consecutiveErrors, err)
			}
			timer := time.NewTimer(100 * time.Millisecond)
			select {
			case <-timer.C:
			case <-ctx.Done():
				timer.Stop()
				return nil
			}
			continue
		}
		consecutiveErrors = 0
		if !l.activeConn.CompareAndSwap(nil, &conn) {
			_ = conn.Close()
			continue
		}
		go l.handleConn(ctx, conn)
	}
}

func (l *Listener) handleConn(ctx context.Context, conn net.Conn) {
	defer conn.Close()
	defer l.activeConn.Store(nil)
	defer l.cfg.Broker.CancelAll()

	l.cfg.Broker.SetSendFn(func(req *RPCRequest) error {
		frame, err := EncodeRequestFrame(req)
		if err != nil {
			return err
		}
		l.writeMu.Lock()
		defer l.writeMu.Unlock()
		return WriteFrame(conn, frame)
	})

	reader := bufio.NewReaderSize(conn, 64*1024)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		frame, err := ReadFrame(reader)
		if err != nil {
			if !errors.Is(err, io.EOF) {
				log.Printf("desktop_rpc: read error: %v", err)
			}
			return
		}
		l.dispatchFrame(ctx, conn, frame)
	}
}

func (l *Listener) dispatchFrame(ctx context.Context, conn net.Conn, f *Frame) {
	switch f.Type {
	case FrameDesktopRPCRequest:
		l.handleIncomingRequest(ctx, conn, f.Payload)
	case FrameDesktopRPCResult:
		l.handleIncomingResult(f.Payload)
	case FrameDesktopEvent:
		l.handleIncomingEvent(f.Payload)
	default:
		log.Printf("desktop_rpc: unknown frame type %q", f.Type)
	}
}

func (l *Listener) handleIncomingRequest(ctx context.Context, conn net.Conn, payload json.RawMessage) {
	var req RPCRequest
	if err := json.Unmarshal(payload, &req); err != nil {
		log.Printf("desktop_rpc: malformed RPCRequest payload: %v", err)
		return
	}
	handler, ok := l.methods[req.Method]
	var res RPCResult
	if !ok {
		details, _ := json.Marshal(map[string]string{"method": req.Method})
		res = RPCResult{
			RequestID: req.RequestID,
			OK:        false,
			Error: &RPCError{
				Code:      ErrCodeInvalidArgument,
				Message:   fmt.Sprintf("daemon does not implement method %q", req.Method),
				Retriable: false,
				Details:   details,
			},
		}
	} else {
		result, rpcErr := handler(ctx, req.Params)
		if rpcErr != nil {
			res = RPCResult{RequestID: req.RequestID, OK: false, Error: rpcErr}
		} else if marshaled, err := json.Marshal(result); err != nil {
			res = RPCResult{
				RequestID: req.RequestID,
				OK:        false,
				Error: &RPCError{
					Code:      ErrCodeInternal,
					Message:   fmt.Sprintf("marshal result: %v", err),
					Retriable: false,
				},
			}
		} else {
			res = RPCResult{RequestID: req.RequestID, OK: true, Result: marshaled}
		}
	}
	frame, err := EncodeResultFrame(&res)
	if err != nil {
		log.Printf("desktop_rpc: encode result for %s: %v", req.RequestID, err)
		return
	}
	l.writeMu.Lock()
	defer l.writeMu.Unlock()
	if err := WriteFrame(conn, frame); err != nil {
		log.Printf("desktop_rpc: write result for %s: %v", req.RequestID, err)
	}
}

func (l *Listener) handleIncomingResult(payload json.RawMessage) {
	var res RPCResult
	if err := json.Unmarshal(payload, &res); err != nil {
		log.Printf("desktop_rpc: malformed RPCResult payload: %v", err)
		return
	}
	if !l.cfg.Broker.Resolve(res.RequestID, &res) {
		log.Printf("desktop_rpc: result for unknown request_id %s", res.RequestID)
	}
}

func (l *Listener) handleIncomingEvent(payload json.RawMessage) {
	var evt DesktopEvent
	if err := json.Unmarshal(payload, &evt); err != nil {
		log.Printf("desktop_rpc: malformed DesktopEvent payload: %v", err)
		return
	}
	if l.cfg.EventSink != nil {
		l.cfg.EventSink(&evt)
	}
}

func (l *Listener) handleSystemPing(_ context.Context, params json.RawMessage) (any, *RPCError) {
	var p SystemPingParams
	if len(params) > 0 {
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, &RPCError{
				Code:      ErrCodeInvalidArgument,
				Message:   "system.ping: malformed params",
				Retriable: false,
			}
		}
	}
	return SystemPingResult{
		Pong:       p.Echo,
		ServerTime: time.Now().UTC().Format(time.RFC3339Nano),
	}, nil
}

func (l *Listener) handleSystemCapabilities(_ context.Context, _ json.RawMessage) (any, *RPCError) {
	return SystemCapabilitiesResult{
		Version:  ProtocolVersion,
		Methods:  append([]string(nil), ProtocolMethods...),
		Platform: l.cfg.Platform,
	}, nil
}

func (l *Listener) cleanupArtifacts() {
	if l.cfg.SockPath != "" {
		_ = os.Remove(l.cfg.SockPath)
	}
	if l.cfg.PidfilePath != "" {
		_ = os.Remove(l.cfg.PidfilePath)
	}
}

func writePidfileAtomic(path string, pid int) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(strconv.Itoa(pid)+"\n"), 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func DefaultPlatform(daemonVersion string) Platform {
	return Platform{
		OS:         mapOS(runtime.GOOS),
		OSVersion:  detectOSVersion(),
		AppVersion: daemonVersion,
	}
}

func mapOS(goos string) string {
	if goos == "darwin" {
		return "macOS"
	}
	return goos
}

func detectOSVersion() string {
	if runtime.GOOS != "darwin" {
		return ""
	}
	out, err := exec.Command("sw_vers", "-productVersion").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
