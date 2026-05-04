package event

import (
	"encoding/json"
	"sync"
	"sync/atomic"
)

// SSEEvent represents a server-sent event to broadcast.
type SSEEvent struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// Subscriber receives events matching an optional filter.
type Subscriber struct {
	ID     uint64
	Events chan SSEEvent
	filter func(SSEEvent) bool
}

// Hub broadcasts events to registered subscribers.
// Subscribers receive only events that pass their filter (or all events if filter is nil).
type Hub struct {
	mu          sync.RWMutex
	subscribers map[uint64]*Subscriber
	nextID      uint64
}

// NewHub creates a new Hub.
func NewHub() *Hub {
	return &Hub{
		subscribers: make(map[uint64]*Subscriber),
	}
}

// Subscribe registers a new subscriber. The optional filter function is called
// for each published event; return true to receive it. If filter is nil, all
// events are delivered. The caller MUST call Unsubscribe when done to avoid
// leaking goroutines.
func (h *Hub) Subscribe(filter func(SSEEvent) bool) *Subscriber {
	id := atomic.AddUint64(&h.nextID, 1)
	sub := &Subscriber{
		ID:     id,
		Events: make(chan SSEEvent, 64),
		filter: filter,
	}
	h.mu.Lock()
	h.subscribers[id] = sub
	h.mu.Unlock()
	return sub
}

// Unsubscribe removes a subscriber and closes its channel.
func (h *Hub) Unsubscribe(sub *Subscriber) {
	h.mu.Lock()
	delete(h.subscribers, sub.ID)
	h.mu.Unlock()
	close(sub.Events)
}

// Publish sends an event to all subscribers whose filter accepts it.
// Sends are non-blocking; slow consumers are skipped.
func (h *Hub) Publish(event SSEEvent) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, sub := range h.subscribers {
		if sub.filter != nil && !sub.filter(event) {
			continue
		}
		select {
		case sub.Events <- event:
		default:
			// drop for slow consumer
		}
	}
}

// FormatSSE serializes an SSEEvent into the SSE wire format.
func FormatSSE(event SSEEvent) ([]byte, error) {
	data, err := json.Marshal(event)
	if err != nil {
		return nil, err
	}
	out := make([]byte, 0, len(data)+8)
	out = append(out, "data: "...)
	out = append(out, data...)
	out = append(out, '\n', '\n')
	return out, nil
}
