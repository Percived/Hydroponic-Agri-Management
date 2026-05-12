package policy

import "time"

// --- ControlPolicy DTOs ---

// CreatePolicyRequest is the request body for creating a control policy.
type CreatePolicyRequest struct {
	PolicyCode    string     `json:"policy_code" binding:"required,min=1,max=64"`
	Name          string     `json:"name" binding:"required,min=1,max=128"`
	PolicyType    string     `json:"policy_type" binding:"required,oneof=THRESHOLD SCHEDULE DURATION"`
	GreenhouseID  uint64     `json:"greenhouse_id" binding:"required"`
	GrowingZoneID *uint64    `json:"growing_zone_id"`
	Priority      *int       `json:"priority"`
	RetryLimit    *uint8     `json:"retry_limit" binding:"omitempty,max=10"`
	TimeoutSec    *uint      `json:"timeout_sec" binding:"omitempty,min=1"`
	Enabled       *bool      `json:"enabled"`
	Version       *string    `json:"version" binding:"omitempty,min=1,max=32"`
	EffectiveFrom *time.Time `json:"effective_from"`
	EffectiveTo   *time.Time `json:"effective_to"`
	ScheduleMode  *string    `json:"schedule_mode" binding:"omitempty,oneof=ONCE DAILY WEEKLY"`
	RunOnceAt     *time.Time `json:"run_once_at"`
	TimeOfDay     *string    `json:"time_of_day" binding:"omitempty,len=8"`
	WeekdaysMask  *uint8     `json:"weekdays_mask"`
	Timezone      *string    `json:"timezone" binding:"omitempty,min=1,max=64"`
}

// UpdatePolicyRequest is the request body for updating a control policy.
type UpdatePolicyRequest struct {
	Name          *string    `json:"name" binding:"omitempty,min=1,max=128"`
	PolicyType    *string    `json:"policy_type" binding:"omitempty,oneof=THRESHOLD SCHEDULE DURATION"`
	GrowingZoneID *uint64    `json:"growing_zone_id"`
	Priority      *int       `json:"priority"`
	RetryLimit    *uint8     `json:"retry_limit" binding:"omitempty,max=10"`
	TimeoutSec    *uint      `json:"timeout_sec" binding:"omitempty,min=1"`
	Enabled       *bool      `json:"enabled"`
	Version       *string    `json:"version" binding:"omitempty,min=1,max=32"`
	EffectiveFrom *time.Time `json:"effective_from"`
	EffectiveTo   *time.Time `json:"effective_to"`
	ScheduleMode  *string    `json:"schedule_mode" binding:"omitempty,oneof=ONCE DAILY WEEKLY"`
	RunOnceAt     *time.Time `json:"run_once_at"`
	TimeOfDay     *string    `json:"time_of_day" binding:"omitempty,len=8"`
	WeekdaysMask  *uint8     `json:"weekdays_mask"`
	Timezone      *string    `json:"timezone" binding:"omitempty,min=1,max=64"`
}

// ControlPolicyResponse is the response body for a control policy.
type ControlPolicyResponse struct {
	ID               uint64                    `json:"id"`
	PolicyCode       string                    `json:"policy_code"`
	Name             string                    `json:"name"`
	PolicyType       string                    `json:"policy_type"`
	GreenhouseID     uint64                    `json:"greenhouse_id"`
	GrowingZoneID    *uint64                   `json:"growing_zone_id"`
	Priority         int                       `json:"priority"`
	RetryLimit       uint8                     `json:"retry_limit"`
	TimeoutSec       uint                      `json:"timeout_sec"`
	Enabled          bool                      `json:"enabled"`
	Version          string                    `json:"version"`
	EffectiveFrom    *time.Time                `json:"effective_from"`
	EffectiveTo      *time.Time                `json:"effective_to"`
	ScheduleMode     *string                   `json:"schedule_mode"`
	RunOnceAt        *time.Time                `json:"run_once_at"`
	TimeOfDay        *string                   `json:"time_of_day"`
	WeekdaysMask     *uint8                    `json:"weekdays_mask"`
	Timezone         string                    `json:"timezone"`
	LastScheduledFor *time.Time                `json:"last_scheduled_for"`
	CreatedBy        *uint64                   `json:"created_by"`
	PublishedBy      *uint64                   `json:"published_by"`
	PublishedAt      *time.Time                `json:"published_at"`
	Conditions       []PolicyConditionResponse `json:"conditions,omitempty"`
	Targets          []PolicyTargetResponse    `json:"targets,omitempty"`
	CreatedAt        time.Time                 `json:"created_at"`
	UpdatedAt        time.Time                 `json:"updated_at"`
}

