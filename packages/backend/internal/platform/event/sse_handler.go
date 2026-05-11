package event

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

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
	"command:dispatched": {internal: "command:dispatched", client: "command_dispatched"},
	"command:acked":      {internal: "command:acked", client: "command_acked"},
}

// SSEHandler creates a Gin handler that streams events from the EventHub via SSE.
// internalType is the EventHub event type to subscribe to (e.g., "alert:created").
// The handler automatically remaps the event type to the frontend-compatible name.
func SSEHandler(hub *Hub, internalType string) gin.HandlerFunc {
	return SSEHandlerMulti(hub, []string{internalType})
}

func SSEHandlerMulti(hub *Hub, internalTypes []string) gin.HandlerFunc {
	internalTypeSet := map[string]struct{}{}
	for _, t := range internalTypes {
		t = strings.TrimSpace(t)
		if t == "" {
			continue
		}
		internalTypeSet[t] = struct{}{}
	}

	return func(c *gin.Context) {
		deviceCodeSet := map[string]struct{}{}
		metricCodeSet := map[string]struct{}{}
		alertLevelSet := map[string]struct{}{}
		alertDeviceCodeSet := map[string]struct{}{}
		hasTelemetry := len(internalTypeSet) == 0
		hasAlert := len(internalTypeSet) == 0
		hasDeviceStatus := len(internalTypeSet) == 0
		hasCommands := len(internalTypeSet) == 0
		if len(internalTypeSet) > 0 {
			_, hasTelemetry = internalTypeSet["telemetry:received"]
			_, hasAlert = internalTypeSet["alert:created"]
			_, hasDeviceStatus = internalTypeSet["device:status"]
			_, hasCommandDispatched := internalTypeSet["command:dispatched"]
			_, hasCommandAcked := internalTypeSet["command:acked"]
			hasCommands = hasCommandDispatched || hasCommandAcked
		}

		if hasTelemetry || hasDeviceStatus || hasCommands {
			for _, v := range strings.Split(c.Query("device_codes"), ",") {
				v = strings.TrimSpace(v)
				if v == "" {
					continue
				}
				deviceCodeSet[v] = struct{}{}
			}
		}
		if hasTelemetry {
			for _, v := range strings.Split(c.Query("metric_codes"), ",") {
				v = strings.TrimSpace(v)
				if v == "" {
					continue
				}
				metricCodeSet[v] = struct{}{}
			}
		}
		if hasAlert {
			for _, v := range strings.Split(c.Query("level"), ",") {
				v = strings.TrimSpace(v)
				if v == "" {
					continue
				}
				alertLevelSet[v] = struct{}{}
			}
			deviceCodes := c.Query("device_codes")
			if deviceCodes == "" {
				deviceCodes = c.Query("device_code")
			}
			for _, v := range strings.Split(deviceCodes, ",") {
				v = strings.TrimSpace(v)
				if v == "" {
					continue
				}
				alertDeviceCodeSet[v] = struct{}{}
			}
		}

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
		_, _ = io.WriteString(w, "retry: 2000\n\n")
		flusher.Flush()

		sub := hub.Subscribe(func(e SSEEvent) bool {
			if len(internalTypeSet) == 0 {
				return true
			}
			_, ok := internalTypeSet[e.Type]
			return ok
		})
		defer hub.Unsubscribe(sub)

		ctx := c.Request.Context()

		for {
			select {
			case evt, ok := <-sub.Events:
				if !ok {
					return
				}
				if evt.Type == "telemetry:received" {
					if len(deviceCodeSet) > 0 {
						if v, ok := getStringField(evt.Data, "device_code"); !ok || v == "" {
							continue
						} else {
							if _, ok := deviceCodeSet[v]; !ok {
								continue
							}
						}
					}
					if len(metricCodeSet) > 0 {
						if v, ok := getStringField(evt.Data, "metric_code"); !ok || v == "" {
							continue
						} else {
							if _, ok := metricCodeSet[v]; !ok {
								continue
							}
						}
					}
				}
				if evt.Type == "alert:created" {
					if len(alertLevelSet) > 0 {
						if v, ok := getStringField(evt.Data, "level"); !ok || v == "" {
							continue
						} else {
							if _, ok := alertLevelSet[v]; !ok {
								continue
							}
						}
					}
					if len(alertDeviceCodeSet) > 0 {
						if v, ok := getStringField(evt.Data, "device_code"); !ok || v == "" {
							continue
						} else {
							if _, ok := alertDeviceCodeSet[v]; !ok {
								continue
							}
						}
					}
				}
				if evt.Type == "device:status" || evt.Type == "command:dispatched" || evt.Type == "command:acked" {
					if len(deviceCodeSet) > 0 {
						if v, ok := getStringField(evt.Data, "device_code"); !ok || v == "" {
							continue
						} else {
							if _, ok := deviceCodeSet[v]; !ok {
								continue
							}
						}
					}
				}

				internalType := evt.Type
				if mapping, ok := eventMappings[internalType]; ok {
					evt.Type = mapping.client
				}
				data, err := FormatSSE(evt)
				if err != nil {
					continue
				}
				id := ""
				if internalType == "telemetry:received" {
					if v, ok := getStringField(evt.Data, "collected_at"); ok && v != "" {
						id = v
					}
				}
				if id == "" && internalType == "device:status" {
					if v, ok := getStringField(evt.Data, "reported_at"); ok && v != "" {
						id = v
					}
				}
				if id == "" && internalType == "command:acked" {
					if v, ok := getUint64Field(evt.Data, "command_id"); ok {
						id = fmt.Sprint(v)
					}
				}
				if id == "" && internalType == "command:dispatched" {
					if v, ok := getUint64Field(evt.Data, "command_id"); ok {
						id = fmt.Sprint(v)
					}
				}
				if id == "" {
					if v, ok := getUint64Field(evt.Data, "id"); ok {
						id = fmt.Sprint(v)
					} else if dataMap, ok := evt.Data.(map[string]interface{}); ok {
						if v, ok := dataMap["id"]; ok {
							id = fmt.Sprint(v)
						}
					}
				}
				if id == "" {
					id = fmt.Sprint(time.Now().UnixMilli())
				}
				if _, err := io.WriteString(w, "id: "+id+"\n"); err != nil {
					return
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

func getStringField(data interface{}, key string) (string, bool) {
	switch v := data.(type) {
	case TelemetrySSEDataV1:
		switch key {
		case "device_code":
			return v.DeviceCode, v.DeviceCode != ""
		case "metric_code":
			return v.MetricCode, v.MetricCode != ""
		case "collected_at":
			return v.CollectedAt, v.CollectedAt != ""
		case "quality_flag":
			return v.QualityFlag, v.QualityFlag != ""
		}
	case *TelemetrySSEDataV1:
		if v != nil {
			return getStringField(*v, key)
		}
	case DeviceStatusSSEDataV1:
		switch key {
		case "device_code":
			return v.DeviceCode, v.DeviceCode != ""
		case "status":
			return v.Status, v.Status != ""
		case "reason":
			return v.Reason, v.Reason != ""
		case "reported_at":
			return v.ReportedAt, v.ReportedAt != ""
		}
	case *DeviceStatusSSEDataV1:
		if v != nil {
			return getStringField(*v, key)
		}
	case CommandAckData:
		switch key {
		case "device_code":
			return v.DeviceCode, v.DeviceCode != ""
		case "ack_code":
			return v.AckCode, v.AckCode != ""
		case "ack_message":
			return v.AckMessage, v.AckMessage != ""
		case "acked_at":
			return v.AckedAt, v.AckedAt != ""
		}
	case *CommandAckData:
		if v != nil {
			return getStringField(*v, key)
		}
	case CommandDispatchedSSEDataV1:
		switch key {
		case "device_code":
			return v.DeviceCode, v.DeviceCode != ""
		case "status":
			return v.Status, v.Status != ""
		case "dispatched_at":
			return v.DispatchedAt, v.DispatchedAt != ""
		case "source_type":
			return v.SourceType, v.SourceType != ""
		case "error_message":
			return v.ErrorMessage, v.ErrorMessage != ""
		}
	case *CommandDispatchedSSEDataV1:
		if v != nil {
			return getStringField(*v, key)
		}
	case map[string]interface{}:
		if s, ok := v[key].(string); ok {
			return s, s != ""
		}
		if b, ok := v[key].([]byte); ok && len(b) > 0 {
			return string(b), true
		}
	}
	return "", false
}

func getUint64Field(data interface{}, key string) (uint64, bool) {
	switch v := data.(type) {
	case TelemetrySSEDataV1:
		if key == "sensor_channel_id" {
			return v.SensorChannelID, v.SensorChannelID != 0
		}
	case *TelemetrySSEDataV1:
		if v != nil {
			return getUint64Field(*v, key)
		}
	case CommandAckData:
		if key == "command_id" {
			return v.CommandID, v.CommandID != 0
		}
	case *CommandAckData:
		if v != nil {
			return getUint64Field(*v, key)
		}
	case CommandDispatchedSSEDataV1:
		switch key {
		case "command_id":
			return v.CommandID, v.CommandID != 0
		case "source_id":
			return v.SourceID, v.SourceID != 0
		}
	case *CommandDispatchedSSEDataV1:
		if v != nil {
			return getUint64Field(*v, key)
		}
	case map[string]interface{}:
		if raw, ok := v[key]; ok {
			switch n := raw.(type) {
			case uint64:
				return n, true
			case uint32:
				return uint64(n), true
			case uint:
				return uint64(n), true
			case int64:
				if n >= 0 {
					return uint64(n), true
				}
			case int:
				if n >= 0 {
					return uint64(n), true
				}
			case float64:
				if n >= 0 {
					return uint64(n), true
				}
			case float32:
				if n >= 0 {
					return uint64(n), true
				}
			case string:
				if n != "" {
					if out, err := strconv.ParseUint(strings.TrimSpace(n), 10, 64); err == nil {
						return out, true
					}
				}
			}
		}
	}
	return 0, false
}

// SSEPing sends a ping comment every interval to keep the connection alive.
// Deprecated: modern browsers handle SSE keepalive natively. Use if needed.
func SSEPing(w io.Writer, flusher http.Flusher) {
	msg := fmt.Sprintf(":ping %s\n\n", "keepalive")
	w.Write([]byte(msg))
	flusher.Flush()
}
