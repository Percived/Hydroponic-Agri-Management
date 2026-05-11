package main

import (
	"encoding/json"
	"log"
	"sync"
)

// ──────────────────── SSE 事件类型常量 ────────────────────

const (
	SSEEventEnv       = "env"
	SSEEventTelemetry = "telemetry"
	SSEEventHeartbeat = "heartbeat"
	SSEEventCmd       = "cmd"
	SSEEventAck       = "ack"
	SSEEventStatus    = "status"
	SSEEventLog       = "log"
)

// ──────────────────── SSE 事件结构体 ────────────────────

// SSEEvent represents a single SSE event.
type SSEEvent struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// ──────────────────── SSE Hub ────────────────────

// SSEHub manages SSE subscriber channels and fans out events.
type SSEHub struct {
	mu          sync.RWMutex
	subscribers map[chan []byte]struct{}
}

// NewSSEHub creates a new SSE hub.
func NewSSEHub() *SSEHub {
	return &SSEHub{
		subscribers: make(map[chan []byte]struct{}),
	}
}

// Subscribe registers a new subscriber channel (buffer 64 events).
// The caller is responsible for reading from this channel.
func (h *SSEHub) Subscribe() chan []byte {
	ch := make(chan []byte, 64)
	h.mu.Lock()
	defer h.mu.Unlock()
	h.subscribers[ch] = struct{}{}
	return ch
}

// Unsubscribe removes a subscriber channel and closes it.
func (h *SSEHub) Unsubscribe(ch chan []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.subscribers[ch]; ok {
		delete(h.subscribers, ch)
		close(ch)
	}
}

// Publish sends an SSE event to all subscribers.
// Non-blocking: if a subscriber's buffer is full, the event is dropped for that subscriber.
func (h *SSEHub) Publish(event SSEEvent) {
	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("❌ SSE 序列化失败: %v", err)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for ch := range h.subscribers {
		select {
		case ch <- data:
			// sent successfully
		default:
			// buffer full, drop event for this slow client
		}
	}
}

// PublishEnv publishes an environment snapshot event.
func (h *SSEHub) PublishEnv(snapshot SimSnapshot) {
	h.Publish(SSEEvent{Type: SSEEventEnv, Data: snapshot})
}

// PublishTelemetry publishes a telemetry event with channel count.
func (h *SSEHub) PublishTelemetry(channelCount int) {
	h.Publish(SSEEvent{Type: SSEEventTelemetry, Data: map[string]int{"channel_count": channelCount}})
}

// PublishHeartbeat publishes a heartbeat event with device type.
func (h *SSEHub) PublishHeartbeat(deviceType string) {
	h.Publish(SSEEvent{Type: SSEEventHeartbeat, Data: map[string]string{"device_type": deviceType}})
}

// PublishCmd publishes a command received event.
func (h *SSEHub) PublishCmd(commandID uint64, cmdType, state string, value float64) {
	h.Publish(SSEEvent{Type: SSEEventCmd, Data: map[string]interface{}{
		"command_id":   commandID,
		"command_type": cmdType,
		"state":        state,
		"value":        value,
	}})
}

// PublishAck publishes an ACK event.
func (h *SSEHub) PublishAck(commandID uint64, ackCode string) {
	h.Publish(SSEEvent{Type: SSEEventAck, Data: map[string]interface{}{
		"command_id": commandID,
		"ack_code":   ackCode,
	}})
}

// PublishStatus publishes a simulator status change event.
func (h *SSEHub) PublishStatus(status string) {
	h.Publish(SSEEvent{Type: SSEEventStatus, Data: map[string]string{"status": status}})
}

// PublishLog publishes a log event.
func (h *SSEHub) PublishLog(level, message string) {
	h.Publish(SSEEvent{Type: SSEEventLog, Data: map[string]string{
		"level":   level,
		"message": message,
	}})
}

// ──────────────────── 条件发射辅助 ────────────────────

// emitIfHub calls fn if hub is non-nil. Convenience for optional SSE emission.
func emitIfHub(hub *SSEHub, fn func()) {
	if hub != nil {
		fn()
	}
}
