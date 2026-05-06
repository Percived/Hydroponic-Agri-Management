package telemetry

import "time"

const (
	QualityFlagNormal        = "normal"
	QualityFlagMissing       = "missing"
	QualityFlagOutOfRange    = "out_of_range"
	QualityFlagDeviceOffline = "device_offline"
)

type TelemetryRecord struct {
	ID              uint64    `gorm:"primaryKey;autoIncrement"`
	SensorChannelID uint64    `gorm:"column:sensor_channel_id;not null"`
	MetricCode      string    `gorm:"column:metric_code;size:32;not null"`
	Value           float64   `gorm:"type:decimal(12,4);not null"`
	RawValue        *float64  `gorm:"column:raw_value;type:decimal(12,4)"`
	QualityFlag     string    `gorm:"column:quality_flag;size:16;default:normal"`
	CollectedAt     time.Time `gorm:"column:collected_at;not null"`
	IngestedAt      time.Time `gorm:"column:ingested_at;autoCreateTime:milli"`
	BatchID         *uint64   `gorm:"column:batch_id"`
	CreatedAt       time.Time `gorm:"autoCreateTime:milli"`
}

func (TelemetryRecord) TableName() string { return "telemetry_records" }
