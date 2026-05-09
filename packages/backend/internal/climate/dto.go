package climate

import "time"

// --- ClimateProfile DTOs ---

// CreateClimateProfileRequest is the request body for creating a climate profile.
type CreateClimateProfileRequest struct {
	GreenhouseID           uint64 `json:"greenhouse_id" binding:"required"`
	Code                   string `json:"code" binding:"required,min=1,max=64"`
	Name                   string `json:"name" binding:"required,min=1,max=128"`
	Description            string `json:"description" binding:"max=255"`
	TriggerMetricCode      string `json:"trigger_metric_code" binding:"required,min=1,max=32"`
	TriggerSensorChannelID uint64 `json:"trigger_sensor_channel_id" binding:"required"`
	Enabled                *bool  `json:"enabled"`
}

// UpdateClimateProfileRequest is the request body for updating a climate profile.
type UpdateClimateProfileRequest struct {
	Name                   *string `json:"name" binding:"omitempty,min=1,max=128"`
	Description            *string `json:"description" binding:"omitempty,max=255"`
	TriggerMetricCode      *string `json:"trigger_metric_code" binding:"omitempty,min=1,max=32"`
	TriggerSensorChannelID *uint64 `json:"trigger_sensor_channel_id"`
	Enabled                *bool   `json:"enabled"`
}

// ClimateProfileResponse is the response body for a climate profile.
type ClimateProfileResponse struct {
	ID                     uint64                 `json:"id"`
	GreenhouseID           uint64                 `json:"greenhouse_id"`
	Code                   string                 `json:"code"`
	Name                   string                 `json:"name"`
	Description            string                 `json:"description"`
	TriggerMetricCode      string                 `json:"trigger_metric_code"`
	TriggerSensorChannelID *uint64                `json:"trigger_sensor_channel_id"`
	Enabled                bool                   `json:"enabled"`
	StagesCount            int                    `json:"stages_count"`
	CreatedAt              time.Time              `json:"created_at"`
	UpdatedAt              time.Time              `json:"updated_at"`
	Stages                 []ClimateStageResponse `json:"stages,omitempty"`
}

// ClimateProfileListResponse is the paginated list response for climate profiles.
type ClimateProfileListResponse struct {
	Items    []ClimateProfileResponse `json:"items"`
	Total    int64                    `json:"total"`
	Page     int                      `json:"page"`
	PageSize int                      `json:"page_size"`
}

// --- ClimateStage DTOs ---

// CreateClimateStageRequest is the request body for creating a climate stage.
type CreateClimateStageRequest struct {
	StageLevel       uint8    `json:"stage_level" binding:"required,min=1"`
	Name             string   `json:"name" binding:"required,min=1,max=64"`
	TriggerOperator  string   `json:"trigger_operator" binding:"required,oneof=> >= < <=`
	TriggerThreshold float64  `json:"trigger_threshold" binding:"required"`
	Hysteresis       *float64 `json:"hysteresis"`
}

// UpdateClimateStageRequest is the request body for updating a climate stage.
type UpdateClimateStageRequest struct {
	Name             *string  `json:"name" binding:"omitempty,min=1,max=64"`
	TriggerOperator  *string  `json:"trigger_operator" binding:"omitempty,oneof=> >= < <=`
	TriggerThreshold *float64 `json:"trigger_threshold"`
	Hysteresis       *float64 `json:"hysteresis"`
}

// ClimateStageResponse is the response body for a climate stage.
type ClimateStageResponse struct {
	ID               uint64                       `json:"id"`
	ProfileID        uint64                       `json:"profile_id"`
	StageLevel       uint8                        `json:"stage_level"`
	Name             string                       `json:"name"`
	TriggerOperator  string                       `json:"trigger_operator"`
	TriggerThreshold float64                      `json:"trigger_threshold"`
	Hysteresis       float64                      `json:"hysteresis"`
	ActionCount      int                          `json:"action_count"`
	CreatedAt        time.Time                    `json:"created_at"`
	UpdatedAt        time.Time                    `json:"updated_at"`
	Actions          []ClimateStageActionResponse `json:"actions,omitempty"`
}

// --- ClimateStageAction DTOs ---

