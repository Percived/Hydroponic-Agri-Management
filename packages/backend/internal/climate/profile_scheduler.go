package climate

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"hydroponic-backend/internal/command"
	"hydroponic-backend/internal/platform/event"
	mqttpkg "hydroponic-backend/internal/platform/mqtt"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"gorm.io/gorm"
)

const (
	DefaultCooldownSec = 60
)

// ProfileScheduler listens for telemetry events and evaluates climate profiles
// to determine if a stage transition should occur.
type ProfileScheduler struct {
	db         *gorm.DB
	hub        *event.Hub
	mqttClient mqtt.Client
	log        *slog.Logger

	cooldownSec int

	cooldowns   map[uint64]time.Time // key = profileID
	cooldownsMu sync.RWMutex
}

// NewProfileScheduler creates a new ProfileScheduler.
func NewProfileScheduler(db *gorm.DB, hub *event.Hub, mqttClient mqtt.Client, log *slog.Logger) *ProfileScheduler {
	return &ProfileScheduler{
		db:          db,
		hub:         hub,
		mqttClient:  mqttClient,
		log:         log,
		cooldownSec: DefaultCooldownSec,
		cooldowns:   make(map[uint64]time.Time),
	}
}

// Start launches the event-driven evaluation goroutine.
func (s *ProfileScheduler) Start() {
	go s.runEventDriven()
	s.log.Info("climate profile scheduler started", "cooldown_sec", s.cooldownSec)
}

// runEventDriven subscribes to telemetry:received events and evaluates matching profiles.
func (s *ProfileScheduler) runEventDriven() {
	sub := s.hub.Subscribe(func(e event.SSEEvent) bool {
		return e.Type == "telemetry:received"
	})
	defer s.hub.Unsubscribe(sub)

	for evt := range sub.Events {
		sensorChannelID, metricCode, value, collectedAt, ok := extractClimateTelemetryTrigger(evt.Data)
		if metricCode == "" || value == nil {
			continue
		}
		if !ok {
			continue
		}
		s.evaluateProfilesByChannel(sensorChannelID, metricCode, *value, collectedAt)
	}
}

func (s *ProfileScheduler) evaluateProfilesByChannel(triggerSensorChannelID uint64, triggerMetric string, triggerValue float64, collectedAt *time.Time) {
	var profiles []ClimateProfile
	err := s.db.
		Where("enabled = true AND trigger_sensor_channel_id = ?", triggerSensorChannelID).
		Find(&profiles).Error
	if err != nil {
		s.log.Error("profile_scheduler: failed to load profiles", "error", err)
		return
	}

	for _, p := range profiles {
		s.evaluateProfile(p, triggerSensorChannelID, triggerMetric, triggerValue, collectedAt)
	}
}

// evaluateProfile checks all stages for a profile and transitions to the highest
// matching stage if conditions are met.
func (s *ProfileScheduler) evaluateProfile(p ClimateProfile, triggerSensorChannelID uint64, triggerMetric string, triggerValue float64, collectedAt *time.Time) {
	now := time.Now().UTC()

	// Check cooldown
	cooldownKey := p.ID
	if s.isInCooldown(cooldownKey) {
		return
	}

	// Load stages with enabled actions, ordered by stage_level ASC
	var stages []ClimateStage
	err := s.db.
		Preload("Actions", func(db *gorm.DB) *gorm.DB {
			return db.Where("enabled = true").Order("execution_order asc, id asc")
		}).
		Where("profile_id = ?", p.ID).
		Order("stage_level ASC").
		Find(&stages).Error
	if err != nil {
		s.log.Error("profile_scheduler: failed to load stages", "profile_id", p.ID, "error", err)
		return
	}

	if len(stages) == 0 {
		return
	}

	// Determine the previous stage level (from the latest execution log)
	var fromStageLevel *uint8
	var lastLog ClimateExecutionLog
	err = s.db.Where("profile_id = ?", p.ID).
		Order("executed_at DESC").
		First(&lastLog).Error
	if err == nil {
		fromStageLevel = &lastLog.ToStageLevel
	}

	matchedStage := s.selectStageWithStatefulHysteresis(stages, fromStageLevel, triggerValue)
	if matchedStage == nil {
		return
	}

	// Skip if already at this stage (no transition needed)
	if fromStageLevel != nil && *fromStageLevel == matchedStage.StageLevel {
		return
	}

	// Execute actions for the matched stage
	executedCount := s.executeStageActions(matchedStage, p.ID)

	// Log execution
	execLog := ClimateExecutionLog{
		ProfileID:              p.ID,
		FromStageLevel:         fromStageLevel,
		ToStageLevel:           matchedStage.StageLevel,
		TriggerValue:           triggerValue,
		TriggerSensorChannelID: &triggerSensorChannelID,
		TriggerMetricCode:      &triggerMetric,
		CollectedAt:            collectedAt,
		ExecutedActionsCount:   executedCount,
		ExecutedAt:             now,
	}
	s.db.Create(&execLog)

	s.log.Info("profile_scheduler: stage transition executed",
		"profile_id", p.ID,
		"from", fromStageLevel,
		"to", matchedStage.StageLevel,
		"trigger_value", triggerValue,
		"actions_executed", executedCount,
	)

	// Set cooldown
	s.setCooldown(cooldownKey)
}

