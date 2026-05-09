package device

import (
	"fmt"
	"log/slog"
	"time"

	"hydroponic-backend/internal/alert"
	"hydroponic-backend/internal/platform/event"

	"gorm.io/gorm"
)

const defaultScanInterval = 30 * time.Second

type OfflineDetector struct {
	db           *gorm.DB
	hub          *event.Hub
	log          *slog.Logger
	timeoutSec   int
	scanInterval time.Duration
}

func NewOfflineDetector(db *gorm.DB, hub *event.Hub, log *slog.Logger, timeoutSec int) *OfflineDetector {
	return &OfflineDetector{
		db:           db,
		hub:          hub,
		log:          log,
		timeoutSec:   timeoutSec,
		scanInterval: defaultScanInterval,
	}
}

func (d *OfflineDetector) Start() {
	if d.timeoutSec <= 0 {
		d.log.Warn("offline detector disabled (heartbeat_timeout_sec <= 0)")
		return
	}
	go d.run()
	d.log.Info("offline detector started", "timeout_sec", d.timeoutSec, "scan_interval", d.scanInterval)
}

func (d *OfflineDetector) run() {
	ticker := time.NewTicker(d.scanInterval)
	defer ticker.Stop()
	for range ticker.C {
		d.detectAndMarkOffline()
	}
}

func (d *OfflineDetector) detectAndMarkOffline() {
	cutoff := time.Now().UTC().Add(-time.Duration(d.timeoutSec) * time.Second)
	d.detectSensors(cutoff)
	d.detectActuators(cutoff)
}

type offlineDevice struct {
	ID         uint64
	DeviceCode string
	Name       string
}

func (d *OfflineDetector) detectSensors(cutoff time.Time) {
	var devices []offlineDevice
	err := d.db.Model(&SensorDevice{}).
		Select("id, device_code, name").
		Where("status = ?", StatusOnline).
		Where(
			"(last_seen_at IS NOT NULL AND last_seen_at < ?) OR (last_seen_at IS NULL AND created_at < ?)",
			cutoff, cutoff,
		).
		Find(&devices).Error
	if err != nil {
		d.log.Error("offline detector: failed to query sensor devices", "error", err)
		return
	}

	for _, dev := range devices {
		d.markOffline("sensor_devices", dev.ID, dev.DeviceCode, dev.Name)
	}
}

func (d *OfflineDetector) detectActuators(cutoff time.Time) {
	var devices []offlineDevice
	err := d.db.Model(&ActuatorDevice{}).
		Select("id, device_code, name").
		Where("status = ?", StatusOnline).
		Where(
			"(last_seen_at IS NOT NULL AND last_seen_at < ?) OR (last_seen_at IS NULL AND created_at < ?)",
			cutoff, cutoff,
		).
		Find(&devices).Error
	if err != nil {
		d.log.Error("offline detector: failed to query actuator devices", "error", err)
		return
	}

	for _, dev := range devices {
		d.markOffline("actuator_devices", dev.ID, dev.DeviceCode, dev.Name)
	}
}

func (d *OfflineDetector) markOffline(table string, id uint64, deviceCode, name string) {
	// Race-condition guard: only update if still ONLINE
	result := d.db.Table(table).
		Where("id = ? AND status = ?", id, StatusOnline).
		Update("status", StatusOffline)
	if result.RowsAffected == 0 {
		return
	}

	// Dedup check for existing OPEN DEVICE_OFFLINE alert
	var existing int64
	if err := d.db.Model(&alert.Alert{}).
		Where("type = ? AND status = ? AND message LIKE ?",
			alert.TypeDeviceOffline, alert.StatusOpen, "%["+deviceCode+"]%").
		Count(&existing).Error; err == nil && existing > 0 {
		return
	}

	now := time.Now().UTC()
	a := alert.Alert{
		Type:        alert.TypeDeviceOffline,
		Level:       alert.LevelWarn,
		Message:     fmt.Sprintf("[%s] 设备离线: %s", deviceCode, name),
		Status:      alert.StatusOpen,
		TriggeredAt: now,
	}

	if err := d.db.Create(&a).Error; err != nil {
		d.log.Error("offline detector: failed to create alert", "device_code", deviceCode, "error", err)
		return
	}

	// Create timeline event
	timeline := alert.AlertTimelineEvent{
		AlertID:     a.ID,
		EventType:   alert.EventTriggered,
		EventSource: alert.SourceSystem,
		EventTime:   now,
	}
	d.db.Create(&timeline)

	d.hub.Publish(event.SSEEvent{
		Type: "alert:created",
		Data: alert.BuildAlertSSEDataV1(a, deviceCode, 1),
	})

	d.hub.Publish(event.SSEEvent{
		Type: "device:status",
		Data: map[string]interface{}{
			"device_code": deviceCode,
			"status":      StatusOffline,
			"reason":      "heartbeat_timeout",
		},
	})

	d.log.Info("offline detector: marked device offline",
		"device_code", deviceCode,
		"name", name,
		"table", table,
	)
}
