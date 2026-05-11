package mqtt

import "time"

const (
	ConfigDeliveryStatusPending  = "PENDING"
	ConfigDeliveryStatusSent     = "SENT"
	ConfigDeliveryStatusAcked    = "ACKED"
	ConfigDeliveryStatusRejected = "REJECTED"
	ConfigDeliveryStatusFailed   = "FAILED"
)

type ConfigDelivery struct {
	ID             uint64     `gorm:"primaryKey;autoIncrement"`
	MsgID          string     `gorm:"column:msg_id;size:64;not null"`
	TraceID        string     `gorm:"column:trace_id;size:64;not null;default:''"`
	DeviceCode     string     `gorm:"column:device_code;size:64;not null"`
	ConfigType     string     `gorm:"column:config_type;size:64;not null"`
	Action         string     `gorm:"column:action;size:16;not null"`
	EntityID       uint64     `gorm:"column:entity_id;not null"`
	EntityRev      uint64     `gorm:"column:entity_rev;not null"`
	SchemaVersion  int        `gorm:"column:schema_version;not null;default:1"`
	IssuedAtMS     uint64     `gorm:"column:issued_at_ms;not null"`
	TTLsec         int        `gorm:"column:ttl_sec;not null;default:600"`
	RequireAck     bool       `gorm:"column:require_ack;not null;default:true"`
	RequestPayload string     `gorm:"column:request_payload;type:json;not null"`
	Status         string     `gorm:"column:status;size:16;not null;default:PENDING"`
	RetryCount     int        `gorm:"column:retry_count;not null;default:0"`
	NextRetryAt    *time.Time `gorm:"column:next_retry_at"`
	SentAt         *time.Time `gorm:"column:sent_at"`
	AckedAt        *time.Time `gorm:"column:acked_at"`
	LastErrorCode  string     `gorm:"column:last_error_code;size:64;not null;default:''"`
	LastErrorMsg   string     `gorm:"column:last_error_message;size:255;not null;default:''"`
	AckPayload     *string    `gorm:"column:ack_payload;type:json"`
	AppliedHash    string     `gorm:"column:applied_hash;size:128;not null;default:''"`
	DeviceFWVer    string     `gorm:"column:device_fw_version;size:64;not null;default:''"`
	CreatedAt      time.Time  `gorm:"autoCreateTime:milli"`
	UpdatedAt      time.Time  `gorm:"autoUpdateTime:milli"`
}

func (ConfigDelivery) TableName() string { return "config_deliveries" }
