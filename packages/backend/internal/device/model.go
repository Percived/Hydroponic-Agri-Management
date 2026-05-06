package device

import "time"

const (
	StatusOnline  = "ONLINE"
	StatusOffline = "OFFLINE"
	StatusFault   = "FAULT"
	ProtocolMQTT  = "MQTT"

	ActuatorTypePump    = "PUMP"
	ActuatorTypeAerator = "AERATOR"
	ActuatorTypeFan     = "FAN"
	ActuatorTypeValve   = "VALVE"
	ActuatorTypeShade   = "SHADE"
	ActuatorTypeLED     = "LED"
	ActuatorTypeHeater  = "HEATER"
	ActuatorTypeCO2Gen  = "CO2_GEN"
	ActuatorTypeFogger  = "FOGGER"
)

type SensorDevice struct {
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

func (SensorDevice) TableName() string { return "sensor_devices" }

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
	Enabled             uint8      `gorm:"default:1"`
	LastReportedAt      *time.Time `gorm:"column:last_reported_at"`
	Metadata            string     `gorm:"type:json"`
	CreatedAt           time.Time  `gorm:"autoCreateTime:milli"`
	UpdatedAt           time.Time  `gorm:"autoUpdateTime:milli"`
}

func (SensorChannel) TableName() string { return "sensor_channels" }

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

type ActuatorChannel struct {
	ID               uint64    `gorm:"primaryKey;autoIncrement"`
	ActuatorDeviceID uint64    `gorm:"column:actuator_device_id;not null"`
	ChannelCode      string    `gorm:"size:64;not null"`
	ActuatorType     string    `gorm:"column:actuator_type;size:16;not null"`
	CurrentState     string    `gorm:"column:current_state;size:16;default:OFF"`
	RatedPowerWatt   *float64  `gorm:"column:rated_power_watt;type:decimal(10,2)"`
	Enabled          uint8     `gorm:"default:1"`
	Metadata         string    `gorm:"type:json"`
	CreatedAt        time.Time `gorm:"autoCreateTime:milli"`
	UpdatedAt        time.Time `gorm:"autoUpdateTime:milli"`
}

func (ActuatorChannel) TableName() string { return "actuator_channels" }
