package policy

import "time"

// ControlPolicy represents a control policy for threshold/schedule/duration-based automation.
type ControlPolicy struct {
	ID            uint64     `gorm:"primaryKey;autoIncrement"`
	PolicyCode    string     `gorm:"column:policy_code;size:64;uniqueIndex;not null"`
	Name          string     `gorm:"size:128;not null"`
	PolicyType    string     `gorm:"column:policy_type;size:16;not null"` // THRESHOLD/SCHEDULE/DURATION
	GreenhouseID  uint64     `gorm:"column:greenhouse_id;not null"`
	GrowingZoneID *uint64    `gorm:"column:growing_zone_id"`
	Priority      int        `gorm:"default:100"`
	RetryLimit    uint8      `gorm:"column:retry_limit;default:3"`
	TimeoutSec    uint       `gorm:"column:timeout_sec;default:30"`
	Enabled       bool       `gorm:"default:true"`
	Version       string     `gorm:"size:32;default:v1"`
	EffectiveFrom *time.Time `gorm:"column:effective_from"`
	EffectiveTo   *time.Time `gorm:"column:effective_to"`
	CreatedBy     *uint64    `gorm:"column:created_by"`
	PublishedBy   *uint64    `gorm:"column:published_by"`
	PublishedAt   *time.Time `gorm:"column:published_at"`
	CreatedAt     time.Time  `gorm:"autoCreateTime:milli"`
	UpdatedAt     time.Time  `gorm:"autoUpdateTime:milli"`
	// Associations
	Conditions []PolicyCondition `gorm:"foreignKey:PolicyID"`
	Targets    []PolicyTarget    `gorm:"foreignKey:PolicyID"`
}

func (ControlPolicy) TableName() string { return "control_policies" }

// PolicyCondition defines a single evaluation condition for a policy.
type PolicyCondition struct {
	ID                  uint64    `gorm:"primaryKey;autoIncrement"`
	PolicyID            uint64    `gorm:"column:policy_id;not null"`
	MetricCode          string    `gorm:"column:metric_code;size:32;not null"`
	Operator            string    `gorm:"size:8;not null"` // >/>=/</<=/==/!=
	ThresholdValue      float64   `gorm:"column:threshold_value;type:decimal(12,4);not null"`
	Hysteresis          *float64  `gorm:"type:decimal(12,4)"`
	WindowSec           *uint     `gorm:"column:window_sec"`
	RequiredDurationSec *uint     `gorm:"column:required_duration_sec"`
	Aggregation         string    `gorm:"size:16"` // avg/max/min/last
	Enabled             bool      `gorm:"default:true"`
	CreatedAt           time.Time `gorm:"autoCreateTime:milli"`
	UpdatedAt           time.Time `gorm:"autoUpdateTime:milli"`
}

func (PolicyCondition) TableName() string { return "policy_conditions" }

// PolicyTarget defines a target actuator command that the policy executes.
type PolicyTarget struct {
	ID                uint64    `gorm:"primaryKey;autoIncrement"`
	PolicyID          uint64    `gorm:"column:policy_id;not null"`
	ActuatorChannelID uint64    `gorm:"column:actuator_channel_id;not null"`
	CommandType       string    `gorm:"column:command_type;size:32;not null"`
	CommandPayload    string    `gorm:"column:command_payload;type:json;not null"`
	ExecutionOrder    uint16    `gorm:"column:execution_order;default:1"`
	Enabled           bool      `gorm:"default:true"`
	CreatedAt         time.Time `gorm:"autoCreateTime:milli"`
	UpdatedAt         time.Time `gorm:"autoUpdateTime:milli"`
}

func (PolicyTarget) TableName() string { return "policy_targets" }

// PolicyExecution records each evaluation and execution result of a policy.
type PolicyExecution struct {
	ID                uint64     `gorm:"primaryKey;autoIncrement"`
	PolicyID          uint64     `gorm:"column:policy_id;not null"`
	TriggerSource     string     `gorm:"column:trigger_source;size:16;not null"` // TELEMETRY/SCHEDULE/MANUAL
	TriggerMetricCode string     `gorm:"column:trigger_metric_code;size:32"`
	TriggerValue      *float64   `gorm:"column:trigger_value;type:decimal(12,4)"`
	Decision          string     `gorm:"size:16;not null"` // EXECUTED/SKIPPED/FAILED/CONFLICT
	DecisionReason    string     `gorm:"column:decision_reason;size:255"`
	CommandID         *uint64    `gorm:"column:command_id"`
	BatchID           *uint64    `gorm:"column:batch_id"`
	ExecutedAt        *time.Time `gorm:"column:executed_at"`
	CreatedAt         time.Time  `gorm:"autoCreateTime:milli"`
}

func (PolicyExecution) TableName() string { return "policy_executions" }
