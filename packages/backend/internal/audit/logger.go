package audit

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

func Write(db *gorm.DB, userID uint64, action, targetType string, targetID *uint64, detail interface{}) {
	if db == nil || userID == 0 || action == "" || targetType == "" {
		return
	}

	var b []byte
	if detail != nil {
		if raw, err := json.Marshal(detail); err == nil {
			b = raw
		}
	}

	_ = db.Table("audit_logs").Create(map[string]interface{}{
		"user_id":     userID,
		"action":      action,
		"target_type": targetType,
		"target_id":   targetID,
		"detail":      b,
		"created_at":  time.Now().UTC(),
	}).Error
}
