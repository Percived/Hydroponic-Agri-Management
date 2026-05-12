package device

import (
	"time"

	"gorm.io/gorm"
)

const (
	StatusOnline  = "ONLINE"
	StatusOffline = "OFFLINE"
	StatusFault   = "FAULT"
	ProtocolMQTT  = "MQTT"

	ActuatorTypePump         = "PUMP"
	ActuatorTypeAerator      = "AERATOR"
	ActuatorTypeFan          = "FAN"
	ActuatorTypeValve        = "VALVE"
	ActuatorTypeShade        = "SHADE"
	ActuatorTypeLED          = "LED"
	ActuatorTypeHeater       = "HEATER"
	ActuatorTypeCO2Gen       = "CO2_GEN"
	ActuatorTypeFogger       = "FOGGER"
	ActuatorTypeDosingPump   = "DOSING_PUMP"
	ActuatorTypeChiller      = "CHILLER"
	ActuatorTypeStirrer      = "STIRRER"
	ActuatorTypeDehumidifier = "DEHUMIDIFIER"
	ActuatorTypeDamper       = "DAMPER"
	ActuatorTypeUVSterilizer = "UV_STERILIZER"
	ActuatorTypeOzoneGen     = "OZONE_GENERATOR"
	ActuatorTypeFilter       = "FILTER"
	ActuatorTypeROSystem     = "RO_SYSTEM"
	ActuatorTypeTopUpValve   = "TOP_UP_VALVE"
	ActuatorTypeAlarm        = "ALARM"
	ActuatorTypeCalibValve   = "CALIBRATION_VALVE"
)

type SensorDevice struct {
	ID              uint64     `gorm:"primaryKey;autoIncrement"`
	GreenhouseID    uint64     `gorm:"not null"`
	GrowingZoneID   *uint64    `gorm:"column:growing_zone_id"`
	DeviceCode      string     `gorm:"size:64;uniqueIndex;not null"`
	Name            string     `gorm:"size:64;not null"`
	Model           string     `gorm:"size:64"`
	FirmwareVersion string     `gorm:"column:firmware_version;size:64"`
	Status          string     `gorm:"size:16;default:OFFLINE"`
	LastSeenAt      *time.Time `gorm:"column:last_seen_at"`
	Protocol        string     `gorm:"size:16;default:MQTT"`
	Metadata        string     `gorm:"type:json"`
	CreatedAt       time.Time  `gorm:"autoCreateTime:milli"`
	UpdatedAt       time.Time  `gorm:"autoUpdateTime:milli"`
}

func (SensorDevice) TableName() string { return "sensor_devices" }

func (s *SensorDevice) BeforeCreate(tx *gorm.DB) error {
	if s.Metadata == "" {
		s.Metadata = "{}"
	}
	return nil
}

type SensorChannel struct {
	ID                  uint64     `gorm:"primaryKey;autoIncrement"`
	SensorDeviceID      uint64     `gorm:"column:sensor_device_id;not null"`
	ChannelCode         string     `gorm:"size:64;not null"`
	MetricCode          string     `gorm:"column:metric_code;size:32;not null"`
	Unit                string     `gorm:"size:16;not null"`
	PrecisionDigits     uint8      `gorm:"column:precision_digits;default:2"`
	RangeMin            *float64   `gorm:"column:range_min;type:decimal(12,4)"`
	RangeMax            *float64   `gorm:"column:range_max;type:decimal(12,4)"`
	SamplingIntervalSec uint       `gorm:"column:sampling_interval_sec;default:60"`
	Enabled             bool       `gorm:"default:true"`
	LastReportedAt      *time.Time `gorm:"column:last_reported_at"`
	Metadata            string     `gorm:"type:json"`
	CreatedAt           time.Time  `gorm:"autoCreateTime:milli"`
	UpdatedAt           time.Time  `gorm:"autoUpdateTime:milli"`
}

func (SensorChannel) TableName() string { return "sensor_channels" }

func (s *SensorChannel) BeforeCreate(tx *gorm.DB) error {
	if s.Metadata == "" {
		s.Metadata = "{}"
	}
	return nil
}

type ActuatorDevice struct {
	ID              uint64     `gorm:"primaryKey;autoIncrement"`
	GreenhouseID    uint64     `gorm:"not null"`
	GrowingZoneID   *uint64    `gorm:"column:growing_zone_id"`
	DeviceCode      string     `gorm:"size:64;uniqueIndex;not null"`
	Name            string     `gorm:"size:64;not null"`
	Model           string     `gorm:"size:64"`
	FirmwareVersion string     `gorm:"column:firmware_version;size:64"`
	Status          string     `gorm:"size:16;default:ONLINE"`
	LastSeenAt      *time.Time `gorm:"column:last_seen_at"`
	Protocol        string     `gorm:"size:16;default:MQTT"`
	Metadata        string     `gorm:"type:json"`
	CreatedAt       time.Time  `gorm:"autoCreateTime:milli"`
	UpdatedAt       time.Time  `gorm:"autoUpdateTime:milli"`
}

func (ActuatorDevice) TableName() string { return "actuator_devices" }

func (a *ActuatorDevice) BeforeCreate(tx *gorm.DB) error {
	if a.Metadata == "" {
		a.Metadata = "{}"
	}
	return nil
}

type ActuatorChannel struct {
	ID               uint64    `gorm:"primaryKey;autoIncrement"`
	ActuatorDeviceID uint64    `gorm:"column:actuator_device_id;not null"`
	ChannelCode      string    `gorm:"size:64;not null"`
	ActuatorType     string    `gorm:"column:actuator_type;size:32;not null"`
	CurrentState     string    `gorm:"column:current_state;size:16;default:OFF"`
	CurrentLevel     *float64  `gorm:"column:current_level;type:decimal(8,2)"`
	RatedPowerWatt   *float64  `gorm:"column:rated_power_watt;type:decimal(10,2)"`
	Enabled          bool      `gorm:"default:true"`
	Metadata         string    `gorm:"type:json"`
	CreatedAt        time.Time `gorm:"autoCreateTime:milli"`
	UpdatedAt        time.Time `gorm:"autoUpdateTime:milli"`
}

func (ActuatorChannel) TableName() string { return "actuator_channels" }

func (a *ActuatorChannel) BeforeCreate(tx *gorm.DB) error {
	if a.Metadata == "" {
		a.Metadata = "{}"
	}
	return nil
}
