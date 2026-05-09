package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// ──────────────────── 传感器模拟器 ────────────────────

// sensorSim simulates a sensor device that reads from the environment model.
type sensorSim struct {
	deviceCode string
	channels   []sensorChannelDetail
	cfgByChan  map[uint64]metricConfig
	mqtt       *mqttManager
	env        *Environment
	rngSource  *rand.Rand

	// Stats
	totalTelemetry int64
}

// newSensorSim creates a new sensor simulator.
func newSensorSim(deviceCode string, channels []sensorChannelDetail, cfgByChan map[uint64]metricConfig, mqtt *mqttManager, env *Environment, rng *rand.Rand) *sensorSim {
	return &sensorSim{
		deviceCode: deviceCode,
		channels:   channels,
		cfgByChan:  cfgByChan,
		mqtt:       mqtt,
		env:        env,
		rngSource:  rng,
	}
}

// sendTelemetry reads environment state and publishes telemetry via MQTT.
func (s *sensorSim) sendTelemetry(anomalyRate float64) {
	if len(s.channels) == 0 {
		return
	}

	now := time.Now().UTC().Format(time.RFC3339)
	items := make([]telemetryItem, 0, len(s.channels))

	for _, ch := range s.channels {
		// Read from environment model
		val := s.env.GetSensorReading(ch.MetricCode)
		qualityFlag := "normal"

		// Apply anomaly injection at the telemetry layer (does not pollute environment)
		if s.rngSource.Float64() < anomalyRate {
			cfg, ok := s.cfgByChan[ch.ID]
			if !ok {
				cfg = metricConfig{Code: ch.MetricCode, Anomaly: 20}
			}
			val = val + (s.rngSource.Float64()-0.5)*2*cfg.Anomaly
			qualityFlag = "out_of_range"
			log.Printf("⚠ 注入异常数据: %s = %.1f", ch.MetricCode, val)
		}

		items = append(items, telemetryItem{
			SensorChannelID: ch.ID,
			MetricCode:      ch.MetricCode,
			Value:           round(val, 2),
			QualityFlag:     qualityFlag,
			CollectedAt:     now,
		})
	}

	payload, err := json.Marshal(items)
	if err != nil {
		log.Printf("❌ 序列化遥测数据失败: %v", err)
		return
	}

	if err := s.mqtt.publish(telemetryTopic(s.deviceCode), 1, false, payload); err != nil {
		log.Printf("❌ 遥测上报失败: %v", err)
		return
	}

	s.totalTelemetry++
	if s.totalTelemetry%10 == 1 || s.totalTelemetry == 1 {
		log.Printf("[%s] ✅ 遥测上报 #%d (%d 通道)",
			time.Now().Format("15:04:05"), s.totalTelemetry, len(items))
	}
}

// publishHeartbeat publishes a sensor device heartbeat.
func (s *sensorSim) publishHeartbeat() {
	payload, _ := json.Marshal(heartbeatPayload{
		TS: time.Now().UTC().Format(time.RFC3339),
	})

	if err := s.mqtt.publish(heartbeatTopic(s.deviceCode), 0, false, payload); err != nil {
		log.Printf("❌ 传感器心跳发送失败: %v", err)
	}
}

// publishStatus publishes the device online/offline status.
func (s *sensorSim) publishStatus(status string) {
	payload, _ := json.Marshal(statusPayload{Status: status})

	if err := s.mqtt.publish(statusTopic(s.deviceCode), 1, false, payload); err != nil {
		log.Printf("❌ 传感器状态上报失败: %v", err)
	}
}

// onCommand is the MQTT command callback for the sensor device.
// It sends a simple ACK without modifying the environment.
func (s *sensorSim) onCommand(_ mqtt.Client, msg mqtt.Message) {
	log.Printf("📥 传感器收到命令: topic=%s, payload=%s", msg.Topic(), string(msg.Payload()))

	var cmd incomingCmd
	if err := json.Unmarshal(msg.Payload(), &cmd); err != nil {
		log.Printf("   无法解析命令结构: %v", err)
		return
	}

	cmdID := cmd.CommandID
	if cmdID == 0 {
		cmdID = cmd.InternalCommandID
	}
	if cmdID == 0 {
		log.Printf("   收到无 command_id 的消息，无需 ACK")
		return
	}

	time.Sleep(100 * time.Millisecond)

	ack := ackPayload{
		CommandID:  cmdID,
		AckCode:    "ok",
		AckMessage: fmt.Sprintf("simulated sensor ack for cmd %d", cmdID),
	}
	ackData, _ := json.Marshal(ack)

	if err := s.mqtt.publish(ackTopic(s.deviceCode), 1, false, ackData); err != nil {
		log.Printf("❌ 传感器 ACK 发送失败: %v", err)
		return
	}
	log.Printf("✅ 传感器 ACK: cmd_id=%d", cmdID)
}