// evaluateStageTrigger checks if a telemetry value satisfies a stage's trigger condition with hysteresis.
func (s *ProfileScheduler) evaluateStageTrigger(stage ClimateStage, value float64) bool {
	return compareWithHysteresis(value, stage.TriggerOperator, stage.TriggerThreshold, stage.Hysteresis)
}

// executeStageActions dispatches all enabled actions for a stage via MQTT.
func (s *ProfileScheduler) executeStageActions(stage *ClimateStage, profileID uint64) uint {
	var executedCount uint
	for _, action := range stage.Actions {
		if !action.Enabled {
			continue
		}

		_, err := s.executeAction(action)
		if err != nil {
			s.log.Error("profile_scheduler: failed to execute action",
				"profile_id", profileID,
				"stage_id", stage.ID,
				"action_id", action.ID,
				"error", err,
			)
			continue
		}
		executedCount++
	}
	return executedCount
}

// executeAction creates a control command and publishes it via MQTT.
func (s *ProfileScheduler) executeAction(action ClimateStageAction) (uint64, error) {
	creatorID, err := s.resolveCommandCreator()
	if err != nil {
		return 0, err
	}

	cmd := command.ControlCommand{
		ActuatorChannelID: action.ActuatorChannelID,
		CommandType:       action.CommandType,
		Payload:           action.CommandPayload,
		Status:            "PENDING",
		CreatedBy:         creatorID,
	}

	if err := s.db.Create(&cmd).Error; err != nil {
		return 0, fmt.Errorf("create command: %w", err)
	}

	deviceCode, channelCode, _ := s.lookupActuatorTarget(action.ActuatorChannelID)
	now := time.Now().UTC()

	if s.mqttClient == nil || !s.mqttClient.IsConnected() {
		s.db.Model(&cmd).Update("status", "FAILED")
		s.hub.Publish(event.SSEEvent{
			Type: "command:dispatched",
			Data: event.CommandDispatchedSSEDataV1{
				SchemaVersion: 1,
				CommandID:     cmd.ID,
				DeviceCode:    deviceCode,
				Status:        "FAILED",
				DispatchedAt:  now.Format(time.RFC3339),
				SourceType:    "CLIMATE",
				SourceID:      action.StageID,
				ErrorMessage:  "mqtt offline",
			},
		})
		return cmd.ID, fmt.Errorf("mqtt offline")
	}

	payload := command.BuildDeviceCommandPayload(action.CommandPayload, command.DispatchTargetMeta{
		CommandID:         cmd.ID,
		CommandType:       action.CommandType,
		ActuatorChannelID: action.ActuatorChannelID,
		ChannelCode:       channelCode,
	})
	topic := fmt.Sprintf("%s/%s/%s/%s", mqttpkg.TopicPrefix, deviceCode, mqttpkg.TopicCmdPrefix, action.CommandType)
	token := s.mqttClient.Publish(topic, 1, false, payload)
	if token.Wait() && token.Error() != nil {
		s.db.Model(&cmd).Update("status", "FAILED")
		s.hub.Publish(event.SSEEvent{
			Type: "command:dispatched",
			Data: event.CommandDispatchedSSEDataV1{
				SchemaVersion: 1,
				CommandID:     cmd.ID,
				DeviceCode:    deviceCode,
				Status:        "FAILED",
				DispatchedAt:  now.Format(time.RFC3339),
				SourceType:    "CLIMATE",
				SourceID:      action.StageID,
				ErrorMessage:  token.Error().Error(),
			},
		})
		return cmd.ID, fmt.Errorf("mqtt publish: %w", token.Error())
	}

	s.db.Model(&cmd).Updates(map[string]interface{}{
		"status":  "SENT",
		"sent_at": now,
	})

	s.hub.Publish(event.SSEEvent{
		Type: "command:dispatched",
		Data: event.CommandDispatchedSSEDataV1{
			SchemaVersion: 1,
			CommandID:     cmd.ID,
			DeviceCode:    deviceCode,
			Status:        "SENT",
			DispatchedAt:  now.Format(time.RFC3339),
			SourceType:    "CLIMATE",
			SourceID:      action.StageID,
		},
	})

	return cmd.ID, nil
}

