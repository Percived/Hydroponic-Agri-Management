package command

import "time"

// --- ControlCommand DTOs ---

// CreateCommandRequest is the request body for creating a control command.
type CreateCommandRequest struct {
	ActuatorChannelID uint64                 `json:"actuator_channel_id" binding:"required"`
	BatchID           *uint64                `json:"batch_id"`
	CommandType       string                 `json:"command_type" binding:"required,min=1,max=32"`
	Payload           map[string]interface{} `json:"payload" binding:"required"`
	RequestID         string                 `json:"request_id" binding:"omitempty,max=64"`
}

// SendCommandRequest is the request body for sending/dispatching a command.
type SendCommandRequest struct {
	RequestID string `json:"request_id" binding:"omitempty,max=64"`
}

// AckCommandRequest is the request body for acknowledging a command.
type AckCommandRequest struct {
	AckCode    string                 `json:"ack_code" binding:"required,max=32"`
	AckMessage string                 `json:"ack_message" binding:"max=255"`
	AckPayload map[string]interface{} `json:"ack_payload"`
}

// CommandResponse is the response body for a control command.
type CommandResponse struct {
	ID                uint64                   `json:"id"`
	ActuatorChannelID uint64                   `json:"actuator_channel_id"`
	BatchID           *uint64                  `json:"batch_id"`
	CommandType       string                   `json:"command_type"`
	Payload           string                   `json:"payload"`
	Status            string                   `json:"status"`
	SentAt            *time.Time               `json:"sent_at"`
	AckedAt           *time.Time               `json:"acked_at"`
	RequestID         string                   `json:"request_id"`
	CreatedBy         uint64                   `json:"created_by"`
	CreatedAt         time.Time                `json:"created_at"`
	Receipts          []CommandReceiptResponse `json:"receipts,omitempty"`
}

// CommandListResponse is the paginated list response for control commands.
type CommandListResponse struct {
	Items    []CommandResponse `json:"items"`
	Total    int64             `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
}

// --- ControlCommandReceipt DTOs ---

// CreateReceiptRequest is the request body for creating a command receipt.
type CreateReceiptRequest struct {
	ReceiptSeq    uint                   `json:"receipt_seq" binding:"required,min=1"`
	ReceiptStatus string                 `json:"receipt_status" binding:"required,oneof=ACCEPTED REJECTED PROCESSED FAILED"`
	AckCode       string                 `json:"ack_code" binding:"max=32"`
	AckMessage    string                 `json:"ack_message" binding:"max=255"`
	AckPayload    map[string]interface{} `json:"ack_payload"`
}

// CommandReceiptResponse is the response body for a command receipt.
type CommandReceiptResponse struct {
	ID            uint64     `json:"id"`
	CommandID     uint64     `json:"command_id"`
	ReceiptSeq    uint       `json:"receipt_seq"`
	ReceiptStatus string     `json:"receipt_status"`
	AckCode       string     `json:"ack_code"`
	AckMessage    string     `json:"ack_message"`
	AckPayload    string     `json:"ack_payload"`
	AckAt         *time.Time `json:"ack_at"`
	CreatedAt     time.Time  `json:"created_at"`
}

// ReceiptListResponse is the list response for receipts.
type ReceiptListResponse struct {
	Items []CommandReceiptResponse `json:"items"`
}
