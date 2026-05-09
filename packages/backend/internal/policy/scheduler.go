package policy

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
	DefaultScanInterval = 30 * time.Second
	DefaultCooldownSec  = 60
)

type conditionState struct {
	FirstTrueAt time.Time // when the condition first evaluated to true
}

type Scheduler struct {
	db         *gorm.DB
	hub        *event.Hub
	mqttClient mqtt.Client
	log        *slog.Logger

	scanInterval time.Duration
	cooldownSec  int

	cooldowns   map[string]time.Time
	cooldownsMu sync.RWMutex

	condStates   map[string]*conditionState // key = "policyID:conditionID"
	condStatesMu sync.RWMutex
}

func NewScheduler(db *gorm.DB, hub *event.Hub, mqttClient mqtt.Client, log *slog.Logger) *Scheduler {
	return &Scheduler{
		db:           db,
		hub:          hub,
		mqttClient:   mqttClient,
		log:          log,
		scanInterval: DefaultScanInterval,
		cooldownSec:  DefaultCooldownSec,
		cooldowns:    make(map[string]time.Time),
		condStates:   make(map[string]*conditionState),
	}
}

func (s *Scheduler) Start() {
	go s.runEventDriven()
	go s.runTimerScan()
	s.log.Info("policy scheduler started", "scan_interval", s.scanInterval, "cooldown_sec", s.cooldownSec)
}

func (s *Scheduler) runEventDriven() {
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
		s.evaluateThresholdPolicies(metricCode, *value, data)
	}
}

func (s *Scheduler) runTimerScan() {
	ticker := time.NewTicker(s.scanInterval)
	defer ticker.Stop()
	for range ticker.C {
		s.evaluateScheduledPolicies()
	}
}

func (s *Scheduler) evaluateThresholdPolicies(triggerMetric string, triggerValue float64, evtData map[string]interface{}) {
	now := time.Now().UTC()

	var policies []ControlPolicy
	err := s.db.
		Preload("Conditions", "enabled = true").
		Preload("Targets", "enabled = true").
		Where("control_policies.enabled = true AND policy_type = ?", "THRESHOLD").
		Where("(effective_from IS NULL OR effective_from <= ?)", now).
		Where("(effective_to IS NULL OR effective_to >= ?)", now).
		Joins("JOIN policy_conditions pc ON pc.policy_id = control_policies.id AND pc.enabled = true AND pc.metric_code = ?", triggerMetric).
		Group("control_policies.id").
		Order("priority ASC").
		Find(&policies).Error
	if err != nil {
		s.log.Error("scheduler: failed to load threshold policies", "error", err)
		return
	}

	for _, p := range policies {
		s.evaluateAndExecute(p, "TELEMETRY", triggerMetric, &triggerValue, evtData)
	}
}

func (s *Scheduler) evaluateScheduledPolicies() {
	now := time.Now().UTC()

	var policies []ControlPolicy
	err := s.db.
		Preload("Conditions", "enabled = true").
		Preload("Targets", "enabled = true").
		Where("control_policies.enabled = true AND policy_type = ?", "SCHEDULE").
		Where("(effective_from IS NULL OR effective_from <= ?)", now).
		Where("(effective_to IS NULL OR effective_to >= ?)", now).
		Order("priority ASC").
		Find(&policies).Error
	if err != nil {
		s.log.Error("scheduler: failed to load schedule policies", "error", err)
		return
	}

	for _, p := range policies {
		s.evaluateAndExecute(p, "SCHEDULE", "", nil, nil)
	}
}

func (s *Scheduler) evaluateAndExecute(p ControlPolicy, triggerSource, triggerMetric string, triggerValue *float64, evtData map[string]interface{}) {
	now := time.Now().UTC()

	if len(p.Conditions) == 0 || len(p.Targets) == 0 {
		return
	}

	decision := "SKIPPED"
	var reason string

	defer func() {
		exec := PolicyExecution{
			PolicyID:          p.ID,
			TriggerSource:     triggerSource,
			TriggerMetricCode: triggerMetric,
			TriggerValue:      triggerValue,
			Decision:          decision,
			DecisionReason:    reason,
			ExecutedAt:        &now,
		}
		if decision == "EXECUTED" {
			exec.ExecutedAt = &now
		}
		s.db.Create(&exec)
	}()

	// Check cooldown
	cooldownKey := fmt.Sprintf("%d", p.ID)
	if s.isInCooldown(cooldownKey) {
		reason = "COOLDOWN"
		decision = "SKIPPED"
		return
	}

	// Evaluate conditions
	allMet := true
	for _, cond := range p.Conditions {
		if !cond.Enabled {
			continue
		}
		if !s.evaluateCondition(cond, evtData) {
			allMet = false
			break
		}
	}

	if !allMet {
		reason = "CONDITION_NOT_MET"
		decision = "SKIPPED"
		return
	}

	// Execute targets
	for _, t := range p.Targets {
		if !t.Enabled {
			continue
		}

		_, err := s.executeTarget(t, p)
		if err != nil {
			s.log.Error("scheduler: failed to execute target",
				"policy_id", p.ID,
				"target_id", t.ID,
				"error", err,
			)
			decision = "FAILED"
			reason = fmt.Sprintf("target_execution_failed: %s", err.Error())
			return
		}
	}

	decision = "EXECUTED"
	reason = fmt.Sprintf("%s_trigger", triggerSource)

	// Record cooldown
	s.setCooldown(cooldownKey)

	// Reset condition duration tracking — after execution the cycle restarts
	s.resetPolicyCondStates(p)
}

