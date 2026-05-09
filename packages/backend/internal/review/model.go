package review

import (
	"encoding/json"
	"time"
)

const (
	SnapshotDaily        = "DAILY"
	SnapshotWeekly       = "WEEKLY"
	SnapshotStageSummary = "STAGE_SUMMARY"
	SnapshotFinal        = "FINAL"
)

type BatchReviewSnapshot struct {
	ID           uint64          `gorm:"primaryKey;autoIncrement"`
	BatchID      uint64          `gorm:"column:batch_id;not null"`
	SnapshotType string          `gorm:"column:snapshot_type;size:16;default:DAILY"`
	WindowStart  time.Time       `gorm:"column:window_start;not null"`
	WindowEnd    time.Time       `gorm:"column:window_end;not null"`
	Summary      json.RawMessage `gorm:"type:json;not null"`
	GeneratedAt  time.Time       `gorm:"column:generated_at;not null"`
	CreatedAt    time.Time       `gorm:"autoCreateTime:milli"`
}

func (BatchReviewSnapshot) TableName() string { return "batch_review_snapshots" }
