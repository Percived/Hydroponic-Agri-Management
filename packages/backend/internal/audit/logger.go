package audit

import (
	"hydroponic-backend/internal/platform/auditlog"

	"gorm.io/gorm"
)

type Entry = auditlog.Entry

func Write(db *gorm.DB, userID uint64, action, targetType string, targetID *uint64, detail interface{}) {
	auditlog.Write(db, userID, action, targetType, targetID, detail)
}

func WriteEntry(db *gorm.DB, entry Entry) error {
	return auditlog.WriteEntry(db, entry)
}
