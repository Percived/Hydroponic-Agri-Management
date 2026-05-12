package main

import (
	"encoding/json"
	"math/rand"
	"testing"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type fakeMQTTToken struct {
	err error
}

func (t *fakeMQTTToken) Wait() bool                       { return true }
func (t *fakeMQTTToken) WaitTimeout(_ time.Duration) bool { return true }
func (t *fakeMQTTToken) Done() <-chan struct{}            { ch := make(chan struct{}); close(ch); return ch }
func (t *fakeMQTTToken) Error() error                     { return t.err }

type fakeMQTTClient struct {
	connected bool
	publishes []struct {
		topic   string
		payload interface{}
	}
}

func (c *fakeMQTTClient) IsConnected() bool      { return c.connected }
func (c *fakeMQTTClient) IsConnectionOpen() bool { return c.connected }
func (c *fakeMQTTClient) Connect() mqtt.Token    { return &fakeMQTTToken{} }
func (c *fakeMQTTClient) Disconnect(_ uint)      {}
func (c *fakeMQTTClient) Subscribe(string, byte, mqtt.MessageHandler) mqtt.Token {
	return &fakeMQTTToken{}
}
func (c *fakeMQTTClient) SubscribeMultiple(map[string]byte, mqtt.MessageHandler) mqtt.Token {
	return &fakeMQTTToken{}
}
func (c *fakeMQTTClient) Unsubscribe(...string) mqtt.Token     { return &fakeMQTTToken{} }
func (c *fakeMQTTClient) AddRoute(string, mqtt.MessageHandler) {}
func (c *fakeMQTTClient) OptionsReader() mqtt.ClientOptionsReader {
	return mqtt.ClientOptionsReader{}
}
func (c *fakeMQTTClient) Publish(topic string, _ byte, _ bool, payload interface{}) mqtt.Token {
	c.publishes = append(c.publishes, struct {
		topic   string
		payload interface{}
	}{topic: topic, payload: payload})
	return &fakeMQTTToken{}
}

type fakeMQTTMessage struct {
	topic   string
	payload []byte
}

func (m *fakeMQTTMessage) Duplicate() bool { return false }
func (m *fakeMQTTMessage) Qos() byte       { return 1 }
func (m *fakeMQTTMessage) Retained() bool  { return false }
func (m *fakeMQTTMessage) Topic() string   { return m.topic }
func (m *fakeMQTTMessage) MessageID() uint16 {
	return 1
}
func (m *fakeMQTTMessage) Payload() []byte { return m.payload }
func (m *fakeMQTTMessage) Ack()            {}

func TestActuatorOnCommand_OnlyAppliesTargetChannel(t *testing.T) {
	env := NewEnvironment(rand.New(rand.NewSource(1)))
	client := &fakeMQTTClient{connected: true}
	sim := newActuatorSim("ACT-001", []actuatorChannelDetail{
		{ID: 101, ChannelCode: "fan-a", ActuatorType: ActuatorFAN},
		{ID: 102, ChannelCode: "fan-b", ActuatorType: ActuatorFAN},
	}, &mqttManager{client: client}, env, nil)

	payload, err := json.Marshal(map[string]interface{}{
		"command_id":          99,
		"command_type":        "SWITCH",
		"actuator_channel_id": 101,
		"channel_code":        "fan-a",
		"state":               "ON",
		"value":               80,
	})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	sim.onCommand(nil, &fakeMQTTMessage{
		topic:   "hydroponic/ACT-001/cmd/SWITCH",
		payload: payload,
	})

	states := env.GetActuatorStates()
	if len(states) != 1 {
		t.Fatalf("expected only one actuator state updated, got %d", len(states))
	}
	got, ok := states[101]
	if !ok {
		t.Fatalf("expected target channel 101 updated")
	}
	if got.State != "ON" || got.Value != 80 {
		t.Fatalf("unexpected target state: %+v", got)
	}
	if _, exists := states[102]; exists {
		t.Fatalf("expected non-target channel 102 untouched")
	}
}
