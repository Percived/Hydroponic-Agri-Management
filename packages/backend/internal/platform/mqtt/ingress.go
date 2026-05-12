package mqtt

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"hydroponic-backend/internal/alert"
	"hydroponic-backend/internal/device"
	"hydroponic-backend/internal/platform/config"
	"hydroponic-backend/internal/platform/event"
	"hydroponic-backend/internal/telemetry"

	mqttlib "github.com/eclipse/paho.mqtt.golang"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"gorm.io/gorm"
)

type IngressService struct {
	db        *gorm.DB
	influx    influxdb2.Client
	influxCfg config.InfluxConfig
	hub       *event.Hub
	cache     *telemetry.SensorStatusCache
	client    mqttlib.Client
	log       *slog.Logger
}

func NewIngressService(
	db *gorm.DB,
	influx influxdb2.Client,
	influxCfg config.InfluxConfig,
	hub *event.Hub,
	cache *telemetry.SensorStatusCache,
	client mqttlib.Client,
	log *slog.Logger,
) *IngressService {
	return &IngressService{
		db:        db,
		influx:    influx,
		influxCfg: influxCfg,
		hub:       hub,
		cache:     cache,
		client:    client,
		log:       log,
	}
}

func (s *IngressService) Start() error {
	// Subscribe to all 6 inbound topic patterns using wildcard
	topicFilter := fmt.Sprintf("%s/+/+/#", TopicPrefix)
	token := s.client.Subscribe(topicFilter, 1, s.onMessage)
	if token.Wait() && token.Error() != nil {
		return fmt.Errorf("mqtt subscribe: %w", token.Error())
	}

	s.log.Info("mqtt ingress started", "topic", topicFilter)
	return nil
}

func (s *IngressService) onMessage(_ mqttlib.Client, msg mqttlib.Message) {
	topic := msg.Topic()
	parts := strings.Split(topic, "/")
	if len(parts) < 3 || parts[0] != TopicPrefix {
		s.log.Warn("ingress: unexpected topic format", "topic", topic)
		return
	}

	deviceCode := parts[1]
	topicType := parts[2]

	s.log.Debug("ingress: message received", "device", deviceCode, "type", topicType)

	switch topicType {
	case TopicTelemetry:
		s.handleTelemetry(deviceCode, msg.Payload())
	case TopicStatus:
		s.handleStatus(deviceCode, msg.Payload())
	case TopicHeartbeat:
		s.handleHeartbeat(deviceCode, msg.Payload())
	case TopicErrors:
		s.handleErrors(deviceCode, msg.Payload())
	case TopicDiagnostics:
		s.handleDiagnostics(deviceCode, msg.Payload())
	case TopicAck:
		s.handleAck(deviceCode, msg.Payload())
	case TopicState:
		s.handleState(deviceCode, msg.Payload())
	default:
		s.log.Warn("ingress: unknown topic type", "type", topicType, "device", deviceCode)
	}
}

// lookupDevicePresence checks whether a device_code exists in sensor_devices or actuator_devices.
func (s *IngressService) lookupDevicePresence(deviceCode string) (sensorFound, actuatorFound bool) {
	var count int64
	if err := s.db.Model(&device.SensorDevice{}).Where("device_code = ?", deviceCode).Count(&count).Error; err == nil && count > 0 {
		sensorFound = true
	}
	count = 0
	if err := s.db.Model(&device.ActuatorDevice{}).Where("device_code = ?", deviceCode).Count(&count).Error; err == nil && count > 0 {
		actuatorFound = true
	}
	return
}

