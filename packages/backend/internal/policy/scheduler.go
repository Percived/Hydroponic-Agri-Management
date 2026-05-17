package policy

import (
	"encoding/json"
	"fmt"
	"hydroponic-backend/internal/alert"
	"log/slog"
	"strings"
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
		evtData, ok := extractTelemetryTriggerData(evt.Data)
		if !ok {
			continue
		}
		metricCode, _ := evtData["metric_code"].(string)
		value, ok := toFloat64(evtData["value"])
		if !ok || metricCode == "" || value == nil {
			continue
		}
		s.evaluateThresholdPolicies(metricCode, *value, evtData)
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
		Preload("Targets", func(db *gorm.DB) *gorm.DB {
			return db.Where("enabled = true").Order("execution_order asc, id asc")
		}).
		Where("control_policies.enabled = true AND policy_type = ?", "THRESHOLD").
		Where("published_at IS NOT NULL").
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
		Preload("Targets", func(db *gorm.DB) *gorm.DB {
			return db.Where("enabled = true").Order("execution_order asc, id asc")
		}).
		Where("control_policies.enabled = true AND policy_type = ?", "SCHEDULE").
		Where("published_at IS NOT NULL").
		Order("priority ASC").
		Find(&policies).Error
	if err != nil {
		s.log.Error("scheduler: failed to load schedule policies", "error", err)
		return
	}

	for _, p := range policies {
		s.evaluateScheduledPolicy(p, now)
	}
}

func (s *Scheduler) evaluateScheduledPolicy(p ControlPolicy, now time.Time) {
	if p.ScheduleMode == nil {
		if p.LastScheduledFor == nil {
			if err := s.db.Model(&ControlPolicy{}).Where("id = ?", p.ID).Update("last_scheduled_for", now).Error; err != nil {
				s.log.Error("scheduler: failed to mark unconfigured schedule policy", "policy_id", p.ID, "error", err)
				return
			}
			s.recordPolicyExecution(p.ID, "SCHEDULE", nil, "SKIPPED", "schedule_not_configured", nil)
		}
		return
	}

	scheduledFor, due, err := s.findDueScheduledFor(p, now)
	if err != nil {
		s.recordPolicyExecution(p.ID, "SCHEDULE", nil, "SKIPPED", "schedule_not_configured", nil)
		return
	}
	if !due {
		return
	}
	claimed, err := s.claimScheduleSlot(p.ID, *scheduledFor)
	if err != nil {
		s.log.Error("scheduler: failed to claim scheduled slot", "policy_id", p.ID, "scheduled_for", scheduledFor, "error", err)
		return
	}
	if !claimed {
		return
	}
	if !s.withinEffectiveWindow(p, *scheduledFor) {
		s.recordPolicyExecution(p.ID, "SCHEDULE", scheduledFor, "SKIPPED", "outside_effective_window", nil)
		return
	}
	if len(p.Targets) == 0 {
		s.recordPolicyExecution(p.ID, "SCHEDULE", scheduledFor, "SKIPPED", "no_targets", nil)
		return
	}

	var lastCommandID *uint64
	successfulTargets := 0
	for _, t := range p.Targets {
		if !t.Enabled {
			continue
		}

		cmdID, execErr := s.executeTarget(t, p)
		if execErr != nil {
			if successfulTargets == 0 {
				s.restoreScheduleSlotClaim(p.ID, p.LastScheduledFor)
			}
			s.log.Error("scheduler: failed to execute scheduled target",
				"policy_id", p.ID,
				"target_id", t.ID,
				"error", execErr,
			)
			s.recordPolicyExecution(p.ID, "SCHEDULE", scheduledFor, "FAILED", fmt.Sprintf("target_execution_failed: %s", execErr.Error()), nil)
			return
		}
		lastCommandID = &cmdID
		successfulTargets++
	}

	s.recordPolicyExecution(p.ID, "SCHEDULE", scheduledFor, "EXECUTED", "schedule_due", lastCommandID)
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

	if p.PolicyType == "THRESHOLD" {
		if err := s.createThresholdAlert(p, triggerMetric, triggerValue, evtData, now); err != nil {
			s.log.Error("scheduler: failed to create threshold alert",
				"policy_id", p.ID,
				"metric_code", triggerMetric,
				"error", err,
			)
		}
	}

	// Record cooldown
	s.setCooldown(cooldownKey)

	// Reset condition duration tracking 鈥?after execution the cycle restarts
	s.resetPolicyCondStates(p)
}

