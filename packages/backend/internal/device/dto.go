package device

type CreateDeviceRequest struct {
	DeviceCode          string  `json:"device_code" binding:"required,min=1,max=64"`
	Name                string  `json:"name" binding:"required,min=1,max=64"`
	Type                string  `json:"type" binding:"required,oneof=SENSOR ACTUATOR"`
	Category            string  `json:"category" binding:"required,min=1,max=32"`
	Protocol            string  `json:"protocol" binding:"required,oneof=MQTT HTTP"`
	GreenhouseID        *uint64 `json:"greenhouse_id"`
	GroupID             *uint64 `json:"group_id"`
	SamplingIntervalSec *uint   `json:"sampling_interval_sec" binding:"omitempty,min=5,max=3600"`
}

type UpdateDeviceRequest struct {
	Name                *string `json:"name" binding:"omitempty,min=1,max=64"`
	Category            *string `json:"category" binding:"omitempty,min=1,max=32"`
	GreenhouseID        *uint64 `json:"greenhouse_id"`
	GroupID             *uint64 `json:"group_id"`
	SamplingIntervalSec *uint   `json:"sampling_interval_sec" binding:"omitempty,min=5,max=3600"`
}

type UpdateDeviceStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=ENABLED DISABLED"`
}

type CreateDeviceGroupRequest struct {
	GreenhouseID uint64 `json:"greenhouse_id" binding:"required"`
	Name         string `json:"name" binding:"required,min=1,max=64"`
	Description  string `json:"description" binding:"max=255"`
}

type UpdateDeviceGroupRequest struct {
	Name        *string `json:"name" binding:"omitempty,min=1,max=64"`
	Description *string `json:"description" binding:"omitempty,max=255"`
}

type CreateGreenhouseRequest struct {
	Name        string  `json:"name" binding:"required,min=1,max=64"`
	Location    *string `json:"location" binding:"omitempty,max=128"`
	Description *string `json:"description" binding:"omitempty,max=255"`
}

type UpdateGreenhouseRequest struct {
	Name        *string `json:"name" binding:"omitempty,min=1,max=64"`
	Location    *string `json:"location" binding:"omitempty,max=128"`
	Description *string `json:"description" binding:"omitempty,max=255"`
}