// createUnknownDeviceAlert creates a DEVICE_DISCOVERED alert for an unregistered device.
// Deduplicates: only creates one open alert per device_code.
func (s *IngressService) createUnknownDeviceAlert(deviceCode string) {
	var existing int64
	if err := s.db.Model(&alert.Alert{}).
		Where("type = ? AND status = ? AND message LIKE ?",
			"DEVICE_DISCOVERED", alert.StatusOpen, "%["+deviceCode+"]%").
		Count(&existing).Error; err == nil && existing > 0 {
		return
	}

	now := time.Now().UTC()
	a := alert.Alert{
		Type:        "DEVICE_DISCOVERED",
		Level:       alert.LevelInfo,
		Message:     fmt.Sprintf("[%s] 发现未知设备，请在系统中注册该设备。", deviceCode),
		Status:      alert.StatusOpen,
		TriggeredAt: now,
	}

	if err := s.db.Create(&a).Error; err != nil {
		s.log.Error("ingress: failed to create device discovered alert", "device_code", deviceCode, "error", err)
		return
	}

	s.db.Create(&alert.AlertTimelineEvent{
		AlertID:     a.ID,
		EventType:   alert.EventTriggered,
		EventSource: alert.SourceSystem,
		EventTime:   now,
	})

	s.hub.Publish(event.SSEEvent{
		Type: "alert:created",
		Data: alert.BuildAlertSSEDataV1(a, deviceCode, 1),
	})
}

// handleTelemetry processes telemetry data: InfluxDB + MySQL + cache + event
func (s *IngressService) handleTelemetry(deviceCode string, payload []byte) {
	sensorFound, _ := s.lookupDevicePresence(deviceCode)
	if !sensorFound {
		s.log.Warn("ingress: telemetry from unknown device, discarding", "device_code", deviceCode)
		s.createUnknownDeviceAlert(deviceCode)
		return
	}

	var items []telemetry.IngestTelemetryRequest
	if err := json.Unmarshal(payload, &items); err != nil {
		// Try single record
		var single telemetry.IngestTelemetryRequest
		if err2 := json.Unmarshal(payload, &single); err2 != nil {
			s.log.Error("ingress: invalid telemetry payload", "device", deviceCode, "error", err)
			return
		}
		items = []telemetry.IngestTelemetryRequest{single}
	}

	records := make([]telemetry.TelemetryRecord, 0, len(items))
	now := time.Now().UTC()

	for _, item := range items {
		collectedAt := now
		if item.CollectedAt != "" {
			if t, err := time.Parse(time.RFC3339, item.CollectedAt); err == nil {
				collectedAt = t.UTC()
			} else if t, err := time.Parse(time.RFC3339Nano, item.CollectedAt); err == nil {
				collectedAt = t.UTC()
			}
		}

		qualityFlag := item.QualityFlag
		if qualityFlag == "" {
			qualityFlag = telemetry.QualityFlagNormal
		}

		rec := telemetry.TelemetryRecord{
			SensorChannelID: item.SensorChannelID,
			MetricCode:      item.MetricCode,
			Value:           item.Value,
			RawValue:        item.RawValue,
			QualityFlag:     qualityFlag,
			CollectedAt:     collectedAt,
			BatchID:         item.BatchID,
		}
		records = append(records, rec)
	}

	if len(records) == 0 {
		return
	}

	// 1. Write to MySQL
	if err := s.db.Create(&records).Error; err != nil {
		s.log.Error("ingress: failed to write telemetry to mysql", "device", deviceCode, "error", err)
	}

	// 2. Write to InfluxDB
	s.writeToInflux(records)

	// 3. Update memory cache
	if s.cache != nil {
		for _, rec := range records {
			s.cache.Set(telemetry.CachedRecord{
				SensorChannelID: rec.SensorChannelID,
				MetricCode:      rec.MetricCode,
				Value:           rec.Value,
				QualityFlag:     rec.QualityFlag,
				CollectedAtUnix: rec.CollectedAt.Unix(),
			})
		}
	}

	// 4. Publish event for SSE + policy evaluation
	for _, rec := range records {
		s.hub.Publish(event.SSEEvent{
			Type: "telemetry:received",
			Data: event.TelemetrySSEDataV1{
				SchemaVersion:   1,
				SensorChannelID: rec.SensorChannelID,
				MetricCode:      rec.MetricCode,
				Value:           rec.Value,
				QualityFlag:     rec.QualityFlag,
				CollectedAt:     rec.CollectedAt.Format(time.RFC3339),
				DeviceCode:      deviceCode,
			},
		})
	}
}