// ControlPolicyListResponse is the paginated list response for control policies.
type ControlPolicyListResponse struct {
	Items    []ControlPolicyResponse `json:"items"`
	Total    int64                   `json:"total"`
	Page     int                     `json:"page"`
	PageSize int                     `json:"page_size"`
}

// --- PolicyCondition DTOs ---

// CreatePolicyConditionRequest is the request body for creating a policy condition.
type CreatePolicyConditionRequest struct {
	MetricCode          string   `json:"metric_code" binding:"required,min=1,max=32"`
	Operator            string   `json:"operator" binding:"required,oneof=> >= < <= == !="`
	ThresholdValue      float64  `json:"threshold_value" binding:"required"`
	Hysteresis          *float64 `json:"hysteresis"`
	WindowSec           *uint    `json:"window_sec"`
	RequiredDurationSec *uint    `json:"required_duration_sec"`
	Aggregation         *string  `json:"aggregation" binding:"omitempty,oneof=avg max min last"`
	Enabled             *bool    `json:"enabled"`
}

// UpdatePolicyConditionRequest is the request body for updating a policy condition.
type UpdatePolicyConditionRequest struct {
	Operator            *string  `json:"operator" binding:"omitempty,oneof=> >= < <= == !="`
	ThresholdValue      *float64 `json:"threshold_value"`
	Hysteresis          *float64 `json:"hysteresis"`
	WindowSec           *uint    `json:"window_sec"`
	RequiredDurationSec *uint    `json:"required_duration_sec"`
	Aggregation         *string  `json:"aggregation" binding:"omitempty,oneof=avg max min last"`
	Enabled             *bool    `json:"enabled"`
}

