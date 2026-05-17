package climate

import (
	"fmt"
	"io"
	"log/slog"
	"testing"
	"time"

	"hydroponic-backend/internal/auth"
	"hydroponic-backend/internal/command"
	"hydroponic-backend/internal/device"
	"hydroponic-backend/internal/platform/event"

	mqttlib "github.com/eclipse/paho.mqtt.golang"
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
	publishes []struct {
		topic   string
		payload interface{}
	}
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
	c.publishes = append(c.publishes, struct {
		topic   string
		payload interface{}
	}{topic: topic, payload: payload})
	return &fakeToken{}
}

func TestProfileSchedulerEventDrivenWithTypedTelemetryPayload(t *testing.T) {
	db := openProfileSchedulerTestDB(t)
	hub := event.NewHub()
	profileID := seedClimateProfile(t, db)
	s := NewProfileScheduler(db, hub, &fakeMQTTClient{connected: true}, climateTestLogger())
	s.Start()

	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		hub.Publish(event.SSEEvent{
			Type: "telemetry:received",
			Data: event.TelemetrySSEDataV1{
				SchemaVersion:   1,
				SensorChannelID: 101,
				MetricCode:      "TEMP",
				Value:           35,
				QualityFlag:     "normal",
				CollectedAt:     time.Now().UTC().Format(time.RFC3339),
				DeviceCode:      "SENSOR-001",
			},
		})

		var execCount int64
		if err := db.Model(&ClimateExecutionLog{}).Where("profile_id = ?", profileID).Count(&execCount).Error; err != nil {
			t.Fatalf("count executions: %v", err)
		}
		if execCount > 0 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	var logs []ClimateExecutionLog
	if err := db.Where("profile_id = ?", profileID).Find(&logs).Error; err != nil {
		t.Fatalf("load executions: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("expected typed telemetry event to trigger one climate execution, got %d", len(logs))
	}

	var commands []command.ControlCommand
	if err := db.Find(&commands).Error; err != nil {
		t.Fatalf("load commands: %v", err)
	}
	if len(commands) != 1 {
		t.Fatalf("expected one command to be dispatched, got %d", len(commands))
	}
}

func TestProfileSchedulerExecuteActionUsesFallbackCreator(t *testing.T) {
	db := openProfileSchedulerTestDB(t)
	_, _, action := seedClimateProfileWithAction(t, db)
	s := NewProfileScheduler(db, event.NewHub(), &fakeMQTTClient{connected: true}, climateTestLogger())

	cmdID, err := s.executeAction(action)
	if err != nil {
		t.Fatalf("execute action: %v", err)
	}

	var cmd command.ControlCommand
	if err := db.First(&cmd, cmdID).Error; err != nil {
		t.Fatalf("load command: %v", err)
	}
	var expectedCreator auth.User
	if err := db.Order("id asc").First(&expectedCreator).Error; err != nil {
		t.Fatalf("load fallback user: %v", err)
	}
	if cmd.CreatedBy != expectedCreator.ID {
		t.Fatalf("expected climate auto command created_by=%d, got %d", expectedCreator.ID, cmd.CreatedBy)
	}
}

func openProfileSchedulerTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&auth.User{},
		&ClimateProfile{},
		&ClimateStage{},
		&ClimateStageAction{},
		&ClimateExecutionLog{},
		&command.ControlCommand{},
		&device.ActuatorDevice{},
		&device.ActuatorChannel{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	admin := auth.User{
		Username:     "admin",
		PasswordHash: "test-hash",
		Status:       auth.UserStatusEnabled,
	}
	if err := db.Create(&admin).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
	return db
}

func seedClimateProfile(t *testing.T, db *gorm.DB) uint64 {
	t.Helper()
	profileID, _, _ := seedClimateProfileWithAction(t, db)
	return profileID
}

func seedClimateProfileWithAction(t *testing.T, db *gorm.DB) (uint64, uint64, ClimateStageAction) {
	t.Helper()

	actuator := device.ActuatorDevice{
		DeviceCode:   "ACT-CLIMATE-001",
		Name:         "climate actuator",
		GreenhouseID: 1,
	}
	if err := db.Create(&actuator).Error; err != nil {
		t.Fatalf("create actuator: %v", err)
	}

	channel := device.ActuatorChannel{
		ActuatorDeviceID: actuator.ID,
		ChannelCode:      "fan-1",
		ActuatorType:     device.ActuatorTypeFan,
		Enabled:          true,
	}
	if err := db.Create(&channel).Error; err != nil {
		t.Fatalf("create channel: %v", err)
	}

	triggerChannelID := uint64(101)
	profile := ClimateProfile{
		GreenhouseID:           1,
		Code:                   "CLIMATE-001",
		Name:                   "climate profile",
		TriggerMetricCode:      "TEMP",
		TriggerSensorChannelID: &triggerChannelID,
		Enabled:                true,
	}
	if err := db.Create(&profile).Error; err != nil {
		t.Fatalf("create profile: %v", err)
	}

	stage := ClimateStage{
		ProfileID:        profile.ID,
		StageLevel:       1,
		Name:             "high temp",
		TriggerOperator:  ">",
		TriggerThreshold: 30,
		Hysteresis:       0,
	}
	if err := db.Create(&stage).Error; err != nil {
		t.Fatalf("create stage: %v", err)
	}

	action := ClimateStageAction{
		StageID:           stage.ID,
		ActuatorChannelID: channel.ID,
		CommandType:       "SWITCH",
		CommandPayload:    `{"state":"ON"}`,
		ExecutionOrder:    1,
		Enabled:           true,
	}
	if err := db.Create(&action).Error; err != nil {
		t.Fatalf("create action: %v", err)
	}

	return profile.ID, channel.ID, action
}

func climateTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
