package climate

import "time"

// ClimateProfile represents a climate control profile for a greenhouse.
type ClimateProfile struct {
	ID                     uint64    `gorm:"primaryKey;autoIncrement"`
	GreenhouseID           uint64    `gorm:"column:greenhouse_id;not null"`
	Code                   string    `gorm:"size:64;not null"`
	Name                   string    `gorm:"size:128;not null"`
	Description            string    `gorm:"size:255"`
	TriggerMetricCode      string    `gorm:"column:trigger_metric_code;size:32;not null"`
	TriggerSensorChannelID *uint64   `gorm:"column:trigger_sensor_channel_id"`
	Enabled                bool      `gorm:"default:true"`
	CreatedAt              time.Time `gorm:"autoCreateTime:milli"`
	UpdatedAt              time.Time `gorm:"autoUpdateTime:milli"`
	// Associations
	Stages []ClimateStage `gorm:"foreignKey:ProfileID"`
}

func (ClimateProfile) TableName() string { return "climate_profiles" }

// ClimateStage represents a stage within a climate profile, triggered by threshold conditions.
type ClimateStage struct {
	ID               uint64    `gorm:"primaryKey;autoIncrement"`
	ProfileID        uint64    `gorm:"column:profile_id;not null"`
	StageLevel       uint8     `gorm:"column:stage_level;not null"`
	Name             string    `gorm:"size:64;not null"`
	TriggerOperator  string    `gorm:"column:trigger_operator;size:4;not null"` // >/>=/</<=
	TriggerThreshold float64   `gorm:"column:trigger_threshold;type:decimal(12,4);not null"`
	Hysteresis       float64   `gorm:"type:decimal(12,4);default:1.0"`
	CreatedAt        time.Time `gorm:"autoCreateTime:milli"`
	UpdatedAt        time.Time `gorm:"autoUpdateTime:milli"`
	// Associations
	Actions []ClimateStageAction `gorm:"foreignKey:StageID"`
}

func (ClimateStage) TableName() string { return "climate_stages" }

// ClimateStageAction represents an action triggered when a stage becomes active.
type ClimateStageAction struct {
	ID                uint64    `gorm:"primaryKey;autoIncrement"`
	StageID           uint64    `gorm:"column:stage_id;not null"`
	ActuatorChannelID uint64    `gorm:"column:actuator_channel_id;not null"`
	CommandType       string    `gorm:"column:command_type;size:32;not null"` // SWITCH/SET_SPEED/SET_ANGLE
	CommandPayload    string    `gorm:"column:command_payload;type:json;not null"`
	ExecutionOrder    uint16    `gorm:"column:execution_order;default:1"`
	Enabled           bool      `gorm:"default:true"`
	CreatedAt         time.Time `gorm:"autoCreateTime:milli"`
	UpdatedAt         time.Time `gorm:"autoUpdateTime:milli"`
}

func (ClimateStageAction) TableName() string { return "climate_stage_actions" }

// ClimateExecutionLog records each execution of a climate profile stage transition.
type ClimateExecutionLog struct {
	ID                     uint64     `gorm:"primaryKey;autoIncrement"`
	ProfileID              uint64     `gorm:"column:profile_id;not null"`
	FromStageLevel         *uint8     `gorm:"column:from_stage_level"`
	ToStageLevel           uint8      `gorm:"column:to_stage_level;not null"`
	TriggerValue           float64    `gorm:"column:trigger_value;type:decimal(12,4);not null"`
	TriggerSensorChannelID *uint64    `gorm:"column:trigger_sensor_channel_id"`
	TriggerMetricCode      *string    `gorm:"column:trigger_metric_code;size:32"`
	CollectedAt            *time.Time `gorm:"column:collected_at"`
	ExecutedActionsCount   uint       `gorm:"column:executed_actions_count;default:0"`
	ExecutedAt             time.Time  `gorm:"column:executed_at;not null"`
	CreatedAt              time.Time  `gorm:"autoCreateTime:milli"`
}

func (ClimateExecutionLog) TableName() string { return "climate_execution_logs" }