// PolicyConditionResponse is the response body for a policy condition.
type PolicyConditionResponse struct {
	ID                  uint64    `json:"id"`
	PolicyID            uint64    `json:"policy_id"`
	MetricCode          string    `json:"metric_code"`
	Operator            string    `json:"operator"`
	ThresholdValue      float64   `json:"threshold_value"`
	Hysteresis          *float64  `json:"hysteresis"`
	WindowSec           *uint     `json:"window_sec"`
	RequiredDurationSec *uint     `json:"required_duration_sec"`
	Aggregation         string    `json:"aggregation"`
	Enabled             bool      `json:"enabled"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// --- PolicyTarget DTOs ---

// CreatePolicyTargetRequest is the request body for creating a policy target.
type CreatePolicyTargetRequest struct {
	ActuatorChannelID uint64                 `json:"actuator_channel_id" binding:"required"`
	CommandType       string                 `json:"command_type" binding:"required,min=1,max=32"`
	CommandPayload    map[string]interface{} `json:"command_payload" binding:"required"`
	ExecutionOrder    *uint16                `json:"execution_order"`
	Enabled           *bool                  `json:"enabled"`
}

// UpdatePolicyTargetRequest is the request body for updating a policy target.
type UpdatePolicyTargetRequest struct {
	ActuatorChannelID *uint64                `json:"actuator_channel_id"`
	CommandType       *string                `json:"command_type" binding:"omitempty,min=1,max=32"`
	CommandPayload    map[string]interface{} `json:"command_payload"`
	ExecutionOrder    *uint16                `json:"execution_order"`
	Enabled           *bool                  `json:"enabled"`
}

// PolicyTargetResponse is the response body for a policy target.
type PolicyTargetResponse struct {
	ID                uint64    `json:"id"`
	PolicyID          uint64    `json:"policy_id"`
	ActuatorChannelID uint64    `json:"actuator_channel_id"`
	CommandType       string    `json:"command_type"`
	CommandPayload    string    `json:"command_payload"`
	ExecutionOrder    uint16    `json:"execution_order"`
	Enabled           bool      `json:"enabled"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// --- PolicyExecution DTOs ---

// PolicyExecutionResponse is the response body for a policy execution.
type PolicyExecutionResponse struct {
	ID                uint64     `json:"id"`
	PolicyID          uint64     `json:"policy_id"`
	PolicyName        string     `json:"policy_name,omitempty"`
	TriggerSource     string     `json:"trigger_source"`
	TriggerMetricCode string     `json:"trigger_metric_code"`
	TriggerValue      *float64   `json:"trigger_value"`
	Decision          string     `json:"decision"`
	DecisionReason    string     `json:"decision_reason"`
	CommandID         *uint64    `json:"command_id"`
	BatchID           *uint64    `json:"batch_id"`
	ExecutedAt        *time.Time `json:"executed_at"`
	CreatedAt         time.Time  `json:"created_at"`
}

// PolicyExecutionListResponse is the paginated list response for policy executions.
type PolicyExecutionListResponse struct {
	Items    []PolicyExecutionResponse `json:"items"`
	Total    int64                     `json:"total"`
	Page     int                       `json:"page"`
	PageSize int                       `json:"page_size"`
}

// --- Nested creation: full policy with conditions and targets ---

// ConditionWithTargets represents a nested condition + targets structure.
type ConditionWithTargets struct {
	MetricCode          string                      `json:"metric_code" binding:"required,min=1,max=32"`
	Operator            string                      `json:"operator" binding:"required,oneof=> >= < <= == !="`
	ThresholdValue      float64                     `json:"threshold_value" binding:"required"`
	Hysteresis          *float64                    `json:"hysteresis"`
	WindowSec           *uint                       `json:"window_sec"`
	RequiredDurationSec *uint                       `json:"required_duration_sec"`
	Aggregation         *string                     `json:"aggregation" binding:"omitempty,oneof=avg max min last"`
	Enabled             *bool                       `json:"enabled"`
	Targets             []CreatePolicyTargetRequest `json:"targets"`
}

// CreatePolicyWithNestedRequest creates a full policy with nested conditions and targets.
type CreatePolicyWithNestedRequest struct {
	PolicyCode    string                 `json:"policy_code" binding:"required,min=1,max=64"`
	Name          string                 `json:"name" binding:"required,min=1,max=128"`
	PolicyType    string                 `json:"policy_type" binding:"required,oneof=THRESHOLD SCHEDULE DURATION"`
	GreenhouseID  uint64                 `json:"greenhouse_id" binding:"required"`
	GrowingZoneID *uint64                `json:"growing_zone_id"`
	Priority      *int                   `json:"priority"`
	RetryLimit    *uint8                 `json:"retry_limit"`
	TimeoutSec    *uint                  `json:"timeout_sec"`
	Enabled       *bool                  `json:"enabled"`
	Version       *string                `json:"version"`
	EffectiveFrom *time.Time             `json:"effective_from"`
	EffectiveTo   *time.Time             `json:"effective_to"`
	ScheduleMode  *string                `json:"schedule_mode" binding:"omitempty,oneof=ONCE DAILY WEEKLY"`
	RunOnceAt     *time.Time             `json:"run_once_at"`
	TimeOfDay     *string                `json:"time_of_day" binding:"omitempty,len=8"`
	WeekdaysMask  *uint8                 `json:"weekdays_mask"`
	Timezone      *string                `json:"timezone" binding:"omitempty,min=1,max=64"`
	Conditions    []ConditionWithTargets `json:"conditions"`
}

// PublishPolicyRequest is the request body for publishing a policy.
type PublishPolicyRequest struct {
	Version *string `json:"version" binding:"omitempty,min=1,max=32"`
}

// ExecutePolicyRequest is the request body for manually executing a policy.
type ExecutePolicyRequest struct {
	TriggerSource     string   `json:"trigger_source" binding:"required,oneof=MANUAL TELEMETRY SCHEDULE"`
	TriggerMetricCode string   `json:"trigger_metric_code" binding:"omitempty,min=1,max=32"`
	TriggerValue      *float64 `json:"trigger_value"`
}