func (s *IngressService) writeToInflux(records []telemetry.TelemetryRecord) {
	if s.influx == nil || s.influxCfg.Org == "" || s.influxCfg.Bucket == "" {
		return
	}

	writeAPI := s.influx.WriteAPI(s.influxCfg.Org, s.influxCfg.Bucket)
	for _, rec := range records {
		p := influxdb2.NewPointWithMeasurement("telemetry").
			AddTag("sensor_channel_id", fmt.Sprintf("%d", rec.SensorChannelID)).
			AddTag("metric_code", rec.MetricCode).
			AddTag("quality_flag", rec.QualityFlag).
			AddField("value", rec.Value).
			SetTime(rec.CollectedAt)

		if rec.RawValue != nil {
			p.AddField("raw_value", *rec.RawValue)
		}
		writeAPI.WritePoint(p)
	}
	writeAPI.Flush()
}

// handleStatus processes device status changes
func (s *IngressService) handleStatus(deviceCode string, payload []byte) {
	var status struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(payload, &status); err != nil {
		s.log.Error("ingress: invalid status payload", "device", deviceCode, "error", err)
		return
	}

	status.Status = strings.ToUpper(strings.TrimSpace(status.Status))
	if status.Status != device.StatusOnline && status.Status != device.StatusOffline && status.Status != device.StatusFault {
		s.log.Warn("ingress: invalid status value, discarding", "device_code", deviceCode, "status", status.Status)
		return
	}

	sensorResult := s.db.Model(&device.SensorDevice{}).
		Where("device_code = ?", deviceCode).
		Update("status", status.Status)
	actuatorResult := s.db.Model(&device.ActuatorDevice{}).
		Where("device_code = ?", deviceCode).
		Update("status", status.Status)
	if sensorResult.RowsAffected == 0 && actuatorResult.RowsAffected == 0 {
		s.log.Warn("ingress: status from unknown device, discarding", "device_code", deviceCode)
		return
	}

	now := time.Now().UTC()
	s.hub.Publish(event.SSEEvent{
		Type: "device:status",
		Data: event.DeviceStatusSSEDataV1{
			SchemaVersion: 1,
			DeviceCode:    deviceCode,
			Status:        status.Status,
			ReportedAt:    now.Format(time.RFC3339),
		},
	})
}

// handleHeartbeat updates device last_seen_at and processes device metadata
func (s *IngressService) handleHeartbeat(deviceCode string, payload []byte) {
	sensorFound, actuatorFound := s.lookupDevicePresence(deviceCode)
	if !sensorFound && !actuatorFound {
		s.log.Warn("ingress: heartbeat from unknown device", "device_code", deviceCode)
		s.createUnknownDeviceAlert(deviceCode)
		return
	}

	now := time.Now().UTC()
	if sensorFound {
		s.db.Model(&device.SensorDevice{}).
			Where("device_code = ?", deviceCode).
			Updates(map[string]interface{}{
				"last_seen_at": now,
				"status":       device.StatusOnline,
			})
	}
	if actuatorFound {
		s.db.Model(&device.ActuatorDevice{}).
			Where("device_code = ?", deviceCode).
			Updates(map[string]interface{}{
				"last_seen_at": now,
				"status":       device.StatusOnline,
			})
	}

	var openOffline []alert.Alert
	if err := s.db.
		Where("type = ? AND status = ? AND message LIKE ?",
			alert.TypeDeviceOffline, alert.StatusOpen, "%["+deviceCode+"]%").
		Find(&openOffline).Error; err != nil {
		return
	}
	if len(openOffline) == 0 {
		return
	}

	_ = s.db.Transaction(func(tx *gorm.DB) error {
		for _, a := range openOffline {
			res := tx.Model(&alert.Alert{}).
				Where("id = ? AND status = ?", a.ID, alert.StatusOpen).
				Updates(map[string]interface{}{
					"status":      alert.StatusResolved,
					"resolved_at": now,
				})
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected == 0 {
				continue
			}
			payload := `{"reason":"heartbeat"}`
			_ = tx.Create(&alert.AlertTimelineEvent{
				AlertID:      a.ID,
				EventType:    alert.EventResolved,
				EventSource:  alert.SourceSystem,
				EventPayload: payload,
				EventTime:    now,
			}).Error
		}
		return nil
	})
}

