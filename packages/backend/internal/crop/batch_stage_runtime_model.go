package crop

import "time"

type BatchStageRuntime struct {
	ID                 uint64     `gorm:"primaryKey;autoIncrement"`
	BatchID            uint64     `gorm:"column:batch_id;uniqueIndex;not null"`
	CurrentStagePlan   *uint64    `gorm:"column:current_stage_plan_id"`
	CurrentGrowthStage *uint64    `gorm:"column:current_growth_stage_id"`
	LastSwitchedAt     *time.Time `gorm:"column:last_switched_at"`
	CreatedAt          time.Time  `gorm:"autoCreateTime:milli"`
	UpdatedAt          time.Time  `gorm:"autoUpdateTime:milli"`
}

func (BatchStageRuntime) TableName() string { return "batch_stage_runtime" }
