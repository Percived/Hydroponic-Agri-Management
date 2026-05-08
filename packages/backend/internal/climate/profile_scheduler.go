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
		data, ok := evt.Data.(map[string]interface{})
		if !ok {
			continue
		}
		metricCode, _ := data["metric_code"].(string)
		value, _ := toFloat64(data["value"])
		if metricCode == "" || value == nil {
			continue
		}
		s.evaluateProfiles(metricCode, *value)
	}
}

// evaluateProfiles loads enabled climate profiles matching the trigger metric and evaluates each.
func (s *ProfileScheduler) evaluateProfiles(triggerMetric string, triggerValue float64) {
	var profiles []ClimateProfile
	err := s.db.
		Where("enabled = true AND trigger_metric_code = ?", triggerMetric).
		Find(&profiles).Error
	if err != nil {
		s.log.Error("profile_scheduler: failed to load profiles", "error", err)
		return
	}

	for _, p := range profiles {
		s.evaluateProfile(p, triggerValue)
	}
}

// evaluateProfile checks all stages for a profile and transitions to the highest
// matching stage if conditions are met.
func (s *ProfileScheduler) evaluateProfile(p ClimateProfile, triggerValue float64) {
	now := time.Now().UTC()

	// Check cooldown
	cooldownKey := p.ID
	if s.isInCooldown(cooldownKey) {
		return
	}

	// Load stages with enabled actions, ordered by stage_level ASC
	var stages []ClimateStage
	err := s.db.
		Preload("Actions", "enabled = true").
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

	// Find the highest matching stage
	var matchedStage *ClimateStage
	for i := range stages {
		if s.evaluateStageTrigger(stages[i], triggerValue) {
			matchedStage = &stages[i]
		}
	}

	if matchedStage == nil {
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

	// Skip if already at this stage (no transition needed)
	if fromStageLevel != nil && *fromStageLevel == matchedStage.StageLevel {
		return
	}

	// Execute actions for the matched stage
	executedCount := s.executeStageActions(matchedStage, p.ID)

	// Log execution
	execLog := ClimateExecutionLog{
		ProfileID:            p.ID,
		FromStageLevel:       fromStageLevel,
		ToStageLevel:         matchedStage.StageLevel,
		TriggerValue:         triggerValue,
		ExecutedActionsCount: executedCount,
		ExecutedAt:           now,
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
	cmd := command.ControlCommand{
		ActuatorChannelID: action.ActuatorChannelID,
		CommandType:       action.CommandType,
		Payload:           action.CommandPayload,
		Status:            "PENDING",
	}

	if err := s.db.Create(&cmd).Error; err != nil {
		return 0, fmt.Errorf("create command: %w", err)
	}

	if s.mqttClient == nil || !s.mqttClient.IsConnected() {
		now := time.Now().UTC()
		s.db.Model(&cmd).Updates(map[string]interface{}{
			"status":  "SENT",
			"sent_at": now,
		})
		return cmd.ID, nil
	}

	deviceCode, err := s.lookupActuatorDeviceCode(action.ActuatorChannelID)
	if err != nil {
		return cmd.ID, fmt.Errorf("device lookup: %w", err)
	}

	topic := fmt.Sprintf("%s/%s/%s/%s", mqttpkg.TopicPrefix, deviceCode, mqttpkg.TopicCmdPrefix, action.CommandType)
	token := s.mqttClient.Publish(topic, 1, false, action.CommandPayload)
	if token.Wait() && token.Error() != nil {
		return cmd.ID, fmt.Errorf("mqtt publish: %w", token.Error())
	}

	now := time.Now().UTC()
	s.db.Model(&cmd).Updates(map[string]interface{}{
		"status":  "SENT",
		"sent_at": now,
	})

	s.hub.Publish(event.SSEEvent{
		Type: "command:dispatched",
		Data: map[string]interface{}{
			"command_id": cmd.ID,
			"status":     "SENT",
			"profile_id": action.StageID,
		},
	})

	return cmd.ID, nil
}

// lookupActuatorDeviceCode resolves an actuator channel ID to its device code.
func (s *ProfileScheduler) lookupActuatorDeviceCode(actuatorChannelID uint64) (string, error) {
	var result struct {
		DeviceCode string
	}
	err := s.db.Table("actuator_channels").
		Select("actuator_devices.device_code").
		Joins("JOIN actuator_devices ON actuator_devices.id = actuator_channels.actuator_device_id").
		Where("actuator_channels.id = ?", actuatorChannelID).
		Scan(&result).Error
	if err != nil {
		return "", err
	}
	if result.DeviceCode == "" {
		return "", fmt.Errorf("device not found for channel %d", actuatorChannelID)
	}
	return result.DeviceCode, nil
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
