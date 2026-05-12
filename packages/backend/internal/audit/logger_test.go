package audit

import (
	"encoding/json"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestWriteEntry_PersistsExtendedAuditFields(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&AuditLog{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	targetID := uint64(9)
	err = WriteEntry(db, Entry{
		UserID:     7,
		Action:     "CONTROL_CMD",
		TargetType: "COMMAND",
		TargetID:   &targetID,
		Detail: map[string]interface{}{
			"command_type": "SWITCH",
			"status":       "SENT",
		},
		RequestID:  "req-audit-1",
		BeforeData: map[string]interface{}{"status": "PENDING"},
		AfterData:  map[string]interface{}{"status": "SENT"},
	})
	if err != nil {
		t.Fatalf("write entry: %v", err)
	}

	var log AuditLog
	if err := db.First(&log).Error; err != nil {
		t.Fatalf("load audit log: %v", err)
	}

	if log.RequestID != "req-audit-1" {
		t.Fatalf("expected request_id req-audit-1, got %s", log.RequestID)
	}
	if log.TargetID == nil || *log.TargetID != targetID {
		t.Fatalf("expected target_id %d", targetID)
	}
	assertJSONContains(t, log.Detail, `"status":"SENT"`)
	assertJSONContains(t, log.BeforeData, `"status":"PENDING"`)
	assertJSONContains(t, log.AfterData, `"status":"SENT"`)
}

func assertJSONContains(t *testing.T, raw json.RawMessage, want string) {
	t.Helper()
	if len(raw) == 0 {
		t.Fatalf("expected JSON payload, got empty")
	}
	if string(raw) == "" {
		t.Fatalf("expected JSON payload string")
	}
	if !json.Valid(raw) {
		t.Fatalf("expected valid JSON, got %s", string(raw))
	}
	if got := string(raw); !contains(got, want) {
		t.Fatalf("expected %s to contain %s", got, want)
	}
}

func contains(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
