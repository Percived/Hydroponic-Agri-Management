package device

// ---- SensorDevice ----

type CreateSensorDeviceRequest struct {
	DeviceCode      string  `json:"device_code" binding:"required,max=64"`
	Name            string  `json:"name" binding:"required,max=64"`
	Model           string  `json:"model" binding:"max=64"`
	FirmwareVersion string  `json:"firmware_version" binding:"max=64"`
	GreenhouseID    uint64  `json:"greenhouse_id" binding:"required"`
	GrowingZoneID   *uint64 `json:"growing_zone_id"`
	Protocol        string  `json:"protocol" binding:"max=16"`
	Metadata        string  `json:"metadata"`
}

type UpdateSensorDeviceRequest struct {
	Name            *string `json:"name" binding:"omitempty,max=64"`
	Model           *string `json:"model" binding:"omitempty,max=64"`
	FirmwareVersion *string `json:"firmware_version" binding:"omitempty,max=64"`
	GrowingZoneID   *uint64 `json:"growing_zone_id"`
	Status          *string `json:"status"`
	Metadata        *string `json:"metadata"`
}

type SensorDeviceResponse struct {
	ID              uint64  `json:"id"`
	GreenhouseID    uint64  `json:"greenhouse_id"`
	GrowingZoneID   *uint64 `json:"growing_zone_id"`
	DeviceCode      string  `json:"device_code"`
	Name            string  `json:"name"`
	Model           string  `json:"model"`
	FirmwareVersion string  `json:"firmware_version"`
	Status          string  `json:"status"`
	LastSeenAt      *string `json:"last_seen_at"`
	Protocol        string  `json:"protocol"`
	Metadata        string  `json:"metadata"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
}

// ---- SensorChannel ----

type CreateSensorChannelRequest struct {
	SensorDeviceID      uint64   `json:"sensor_device_id" binding:"required"`
	ChannelCode         string   `json:"channel_code" binding:"required,max=64"`
	MetricCode          string   `json:"metric_code" binding:"required,max=32"`
	Unit                string   `json:"unit" binding:"required,max=16"`
	PrecisionDigits     uint8    `json:"precision_digits"`
	RangeMin            *float64 `json:"range_min"`
	RangeMax            *float64 `json:"range_max"`
	SamplingIntervalSec uint     `json:"sampling_interval_sec"`
	Metadata            string   `json:"metadata"`
}

type UpdateSensorChannelRequest struct {
	ChannelCode         *string  `json:"channel_code" binding:"omitempty,max=64"`
	MetricCode          *string  `json:"metric_code" binding:"omitempty,max=32"`
	Unit                *string  `json:"unit" binding:"omitempty,max=16"`
	PrecisionDigits     *uint8   `json:"precision_digits"`
	RangeMin            *float64 `json:"range_min"`
	RangeMax            *float64 `json:"range_max"`
	SamplingIntervalSec *uint    `json:"sampling_interval_sec"`
	Enabled             *bool    `json:"enabled"`
	Metadata            *string  `json:"metadata"`
}

type SensorChannelResponse struct {
	ID                  uint64   `json:"id"`
	SensorDeviceID      uint64   `json:"sensor_device_id"`
	ChannelCode         string   `json:"channel_code"`
	MetricCode          string   `json:"metric_code"`
	Unit                string   `json:"unit"`
	PrecisionDigits     uint8    `json:"precision_digits"`
	RangeMin            *float64 `json:"range_min"`
	RangeMax            *float64 `json:"range_max"`
	SamplingIntervalSec uint     `json:"sampling_interval_sec"`
	Enabled             bool     `json:"enabled"`
	LastReportedAt      *string  `json:"last_reported_at"`
	Metadata            string   `json:"metadata"`
	CreatedAt           string   `json:"created_at"`
	UpdatedAt           string   `json:"updated_at"`
}

// ---- ActuatorDevice ----

type CreateActuatorDeviceRequest struct {
	DeviceCode      string  `json:"device_code" binding:"required,max=64"`
	Name            string  `json:"name" binding:"required,max=64"`
	Model           string  `json:"model" binding:"max=64"`
	FirmwareVersion string  `json:"firmware_version" binding:"max=64"`
	GreenhouseID    uint64  `json:"greenhouse_id" binding:"required"`
	GrowingZoneID   *uint64 `json:"growing_zone_id"`
	Protocol        string  `json:"protocol" binding:"max=16"`
	Metadata        string  `json:"metadata"`
}

type UpdateActuatorDeviceRequest struct {
	Name            *string `json:"name" binding:"omitempty,max=64"`
	Model           *string `json:"model" binding:"omitempty,max=64"`
	FirmwareVersion *string `json:"firmware_version" binding:"omitempty,max=64"`
	GrowingZoneID   *uint64 `json:"growing_zone_id"`
	Status          *string `json:"status"`
	Metadata        *string `json:"metadata"`
}

type ActuatorDeviceResponse struct {
	ID              uint64  `json:"id"`
	GreenhouseID    uint64  `json:"greenhouse_id"`
	GrowingZoneID   *uint64 `json:"growing_zone_id"`
	DeviceCode      string  `json:"device_code"`
	Name            string  `json:"name"`
	Model           string  `json:"model"`
	FirmwareVersion string  `json:"firmware_version"`
	Status          string  `json:"status"`
	LastSeenAt      *string `json:"last_seen_at"`
	Protocol        string  `json:"protocol"`
	Metadata        string  `json:"metadata"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
}

