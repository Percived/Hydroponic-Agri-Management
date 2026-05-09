package event

type CommandAckData struct {
	CommandID  uint64                 `json:"command_id"`
	DeviceCode string                 `json:"device_code"`
	AckCode    string                 `json:"ack_code"`
	AckMessage string                 `json:"ack_message"`
	AckPayload map[string]interface{} `json:"ack_payload"`
}
