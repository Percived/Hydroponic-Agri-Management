package alert

import "time"

const (
	TypeThreshold     = "THRESHOLD"
	TypeDeviceOffline = "DEVICE_OFFLINE"
	TypeSystem        = "SYSTEM"

	LevelInfo     = "INFO"
	LevelWarn     = "WARN"
	LevelCritical = "CRITICAL"

	StatusOpen         = "OPEN"
	StatusAcknowledged = "ACKNOWLEDGED"
	StatusResolved     = "RESOLVED"
	StatusIgnored      = "IGNORED"

	EventTriggered    = "TRIGGERED"
	EventAutoAction   = "AUTO_ACTION"
	EventManualAction = "MANUAL_ACTION"
	EventAcknowledged = "ACKNOWLEDGED"
	EventResolved     = "RESOLVED"
	EventComment      = "COMMENT"

	SourceSystem = "SYSTEM"
	SourceManual = "MANUAL"
)

type Alert struct {
	ID                uint64     `gorm:"primaryKey;autoIncrement"`
	Type              string     `gorm:"size:32;not null"`
	Level             string     `gorm:"size:16;not null"`
	MetricCode        string     `gorm:"column:metric_code;size:32"`
	SensorChannelID   *uint64    `gorm:"column:sensor_channel_id"`
	ActuatorChannelID *uint64    `gorm:"column:actuator_channel_id"`
	BatchID           *uint64    `gorm:"column:batch_id"`
	TriggerValue      *float64   `gorm:"column:trigger_value;type:decimal(12,4)"`
	Message           string     `gorm:"size:255;not null"`
	Status            string     `gorm:"size:16;default:OPEN"`
	TriggeredAt       time.Time  `gorm:"column:triggered_at;not null"`
	ResolvedAt        *time.Time `gorm:"column:resolved_at"`
	ResolvedBy        *uint64    `gorm:"column:resolved_by"`
	CreatedAt         time.Time  `gorm:"autoCreateTime:milli"`
	UpdatedAt         time.Time  `gorm:"autoUpdateTime:milli"`
	// Associations
	TimelineEvents []AlertTimelineEvent `gorm:"foreignKey:AlertID"`
}

func (Alert) TableName() string { return "alerts" }

type AlertTimelineEvent struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement"`
	AlertID      uint64    `gorm:"column:alert_id;not null"`
	EventType    string    `gorm:"column:event_type;size:32;not null"`
	EventSource  string    `gorm:"column:event_source;size:16;not null"`
	OperatorID   *uint64   `gorm:"column:operator_id"`
	Comment      string    `gorm:"size:255"`
	EventPayload string    `gorm:"column:event_payload;type:json"`
	EventTime    time.Time `gorm:"column:event_time;not null"`
	CreatedAt    time.Time `gorm:"autoCreateTime:milli"`
}

func (AlertTimelineEvent) TableName() string { return "alert_timeline_events" }
