package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// ──────────────────── 执行器模拟器 ────────────────────

// actuatorSim simulates an actuator device that receives commands via MQTT
// and applies their effects to the environment model.
type actuatorSim struct {
	deviceCode string
	channels   []actuatorChannelDetail
	mqtt       *mqttManager
	env        *Environment
	hub        *SSEHub // optional, nil in CLI mode

	// Map channelID → actuatorChannelDetail for quick lookup
	chByID map[uint64]actuatorChannelDetail

	// Stats
	totalCmdACK int64
}

// newActuatorSim creates a new actuator simulator.
func newActuatorSim(deviceCode string, channels []actuatorChannelDetail, mqtt *mqttManager, env *Environment, hub *SSEHub) *actuatorSim {
	chByID := make(map[uint64]actuatorChannelDetail, len(channels))
	for _, ch := range channels {
		chByID[ch.ID] = ch
	}
	return &actuatorSim{
		deviceCode: deviceCode,
		channels:   channels,
		chByID:     chByID,
		mqtt:       mqtt,
		env:        env,
		hub:        hub,
	}
}

// channelIDs returns all actuator channel IDs.
func (a *actuatorSim) channelIDs() []uint64 {
	ids := make([]uint64, 0, len(a.channels))
	for _, ch := range a.channels {
		ids = append(ids, ch.ID)
	}
	return ids
}

// onCommand is the MQTT command callback. It parses the command payload,
// updates the environment model, and sends an ACK.
func (a *actuatorSim) onCommand(_ mqtt.Client, msg mqtt.Message) {
	log.Printf("📥 执行器收到命令: topic=%s, payload=%s", msg.Topic(), string(msg.Payload()))

	// Parse the incoming command (supports both wrapped and unwrapped payloads)
	var cmd incomingCmd
	if err := json.Unmarshal(msg.Payload(), &cmd); err != nil {
		log.Printf("   无法解析命令结构: %v", err)
		return
	}

	// Determine the effective command ID
	cmdID := cmd.CommandID
	if cmdID == 0 {
		cmdID = cmd.InternalCommandID
	}
	if cmdID == 0 {
		log.Printf("   收到无 command_id 的消息，无需 ACK")
		return
	}

	// Determine the effective command type
	cmdType := cmd.CommandType
	if cmdType == "" {
		cmdType = cmd.InternalCommandType
	}

	// Parse state/value from payload
	var statePayload cmdPayloadState
	if err := json.Unmarshal(msg.Payload(), &statePayload); err != nil {
		log.Printf("   无法解析命令 payload: %v", err)
	}

	// Emit command event to SSE
	emitIfHub(a.hub, func() { a.hub.PublishCmd(cmdID, cmdType, statePayload.State, statePayload.Value) })

	// Simulate execution delay
	time.Sleep(100 * time.Millisecond)

	// Apply to all actuator channels (the command targets a specific channel,
	// but in the simulator we apply it to all channels of this device)
	// In practice, the backend dispatches per-channel, so the topic is device-specific
	for _, ch := range a.channels {
		a.env.UpdateActuatorState(ch.ID, ch.ActuatorType, statePayload.State, statePayload.Value)
		log.Printf("   执行器通道 [%d] %s → %s (value=%.0f%%)", ch.ID, ch.ActuatorType, statePayload.State, statePayload.Value)
	}

	// Send ACK
	ack := ackPayload{
		CommandID:  cmdID,
		AckCode:    "ok",
		AckMessage: fmt.Sprintf("simulated %s completed", cmdType),
	}
	ackData, _ := json.Marshal(ack)

	if err := a.mqtt.publish(ackTopic(a.deviceCode), 1, false, ackData); err != nil {
		log.Printf("❌ ACK 发送失败: %v", err)
		return
	}
	a.totalCmdACK++

	// Emit ACK event to SSE
	emitIfHub(a.hub, func() { a.hub.PublishAck(cmdID, ack.AckCode) })

	log.Printf("✅ 已发送 ACK: cmd_id=%d, code=%s", cmdID, ack.AckCode)
}

// publishHeartbeat publishes an actuator device heartbeat with current states.
func (a *actuatorSim) publishHeartbeat() {
	emitIfHub(a.hub, func() { a.hub.PublishHeartbeat("actuator") })

	states := a.env.GetActuatorStates()

	type actuatorHeartbeatPayload struct {
		TS       string                          `json:"ts"`
		Channels map[uint64]actuatorRuntimeState `json:"channels"`
	}

	payload, _ := json.Marshal(actuatorHeartbeatPayload{
		TS:       time.Now().UTC().Format(time.RFC3339),
		Channels: states,
	})

	if err := a.mqtt.publish(heartbeatTopic(a.deviceCode), 0, false, payload); err != nil {
		log.Printf("❌ 执行器心跳发送失败: %v", err)
	}
}

// publishStatus publishes the actuator device online/offline status.
func (a *actuatorSim) publishStatus(status string) {
	payload, _ := json.Marshal(statusPayload{Status: status})

	if err := a.mqtt.publish(statusTopic(a.deviceCode), 1, false, payload); err != nil {
		log.Printf("❌ 执行器状态上报失败: %v", err)
	}
}
