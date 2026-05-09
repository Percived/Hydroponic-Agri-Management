package command

import "time"

// ControlCommand represents a control command dispatched to an actuator channel.
type ControlCommand struct {
	ID                uint64     `gorm:"primaryKey;autoIncrement"`
	ActuatorChannelID uint64     `gorm:"column:actuator_channel_id;not null"`
	BatchID           *uint64    `gorm:"column:batch_id"`
	CommandType       string     `gorm:"column:command_type;size:32;not null"`
	Payload           string     `gorm:"type:json;not null"`
	Status            string     `gorm:"size:16;default:PENDING"` // PENDING/QUEUED/SENT/ACKED/TIMEOUT/FAILED
	SentAt            *time.Time `gorm:"column:sent_at"`
	AckedAt           *time.Time `gorm:"column:acked_at"`
	RequestID         string     `gorm:"column:request_id;size:64"`
	CreatedBy         uint64     `gorm:"column:created_by;not null"`
	CreatedAt         time.Time  `gorm:"autoCreateTime:milli"`
	// Associations
	Receipts []ControlCommandReceipt `gorm:"foreignKey:CommandID"`
}

func (ControlCommand) TableName() string { return "control_commands" }

// ControlCommandReceipt represents an acknowledgement receipt for a sent command.
type ControlCommandReceipt struct {
	ID            uint64     `gorm:"primaryKey;autoIncrement"`
	CommandID     uint64     `gorm:"column:command_id;not null"`
	ReceiptSeq    uint       `gorm:"column:receipt_seq;default:1"`
	ReceiptStatus string     `gorm:"column:receipt_status;size:16;not null"`
	AckCode       string     `gorm:"column:ack_code;size:32"`
	AckMessage    string     `gorm:"column:ack_message;size:255"`
	AckPayload    string     `gorm:"column:ack_payload;type:json"`
	AckAt         *time.Time `gorm:"column:ack_at"`
	CreatedAt     time.Time  `gorm:"autoCreateTime:milli"`
}

func (ControlCommandReceipt) TableName() string { return "control_command_receipts" }
