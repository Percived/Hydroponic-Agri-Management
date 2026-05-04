package telemetry

import "time"

type Metric struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement"`
	Code      string    `gorm:"size:32;uniqueIndex;not null"`
	Name      string    `gorm:"size:64;not null"`
	Unit      string    `gorm:"size:16;not null"`
	MinValue  *float64  `gorm:"type:decimal(12,4)"`
	MaxValue  *float64  `gorm:"type:decimal(12,4)"`
	CreatedAt time.Time `gorm:"autoCreateTime:milli"`
	UpdatedAt time.Time `gorm:"autoUpdateTime:milli"`
}

func (Metric) TableName() string { return "metrics" }

type TelemetryData struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement"`
	DeviceID    uint64    `gorm:"not null;index"`
	MetricID    uint64    `gorm:"not null;index"`
	Value       float64   `gorm:"type:decimal(12,4);not null"`
	RawValue    *float64  `gorm:"type:decimal(12,4)"`
	Quality     uint8     `gorm:"not null;default:0"`
	CollectedAt time.Time `gorm:"not null;index"`
	CreatedAt   time.Time `gorm:"autoCreateTime:milli"`
}

func (TelemetryData) TableName() string { return "telemetry_data" }

type SystemConfig struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement"`
	ConfigKey   string    `gorm:"size:64;uniqueIndex;not null"`
	ConfigValue string    `gorm:"size:255;not null"`
	Description string    `gorm:"size:255"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime:milli"`
}

func (SystemConfig) TableName() string { return "system_configs" }
