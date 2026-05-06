package metric

import "time"

type MetricDefinition struct {
	ID              uint64    `gorm:"primaryKey;autoIncrement"`
	Code            string    `gorm:"size:32;uniqueIndex;not null"`
	Name            string    `gorm:"size:64;not null"`
	Unit            string    `gorm:"size:16;not null"`
	PrecisionDigits uint8     `gorm:"column:precision_digits;default:2"`
	NormalRangeMin  *float64  `gorm:"column:normal_range_min;type:decimal(12,4)"`
	NormalRangeMax  *float64  `gorm:"column:normal_range_max;type:decimal(12,4)"`
	IsCore          uint8     `gorm:"column:is_core;default:0"`
	Status          string    `gorm:"size:16;default:ENABLED"`
	CreatedAt       time.Time `gorm:"autoCreateTime:milli"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime:milli"`
}

func (MetricDefinition) TableName() string { return "metric_definitions" }
