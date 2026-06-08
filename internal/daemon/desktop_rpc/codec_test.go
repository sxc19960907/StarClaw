package desktop_rpc

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestFrameRoundTrip(t *testing.T) {
	t.Parallel()
	in := &Frame{
		Type:    FrameDesktopRPCRequest,
		Payload: json.RawMessage(`{"request_id":"drpc_1","method":"system.ping"}`),
	}
	var buf bytes.Buffer
	if err := WriteFrame(&buf, in); err != nil {
		t.Fatalf("WriteFrame: %v", err)
	}
	got, err := ReadFrame(bufio.NewReader(&buf))
	if err != nil {
		t.Fatalf("ReadFrame: %v", err)
	}
	if got.Type != in.Type {
		t.Fatalf("type = %q, want %q", got.Type, in.Type)
	}
	if string(got.Payload) != string(in.Payload) {
		t.Fatalf("payload = %s, want %s", got.Payload, in.Payload)
	}
}

func TestReadFrameRejectsInvalidFrames(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		raw  []byte
		want error
	}{
		{name: "clean eof", raw: nil, want: io.EOF},
		{name: "short prefix", raw: []byte{0, 0}, want: io.ErrUnexpectedEOF},
		{name: "empty", raw: []byte{0, 0, 0, 0}, want: ErrEmptyFrame},
		{name: "too large", raw: prefixOnly(MaxFrameBodyBytes + 1), want: ErrFrameTooLarge},
		{name: "partial body", raw: append(prefixOnly(16), []byte("{}")...), want: io.ErrUnexpectedEOF},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := ReadFrame(bufio.NewReader(bytes.NewReader(tt.raw)))
			if !errors.Is(err, tt.want) {
				t.Fatalf("error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestReadFrameRejectsMalformedJSON(t *testing.T) {
	t.Parallel()
	body := []byte("{not json")
	raw := append(prefixOnly(uint32(len(body))), body...)
	_, err := ReadFrame(bufio.NewReader(bytes.NewReader(raw)))
	if err == nil || !strings.Contains(err.Error(), "decode frame envelope") {
		t.Fatalf("error = %v, want decode frame envelope", err)
	}
}

func TestWriteFrameRejectsOversizedFrame(t *testing.T) {
	t.Parallel()
	frame := &Frame{
		Type:    FrameDesktopRPCRequest,
		Payload: json.RawMessage(`"` + strings.Repeat("a", MaxFrameBodyBytes) + `"`),
	}
	var buf bytes.Buffer
	err := WriteFrame(&buf, frame)
	if !errors.Is(err, ErrFrameTooLarge) {
		t.Fatalf("error = %v, want ErrFrameTooLarge", err)
	}
	if buf.Len() != 0 {
		t.Fatalf("oversized frame wrote %d bytes", buf.Len())
	}
}

func TestEncodeFrameHelpers(t *testing.T) {
	t.Parallel()
	reqFrame, err := EncodeRequestFrame(&RPCRequest{RequestID: "drpc_a", Method: MethodSystemPing})
	if err != nil {
		t.Fatalf("EncodeRequestFrame: %v", err)
	}
	if reqFrame.Type != FrameDesktopRPCRequest {
		t.Fatalf("request type = %q", reqFrame.Type)
	}
	resFrame, err := EncodeResultFrame(&RPCResult{RequestID: "drpc_a", OK: true})
	if err != nil {
		t.Fatalf("EncodeResultFrame: %v", err)
	}
	if resFrame.Type != FrameDesktopRPCResult {
		t.Fatalf("result type = %q", resFrame.Type)
	}
	evtFrame, err := EncodeEventFrame(&DesktopEvent{Event: "desktop_online"})
	if err != nil {
		t.Fatalf("EncodeEventFrame: %v", err)
	}
	if evtFrame.Type != FrameDesktopEvent {
		t.Fatalf("event type = %q", evtFrame.Type)
	}
}

func prefixOnly(n uint32) []byte {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, n)
	return buf
}
