package alert

import "time"

// --- Request DTOs ---

type CreateAlertRequest struct {
	Type              string    `json:"type" binding:"required,oneof=THRESHOLD DEVICE_OFFLINE SYSTEM"`
	Level             string    `json:"level" binding:"required,oneof=INFO WARN CRITICAL"`
	MetricCode        string    `json:"metric_code" binding:"omitempty,max=32"`
	SensorChannelID   *uint64   `json:"sensor_channel_id"`
	ActuatorChannelID *uint64   `json:"actuator_channel_id"`
	TriggerValue      *float64  `json:"trigger_value"`
	Message           string    `json:"message" binding:"required,max=255"`
	TriggeredAt       time.Time `json:"triggered_at" binding:"required"`
}

type UpdateAlertStatusRequest struct {
	Status     string     `json:"status" binding:"required,oneof=OPEN ACKNOWLEDGED RESOLVED IGNORED"`
	ResolvedAt *time.Time `json:"resolved_at"`
	Comment    string     `json:"comment" binding:"omitempty,max=255"`
	ResolvedBy *uint64    `json:"resolved_by"`
}

type CreateTimelineEventRequest struct {
	EventType    string    `json:"event_type" binding:"required,oneof=TRIGGERED AUTO_ACTION MANUAL_ACTION ACKNOWLEDGED RESOLVED COMMENT"`
	EventSource  string    `json:"event_source" binding:"required,oneof=SYSTEM MANUAL"`
	OperatorID   *uint64   `json:"operator_id"`
	Comment      string    `json:"comment" binding:"omitempty,max=255"`
	EventPayload string    `json:"event_payload"`
	EventTime    time.Time `json:"event_time" binding:"required"`
}

// --- Response DTOs ---

type AlertResponse struct {
	ID                uint64   `json:"id"`
	Type              string   `json:"type"`
	Level             string   `json:"level"`
	MetricCode        string   `json:"metric_code"`
	SensorChannelID   *uint64  `json:"sensor_channel_id"`
	ActuatorChannelID *uint64  `json:"actuator_channel_id"`
	TriggerValue      *float64 `json:"trigger_value"`
	Message           string   `json:"message"`
	Status            string   `json:"status"`
	TriggeredAt       string   `json:"triggered_at"`
	ResolvedAt        *string  `json:"resolved_at"`
	ResolvedBy        *uint64  `json:"resolved_by"`
	TimelineCount     int64    `json:"timeline_count"`
	CreatedAt         string   `json:"created_at"`
	UpdatedAt         string   `json:"updated_at"`
}

type AlertTimelineEventResponse struct {
	ID           uint64  `json:"id"`
	AlertID      uint64  `json:"alert_id"`
	EventType    string  `json:"event_type"`
	EventSource  string  `json:"event_source"`
	OperatorID   *uint64 `json:"operator_id"`
	Comment      string  `json:"comment"`
	EventPayload string  `json:"event_payload"`
	EventTime    string  `json:"event_time"`
	CreatedAt    string  `json:"created_at"`
}

type AlertStatsResponse struct {
	OpenCount         int64 `json:"open_count"`
	AcknowledgedCount int64 `json:"acknowledged_count"`
	ResolvedCount     int64 `json:"resolved_count"`
	IgnoredCount      int64 `json:"ignored_count"`
	InfoCount         int64 `json:"info_count"`
	WarnCount         int64 `json:"warn_count"`
	CriticalCount     int64 `json:"critical_count"`
}

func timeToStr(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}

func timePtrToStr(t *time.Time) *string {
	if t == nil || t.IsZero() {
		return nil
	}
	s := t.Format(time.RFC3339)
	return &s
}
