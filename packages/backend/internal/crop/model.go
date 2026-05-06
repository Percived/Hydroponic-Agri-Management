package crop

import "time"

const (
	BatchStatusPlanned    = "PLANNED"
	BatchStatusRunning    = "RUNNING"
	BatchStatusHarvesting = "HARVESTING"
	BatchStatusCompleted  = "COMPLETED"
	BatchStatusAborted    = "ABORTED"

	HarvestGradeA     = "A"
	HarvestGradeB     = "B"
	HarvestGradeC     = "C"
	HarvestGradeWaste = "Waste"
)

type CropVariety struct {
	ID               uint64    `gorm:"primaryKey;autoIncrement"`
	Code             string    `gorm:"size:32;uniqueIndex;not null"`
	Name             string    `gorm:"size:64;not null"`
	Description      string    `gorm:"size:255"`
	DefaultCycleDays *uint     `gorm:"column:default_cycle_days"`
	CreatedAt        time.Time `gorm:"autoCreateTime:milli"`
	UpdatedAt        time.Time `gorm:"autoUpdateTime:milli"`
}

func (CropVariety) TableName() string { return "crop_varieties" }

type GrowthStage struct {
	ID                  uint64    `gorm:"primaryKey;autoIncrement"`
	Code                string    `gorm:"size:32;uniqueIndex;not null"`
	Name                string    `gorm:"size:64;not null"`
	SortOrder           uint      `gorm:"column:sort_order;default:0"`
	DefaultDurationDays *uint     `gorm:"column:default_duration_days"`
	CreatedAt           time.Time `gorm:"autoCreateTime:milli"`
	UpdatedAt           time.Time `gorm:"autoUpdateTime:milli"`
}

func (GrowthStage) TableName() string { return "growth_stages" }

type CropBatch struct {
	ID                uint64     `gorm:"primaryKey;autoIncrement"`
	BatchNo           string     `gorm:"column:batch_no;size:64;uniqueIndex;not null"`
	GreenhouseID      uint64     `gorm:"column:greenhouse_id;not null"`
	GrowingZoneID     *uint64    `gorm:"column:growing_zone_id"`
	CropVarietyID     uint64     `gorm:"column:crop_variety_id;not null"`
	Status            string     `gorm:"size:16;default:PLANNED"`
	PlantingDensity   *float64   `gorm:"column:planting_density;type:decimal(8,2)"`
	TotalPlants       *uint      `gorm:"column:total_plants"`
	StartedAt         *time.Time `gorm:"column:started_at"`
	EndedAt           *time.Time `gorm:"column:ended_at"`
	ExpectedHarvestAt *time.Time `gorm:"column:expected_harvest_at"`
	RecipeVersion     string     `gorm:"column:recipe_version;size:32"`
	PolicyVersion     string     `gorm:"column:policy_version;size:32"`
	Note              string     `gorm:"size:255"`
	CreatedBy         *uint64    `gorm:"column:created_by"`
	CreatedAt         time.Time  `gorm:"autoCreateTime:milli"`
	UpdatedAt         time.Time  `gorm:"autoUpdateTime:milli"`
}

func (CropBatch) TableName() string { return "crop_batches" }

type BatchStagePlan struct {
	ID            uint64    `gorm:"primaryKey;autoIncrement"`
	BatchID       uint64    `gorm:"column:batch_id;not null"`
	GrowthStageID uint64    `gorm:"column:growth_stage_id;not null"`
	StageStartAt  time.Time `gorm:"column:stage_start_at;not null"`
	StageEndAt    time.Time `gorm:"column:stage_end_at;not null"`
	TargetECMin   *float64  `gorm:"column:target_ec_min;type:decimal(12,4)"`
	TargetECMax   *float64  `gorm:"column:target_ec_max;type:decimal(12,4)"`
	TargetPHMin   *float64  `gorm:"column:target_ph_min;type:decimal(12,4)"`
	TargetPHMax   *float64  `gorm:"column:target_ph_max;type:decimal(12,4)"`
	CreatedAt     time.Time `gorm:"autoCreateTime:milli"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime:milli"`
}

func (BatchStagePlan) TableName() string { return "batch_stage_plans" }

type HarvestRecord struct {
	ID              uint64    `gorm:"primaryKey;autoIncrement"`
	BatchID         uint64    `gorm:"column:batch_id;not null"`
	HarvestedAt     time.Time `gorm:"column:harvested_at;not null"`
	HarvestWeightKg float64   `gorm:"column:harvest_weight_kg;type:decimal(10,3);not null"`
	Grade           string    `gorm:"size:8;default:A"`
	GradeWeightKg   float64   `gorm:"column:grade_weight_kg;type:decimal(10,3);not null"`
	Note            string    `gorm:"size:255"`
	HarvestedBy     *uint64   `gorm:"column:harvested_by"`
	CreatedAt       time.Time `gorm:"autoCreateTime:milli"`
}

func (HarvestRecord) TableName() string { return "harvest_records" }