// handleErrors processes device error reports and creates alerts
func (s *IngressService) handleErrors(deviceCode string, payload []byte) {
	var errReport struct {
		Level           string  `json:"level"`
		MetricCode      string  `json:"metric_code"`
		Message         string  `json:"message"`
		SensorChannelID *uint64 `json:"sensor_channel_id"`
	}
	if err := json.Unmarshal(payload, &errReport); err != nil {
		s.log.Error("ingress: invalid error payload", "device", deviceCode, "error", err)
		return
	}

	alertLevel := errReport.Level
	if alertLevel == "" {
		alertLevel = alert.LevelWarn
	}

	now := time.Now().UTC()
	a := alert.Alert{
		Type:        "DEVICE_ERROR",
		Level:       alertLevel,
		MetricCode:  errReport.MetricCode,
		Message:     fmt.Sprintf("[%s] %s", deviceCode, errReport.Message),
		Status:      alert.StatusOpen,
		TriggeredAt: now,
	}
	if errReport.SensorChannelID != nil {
		a.SensorChannelID = errReport.SensorChannelID
	}

	if err := s.db.Create(&a).Error; err != nil {
		s.log.Error("ingress: failed to create alert", "device", deviceCode, "error", err)
		return
	}

	s.db.Create(&alert.AlertTimelineEvent{
		AlertID:     a.ID,
		EventType:   alert.EventTriggered,
		EventSource: alert.SourceSystem,
		EventTime:   now,
	})

	s.hub.Publish(event.SSEEvent{Type: "alert:created", Data: alert.BuildAlertSSEDataV1(a, deviceCode, 1)})
}

// handleDiagnostics processes diagnostic data (log and store)
func (s *IngressService) handleDiagnostics(deviceCode string, payload []byte) {
	s.log.Info("ingress: diagnostics received", "device", deviceCode, "payload_len", len(payload))
	// Future: store in a diagnostics table or forward to monitoring system
}

// channelStateItem is a single channel state update within a handleState payload.
type channelStateItem struct {
	ChannelCode string   `json:"channel_code"`
	State       string   `json:"state"`
	Level       *float64 `json:"level"`
}

// handleState processes per-channel state updates from devices.
// Expected payload: {"channels":[{"channel_code":"fan-01","state":"ON"},...]}
func (s *IngressService) handleState(deviceCode string, payload []byte) {
	var req struct {
		Channels []channelStateItem `json:"channels"`
	}
	if err := json.Unmarshal(payload, &req); err != nil {
		s.log.Error("ingress: invalid state payload", "device", deviceCode, "error", err)
		return
	}
	if len(req.Channels) == 0 {
		return
	}

	updated := 0
	for _, ch := range req.Channels {
		if ch.ChannelCode == "" {
			continue
		}

		channelID, found, err := s.resolveActuatorChannelID(deviceCode, ch.ChannelCode)
		if err != nil {
			s.log.Error("ingress: failed to resolve actuator channel", "device", deviceCode, "channel", ch.ChannelCode, "error", err)
			continue
		}
		if !found {
			s.log.Warn("ingress: state for unknown channel", "device", deviceCode, "channel", ch.ChannelCode)
			continue
		}

		updates := map[string]interface{}{"current_state": ch.State}
		if ch.Level != nil {
			updates["current_level"] = *ch.Level
		}
		result := s.db.Model(&device.ActuatorChannel{}).
			Where("id = ?", channelID).
			Updates(updates)
		if result.Error != nil {
			s.log.Error("ingress: failed to update actuator channel state", "device", deviceCode, "channel", ch.ChannelCode, "channel_id", channelID, "error", result.Error)
			continue
		}
		if result.RowsAffected > 0 {
			updated++
		}
	}
	s.log.Debug("ingress: state updated", "device", deviceCode, "channels_updated", updated)
}