func (s *Scheduler) recordPolicyExecution(policyID uint64, triggerSource string, executedAt *time.Time, decision, reason string, commandID *uint64) {
	exec := PolicyExecution{
		PolicyID:       policyID,
		TriggerSource:  triggerSource,
		Decision:       decision,
		DecisionReason: reason,
		CommandID:      commandID,
	}
	if executedAt != nil && decision == "EXECUTED" {
		exec.ExecutedAt = executedAt
	}
	if err := s.db.Create(&exec).Error; err != nil {
		s.log.Error("scheduler: failed to create policy execution", "policy_id", policyID, "decision", decision, "error", err)
	}
}

func (s *Scheduler) withinEffectiveWindow(p ControlPolicy, scheduledFor time.Time) bool {
	if p.EffectiveFrom != nil && scheduledFor.Before(p.EffectiveFrom.UTC()) {
		return false
	}
	if p.EffectiveTo != nil && scheduledFor.After(p.EffectiveTo.UTC()) {
		return false
	}
	return true
}

func (s *Scheduler) findDueScheduledFor(p ControlPolicy, now time.Time) (*time.Time, bool, error) {
	if p.ScheduleMode == nil {
		return nil, false, fmt.Errorf("schedule mode not configured")
	}

	loc := time.UTC
	if p.Timezone != "" {
		loaded, err := time.LoadLocation(p.Timezone)
		if err != nil {
			return nil, false, err
		}
		loc = loaded
	}

	switch *p.ScheduleMode {
	case "ONCE":
		if p.RunOnceAt == nil {
			return nil, false, fmt.Errorf("run_once_at required")
		}
		slot := p.RunOnceAt.UTC()
		return &slot, !slot.After(now), nil
	case "DAILY":
		if p.TimeOfDay == nil {
			return nil, false, fmt.Errorf("time_of_day required")
		}
		slot, err := buildLatestDailyScheduleSlot(now, loc, *p.TimeOfDay)
		if err != nil {
			return nil, false, err
		}
		return &slot, true, nil
	case "WEEKLY":
		if p.TimeOfDay == nil || p.WeekdaysMask == nil || *p.WeekdaysMask == 0 {
			return nil, false, fmt.Errorf("weekly schedule requires time_of_day and weekdays_mask")
		}
		slot, ok, err := buildLatestWeeklyScheduleSlot(now, loc, *p.TimeOfDay, *p.WeekdaysMask)
		if err != nil {
			return nil, false, err
		}
		if !ok {
			return nil, false, nil
		}
		return &slot, true, nil
	default:
		return nil, false, fmt.Errorf("unsupported schedule mode %q", *p.ScheduleMode)
	}
}

func buildLatestDailyScheduleSlot(now time.Time, loc *time.Location, timeOfDay string) (time.Time, error) {
	slotClock, err := time.ParseInLocation("15:04:05", timeOfDay, loc)
	if err != nil {
		return time.Time{}, err
	}
	nowLocal := now.In(loc)
	slotLocal := time.Date(nowLocal.Year(), nowLocal.Month(), nowLocal.Day(), slotClock.Hour(), slotClock.Minute(), slotClock.Second(), 0, loc)
	if slotLocal.After(nowLocal) {
		slotLocal = slotLocal.AddDate(0, 0, -1)
	}
	return slotLocal.UTC(), nil
}

func buildLatestWeeklyScheduleSlot(now time.Time, loc *time.Location, timeOfDay string, weekdaysMask uint8) (time.Time, bool, error) {
	slotClock, err := time.ParseInLocation("15:04:05", timeOfDay, loc)
	if err != nil {
		return time.Time{}, false, err
	}
	nowLocal := now.In(loc)
	for offset := 0; offset < 7; offset++ {
		candidateDate := nowLocal.AddDate(0, 0, -offset)
		if !weekdayMatched(weekdaysMask, candidateDate.Weekday()) {
			continue
		}
		slotLocal := time.Date(candidateDate.Year(), candidateDate.Month(), candidateDate.Day(), slotClock.Hour(), slotClock.Minute(), slotClock.Second(), 0, loc)
		if slotLocal.After(nowLocal) {
			continue
		}
		return slotLocal.UTC(), true, nil
	}
	return time.Time{}, false, nil
}