// ---- ActuatorChannel ----

type CreateActuatorChannelRequest struct {
	ActuatorDeviceID uint64   `json:"actuator_device_id" binding:"required"`
	ChannelCode      string   `json:"channel_code" binding:"required,max=64"`
	ActuatorType     string   `json:"actuator_type" binding:"required,max=32"`
	RatedPowerWatt   *float64 `json:"rated_power_watt"`
	Metadata         string   `json:"metadata"`
}

type UpdateActuatorChannelRequest struct {
	ChannelCode    *string  `json:"channel_code" binding:"omitempty,max=64"`
	ActuatorType   *string  `json:"actuator_type" binding:"omitempty,max=32"`
	CurrentState   *string  `json:"current_state" binding:"omitempty,max=16"`
	CurrentLevel   *float64 `json:"current_level"`
	RatedPowerWatt *float64 `json:"rated_power_watt"`
	Enabled        *bool    `json:"enabled"`
	Metadata       *string  `json:"metadata"`
}

type ActuatorChannelResponse struct {
	ID               uint64   `json:"id"`
	ActuatorDeviceID uint64   `json:"actuator_device_id"`
	ChannelCode      string   `json:"channel_code"`
	ActuatorType     string   `json:"actuator_type"`
	CurrentState     string   `json:"current_state"`
	CurrentLevel     *float64 `json:"current_level"`
	RatedPowerWatt   *float64 `json:"rated_power_watt"`
	Enabled          bool     `json:"enabled"`
	Metadata         string   `json:"metadata"`
	CreatedAt        string   `json:"created_at"`
	UpdatedAt        string   `json:"updated_at"`
}

// ---- Batch Registration ----

type RegisterDeviceChannelItem struct {
	ChannelCode         string   `json:"channel_code" binding:"required,max=64"`
	MetricCode          string   `json:"metric_code" binding:"max=32"`
	Unit                string   `json:"unit" binding:"max=16"`
	RangeMin            *float64 `json:"range_min"`
	RangeMax            *float64 `json:"range_max"`
	SamplingIntervalSec uint     `json:"sampling_interval_sec"`
	ActuatorType        string   `json:"actuator_type" binding:"max=32"`
	RatedPowerWatt      *float64 `json:"rated_power_watt"`
}

type RegisterDeviceRequest struct {
	DeviceCode      string                      `json:"device_code" binding:"required,max=64"`
	Name            string                      `json:"name" binding:"required,max=64"`
	Model           string                      `json:"model" binding:"max=64"`
	FirmwareVersion string                      `json:"firmware_version" binding:"max=64"`
	GreenhouseID    uint64                      `json:"greenhouse_id" binding:"required"`
	GrowingZoneID   *uint64                     `json:"growing_zone_id"`
	Protocol        string                      `json:"protocol" binding:"max=16"`
	DeviceType      string                      `json:"device_type" binding:"required,oneof=sensor actuator"`
	Channels        []RegisterDeviceChannelItem `json:"channels"`
}

type RegisterDeviceResponse struct {
	DeviceID   uint64   `json:"device_id"`
	ChannelIDs []uint64 `json:"channel_ids,omitempty"`
}

// ---- Device Self-Discovery ----

type DeviceSelfResponse struct {
	DeviceType string      `json:"device_type"`
	Device     interface{} `json:"device"`
	Channels   interface{} `json:"channels"`
}