func (s *ProfileScheduler) resolveCommandCreator() (uint64, error) {
	var fallback struct {
		ID uint64
	}
	if err := s.db.Table("users").Select("id").Order("id asc").Limit(1).Scan(&fallback).Error; err != nil {
		return 0, fmt.Errorf("resolve command creator: %w", err)
	}
	if fallback.ID == 0 {
		return 0, fmt.Errorf("resolve command creator: no available user")
	}
	return fallback.ID, nil
}

// lookupActuatorDeviceCode resolves an actuator channel ID to its device code.
func (s *ProfileScheduler) lookupActuatorTarget(actuatorChannelID uint64) (string, string, error) {
	var result struct {
		DeviceCode  string
		ChannelCode string
	}
	err := s.db.Table("actuator_channels").
		Select("actuator_devices.device_code, actuator_channels.channel_code").
		Joins("JOIN actuator_devices ON actuator_devices.id = actuator_channels.actuator_device_id").
		Where("actuator_channels.id = ?", actuatorChannelID).
		Scan(&result).Error
	if err != nil {
		return "", "", err
	}
	if result.DeviceCode == "" {
		return "", "", fmt.Errorf("device not found for channel %d", actuatorChannelID)
	}
	return result.DeviceCode, result.ChannelCode, nil
}

// --- Cooldown helpers ---

func (s *ProfileScheduler) isInCooldown(key uint64) bool {
	s.cooldownsMu.RLock()
	defer s.cooldownsMu.RUnlock()
	last, ok := s.cooldowns[key]
	if !ok {
		return false
	}
	return time.Since(last) < time.Duration(s.cooldownSec)*time.Second
}

func (s *ProfileScheduler) setCooldown(key uint64) {
	s.cooldownsMu.Lock()
	defer s.cooldownsMu.Unlock()
	s.cooldowns[key] = time.Now().UTC()
}

// --- Value comparison helpers ---

// compareWithHysteresis evaluates whether a value satisfies stage trigger condition, with hysteresis.
func compareWithHysteresis(val float64, operator string, threshold, hysteresis float64) bool {
	matched := compareValues(val, threshold, operator)
	if matched && hysteresis != 0 {
		switch operator {
		case ">":
			matched = val > threshold+hysteresis
		case ">=":
			matched = val >= threshold+hysteresis
		case "<":
			matched = val < threshold-hysteresis
		case "<=":
			matched = val <= threshold-hysteresis
		}
	}
	return matched
}

func compareValues(actual, threshold float64, operator string) bool {
	switch operator {
	case ">":
		return actual > threshold
	case ">=":
		return actual >= threshold
	case "<":
		return actual < threshold
	case "<=":
		return actual <= threshold
	default:
		return false
	}
}

func toFloat64(v interface{}) (*float64, bool) {
	switch val := v.(type) {
	case float64:
		return &val, true
	case float32:
		f := float64(val)
		return &f, true
	case int:
		f := float64(val)
		return &f, true
	case int64:
		f := float64(val)
		return &f, true
	case json.Number:
		f, err := val.Float64()
		if err != nil {
			return nil, false
		}
		return &f, true
	default:
		return nil, false
	}
}