func (s *Scheduler) claimScheduleSlot(policyID uint64, scheduledFor time.Time) (bool, error) {
	result := s.db.Model(&ControlPolicy{}).
		Where("id = ? AND (last_scheduled_for IS NULL OR last_scheduled_for < ?)", policyID, scheduledFor).
		Update("last_scheduled_for", scheduledFor)
	return result.RowsAffected > 0, result.Error
}

func (s *Scheduler) restoreScheduleSlotClaim(policyID uint64, previous *time.Time) {
	updates := map[string]interface{}{"last_scheduled_for": previous}
	if err := s.db.Model(&ControlPolicy{}).Where("id = ?", policyID).Updates(updates).Error; err != nil {
		s.log.Error("scheduler: failed to restore schedule slot claim", "policy_id", policyID, "error", err)
	}
}

func weekdayMatched(mask uint8, weekday time.Weekday) bool {
	if weekday == time.Sunday {
		return mask&(1<<6) != 0
	}
	if weekday < time.Monday || weekday > time.Saturday {
		return false
	}
	return mask&(1<<(weekday-1)) != 0
}

func (s *Scheduler) evaluateCondition(cond PolicyCondition, evtData map[string]interface{}) bool {
	needsWindow := cond.WindowSec != nil && *cond.WindowSec > 0

	if needsWindow || evtData == nil {
		return s.evaluateConditionFromDB(cond)
	}

	metricCode, _ := evtData["metric_code"].(string)
	if metricCode != cond.MetricCode {
		return s.evaluateConditionFromDB(cond)
	}

	val, ok := toFloat64(evtData["value"])
	if !ok {
		return s.evaluateConditionFromDB(cond)
	}

	rawResult := compareWithHysteresis(*val, cond)

	if cond.RequiredDurationSec == nil || *cond.RequiredDurationSec == 0 {
		return rawResult
	}

	key := fmt.Sprintf("%d:%d", cond.PolicyID, cond.ID)
	now := time.Now().UTC()

	if !rawResult {
		s.resetCondState(key)
		return false
	}

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

	var query *gorm.DB
	switch cond.Aggregation {
	case "avg":
		query = s.db.Table("telemetry_records").Select("AVG(value) as value")
	case "max":
		query = s.db.Table("telemetry_records").Select("MAX(value) as value")
	case "min":
		query = s.db.Table("telemetry_records").Select("MIN(value) as value")
	case "last", "":
		query = s.db.Table("telemetry_records").Select("value").Order("collected_at DESC").Limit(1)
	}

	query = query.Where("metric_code = ?", cond.MetricCode)

	if cond.WindowSec != nil && *cond.WindowSec > 0 {
		query = query.Where("collected_at >= ?", time.Now().UTC().Add(-time.Duration(*cond.WindowSec)*time.Second))
	}

	if err := query.Scan(&row).Error; err != nil {
		s.log.Error("scheduler: failed to query telemetry for condition", "metric", cond.MetricCode, "error", err)
		return false
	}

	return compareValues(row.Value, cond.ThresholdValue, cond.Operator)
}

func (s *Scheduler) executeTarget(t PolicyTarget, p ControlPolicy) (uint64, error) {
	creatorID, err := s.resolveCommandCreator(p)
	if err != nil {
		return 0, err
	}

	cmd := command.ControlCommand{
		ActuatorChannelID: t.ActuatorChannelID,
		CommandType:       t.CommandType,
		Payload:           t.CommandPayload,
		Status:            "PENDING",
		CreatedBy:         creatorID,
	}

	if err := s.db.Create(&cmd).Error; err != nil {
		return 0, fmt.Errorf("create command: %w", err)
	}

	deviceCode, channelCode, _ := s.lookupActuatorTarget(t.ActuatorChannelID)
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
				SourceType:    "POLICY",
				SourceID:      p.ID,
				ErrorMessage:  "mqtt offline",
			},
		})
		return cmd.ID, fmt.Errorf("mqtt offline")
	}

	payload := command.BuildDeviceCommandPayload(t.CommandPayload, command.DispatchTargetMeta{
		CommandID:         cmd.ID,
		CommandType:       t.CommandType,
		ActuatorChannelID: t.ActuatorChannelID,
		ChannelCode:       channelCode,
	})
	topic := fmt.Sprintf("%s/%s/%s/%s", mqttpkg.TopicPrefix, deviceCode, mqttpkg.TopicCmdPrefix, t.CommandType)
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
				SourceType:    "POLICY",
				SourceID:      p.ID,
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
			SourceType:    "POLICY",
			SourceID:      p.ID,
		},
	})

	return cmd.ID, nil
}

