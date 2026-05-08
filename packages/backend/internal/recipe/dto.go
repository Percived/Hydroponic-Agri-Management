package recipe

// ---------- NutrientRecipe ----------

type CreateRecipeRequest struct {
	RecipeCode    string                     `json:"recipe_code" binding:"required,min=1,max=64"`
	Name          string                     `json:"name" binding:"required,min=1,max=128"`
	CropVarietyID *uint64                    `json:"crop_variety_id"`
	Description   string                     `json:"description"`
	Version       string                     `json:"version"`
	Status        string                     `json:"status"`
	EffectiveFrom *string                    `json:"effective_from"`
	EffectiveTo   *string                    `json:"effective_to"`
	StageTargets  []CreateStageTargetRequest `json:"stage_targets"`
	IonTargets    []CreateIonTargetRequest   `json:"ion_targets"`
}

type UpdateRecipeRequest struct {
	Name          *string `json:"name" binding:"omitempty,min=1,max=128"`
	CropVarietyID *uint64 `json:"crop_variety_id"`
	Description   *string `json:"description"`
	Version       *string `json:"version"`
	Status        *string `json:"status" binding:"omitempty,oneof=DRAFT ACTIVE ARCHIVED"`
	EffectiveFrom *string `json:"effective_from"`
	EffectiveTo   *string `json:"effective_to"`
}

type PublishRecipeRequest struct {
	Version string `json:"version" binding:"required,min=1,max=32"`
}

type NutrientRecipeResponse struct {
	ID            uint64  `json:"id"`
	RecipeCode    string  `json:"recipe_code"`
	Name          string  `json:"name"`
	CropVarietyID *uint64 `json:"crop_variety_id"`
	Description   string  `json:"description"`
	Version       string  `json:"version"`
	Status        string  `json:"status"`
	EffectiveFrom *string `json:"effective_from"`
	EffectiveTo   *string `json:"effective_to"`
	CreatedBy     *uint64 `json:"created_by"`
	PublishedBy   *uint64 `json:"published_by"`
	PublishedAt   *string `json:"published_at"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

// ---------- RecipeStageTarget ----------

type CreateStageTargetRequest struct {
	GrowthStageID *uint64  `json:"growth_stage_id"`
	MetricCode    string   `json:"metric_code" binding:"required,min=1,max=32"`
	TargetMin     *float64 `json:"target_min"`
	TargetMax     *float64 `json:"target_max"`
	Tolerance     *float64 `json:"tolerance"`
	Unit          string   `json:"unit"`
	Enabled       *bool    `json:"enabled"`
}

type RecipeStageTargetResponse struct {
	ID            uint64   `json:"id"`
	RecipeID      uint64   `json:"recipe_id"`
	GrowthStageID *uint64  `json:"growth_stage_id"`
	MetricCode    string   `json:"metric_code"`
	TargetMin     *float64 `json:"target_min"`
	TargetMax     *float64 `json:"target_max"`
	Tolerance     *float64 `json:"tolerance"`
	Unit          string   `json:"unit"`
	Enabled       bool     `json:"enabled"`
	CreatedAt     string   `json:"created_at"`
	UpdatedAt     string   `json:"updated_at"`
}

// ---------- RecipeIonTarget ----------

type CreateIonTargetRequest struct {
	GrowthStageID *uint64  `json:"growth_stage_id"`
	IonCode       string   `json:"ion_code" binding:"required,min=1,max=8"`
	TargetMinMgL  *float64 `json:"target_min_mg_l"`
	TargetMaxMgL  *float64 `json:"target_max_mg_l"`
	Enabled       *bool    `json:"enabled"`
}

type RecipeIonTargetResponse struct {
	ID            uint64   `json:"id"`
	RecipeID      uint64   `json:"recipe_id"`
	GrowthStageID *uint64  `json:"growth_stage_id"`
	IonCode       string   `json:"ion_code"`
	TargetMinMgL  *float64 `json:"target_min_mg_l"`
	TargetMaxMgL  *float64 `json:"target_max_mg_l"`
	Enabled       bool     `json:"enabled"`
	CreatedAt     string   `json:"created_at"`
	UpdatedAt     string   `json:"updated_at"`
}

// ---------- RecipeTargets (combined) ----------

type RecipeTargetsResponse struct {
	Recipe       NutrientRecipeResponse      `json:"recipe"`
	StageTargets []RecipeStageTargetResponse `json:"stage_targets"`
	IonTargets   []RecipeIonTargetResponse   `json:"ion_targets"`
}

type UpdateRecipeTargetsRequest struct {
	StageTargets []CreateStageTargetRequest `json:"stage_targets"`
	IonTargets   []CreateIonTargetRequest   `json:"ion_targets"`
}

// ---------- BatchRecipeBinding ----------

type CreateBindingRequest struct {
	BatchID       uint64  `json:"batch_id" binding:"required"`
	RecipeID      uint64  `json:"recipe_id" binding:"required"`
	BindingType   string  `json:"binding_type"`
	Version       string  `json:"version"`
	EffectiveFrom string  `json:"effective_from" binding:"required"`
	EffectiveTo   *string `json:"effective_to"`
	Status        string  `json:"status"`
}

type UpdateBindingRequest struct {
	BindingType   *string `json:"binding_type" binding:"omitempty,oneof=PRIMARY SECONDARY"`
	Version       *string `json:"version"`
	EffectiveFrom *string `json:"effective_from"`
	EffectiveTo   *string `json:"effective_to"`
	Status        *string `json:"status" binding:"omitempty,oneof=ACTIVE INACTIVE"`
}

type BatchRecipeBindingResponse struct {
	ID            uint64  `json:"id"`
	BatchID       uint64  `json:"batch_id"`
	RecipeID      uint64  `json:"recipe_id"`
	BindingType   string  `json:"binding_type"`
	Version       string  `json:"version"`
	EffectiveFrom string  `json:"effective_from"`
	EffectiveTo   *string `json:"effective_to"`
	Status        string  `json:"status"`
	CreatedBy     *uint64 `json:"created_by"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

// ---------- Common List Response ----------

type RecipeListResponse struct {
	Items    interface{} `json:"items"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}
