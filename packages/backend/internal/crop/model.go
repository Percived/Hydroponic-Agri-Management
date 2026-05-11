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

	DeviceTypeSensor   = "sensor"
	DeviceTypeActuator = "actuator"
)

// BatchStatusTransitions defines legal status transitions.
// PLANNED → RUNNING → HARVESTING → COMPLETED
// Any status → ABORTED
var BatchStatusTransitions = map[string][]string{
	BatchStatusPlanned:    {BatchStatusRunning, BatchStatusAborted},
	BatchStatusRunning:    {BatchStatusHarvesting, BatchStatusAborted},
	BatchStatusHarvesting: {BatchStatusCompleted, BatchStatusAborted},
	BatchStatusCompleted:  {},
	BatchStatusAborted:    {},
}

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
	ActiveRecipeID    *uint64    `gorm:"column:active_recipe_id"`
	ActivePolicyID    *uint64    `gorm:"column:active_policy_id"`
	ActiveClimateID   *uint64    `gorm:"column:active_climate_profile_id"`
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
	RecipeID      *uint64   `gorm:"column:recipe_id"`
	PolicyID      *uint64   `gorm:"column:policy_id"`
	ClimateID     *uint64   `gorm:"column:climate_profile_id"`
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

// BatchDevice binds a sensor or actuator device to a batch.
type BatchDevice struct {
	ID         uint64     `gorm:"primaryKey;autoIncrement"`
	BatchID    uint64     `gorm:"column:batch_id;not null"`
	DeviceType string     `gorm:"column:device_type;size:16;not null"` // sensor / actuator
	DeviceID   uint64     `gorm:"column:device_id;not null"`
	BoundAt    time.Time  `gorm:"column:bound_at;not null"`
	UnboundAt  *time.Time `gorm:"column:unbound_at"`
	IsActive   bool       `gorm:"column:is_active;default:1"`
}

func (BatchDevice) TableName() string { return "batch_devices" }

// PlantingRecord stores planting/transplanting details. 1:1 with CropBatch.
type PlantingRecord struct {
	ID                      uint64     `gorm:"primaryKey;autoIncrement"`
	BatchID                 uint64     `gorm:"column:batch_id;uniqueIndex;not null"`
	SeedSource              string     `gorm:"column:seed_source;size:128"`
	SeedBatchNo             string     `gorm:"column:seed_batch_no;size:64"`
	SeedlingAgeDays         *uint      `gorm:"column:seedling_age_days"`
	SeededAt                *time.Time `gorm:"column:seeded_at"`
	PlantedAt               *time.Time `gorm:"column:planted_at"`
	ActualPlantCount        *uint      `gorm:"column:actual_plant_count"`
	InitialEC               *float64   `gorm:"column:initial_ec;type:decimal(12,4)"`
	InitialPH               *float64   `gorm:"column:initial_ph;type:decimal(12,4)"`
	InitialWaterTemp        *float64   `gorm:"column:initial_water_temp;type:decimal(12,4)"`
	InitialNutrientRecipeID *uint64    `gorm:"column:initial_nutrient_recipe_id"`
	PlantedBy               *uint64    `gorm:"column:planted_by"`
	Note                    string     `gorm:"size:255"`
	CreatedAt               time.Time  `gorm:"autoCreateTime:milli"`
	UpdatedAt               time.Time  `gorm:"autoUpdateTime:milli"`
}

func (PlantingRecord) TableName() string { return "planting_records" }
