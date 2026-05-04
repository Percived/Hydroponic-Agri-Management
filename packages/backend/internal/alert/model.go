package alert

import "time"

const (
	StatusOpen   = "OPEN"
	StatusAck    = "ACK"
	StatusClosed = "CLOSED"
)

type Alert struct {
	ID          uint64     `gorm:"primaryKey;autoIncrement"`
	Type        string     `gorm:"size:32;not null"`
	Level       string     `gorm:"size:16;not null"`
	MetricID    *uint64    `gorm:""`
	DeviceID    uint64     `gorm:"not null;index"`
	Value       *float64   `gorm:"type:decimal(12,4)"`
	Message     string     `gorm:"size:255;not null"`
	Status      string     `gorm:"size:16;not null;default:OPEN"`
	TriggeredAt time.Time  `gorm:"not null"`
	ResolvedAt  *time.Time `gorm:""`
}

func (Alert) TableName() string { return "alerts" }
