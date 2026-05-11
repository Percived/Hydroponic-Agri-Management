package mqtt

import (
	"encoding/json"
	"fmt"
)

type LegacyCommandAck struct {
	CommandID  uint64                 `json:"command_id"`
	AckCode    string                 `json:"ack_code"`
	AckMessage string                 `json:"ack_message"`
	AckPayload map[string]interface{} `json:"ack_payload"`
}

type AckEnvelopeV1 struct {
	SchemaVersion int                    `json:"schema_version"`
	AckType       string                 `json:"ack_type"`
	MsgID         string                 `json:"msg_id"`
	TraceID       string                 `json:"trace_id"`
	Result        string                 `json:"result"`
	ErrorCode     string                 `json:"error_code"`
	ErrorMessage  string                 `json:"error_message"`
	DeviceTSms    uint64                 `json:"device_ts_ms"`
	Payload       map[string]interface{} `json:"payload"`
}

type ParsedAck struct {
	Kind   string
	V1     *AckEnvelopeV1
	Legacy *LegacyCommandAck
}

func ParseAckPayload(payload []byte) (ParsedAck, error) {
	var probe map[string]interface{}
	if err := json.Unmarshal(payload, &probe); err != nil {
		return ParsedAck{}, err
	}

	if _, ok := probe["schema_version"]; ok {
		var env AckEnvelopeV1
		if err := json.Unmarshal(payload, &env); err != nil {
			return ParsedAck{}, err
		}
		if env.SchemaVersion <= 0 {
			return ParsedAck{}, fmt.Errorf("invalid schema_version")
		}
		switch env.AckType {
		case "config":
			return ParsedAck{Kind: "v1_config", V1: &env}, nil
		case "command":
			return ParsedAck{Kind: "v1_command", V1: &env}, nil
		default:
			break
		}
	}

	var legacy LegacyCommandAck
	if err := json.Unmarshal(payload, &legacy); err != nil {
		return ParsedAck{}, err
	}
	if legacy.CommandID == 0 {
		return ParsedAck{}, fmt.Errorf("invalid legacy command_id")
	}
	return ParsedAck{Kind: "legacy_command", Legacy: &legacy}, nil
}
