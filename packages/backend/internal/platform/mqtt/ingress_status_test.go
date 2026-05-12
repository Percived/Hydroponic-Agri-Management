package mqtt

import (
	"io"
	"log/slog"
	"testing"

	"hydroponic-backend/internal/device"
	"hydroponic-backend/internal/platform/config"
	"hydroponic-backend/internal/platform/event"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestIngressService_HandleStatus_InvalidStatusIgnored(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&device.SensorDevice{}, &device.ActuatorDevice{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	if err := db.Create(&device.SensorDevice{
		GreenhouseID: 1,
		DeviceCode:   "dev_001",
		Name:         "s1",
		Status:       device.StatusOnline,
	}).Error; err != nil {
		t.Fatalf("create: %v", err)
	}

	hub := event.NewHub()
	sub := hub.Subscribe(func(e event.SSEEvent) bool { return e.Type == "device:status" })
	defer hub.Unsubscribe(sub)

	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	ing := NewIngressService(db, nil, config.InfluxConfig{}, hub, nil, nil, log)

	ing.handleStatus("dev_001", []byte(`{"status":"BAD"}`))

	var d device.SensorDevice
	if err := db.Where("device_code = ?", "dev_001").First(&d).Error; err != nil {
		t.Fatalf("reload: %v", err)
	}
	if d.Status != device.StatusOnline {
		t.Fatalf("expected status unchanged, got %s", d.Status)
	}
	select {
	case <-sub.Events:
		t.Fatalf("expected no device:status event for invalid status")
	default:
	}
}

func TestIngressService_HandleState_UpdatesActuatorChannelByDeviceAndChannelCode(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&device.ActuatorDevice{}, &device.ActuatorChannel{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	actuator := device.ActuatorDevice{
		GreenhouseID: 1,
		DeviceCode:   "test-action_1",
		Name:         "relay",
		Status:       device.StatusOnline,
	}
	if err := db.Create(&actuator).Error; err != nil {
		t.Fatalf("create actuator: %v", err)
	}
	level := 50.0
	channel := device.ActuatorChannel{
		ActuatorDeviceID: actuator.ID,
		ChannelCode:      "ch2",
		ActuatorType:     device.ActuatorTypePump,
		CurrentState:     "ON",
		CurrentLevel:     &level,
		Enabled:          true,
	}
	if err := db.Create(&channel).Error; err != nil {
		t.Fatalf("create channel: %v", err)
	}

	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	ing := NewIngressService(db, nil, config.InfluxConfig{}, event.NewHub(), nil, nil, log)

	ing.handleState("test-action_1", []byte(`{"channels":[{"channel_code":"ch2","state":"OFF"}]}`))

	var reloaded device.ActuatorChannel
	if err := db.First(&reloaded, channel.ID).Error; err != nil {
		t.Fatalf("reload channel: %v", err)
	}
	if reloaded.CurrentState != "OFF" {
		t.Fatalf("expected current_state OFF, got %s", reloaded.CurrentState)
	}
	if reloaded.CurrentLevel == nil || *reloaded.CurrentLevel != 50 {
		t.Fatalf("expected current_level unchanged, got %+v", reloaded.CurrentLevel)
	}
}