func (s *Scheduler) evaluateCondition(cond PolicyCondition, evtData map[string]interface{}) bool {
	// Step 1: evaluate the raw condition (real-time event or DB fallback)
	var rawResult bool
	if evtData != nil {
		metricCode, _ := evtData["metric_code"].(string)
		if metricCode == cond.MetricCode {
			val, ok := toFloat64(evtData["value"])
			if ok {
				rawResult = compareWithHysteresis(*val, cond)
			}
		}
	}
	if evtData == nil || evtData["metric_code"].(string) != cond.MetricCode {
		rawResult = s.evaluateConditionFromDB(cond)
	}

	// Step 2: apply RequiredDurationSec if set
	if cond.RequiredDurationSec == nil || *cond.RequiredDurationSec == 0 {
		return rawResult
	}

	key := fmt.Sprintf("%d:%d", cond.PolicyID, cond.ID)
	now := time.Now().UTC()

	if !rawResult {
		// Condition no longer holds → reset tracking
		s.resetCondState(key)
		return false
	}

	// Condition holds → check / advance duration
	return s.trackCondDuration(key, *cond.RequiredDurationSec, now)
}

func compareWithHysteresis(val float64, cond PolicyCondition) bool {
	matched := compareValues(val, cond.ThresholdValue, cond.Operator)
	if matched && cond.Hysteresis != nil {
		hystVal := *cond.Hysteresis
		switch cond.Operator {
		case ">":
			matched = val > cond.ThresholdValue+hystVal
		case ">=":
			matched = val >= cond.ThresholdValue+hystVal
		case "<":
			matched = val < cond.ThresholdValue-hystVal
		case "<=":
			matched = val <= cond.ThresholdValue-hystVal
		}
	}
	return matched
}

func (s *Scheduler) evaluateConditionFromDB(cond PolicyCondition) bool {
	type telemetryRow struct {
		Value float64
	}
	var row telemetryRow
	query := s.db.Table("telemetry_records").
		Select("value").
		Where("metric_code = ?", cond.MetricCode)

	if cond.WindowSec != nil && *cond.WindowSec > 0 {
		query = query.Where("collected_at >= ?", time.Now().UTC().Add(-time.Duration(*cond.WindowSec)*time.Second))
	}

	switch cond.Aggregation {
	case "avg":
		query = query.Select("AVG(value) as value")
	case "max":
		query = query.Select("MAX(value) as value")
	case "min":
		query = query.Select("MIN(value) as value")
	case "last", "":
		query = query.Order("collected_at DESC").Limit(1)
	}

	if err := query.Scan(&row).Error; err != nil {
		s.log.Error("scheduler: failed to query telemetry for condition", "metric", cond.MetricCode, "error", err)
		return false
	}

	return compareValues(row.Value, cond.ThresholdValue, cond.Operator)
}

func (s *Scheduler) executeTarget(t PolicyTarget, p ControlPolicy) (uint64, error) {
	cmd := command.ControlCommand{
		ActuatorChannelID: t.ActuatorChannelID,
		CommandType:       t.CommandType,
		Payload:           t.CommandPayload,
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

	deviceCode, err := s.lookupActuatorDeviceCode(t.ActuatorChannelID)
	if err != nil {
		return cmd.ID, fmt.Errorf("device lookup: %w", err)
	}

	topic := fmt.Sprintf("%s/%s/%s/%s", mqttpkg.TopicPrefix, deviceCode, mqttpkg.TopicCmdPrefix, t.CommandType)
	token := s.mqttClient.Publish(topic, 1, false, t.CommandPayload)
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
			"policy_id":  p.ID,
		},
	})

	return cmd.ID, nil
}

func (s *Scheduler) lookupActuatorDeviceCode(actuatorChannelID uint64) (string, error) {
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

func (s *Scheduler) isInCooldown(key string) bool {
	s.cooldownsMu.RLock()
	defer s.cooldownsMu.RUnlock()
	last, ok := s.cooldowns[key]
	if !ok {
		return false
	}
	return time.Since(last) < time.Duration(s.cooldownSec)*time.Second
}

func (s *Scheduler) setCooldown(key string) {
	s.cooldownsMu.Lock()
	defer s.cooldownsMu.Unlock()
	s.cooldowns[key] = time.Now().UTC()
}

// ── RequiredDurationSec tracking ──

// trackCondDuration records the first-true time for a condition and returns
// whether the required duration has elapsed.
func (s *Scheduler) trackCondDuration(key string, requiredSec uint, now time.Time) bool {
	s.condStatesMu.Lock()
	defer s.condStatesMu.Unlock()

	state, exists := s.condStates[key]
	if !exists {
		s.condStates[key] = &conditionState{FirstTrueAt: now}
		return false
	}

	return now.Sub(state.FirstTrueAt) >= time.Duration(requiredSec)*time.Second
}

func (s *Scheduler) resetCondState(key string) {
	s.condStatesMu.Lock()
	defer s.condStatesMu.Unlock()
	delete(s.condStates, key)
}

// resetPolicyCondStates clears all condition duration tracking for a policy.
func (s *Scheduler) resetPolicyCondStates(p ControlPolicy) {
	s.condStatesMu.Lock()
	defer s.condStatesMu.Unlock()
	for _, cond := range p.Conditions {
		delete(s.condStates, fmt.Sprintf("%d:%d", p.ID, cond.ID))
	}
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
	case "==":
		return actual == threshold
	case "!=":
		return actual != threshold
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
