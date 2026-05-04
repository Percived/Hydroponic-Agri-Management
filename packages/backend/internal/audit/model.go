package audit

import (
	"encoding/json"
	"time"
)

type AuditLog struct {
	ID         uint64          `gorm:"primaryKey;autoIncrement"`
	UserID     uint64          `gorm:"not null;index"`
	Action     string          `gorm:"size:64;not null"`
	TargetType string          `gorm:"size:64;not null"`
	TargetID   *uint64         `gorm:""`
	Detail     json.RawMessage `gorm:"type:json"`
	CreatedAt  time.Time       `gorm:"autoCreateTime:milli"`
}

func (AuditLog) TableName() string { return "audit_logs" }