func (s *IngressService) resolveActuatorChannelID(deviceCode, channelCode string) (uint64, bool, error) {
	var row struct {
		ID uint64
	}
	err := s.db.Table("actuator_channels").
		Select("actuator_channels.id").
		Joins("JOIN actuator_devices ON actuator_devices.id = actuator_channels.actuator_device_id").
		Where("actuator_devices.device_code = ? AND actuator_channels.channel_code = ?", deviceCode, channelCode).
		Take(&row).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, false, nil
		}
		return 0, false, err
	}
	return row.ID, true, nil
}

// handleAck processes command acknowledgements
func (s *IngressService) handleAck(deviceCode string, payload []byte) {
	parsed, err := ParseAckPayload(payload)
	if err != nil {
		s.log.Error("ingress: invalid ack payload", "device", deviceCode, "error", err)
		return
	}

	now := time.Now().UTC()
	repo := NewConfigDeliveryRepo(s.db)

	if parsed.Kind == "v1_config" {
		env := parsed.V1
		if env == nil || env.MsgID == "" {
			return
		}
		ackedAt := now
		if env.DeviceTSms > 0 {
			ackedAt = time.UnixMilli(int64(env.DeviceTSms)).UTC()
		}

		ackPayloadJSON := "{}"
		if env.Payload != nil {
			if b, err := json.Marshal(env.Payload); err == nil {
				ackPayloadJSON = string(b)
			}
		}

		fwVersion := ""
		appliedHash := ""
		if env.Payload != nil {
			fwVersion, _ = env.Payload["fw_version"].(string)
			appliedHash, _ = env.Payload["applied_hash"].(string)
		}

		switch env.Result {
		case "ACKED":
			_, _ = repo.MarkAckedByMsgID(env.MsgID, ackedAt, ackPayloadJSON, fwVersion, appliedHash)
		case "REJECTED":
			_, _ = repo.MarkRejectedByMsgID(env.MsgID, ackedAt, ackPayloadJSON, env.ErrorCode, env.ErrorMessage, fwVersion, appliedHash)
		default:
			_, _ = repo.MarkAckFailedByMsgID(env.MsgID, ackedAt, ackPayloadJSON, env.ErrorCode, env.ErrorMessage, fwVersion, appliedHash)
		}
		return
	}

	if parsed.Kind == "v1_command" {
		env := parsed.V1
		if env == nil || env.Payload == nil {
			return
		}
		legacy := LegacyCommandAck{}
		if v, ok := env.Payload["command_id"]; ok {
			switch n := v.(type) {
			case float64:
				legacy.CommandID = uint64(n)
			case uint64:
				legacy.CommandID = n
			case int:
				if n > 0 {
					legacy.CommandID = uint64(n)
				}
			case int64:
				if n > 0 {
					legacy.CommandID = uint64(n)
				}
			}
		}
		if legacy.CommandID == 0 {
			return
		}
		if v, ok := env.Payload["ack_code"].(string); ok {
			legacy.AckCode = v
		}
		if v, ok := env.Payload["ack_message"].(string); ok {
			legacy.AckMessage = v
		}
		if v, ok := env.Payload["ack_payload"].(map[string]interface{}); ok {
			legacy.AckPayload = v
		}
		parsed = ParsedAck{Kind: "legacy_command", Legacy: &legacy}
	}

	if parsed.Kind != "legacy_command" || parsed.Legacy == nil {
		return
	}

	ack := parsed.Legacy

	s.hub.Publish(event.SSEEvent{
		Type: "command:acked",
		Data: event.CommandAckData{
			SchemaVersion: 1,
			CommandID:     ack.CommandID,
			DeviceCode:    deviceCode,
			AckCode:       ack.AckCode,
			AckMessage:    ack.AckMessage,
			AckPayload:    ack.AckPayload,
			AckedAt:       now.Format(time.RFC3339),
		},
	})

	s.db.Model(&struct{ ID uint64 }{}).
		Table("control_commands").
		Where("id = ?", ack.CommandID).
		Updates(map[string]interface{}{
			"status":   "ACKED",
			"acked_at": now,
		})
}
