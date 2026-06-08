package daemon

import (
	"strings"
	"time"

	"github.com/starclaw/starclaw/internal/agent"
)

func channelLabel(p ReplyDeliveryResultPayload) string {
	if p.ThreadID != "" {
		if i := strings.IndexByte(p.ThreadID, '-'); i > 0 {
			return p.Channel + " " + p.ThreadID[:i]
		}
	}
	return p.Channel
}

func formatDeliveryFailure(p ReplyDeliveryResultPayload) string {
	reason := p.Reason
	if reason == "" {
		reason = "delivery failed"
	}
	where := channelLabel(p)
	if p.Class == ClassPermanent {
		return "reply to " + where + " FAILED: " + reason +
			" - the user did not see it, and the bot will not receive or deliver messages there until re-added/re-authorized."
	}
	return "reply to " + where + " may not have been delivered (" + reason + "); a retry is in progress."
}

func newDeliveryResultConsumer(store *SystemEventStore, idx *ReplyRouteIndex) func(ReplyDeliveryResultPayload, string) {
	return func(p ReplyDeliveryResultPayload, messageID string) {
		if p.OK || store == nil || idx == nil {
			return
		}
		routeKey := idx.Get(messageID)
		if routeKey == "" {
			return
		}
		store.Enqueue(routeKey, agent.SystemEvent{
			Text:       formatDeliveryFailure(p),
			ContextKey: "delivery-fail:" + p.Channel + ":" + p.ThreadID,
			Trusted:    true,
			TS:         time.Now(),
		})
	}
}

func HandleReplyDeliveryResult(store *SystemEventStore, idx *ReplyRouteIndex, p ReplyDeliveryResultPayload, messageID string) {
	newDeliveryResultConsumer(store, idx)(p, messageID)
}
