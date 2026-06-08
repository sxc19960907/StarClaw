package cloudflow

import (
	"context"
	"errors"
)

var ErrProviderNotConfigured = errors.New("cloudflow: provider not configured")

type Request struct {
	Query        string
	WorkflowType string
	Strategy     string
	SessionID    string
	UserContext  string
	ExtraContext map[string]any
}

type Result struct {
	FinalText  string
	WorkflowID string
	TaskID     string
	LocalOnly  bool
}

type Event struct {
	Type    string
	AgentID string
	Status  string
	Message string
}

type EventSink func(Event)

type Provider interface {
	Dispatch(ctx context.Context, req Request, sink EventSink) (Result, error)
}

type LocalProvider struct{}

func (p LocalProvider) Dispatch(ctx context.Context, req Request, sink EventSink) (Result, error) {
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	if sink != nil {
		sink(Event{Type: "cloudflow.local", Status: "completed", Message: "Cloudflow provider is not configured; using local workflow execution."})
	}
	return Result{
		FinalText:  "Cloudflow provider is not configured; use the local StarClaw workflow path.",
		WorkflowID: "",
		TaskID:     "",
		LocalOnly:  true,
	}, ErrProviderNotConfigured
}
