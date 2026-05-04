package device

import "time"

const (
	DeviceStatusEnabled  = "ENABLED"
	DeviceStatusDisabled = "DISABLED"

	DeviceTypeSensor   = "SENSOR"
	DeviceTypeActuator = "ACTUATOR"

	ProtocolMQTT = "MQTT"
	ProtocolHTTP = "HTTP"
)

type Greenhouse struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement"`
	Name        string    `gorm:"size:64;not null"`
	Location    string    `gorm:"size:128"`
	Description string    `gorm:"size:255"`
	CreatedAt   time.Time `gorm:"autoCreateTime:milli"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime:milli"`
}

func (Greenhouse) TableName() string { return "greenhouses" }

type DeviceGroup struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement"`
	GreenhouseID uint64    `gorm:"not null"`
	Name         string    `gorm:"size:64;not null"`
	Description  string    `gorm:"size:255"`
	CreatedAt    time.Time `gorm:"autoCreateTime:milli"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime:milli"`
}

func (DeviceGroup) TableName() string { return "device_groups" }

type Device struct {
	ID                  uint64     `gorm:"primaryKey;autoIncrement"`
	DeviceCode          string     `gorm:"size:64;uniqueIndex;not null"`
	Name                string     `gorm:"size:64;not null"`
	Type                string     `gorm:"size:16;not null"`
	Category            string     `gorm:"size:32;not null"`
	GreenhouseID        *uint64    `gorm:""`
	GroupID             *uint64    `gorm:""`
	Status              string     `gorm:"size:16;default:ENABLED"`
	Protocol            string     `gorm:"size:16;not null"`
	SamplingIntervalSec uint       `gorm:"not null;default:60"`
	LastSeenAt          *time.Time `gorm:""`
	CreatedAt           time.Time  `gorm:"autoCreateTime:milli"`
	UpdatedAt           time.Time  `gorm:"autoUpdateTime:milli"`
}

func (Device) TableName() string { return "devices" }