func extractClimateTelemetryTrigger(data interface{}) (uint64, string, *float64, *time.Time, bool) {
	switch payload := data.(type) {
	case event.TelemetrySSEDataV1:
		value := payload.Value
		collectedAt := parseRFC3339Time(payload.CollectedAt)
		return payload.SensorChannelID, payload.MetricCode, &value, collectedAt, true
	case *event.TelemetrySSEDataV1:
		if payload == nil {
			return 0, "", nil, nil, false
		}
		value := payload.Value
		collectedAt := parseRFC3339Time(payload.CollectedAt)
		return payload.SensorChannelID, payload.MetricCode, &value, collectedAt, true
	case map[string]interface{}:
		sensorChannelID, ok := toUint64(payload["sensor_channel_id"])
		if !ok {
			return 0, "", nil, nil, false
		}
		metricCode, _ := payload["metric_code"].(string)
		value, ok := toFloat64(payload["value"])
		if !ok {
			return 0, "", nil, nil, false
		}
		collectedAt := parseRFC3339Time(payload["collected_at"])
		return sensorChannelID, metricCode, value, collectedAt, true
	default:
		return 0, "", nil, nil, false
	}
}

func toUint64(v interface{}) (uint64, bool) {
	switch val := v.(type) {
	case uint64:
		return val, true
	case uint32:
		return uint64(val), true
	case uint:
		return uint64(val), true
	case int:
		if val < 0 {
			return 0, false
		}
		return uint64(val), true
	case int64:
		if val < 0 {
			return 0, false
		}
		return uint64(val), true
	case float64:
		if val < 0 {
			return 0, false
		}
		return uint64(val), true
	case json.Number:
		i, err := val.Int64()
		if err != nil || i < 0 {
			return 0, false
		}
		return uint64(i), true
	case string:
		i, err := json.Number(val).Int64()
		if err != nil || i < 0 {
			return 0, false
		}
		return uint64(i), true
	default:
		return 0, false
	}
}

func parseRFC3339Time(v interface{}) *time.Time {
	s, ok := v.(string)
	if !ok || s == "" {
		return nil
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		tt := t.UTC()
		return &tt
	}
	if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
		tt := t.UTC()
		return &tt
	}
	return nil
}

func selectStageDirection(op string) (string, bool) {
	switch op {
	case ">", ">=":
		return "up", true
	case "<", "<=":
		return "down", true
	default:
		return "", false
	}
}

func (s *ProfileScheduler) selectStageWithStatefulHysteresis(stages []ClimateStage, currentStageLevel *uint8, value float64) *ClimateStage {
	if len(stages) == 0 {
		return nil
	}
	dir, ok := selectStageDirection(stages[0].TriggerOperator)
	if !ok {
		return nil
	}

	var current *ClimateStage
	if currentStageLevel != nil {
		for i := range stages {
			if stages[i].StageLevel == *currentStageLevel {
				current = &stages[i]
				break
			}
		}
	}

	var candidate *ClimateStage
	for i := range stages {
		if compareValues(value, stages[i].TriggerThreshold, stages[i].TriggerOperator) {
			candidate = &stages[i]
		}
	}
	if candidate == nil {
		return nil
	}

	if current != nil {
		if candidate.StageLevel == current.StageLevel {
			return candidate
		}
		if dir == "up" {
			if candidate.StageLevel > current.StageLevel {
				if value < candidate.TriggerThreshold+candidate.Hysteresis {
					return nil
				}
			} else {
				if value >= current.TriggerThreshold-current.Hysteresis {
					return nil
				}
			}
		} else {
			if candidate.StageLevel > current.StageLevel {
				if value > candidate.TriggerThreshold-candidate.Hysteresis {
					return nil
				}
			} else {
				if value <= current.TriggerThreshold+current.Hysteresis {
					return nil
				}
			}
		}
	} else {
		if dir == "up" {
			if value < candidate.TriggerThreshold+candidate.Hysteresis {
				return nil
			}
		} else {
			if value > candidate.TriggerThreshold-candidate.Hysteresis {
				return nil
			}
		}
	}

	return candidate
}
