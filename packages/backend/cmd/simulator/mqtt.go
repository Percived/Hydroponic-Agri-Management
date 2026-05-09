package main

import (
	"fmt"
	"log"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// mqttManager handles MQTT connection, publishing, and subscriptions.
type mqttManager struct {
	client mqtt.Client
}

// newMQTTManager creates and connects an MQTT client.
func newMQTTManager(broker, user, pass, clientID string) (*mqttManager, error) {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetClientID(clientID)
	opts.SetAutoReconnect(true)
	opts.SetConnectRetry(true)
	opts.SetConnectRetryInterval(3 * time.Second)
	opts.SetMaxReconnectInterval(10 * time.Second)

	if user != "" {
		opts.SetUsername(user)
		opts.SetPassword(pass)
	}

	opts.OnConnect = func(_ mqtt.Client) {
		log.Println("🔗 MQTT 连接成功")
	}
	opts.OnConnectionLost = func(_ mqtt.Client, err error) {
		log.Printf("⚠ MQTT 连接断开: %v", err)
	}
	opts.OnReconnecting = func(_ mqtt.Client, _ *mqtt.ClientOptions) {
		log.Println("🔄 MQTT 正在重连...")
	}

	client := mqtt.NewClient(opts)
	token := client.Connect()
	if token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}

	return &mqttManager{client: client}, nil
}

// subscribe subscribes to an MQTT topic with the given handler.
func (m *mqttManager) subscribe(topic string, handler mqtt.MessageHandler) error {
	token := m.client.Subscribe(topic, 1, handler)
	if token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

// publish sends a message to an MQTT topic.
func (m *mqttManager) publish(topic string, qos byte, retained bool, payload []byte) error {
	token := m.client.Publish(topic, qos, retained, payload)
	if token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

// disconnect gracefully closes the MQTT connection.
func (m *mqttManager) disconnect(quiesce uint) {
	m.client.Disconnect(quiesce)
}

// ──────────────── MQTT 主题构建函数 ────────────────

func telemetryTopic(deviceCode string) string {
	return fmt.Sprintf("%s/%s/%s", topicPrefix, deviceCode, topicTelemetry)
}

func heartbeatTopic(deviceCode string) string {
	return fmt.Sprintf("%s/%s/%s", topicPrefix, deviceCode, topicHeartbeat)
}

func statusTopic(deviceCode string) string {
	return fmt.Sprintf("%s/%s/%s", topicPrefix, deviceCode, topicStatus)
}

func ackTopic(deviceCode string) string {
	return fmt.Sprintf("%s/%s/%s", topicPrefix, deviceCode, topicAck)
}

func cmdTopic(deviceCode string) string {
	return fmt.Sprintf("%s/%s/%s/#", topicPrefix, deviceCode, topicCmdPrefix)
}

func mqttClientID(prefix, deviceCode string) string {
	return fmt.Sprintf("%s-%s-%d", prefix, deviceCode, os.Getpid())
}
