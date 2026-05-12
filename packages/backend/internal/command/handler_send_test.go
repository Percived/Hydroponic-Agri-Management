package command

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"hydroponic-backend/internal/device"
	"hydroponic-backend/internal/platform/event"

	mqttlib "github.com/eclipse/paho.mqtt.golang"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type fakeToken struct {
	err error
}

func (t *fakeToken) Wait() bool                       { return true }
func (t *fakeToken) WaitTimeout(_ time.Duration) bool { return true }
func (t *fakeToken) Done() <-chan struct{}            { ch := make(chan struct{}); close(ch); return ch }
func (t *fakeToken) Error() error                     { return t.err }

type fakeMQTTClient struct {
	connected bool
	topic     string
	payload   interface{}
}

func (c *fakeMQTTClient) IsConnected() bool      { return c.connected }
func (c *fakeMQTTClient) IsConnectionOpen() bool { return c.connected }
func (c *fakeMQTTClient) Connect() mqttlib.Token { return &fakeToken{} }
func (c *fakeMQTTClient) Disconnect(_ uint)      {}
func (c *fakeMQTTClient) Subscribe(string, byte, mqttlib.MessageHandler) mqttlib.Token {
	return &fakeToken{}
}
func (c *fakeMQTTClient) SubscribeMultiple(map[string]byte, mqttlib.MessageHandler) mqttlib.Token {
	return &fakeToken{}
}
func (c *fakeMQTTClient) Unsubscribe(...string) mqttlib.Token     { return &fakeToken{} }
func (c *fakeMQTTClient) AddRoute(string, mqttlib.MessageHandler) {}
func (c *fakeMQTTClient) OptionsReader() mqttlib.ClientOptionsReader {
	return mqttlib.ClientOptionsReader{}
}
func (c *fakeMQTTClient) Publish(topic string, _ byte, _ bool, payload interface{}) mqttlib.Token {
	c.topic = topic
	c.payload = payload
	return &fakeToken{}
}

func TestSendCommand_MQTTNotConnected_MarksFailedAndCreatesReceipt(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&ControlCommand{}, &ControlCommandReceipt{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	cmd := ControlCommand{
		ActuatorChannelID: 1,
		CommandType:       "SWITCH",
		Payload:           `{"state":"ON"}`,
		Status:            "PENDING",
		CreatedBy:         1,
	}
	if err := db.Create(&cmd).Error; err != nil {
		t.Fatalf("create cmd: %v", err)
	}

	h := NewHandler(db, nil, event.NewHub())

	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)
	r.POST("/commands/:id/send", h.SendCommand)

	req := httptest.NewRequest(http.MethodPost, "/commands/"+itoa(cmd.ID)+"/send", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: itoa(cmd.ID)}}

	h.SendCommand(c)

	var reloaded ControlCommand
	if err := db.First(&reloaded, cmd.ID).Error; err != nil {
		t.Fatalf("reload cmd: %v", err)
	}
	if reloaded.Status != "FAILED" {
		t.Fatalf("expected FAILED, got %s", reloaded.Status)
	}
	if reloaded.SentAt != nil {
		t.Fatalf("expected sent_at nil")
	}

	var rcpt ControlCommandReceipt
	if err := db.Where("command_id = ?", cmd.ID).First(&rcpt).Error; err != nil {
		t.Fatalf("receipt: %v", err)
	}
	if rcpt.ReceiptStatus != "FAILED" {
		t.Fatalf("expected receipt FAILED, got %s", rcpt.ReceiptStatus)
	}
	if rcpt.AckAt == nil || time.Since(*rcpt.AckAt) > time.Minute {
		t.Fatalf("expected ack_at set")
	}
}

func TestDispatchMQTT_IncludesTargetActuatorChannelMetadata(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&ControlCommand{}, &device.ActuatorDevice{}, &device.ActuatorChannel{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	dev := device.ActuatorDevice{
		ID:           10,
		GreenhouseID: 1,
		DeviceCode:   "ACT-001",
		Name:         "actuator-1",
	}
	if err := db.Create(&dev).Error; err != nil {
		t.Fatalf("create actuator device: %v", err)
	}
	ch := device.ActuatorChannel{
		ID:               101,
		ActuatorDeviceID: dev.ID,
		ChannelCode:      "fan-a",
		ActuatorType:     device.ActuatorTypeFan,
	}
	if err := db.Create(&ch).Error; err != nil {
		t.Fatalf("create actuator channel: %v", err)
	}

	client := &fakeMQTTClient{connected: true}
	h := NewHandler(db, client, event.NewHub())

	cmd := ControlCommand{
		ID:                99,
		ActuatorChannelID: ch.ID,
		CommandType:       "SWITCH",
		Payload:           `{"state":"ON","value":80}`,
		Status:            "PENDING",
		CreatedBy:         1,
	}

	deviceCode, err := h.dispatchMQTT(cmd)
	if err != nil {
		t.Fatalf("dispatchMQTT: %v", err)
	}
	if deviceCode != dev.DeviceCode {
		t.Fatalf("expected device code %s, got %s", dev.DeviceCode, deviceCode)
	}
	if client.topic != "hydroponic/ACT-001/cmd/SWITCH" {
		t.Fatalf("unexpected topic: %s", client.topic)
	}

	payloadBytes, ok := client.payload.([]byte)
	if !ok {
		t.Fatalf("expected []byte payload, got %T", client.payload)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}

	if got := uint64(payload["_command_id"].(float64)); got != cmd.ID {
		t.Fatalf("expected _command_id %d, got %d", cmd.ID, got)
	}
	if got := payload["actuator_channel_id"]; got == nil {
		t.Fatalf("expected actuator_channel_id in payload")
	}
	if got := payload["channel_code"]; got == nil {
		t.Fatalf("expected channel_code in payload")
	}
}

func itoa(v uint64) string {
	return fmt.Sprintf("%d", v)
}
