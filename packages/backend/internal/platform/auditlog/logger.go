package auditlog

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type Entry struct {
	UserID     uint64
	Action     string
	TargetType string
	TargetID   *uint64
	Detail     interface{}
	RequestID  string
	BeforeData interface{}
	AfterData  interface{}
}

func Write(db *gorm.DB, userID uint64, action, targetType string, targetID *uint64, detail interface{}) {
	_ = WriteEntry(db, Entry{
		UserID:     userID,
		Action:     action,
		TargetType: targetType,
		TargetID:   targetID,
		Detail:     detail,
	})
}

func WriteEntry(db *gorm.DB, entry Entry) error {
	if db == nil || entry.UserID == 0 || entry.Action == "" || entry.TargetType == "" {
		return nil
	}

	detail, err := marshalJSON(entry.Detail)
	if err != nil {
		return err
	}
	beforeData, err := marshalJSON(entry.BeforeData)
	if err != nil {
		return err
	}
	afterData, err := marshalJSON(entry.AfterData)
	if err != nil {
		return err
	}

	return db.Table("audit_logs").Create(map[string]interface{}{
		"user_id":     entry.UserID,
		"action":      entry.Action,
		"target_type": entry.TargetType,
		"target_id":   entry.TargetID,
		"detail":      detail,
		"request_id":  entry.RequestID,
		"before_data": beforeData,
		"after_data":  afterData,
		"created_at":  time.Now().UTC(),
	}).Error
}

func marshalJSON(v interface{}) ([]byte, error) {
	if v == nil {
		return nil, nil
	}
	if raw, ok := v.(json.RawMessage); ok {
		return raw, nil
	}
	return json.Marshal(v)
}
