package pest

import "time"

const (
	SeverityLight    = "LIGHT"
	SeverityModerate = "MODERATE"
	SeveritySevere   = "SEVERE"

	TreatmentChemical   = "CHEMICAL"
	TreatmentBiological = "BIOLOGICAL"
	TreatmentPhysical   = "PHYSICAL"

	ApplicationSpray   = "SPRAY"
	ApplicationDrench  = "DRENCH"
	ApplicationFog     = "FOG"
	ApplicationRelease = "RELEASE"
)

type PestDiseaseObservation struct {
	ID                 uint64    `gorm:"primaryKey;autoIncrement"`
	GreenhouseID       uint64    `gorm:"column:greenhouse_id;not null"`
	GrowingZoneID      *uint64   `gorm:"column:growing_zone_id"`
	BatchID            *uint64   `gorm:"column:batch_id"`
	ObservedAt         time.Time `gorm:"column:observed_at;not null"`
	PestOrDisease      string    `gorm:"column:pest_or_disease;size:64;not null"`
	Severity           string    `gorm:"size:16;not null"`
	AffectedAreaPct    *float64  `gorm:"column:affected_area_pct;type:decimal(5,2)"`
	AffectedPlantCount *uint     `gorm:"column:affected_plant_count"`
	Symptoms           string    `gorm:"size:255"`
	PhotoUrls          string    `gorm:"column:photo_urls;type:json"`
	ObservedBy         *uint64   `gorm:"column:observed_by"`
	CreatedAt          time.Time `gorm:"autoCreateTime:milli"`
}

func (PestDiseaseObservation) TableName() string { return "pest_disease_observations" }

type TreatmentRecord struct {
	ID                   uint64    `gorm:"primaryKey;autoIncrement"`
	ObservationID        *uint64   `gorm:"column:observation_id"`
	GreenhouseID         uint64    `gorm:"column:greenhouse_id;not null"`
	GrowingZoneID        *uint64   `gorm:"column:growing_zone_id"`
	BatchID              *uint64   `gorm:"column:batch_id"`
	TreatmentType        string    `gorm:"column:treatment_type;size:16;not null"`
	ProductName          string    `gorm:"column:product_name;size:128;not null"`
	ActiveIngredient     string    `gorm:"column:active_ingredient;size:128"`
	Dosage               string    `gorm:"size:64;not null"`
	ApplicationMethod    string    `gorm:"column:application_method;size:32;not null"`
	SafetyIntervalDays   *uint     `gorm:"column:safety_interval_days"`
	ReentryIntervalHours *uint     `gorm:"column:reentry_interval_hours"`
	TreatedAt            time.Time `gorm:"column:treated_at;not null"`
	TreatedBy            *uint64   `gorm:"column:treated_by"`
	Note                 string    `gorm:"size:255"`
	CreatedAt            time.Time `gorm:"autoCreateTime:milli"`
}

func (TreatmentRecord) TableName() string { return "treatment_records" }
