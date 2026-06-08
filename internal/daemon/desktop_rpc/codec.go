package desktop_rpc

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

var ErrFrameTooLarge = errors.New("desktop_rpc: frame body exceeds 4 MiB cap")
var ErrEmptyFrame = errors.New("desktop_rpc: zero-length frame")

func ReadFrame(r *bufio.Reader) (*Frame, error) {
	var lenBuf [4]byte
	if _, err := io.ReadFull(r, lenBuf[:]); err != nil {
		return nil, err
	}
	bodyLen := binary.BigEndian.Uint32(lenBuf[:])
	if bodyLen == 0 {
		return nil, ErrEmptyFrame
	}
	if bodyLen > MaxFrameBodyBytes {
		return nil, fmt.Errorf("%w: got %d bytes", ErrFrameTooLarge, bodyLen)
	}
	body := make([]byte, bodyLen)
	if _, err := io.ReadFull(r, body); err != nil {
		return nil, fmt.Errorf("read frame body: %w", err)
	}
	var f Frame
	if err := json.Unmarshal(body, &f); err != nil {
		return nil, fmt.Errorf("decode frame envelope: %w", err)
	}
	return &f, nil
}

func WriteFrame(w io.Writer, f *Frame) error {
	body, err := json.Marshal(f)
	if err != nil {
		return fmt.Errorf("encode frame envelope: %w", err)
	}
	if len(body) > MaxFrameBodyBytes {
		return fmt.Errorf("%w: would write %d bytes", ErrFrameTooLarge, len(body))
	}
	out := make([]byte, 4+len(body))
	binary.BigEndian.PutUint32(out[:4], uint32(len(body)))
	copy(out[4:], body)
	if _, err := w.Write(out); err != nil {
		return fmt.Errorf("write frame: %w", err)
	}
	return nil
}

func EncodeRequestFrame(req *RPCRequest) (*Frame, error) {
	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("encode RPCRequest payload: %w", err)
	}
	return &Frame{Type: FrameDesktopRPCRequest, Payload: payload}, nil
}

func EncodeResultFrame(res *RPCResult) (*Frame, error) {
	payload, err := json.Marshal(res)
	if err != nil {
		return nil, fmt.Errorf("encode RPCResult payload: %w", err)
	}
	return &Frame{Type: FrameDesktopRPCResult, Payload: payload}, nil
}

func EncodeEventFrame(evt *DesktopEvent) (*Frame, error) {
	payload, err := json.Marshal(evt)
	if err != nil {
		return nil, fmt.Errorf("encode DesktopEvent payload: %w", err)
	}
	return &Frame{Type: FrameDesktopEvent, Payload: payload}, nil
}
