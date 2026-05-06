package review

import (
	"encoding/json"
	"time"
)

// --- Request DTOs ---

type CreateSnapshotRequest struct {
	BatchID      uint64          `json:"batch_id" binding:"required"`
	SnapshotType string          `json:"snapshot_type" binding:"required,oneof=DAILY WEEKLY STAGE_SUMMARY"`
	WindowStart  time.Time       `json:"window_start" binding:"required"`
	WindowEnd    time.Time       `json:"window_end" binding:"required"`
	Summary      json.RawMessage `json:"summary" binding:"required"`
	GeneratedAt  time.Time       `json:"generated_at" binding:"required"`
}

type UpdateSnapshotRequest struct {
	SnapshotType *string          `json:"snapshot_type"`
	Summary      *json.RawMessage `json:"summary"`
}

type GenerateReviewRequest struct {
	BatchID      uint64    `json:"batch_id" binding:"required"`
	SnapshotType string    `json:"snapshot_type" binding:"required,oneof=DAILY WEEKLY STAGE_SUMMARY"`
	WindowStart  time.Time `json:"window_start" binding:"required"`
	WindowEnd    time.Time `json:"window_end" binding:"required"`
}

// --- Time helpers ---

func timeToStr(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}
