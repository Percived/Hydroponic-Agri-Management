package crop

// ---------- CropVariety ----------

type CreateCropVarietyRequest struct {
	Code             string `json:"code" binding:"required,min=1,max=32"`
	Name             string `json:"name" binding:"required,min=1,max=64"`
	Description      string `json:"description"`
	DefaultCycleDays *uint  `json:"default_cycle_days"`
}

type UpdateCropVarietyRequest struct {
	Name             *string `json:"name" binding:"omitempty,min=1,max=64"`
	Description      *string `json:"description"`
	DefaultCycleDays *uint   `json:"default_cycle_days"`
}

type CropVarietyResponse struct {
	ID               uint64 `json:"id"`
	Code             string `json:"code"`
	Name             string `json:"name"`
	Description      string `json:"description"`
	DefaultCycleDays *uint  `json:"default_cycle_days"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
}

// ---------- GrowthStage ----------

type CreateGrowthStageRequest struct {
	Code                string `json:"code" binding:"required,min=1,max=32"`
	Name                string `json:"name" binding:"required,min=1,max=64"`
	SortOrder           uint   `json:"sort_order"`
	DefaultDurationDays *uint  `json:"default_duration_days"`
}

type UpdateGrowthStageRequest struct {
	Name                *string `json:"name" binding:"omitempty,min=1,max=64"`
	SortOrder           *uint   `json:"sort_order"`
	DefaultDurationDays *uint   `json:"default_duration_days"`
}

type GrowthStageResponse struct {
	ID                  uint64 `json:"id"`
	Code                string `json:"code"`
	Name                string `json:"name"`
	SortOrder           uint   `json:"sort_order"`
	DefaultDurationDays *uint  `json:"default_duration_days"`
	CreatedAt           string `json:"created_at"`
	UpdatedAt           string `json:"updated_at"`
}

// ---------- CropBatch ----------

type CreateCropBatchRequest struct {
	BatchNo           string   `json:"batch_no" binding:"required,min=1,max=64"`
	GreenhouseID      uint64   `json:"greenhouse_id" binding:"required"`
	GrowingZoneID     *uint64  `json:"growing_zone_id"`
	CropVarietyID     uint64   `json:"crop_variety_id" binding:"required"`
	Status            string   `json:"status"`
	PlantingDensity   *float64 `json:"planting_density"`
	TotalPlants       *uint    `json:"total_plants"`
	StartedAt         *string  `json:"started_at"`
	EndedAt           *string  `json:"ended_at"`
	ExpectedHarvestAt *string  `json:"expected_harvest_at"`
	RecipeVersion     string   `json:"recipe_version"`
	PolicyVersion     string   `json:"policy_version"`
	Note              string   `json:"note"`
}

type UpdateCropBatchRequest struct {
	BatchNo           *string  `json:"batch_no" binding:"omitempty,min=1,max=64"`
	GreenhouseID      *uint64  `json:"greenhouse_id"`
	GrowingZoneID     *uint64  `json:"growing_zone_id"`
	CropVarietyID     *uint64  `json:"crop_variety_id"`
	PlantingDensity   *float64 `json:"planting_density"`
	TotalPlants       *uint    `json:"total_plants"`
	StartedAt         *string  `json:"started_at"`
	EndedAt           *string  `json:"ended_at"`
	ExpectedHarvestAt *string  `json:"expected_harvest_at"`
	RecipeVersion     *string  `json:"recipe_version"`
	PolicyVersion     *string  `json:"policy_version"`
	Note              *string  `json:"note"`
}

type BatchStatusTransitionRequest struct {
	Status string `json:"status" binding:"required,oneof=RUNNING HARVESTING COMPLETED ABORTED"`
	Note   string `json:"note"`
}

type CropBatchResponse struct {
	ID                uint64   `json:"id"`
	BatchNo           string   `json:"batch_no"`
	GreenhouseID      uint64   `json:"greenhouse_id"`
	GrowingZoneID     *uint64  `json:"growing_zone_id"`
	CropVarietyID     uint64   `json:"crop_variety_id"`
	VarietyCode       string   `json:"variety_code"`
	VarietyName       string   `json:"variety_name"`
	Status            string   `json:"status"`
	PlantingDensity   *float64 `json:"planting_density"`
	TotalPlants       *uint    `json:"total_plants"`
	StartedAt         *string  `json:"started_at"`
	EndedAt           *string  `json:"ended_at"`
	ExpectedHarvestAt *string  `json:"expected_harvest_at"`
	RecipeVersion     string   `json:"recipe_version"`
	PolicyVersion     string   `json:"policy_version"`
	Note              string   `json:"note"`
	CreatedBy         *uint64  `json:"created_by"`
	CreatedAt         string   `json:"created_at"`
	UpdatedAt         string   `json:"updated_at"`
}

// ---------- BatchStagePlan ----------

type CreateBatchStagePlanRequest struct {
	BatchID       uint64   `json:"batch_id" binding:"required"`
	GrowthStageID uint64   `json:"growth_stage_id" binding:"required"`
	StageStartAt  string   `json:"stage_start_at" binding:"required"`
	StageEndAt    string   `json:"stage_end_at" binding:"required"`
	TargetECMin   *float64 `json:"target_ec_min"`
	TargetECMax   *float64 `json:"target_ec_max"`
	TargetPHMin   *float64 `json:"target_ph_min"`
	TargetPHMax   *float64 `json:"target_ph_max"`
}

type UpdateBatchStagePlanRequest struct {
	StageStartAt *string  `json:"stage_start_at"`
	StageEndAt   *string  `json:"stage_end_at"`
	TargetECMin  *float64 `json:"target_ec_min"`
	TargetECMax  *float64 `json:"target_ec_max"`
	TargetPHMin  *float64 `json:"target_ph_min"`
	TargetPHMax  *float64 `json:"target_ph_max"`
}

type BatchStagePlanResponse struct {
	ID            uint64   `json:"id"`
	BatchID       uint64   `json:"batch_id"`
	GrowthStageID uint64   `json:"growth_stage_id"`
	StageStartAt  string   `json:"stage_start_at"`
	StageEndAt    string   `json:"stage_end_at"`
	TargetECMin   *float64 `json:"target_ec_min"`
	TargetECMax   *float64 `json:"target_ec_max"`
	TargetPHMin   *float64 `json:"target_ph_min"`
	TargetPHMax   *float64 `json:"target_ph_max"`
	CreatedAt     string   `json:"created_at"`
	UpdatedAt     string   `json:"updated_at"`
}

// ---------- HarvestRecord ----------

type CreateHarvestRecordRequest struct {
	BatchID         uint64  `json:"batch_id" binding:"required"`
	HarvestedAt     string  `json:"harvested_at" binding:"required"`
	HarvestWeightKg float64 `json:"harvest_weight_kg" binding:"required,gt=0"`
	Grade           string  `json:"grade"`
	GradeWeightKg   float64 `json:"grade_weight_kg" binding:"required,gt=0"`
	Note            string  `json:"note"`
}

type HarvestRecordResponse struct {
	ID              uint64  `json:"id"`
	BatchID         uint64  `json:"batch_id"`
	HarvestedAt     string  `json:"harvested_at"`
	HarvestWeightKg float64 `json:"harvest_weight_kg"`
	Grade           string  `json:"grade"`
	GradeWeightKg   float64 `json:"grade_weight_kg"`
	Note            string  `json:"note"`
	HarvestedBy     *uint64 `json:"harvested_by"`
	CreatedAt       string  `json:"created_at"`
}

type HarvestSummaryResponse struct {
	BatchID       uint64                `json:"batch_id"`
	TotalWeightKg float64               `json:"total_weight_kg"`
	Grades        []HarvestGradeSummary `json:"grades"`
}

type HarvestGradeSummary struct {
	Grade    string  `json:"grade"`
	WeightKg float64 `json:"weight_kg"`
	Count    int64   `json:"count"`
}

// ---------- Common List Response ----------

type CropListResponse struct {
	Items    interface{} `json:"items"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}
