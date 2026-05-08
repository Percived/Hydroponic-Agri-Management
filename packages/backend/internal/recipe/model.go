package recipe

import "time"

const (
	RecipeStatusDraft    = "DRAFT"
	RecipeStatusActive   = "ACTIVE"
	RecipeStatusArchived = "ARCHIVED"

	BindingTypePrimary   = "PRIMARY"
	BindingTypeSecondary = "SECONDARY"

	BindingStatusActive   = "ACTIVE"
	BindingStatusInactive = "INACTIVE"
)

type NutrientRecipe struct {
	ID            uint64     `gorm:"primaryKey;autoIncrement"`
	RecipeCode    string     `gorm:"column:recipe_code;size:64;uniqueIndex;not null"`
	Name          string     `gorm:"size:128;not null"`
	CropVarietyID *uint64    `gorm:"column:crop_variety_id"`
	Description   string     `gorm:"size:255"`
	Version       string     `gorm:"size:32;default:v1"`
	Status        string     `gorm:"size:16;default:DRAFT"`
	EffectiveFrom *time.Time `gorm:"column:effective_from"`
	EffectiveTo   *time.Time `gorm:"column:effective_to"`
	CreatedBy     *uint64    `gorm:"column:created_by"`
	PublishedBy   *uint64    `gorm:"column:published_by"`
	PublishedAt   *time.Time `gorm:"column:published_at"`
	CreatedAt     time.Time  `gorm:"autoCreateTime:milli"`
	UpdatedAt     time.Time  `gorm:"autoUpdateTime:milli"`
}

func (NutrientRecipe) TableName() string { return "nutrient_recipes" }

type RecipeStageTarget struct {
	ID            uint64    `gorm:"primaryKey;autoIncrement"`
	RecipeID      uint64    `gorm:"column:recipe_id;not null"`
	GrowthStageID *uint64   `gorm:"column:growth_stage_id"`
	MetricCode    string    `gorm:"column:metric_code;size:32;not null"`
	TargetMin     *float64  `gorm:"column:target_min;type:decimal(12,4)"`
	TargetMax     *float64  `gorm:"column:target_max;type:decimal(12,4)"`
	Tolerance     *float64  `gorm:"type:decimal(12,4)"`
	Unit          string    `gorm:"size:16"`
	Enabled       bool      `gorm:"default:true"`
	CreatedAt     time.Time `gorm:"autoCreateTime:milli"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime:milli"`
}

func (RecipeStageTarget) TableName() string { return "recipe_stage_targets" }

type RecipeIonTarget struct {
	ID            uint64    `gorm:"primaryKey;autoIncrement"`
	RecipeID      uint64    `gorm:"column:recipe_id;not null"`
	GrowthStageID *uint64   `gorm:"column:growth_stage_id"`
	IonCode       string    `gorm:"column:ion_code;size:8;not null"`
	TargetMinMgL  *float64  `gorm:"column:target_min_mg_l;type:decimal(10,4)"`
	TargetMaxMgL  *float64  `gorm:"column:target_max_mg_l;type:decimal(10,4)"`
	Enabled       bool      `gorm:"default:true"`
	CreatedAt     time.Time `gorm:"autoCreateTime:milli"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime:milli"`
}

func (RecipeIonTarget) TableName() string { return "recipe_ion_targets" }

type BatchRecipeBinding struct {
	ID            uint64     `gorm:"primaryKey;autoIncrement"`
	BatchID       uint64     `gorm:"column:batch_id;not null"`
	RecipeID      uint64     `gorm:"column:recipe_id;not null"`
	BindingType   string     `gorm:"column:binding_type;size:16;default:PRIMARY"`
	Version       string     `gorm:"size:32;default:v1"`
	EffectiveFrom time.Time  `gorm:"column:effective_from;not null"`
	EffectiveTo   *time.Time `gorm:"column:effective_to"`
	Status        string     `gorm:"size:16;default:ACTIVE"`
	CreatedBy     *uint64    `gorm:"column:created_by"`
	CreatedAt     time.Time  `gorm:"autoCreateTime:milli"`
	UpdatedAt     time.Time  `gorm:"autoUpdateTime:milli"`
}

func (BatchRecipeBinding) TableName() string { return "batch_recipe_bindings" }
