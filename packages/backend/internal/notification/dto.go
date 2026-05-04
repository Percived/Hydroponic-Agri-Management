package notification

type CreateChannelRequest struct {
	ChannelType   string                 `json:"channel_type" binding:"required,oneof=EMAIL SMS WEBHOOK"`
	Name          string                 `json:"name" binding:"required,min=1,max=64"`
	Config        map[string]interface{} `json:"config" binding:"required"`
	MinAlertLevel string                 `json:"min_alert_level" binding:"omitempty,oneof=INFO WARN CRITICAL"`
	Enabled       *bool                  `json:"enabled" binding:"required"`
}

type UpdateChannelRequest struct {
	Name          *string                `json:"name" binding:"omitempty,min=1,max=64"`
	Config        map[string]interface{} `json:"config"`
	MinAlertLevel *string                `json:"min_alert_level" binding:"omitempty,oneof=INFO WARN CRITICAL"`
	Enabled       *bool                  `json:"enabled"`
}