// CreateClimateStageActionRequest is the request body for creating a stage action.
type CreateClimateStageActionRequest struct {
	ActuatorChannelID uint64                 `json:"actuator_channel_id" binding:"required"`
	CommandType       string                 `json:"command_type" binding:"required,min=1,max=32"`
	CommandPayload    map[string]interface{} `json:"command_payload" binding:"required"`
	ExecutionOrder    *uint16                `json:"execution_order" binding:"omitempty,min=1"`
	Enabled           *bool                  `json:"enabled"`
}

// UpdateClimateStageActionRequest is the request body for updating a stage action.
type UpdateClimateStageActionRequest struct {
	ActuatorChannelID *uint64                `json:"actuator_channel_id"`
	CommandType       *string                `json:"command_type" binding:"omitempty,min=1,max=32"`
	CommandPayload    map[string]interface{} `json:"command_payload"`
	ExecutionOrder    *uint16                `json:"execution_order" binding:"omitempty,min=1"`
	Enabled           *bool                  `json:"enabled"`
}

// ClimateStageActionResponse is the response body for a stage action.
type ClimateStageActionResponse struct {
	ID                uint64    `json:"id"`
	StageID           uint64    `json:"stage_id"`
	ActuatorChannelID uint64    `json:"actuator_channel_id"`
	CommandType       string    `json:"command_type"`
	CommandPayload    string    `json:"command_payload"`
	ExecutionOrder    uint16    `json:"execution_order"`
	Enabled           bool      `json:"enabled"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// --- ClimateExecutionLog DTOs ---

// ClimateExecutionLogResponse is the response body for an execution log.
type ClimateExecutionLogResponse struct {
	ID                     uint64     `json:"id"`
	ProfileID              uint64     `json:"profile_id"`
	ProfileName            string     `json:"profile_name,omitempty"`
	FromStageLevel         *uint8     `json:"from_stage_level"`
	ToStageLevel           uint8      `json:"to_stage_level"`
	TriggerValue           float64    `json:"trigger_value"`
	TriggerSensorChannelID *uint64    `json:"trigger_sensor_channel_id,omitempty"`
	TriggerMetricCode      *string    `json:"trigger_metric_code,omitempty"`
	CollectedAt            *time.Time `json:"collected_at,omitempty"`
	ExecutedActionsCount   uint       `json:"executed_actions_count"`
	ExecutedAt             time.Time  `json:"executed_at"`
	CreatedAt              time.Time  `json:"created_at"`
}

// ClimateExecutionLogListResponse is the paginated list response for execution logs.
type ClimateExecutionLogListResponse struct {
	Items    []ClimateExecutionLogResponse `json:"items"`
	Total    int64                         `json:"total"`
	Page     int                           `json:"page"`
	PageSize int                           `json:"page_size"`
}

// --- Nested creation: full profile with stages and actions ---

// StageWithActions is a nested struct for creating a stage with its actions in one request.
type StageWithActions struct {
	StageLevel       uint8                             `json:"stage_level" binding:"required,min=1"`
	Name             string                            `json:"name" binding:"required,min=1,max=64"`
	TriggerOperator  string                            `json:"trigger_operator" binding:"required,oneof=> >= < <=`
	TriggerThreshold float64                           `json:"trigger_threshold" binding:"required"`
	Hysteresis       *float64                          `json:"hysteresis"`
	Actions          []CreateClimateStageActionRequest `json:"actions"`
}

// CreateClimateProfileWithStagesRequest creates a profile with nested stages and actions.
type CreateClimateProfileWithStagesRequest struct {
	GreenhouseID           uint64             `json:"greenhouse_id" binding:"required"`
	Code                   string             `json:"code" binding:"required,min=1,max=64"`
	Name                   string             `json:"name" binding:"required,min=1,max=128"`
	Description            string             `json:"description" binding:"max=255"`
	TriggerMetricCode      string             `json:"trigger_metric_code" binding:"required,min=1,max=32"`
	TriggerSensorChannelID uint64             `json:"trigger_sensor_channel_id" binding:"required"`
	Enabled                *bool              `json:"enabled"`
	Stages                 []StageWithActions `json:"stages"`
}

// ExecuteClimateProfileRequest is the request body for manually executing a profile.
type ExecuteClimateProfileRequest struct {
	TriggerValue   float64 `json:"trigger_value" binding:"required"`
	FromStageLevel *uint8  `json:"from_stage_level"`
	ToStageLevel   uint8   `json:"to_stage_level" binding:"required,min=1"`
}
