package command

import "encoding/json"

// DispatchTargetMeta carries device-targeting metadata that device-side
// consumers use to route a command to a specific actuator channel.
type DispatchTargetMeta struct {
	CommandID         uint64
	CommandType       string
	ActuatorChannelID uint64
	ChannelCode       string
}

// BuildDeviceCommandPayload enriches a JSON command payload with internal
// command metadata and the target actuator channel. If the payload is not valid
// JSON, it falls back to the raw payload to preserve backward compatibility.
func BuildDeviceCommandPayload(rawPayload string, meta DispatchTargetMeta) []byte {
	var payloadMap map[string]interface{}
	if err := json.Unmarshal([]byte(rawPayload), &payloadMap); err != nil {
		return []byte(rawPayload)
	}

	payloadMap["_command_id"] = meta.CommandID
	payloadMap["_command_type"] = meta.CommandType
	if meta.ActuatorChannelID != 0 {
		payloadMap["actuator_channel_id"] = meta.ActuatorChannelID
	}
	if meta.ChannelCode != "" {
		payloadMap["channel_code"] = meta.ChannelCode
	}

	wrappedPayload, err := json.Marshal(payloadMap)
	if err != nil {
		return []byte(rawPayload)
	}
	return wrappedPayload
}
