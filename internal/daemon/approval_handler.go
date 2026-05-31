package daemon

import (
	"context"
	"encoding/json"

	"github.com/starclaw/starclaw/internal/agent"
)

// DaemonApprovalRequester adapts daemon approval endpoints/events to agent approval requests.
type DaemonApprovalRequester struct {
	broker   *ApprovalBroker
	bus      *EventBus
	channel  string
	threadID string
	agent    string
}

func NewDaemonApprovalRequester(broker *ApprovalBroker, bus *EventBus, channel, threadID, agentName string) *DaemonApprovalRequester {
	return &DaemonApprovalRequester{
		broker:   broker,
		bus:      bus,
		channel:  channel,
		threadID: threadID,
		agent:    agentName,
	}
}

func (r *DaemonApprovalRequester) RequestApproval(ctx context.Context, req agent.ApprovalRequest) (agent.ApprovalDecision, error) {
	if r.broker == nil {
		return agent.ApprovalDeny, nil
	}
	approval := ApprovalRequest{
		Channel:   r.channel,
		ThreadID:  r.threadID,
		RequestID: NewApprovalRequestID(),
		Tool:      req.Tool,
		Args:      req.Args,
		Agent:     r.agent,
		Reason:    req.Reason,
	}
	r.publish(EventApprovalNeeded, approval)

	decision, err := r.broker.WaitForApproval(ctx, approval)
	resolved := ApprovalResolvedPayload{
		RequestID:  approval.RequestID,
		Decision:   decision,
		ResolvedBy: "daemon",
	}
	if err != nil {
		resolved.Decision = DecisionDeny
	}
	r.publish(EventApprovalResolved, resolved)
	if err != nil {
		return agent.ApprovalDeny, err
	}
	if decision == DecisionAllow {
		return agent.ApprovalAllow, nil
	}
	return agent.ApprovalDeny, nil
}

func (r *DaemonApprovalRequester) publish(eventType string, payload any) {
	if r.bus == nil {
		return
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return
	}
	r.bus.Publish(Event{Type: eventType, Data: string(data)})
}
