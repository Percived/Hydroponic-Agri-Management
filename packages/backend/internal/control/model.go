package control

import (
	"encoding/json"
	"time"
)

const (
	CommandStatusPending  = "PENDING"
	CommandStatusSent     = "SENT"
	CommandStatusExecuted = "EXECUTED"
	CommandStatusFailed   = "FAILED"
)

type ControlCommand struct {
	ID          uint64          `gorm:"primaryKey;autoIncrement"`
	DeviceID    uint64          `gorm:"not null;index"`
	CommandType string          `gorm:"size:32;not null"`
	Payload     json.RawMessage `gorm:"type:json;not null"`
	Status      string          `gorm:"size:16;not null;default:PENDING"`
	SentAt      *time.Time      `gorm:""`
	ExecutedAt  *time.Time      `gorm:""`
	CreatedBy   uint64          `gorm:"not null"`
	CreatedAt   time.Time       `gorm:"autoCreateTime:milli"`
}

func (ControlCommand) TableName() string { return "control_commands" }

type ControlRule struct {
	ID             uint64          `gorm:"primaryKey;autoIncrement"`
	Name           string          `gorm:"size:64;not null"`
	MetricID       uint64          `gorm:"not null;index"`
	Operator       string          `gorm:"size:4;not null"`
	Threshold      float64         `gorm:"type:decimal(12,4);not null"`
	Action         json.RawMessage `gorm:"type:json;not null"`
	TargetDeviceID uint64          `gorm:"not null"`
	Enabled        bool            `gorm:"not null;default:true"`
	CreatedBy      uint64          `gorm:"not null"`
	CreatedAt      time.Time       `gorm:"autoCreateTime:milli"`
	UpdatedAt      time.Time       `gorm:"autoUpdateTime:milli"`
}

func (ControlRule) TableName() string { return "control_rules" }

type ControlTemplate struct {
	ID          uint64          `gorm:"primaryKey;autoIncrement"`
	Name        string          `gorm:"size:64;not null"`
	Description string          `gorm:"size:255"`
	Content     json.RawMessage `gorm:"type:json;not null"`
	CreatedBy   uint64          `gorm:"not null"`
	CreatedAt   time.Time       `gorm:"autoCreateTime:milli"`
	UpdatedAt   time.Time       `gorm:"autoUpdateTime:milli"`
}

func (ControlTemplate) TableName() string { return "control_templates" }

type metricRef struct {
	ID   uint64
	Code string
}
