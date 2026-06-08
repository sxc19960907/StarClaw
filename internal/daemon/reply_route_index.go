package daemon

import (
	"container/list"
	"sync"
)

const defaultReplyRouteIndexCap = 256

type ReplyRouteIndex struct {
	mu    sync.Mutex
	cap   int
	items map[string]*list.Element
	order *list.List
}

type replyRouteEntry struct {
	messageID string
	routeKey  string
}

func NewReplyRouteIndex(capItems int) *ReplyRouteIndex {
	if capItems <= 0 {
		capItems = defaultReplyRouteIndexCap
	}
	return &ReplyRouteIndex{
		cap:   capItems,
		items: make(map[string]*list.Element),
		order: list.New(),
	}
}

func (x *ReplyRouteIndex) Put(messageID, routeKey string) {
	if x == nil || messageID == "" || routeKey == "" {
		return
	}
	x.mu.Lock()
	defer x.mu.Unlock()
	if el, ok := x.items[messageID]; ok {
		el.Value.(*replyRouteEntry).routeKey = routeKey
		x.order.MoveToBack(el)
		return
	}
	el := x.order.PushBack(&replyRouteEntry{messageID: messageID, routeKey: routeKey})
	x.items[messageID] = el
	for x.order.Len() > x.cap {
		front := x.order.Front()
		if front == nil {
			break
		}
		x.order.Remove(front)
		delete(x.items, front.Value.(*replyRouteEntry).messageID)
	}
}

func (x *ReplyRouteIndex) Get(messageID string) string {
	if x == nil || messageID == "" {
		return ""
	}
	x.mu.Lock()
	defer x.mu.Unlock()
	if el, ok := x.items[messageID]; ok {
		return el.Value.(*replyRouteEntry).routeKey
	}
	return ""
}
