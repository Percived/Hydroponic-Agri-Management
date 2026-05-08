package event

import (
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

type eventTypeMapping struct {
	internal string
	client   string
}

var eventMappings = map[string]eventTypeMapping{
	"alert:created":      {internal: "alert:created", client: "new_alert"},
	"telemetry:received": {internal: "telemetry:received", client: "telemetry_update"},
	"device:status":      {internal: "device:status", client: "device_status"},
}

// SSEHandler creates a Gin handler that streams events from the EventHub via SSE.
// internalType is the EventHub event type to subscribe to (e.g., "alert:created").
// The handler automatically remaps the event type to the frontend-compatible name.
func SSEHandler(hub *Hub, internalType string) gin.HandlerFunc {
	mapping, ok := eventMappings[internalType]
	if !ok {
		mapping = eventTypeMapping{internal: internalType, client: internalType}
	}

	return func(c *gin.Context) {
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		c.Header("X-Accel-Buffering", "no")
		c.Header("Access-Control-Allow-Origin", "*")

		w := c.Writer
		flusher, ok := w.(http.Flusher)
		if !ok {
			c.String(http.StatusInternalServerError, "streaming not supported")
			return
		}

		sub := hub.Subscribe(func(e SSEEvent) bool {
			return e.Type == mapping.internal
		})
		defer hub.Unsubscribe(sub)

		ctx := c.Request.Context()

		for {
			select {
			case evt, ok := <-sub.Events:
				if !ok {
					return
				}
				evt.Type = mapping.client
				data, err := FormatSSE(evt)
				if err != nil {
					continue
				}
				if _, err := io.WriteString(w, string(data)); err != nil {
					return
				}
				flusher.Flush()

			case <-ctx.Done():
				return
			}
		}
	}
}

// SSEPing sends a ping comment every interval to keep the connection alive.
// Deprecated: modern browsers handle SSE keepalive natively. Use if needed.
func SSEPing(w io.Writer, flusher http.Flusher) {
	msg := fmt.Sprintf(":ping %s\n\n", "keepalive")
	w.Write([]byte(msg))
	flusher.Flush()
}
