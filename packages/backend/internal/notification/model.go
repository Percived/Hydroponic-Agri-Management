package notification

import (
	"encoding/json"
	"time"
)

const (
	ChannelEmail   = "EMAIL"
	ChannelSMS     = "SMS"
	ChannelWebhook = "WEBHOOK"
)

type NotificationChannel struct {
	ID            uint64          `gorm:"primaryKey;autoIncrement"`
	UserID        uint64          `gorm:"not null;index"`
	ChannelType   string          `gorm:"size:16;not null"`
	Name          string          `gorm:"size:64;not null"`
	Config        json.RawMessage `gorm:"type:json;not null"`
	MinAlertLevel string          `gorm:"size:16;not null;default:WARN"`
	Enabled       bool            `gorm:"not null;default:true"`
	CreatedAt     time.Time       `gorm:"autoCreateTime:milli"`
	UpdatedAt     time.Time       `gorm:"autoUpdateTime:milli"`
}

func (NotificationChannel) TableName() string { return "notification_channels" }
