package alert

import (
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestAlertTimelineEvent_Create_DefaultsEmptyPayloadToJSONObject(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&Alert{}, &AlertTimelineEvent{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	alert := Alert{
		Type:        TypeDeviceOffline,
		Level:       LevelWarn,
		Message:     "[dev_001] device offline",
		Status:      StatusOpen,
		TriggeredAt: time.Now().UTC(),
	}
	if err := db.Create(&alert).Error; err != nil {
		t.Fatalf("create alert: %v", err)
	}

	event := AlertTimelineEvent{
		AlertID:     alert.ID,
		EventType:   EventTriggered,
		EventSource: SourceSystem,
		EventTime:   time.Now().UTC(),
	}
	if err := db.Create(&event).Error; err != nil {
		t.Fatalf("create timeline event: %v", err)
	}

	var saved AlertTimelineEvent
	if err := db.First(&saved, event.ID).Error; err != nil {
		t.Fatalf("reload timeline event: %v", err)
	}
	if saved.EventPayload != "{}" {
		t.Fatalf("expected default event_payload to be {}, got %q", saved.EventPayload)
	}
}
