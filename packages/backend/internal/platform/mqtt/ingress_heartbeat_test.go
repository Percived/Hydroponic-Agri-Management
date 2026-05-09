package mqtt

import (
	"io"
	"log/slog"
	"testing"
	"time"

	"hydroponic-backend/internal/alert"
	"hydroponic-backend/internal/device"
	"hydroponic-backend/internal/platform/config"
	"hydroponic-backend/internal/platform/event"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestIngressService_HandleHeartbeat_AutoResolveOfflineWritesTimeline(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&device.SensorDevice{}, &alert.Alert{}, &alert.AlertTimelineEvent{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if err := db.AutoMigrate(&device.ActuatorDevice{}); err != nil {
		t.Fatalf("migrate actuator: %v", err)
	}

	const deviceCode = "dev_001"
	if err := db.Create(&device.SensorDevice{
		GreenhouseID: 1,
		DeviceCode:   deviceCode,
		Name:         "sensor-1",
		Status:       device.StatusOffline,
	}).Error; err != nil {
		t.Fatalf("create sensor device: %v", err)
	}

	now := time.Now().UTC().Add(-10 * time.Minute)
	offlineAlert := alert.Alert{
		Type:        alert.TypeDeviceOffline,
		Level:       alert.LevelWarn,
		Message:     "[" + deviceCode + "] 设备离线: sensor-1",
		Status:      alert.StatusOpen,
		TriggeredAt: now,
	}
	if err := db.Create(&offlineAlert).Error; err != nil {
		t.Fatalf("create offline alert: %v", err)
	}

	hub := event.NewHub()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	ing := NewIngressService(db, nil, config.InfluxConfig{}, hub, nil, nil, log)

	ing.handleHeartbeat(deviceCode, nil)

	var a alert.Alert
	if err := db.First(&a, offlineAlert.ID).Error; err != nil {
		t.Fatalf("reload alert: %v", err)
	}
	if a.Status != alert.StatusResolved {
		t.Fatalf("expected resolved, got %s", a.Status)
	}
	if a.ResolvedAt == nil {
		t.Fatalf("expected resolved_at set")
	}

	var count int64
	if err := db.Model(&alert.AlertTimelineEvent{}).
		Where("alert_id = ? AND event_type = ?", offlineAlert.ID, alert.EventResolved).
		Count(&count).Error; err != nil {
		t.Fatalf("count timeline: %v", err)
	}
	if count == 0 {
		t.Fatalf("expected resolved timeline event written")
	}
}
