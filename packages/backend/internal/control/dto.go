package control

type CreateCommandRequest struct {
	DeviceID    uint64                 `json:"device_id" binding:"required"`
	CommandType string                 `json:"command_type" binding:"required,min=1,max=32"`
	Payload     map[string]interface{} `json:"payload" binding:"required"`
}

type CreateRuleRequest struct {
	Name           string                 `json:"name" binding:"required,min=1,max=64"`
	MetricCode     string                 `json:"metric_code" binding:"required,min=1,max=32"`
	Operator       string                 `json:"operator" binding:"required,oneof=> >= < <= =="`
	Threshold      float64                `json:"threshold" binding:"required"`
	Action         map[string]interface{} `json:"action" binding:"required"`
	TargetDeviceID uint64                 `json:"target_device_id" binding:"required"`
	Enabled        *bool                  `json:"enabled" binding:"required"`
}

type UpdateRuleRequest struct {
	Name           *string                `json:"name" binding:"omitempty,min=1,max=64"`
	Operator       *string                `json:"operator" binding:"omitempty,oneof=> >= < <= =="`
	Threshold      *float64               `json:"threshold"`
	Action         map[string]interface{} `json:"action"`
	TargetDeviceID *uint64                `json:"target_device_id"`
	Enabled        *bool                  `json:"enabled"`
}

type CreateTemplateRequest struct {
	Name        string                 `json:"name" binding:"required,min=1,max=64"`
	Description string                 `json:"description" binding:"max=255"`
	Content     map[string]interface{} `json:"content" binding:"required"`
}

type ApplyTemplateRequest struct {
	TargetGroupID uint64 `json:"target_group_id" binding:"required"`
}
