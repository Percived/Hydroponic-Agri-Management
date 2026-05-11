package event

type CommandAckData struct {
	SchemaVersion int                    `json:"schema_version"`
	CommandID     uint64                 `json:"command_id"`
	DeviceCode    string                 `json:"device_code"`
	AckCode       string                 `json:"ack_code"`
	AckMessage    string                 `json:"ack_message"`
	AckPayload    map[string]interface{} `json:"ack_payload"`
	AckedAt       string                 `json:"acked_at"`
}