func (s *Scheduler) createThresholdAlert(
	p ControlPolicy,
	triggerMetric string,
	triggerValue *float64,
	evtData map[string]interface{},
	triggeredAt time.Time,
) error {
	message := fmt.Sprintf("阈值策略触发: %s", p.Name)
	if triggerMetric != "" && triggerValue != nil {
		message = fmt.Sprintf("阈值策略触发: %s (%s=%.4f)", p.Name, triggerMetric, *triggerValue)
	}

	var sensorChannelID *uint64
	if id, ok := toUint64(evtData["sensor_channel_id"]); ok {
		sensorChannelID = &id
	}

	a := alert.Alert{
		Type:            alert.TypeThreshold,
		Level:           alert.LevelWarn,
		MetricCode:      triggerMetric,
		SensorChannelID: sensorChannelID,
		TriggerValue:    triggerValue,
		Message:         message,
		Status:          alert.StatusOpen,
		TriggeredAt:     triggeredAt,
	}
	if err := s.db.Create(&a).Error; err != nil {
		return err
	}

	payload, _ := json.Marshal(map[string]interface{}{
		"policy_id":      p.ID,
		"policy_code":    p.PolicyCode,
		"policy_name":    p.Name,
		"trigger_metric": triggerMetric,
		"trigger_value":  triggerValue,
		"sensor_channel": sensorChannelID,
		"trigger_source": "TELEMETRY",
	})
	timeline := alert.AlertTimelineEvent{
		AlertID:      a.ID,
		EventType:    alert.EventTriggered,
		EventSource:  alert.SourceSystem,
		EventPayload: string(payload),
		EventTime:    triggeredAt,
	}
	if err := s.db.Create(&timeline).Error; err != nil {
		return err
	}

	if s.hub != nil {
		deviceCode, _ := evtData["device_code"].(string)
		s.hub.Publish(event.SSEEvent{
			Type: "alert:created",
			Data: alert.BuildAlertSSEDataV1(a, deviceCode, 1),
		})
	}

	return nil
}

func (s *Scheduler) resolveCommandCreator(p ControlPolicy) (uint64, error) {
	if p.PublishedBy != nil && *p.PublishedBy != 0 {
		return *p.PublishedBy, nil
	}
	if p.CreatedBy != nil && *p.CreatedBy != 0 {
		return *p.CreatedBy, nil
	}

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

func (s *Scheduler) lookupActuatorTarget(actuatorChannelID uint64) (string, string, error) {
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

func (s *Scheduler) ResetPolicyRuntime(policyID uint64) {
	cooldownKey := fmt.Sprintf("%d", policyID)

	s.cooldownsMu.Lock()
	delete(s.cooldowns, cooldownKey)
	s.cooldownsMu.Unlock()

	prefix := fmt.Sprintf("%d:", policyID)
	s.condStatesMu.Lock()
	for key := range s.condStates {
		if strings.HasPrefix(key, prefix) {
			delete(s.condStates, key)
		}
	}
	s.condStatesMu.Unlock()
}

// 鈹€鈹€ RequiredDurationSec tracking 鈹€鈹€

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

func toUint64(v interface{}) (uint64, bool) {
	switch val := v.(type) {
	case uint64:
		return val, true
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
	default:
		return 0, false
	}
}

func extractTelemetryTriggerData(data interface{}) (map[string]interface{}, bool) {
	switch payload := data.(type) {
	case event.TelemetrySSEDataV1:
		return map[string]interface{}{
			"metric_code":       payload.MetricCode,
			"value":             payload.Value,
			"sensor_channel_id": payload.SensorChannelID,
			"device_code":       payload.DeviceCode,
			"collected_at":      payload.CollectedAt,
		}, true
	case *event.TelemetrySSEDataV1:
		if payload == nil {
			return nil, false
		}
		return map[string]interface{}{
			"metric_code":       payload.MetricCode,
			"value":             payload.Value,
			"sensor_channel_id": payload.SensorChannelID,
			"device_code":       payload.DeviceCode,
			"collected_at":      payload.CollectedAt,
		}, true
	case map[string]interface{}:
		return payload, true
	default:
		return nil, false
	}
}
