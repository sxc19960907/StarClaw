package cloudflow

import (
	"context"
	"errors"
	"testing"
)

func TestLocalProviderDispatch(t *testing.T) {
	var events []Event
	res, err := LocalProvider{}.Dispatch(context.Background(), Request{Query: "hello", WorkflowType: TypeAuto}, func(evt Event) {
		events = append(events, evt)
	})
	if !errors.Is(err, ErrProviderNotConfigured) {
		t.Fatalf("err = %v, want ErrProviderNotConfigured", err)
	}
	if !res.LocalOnly {
		t.Fatalf("result = %#v, want local only", res)
	}
	if len(events) != 1 || events[0].Type != "cloudflow.local" {
		t.Fatalf("events = %#v", events)
	}
}
