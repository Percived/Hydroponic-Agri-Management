package pest

import "time"

// --- Pest/Disease Observation DTOs ---

type CreateObservationRequest struct {
	GreenhouseID       uint64    `json:"greenhouse_id" binding:"required"`
	GrowingZoneID      *uint64   `json:"growing_zone_id"`
	BatchID            *uint64   `json:"batch_id"`
	ObservedAt         time.Time `json:"observed_at" binding:"required"`
	PestOrDisease      string    `json:"pest_or_disease" binding:"required,min=1,max=64"`
	Severity           string    `json:"severity" binding:"required,oneof=LIGHT MODERATE SEVERE"`
	AffectedAreaPct    *float64  `json:"affected_area_pct"`
	AffectedPlantCount *uint     `json:"affected_plant_count"`
	Symptoms           string    `json:"symptoms" binding:"omitempty,max=255"`
	PhotoUrls          string    `json:"photo_urls"`
	ObservedBy         *uint64   `json:"observed_by"`
}

type UpdateObservationRequest struct {
	PestOrDisease      *string  `json:"pest_or_disease"`
	Severity           *string  `json:"severity"`
	AffectedAreaPct    *float64 `json:"affected_area_pct"`
	AffectedPlantCount *uint    `json:"affected_plant_count"`
	Symptoms           *string  `json:"symptoms"`
	PhotoUrls          *string  `json:"photo_urls"`
}

// --- Treatment Record DTOs ---

type CreateTreatmentRequest struct {
	ObservationID        *uint64   `json:"observation_id"`
	GreenhouseID         uint64    `json:"greenhouse_id" binding:"required"`
	GrowingZoneID        *uint64   `json:"growing_zone_id"`
	BatchID              *uint64   `json:"batch_id"`
	TreatmentType        string    `json:"treatment_type" binding:"required,oneof=CHEMICAL BIOLOGICAL PHYSICAL"`
	ProductName          string    `json:"product_name" binding:"required,min=1,max=128"`
	ActiveIngredient     string    `json:"active_ingredient" binding:"omitempty,max=128"`
	Dosage               string    `json:"dosage" binding:"required,min=1,max=64"`
	ApplicationMethod    string    `json:"application_method" binding:"required,oneof=SPRAY DRENCH FOG RELEASE"`
	SafetyIntervalDays   *uint     `json:"safety_interval_days"`
	ReentryIntervalHours *uint     `json:"reentry_interval_hours"`
	TreatedAt            time.Time `json:"treated_at" binding:"required"`
	TreatedBy            *uint64   `json:"treated_by"`
	Note                 string    `json:"note" binding:"omitempty,max=255"`
}

type UpdateTreatmentRequest struct {
	TreatmentType        *string `json:"treatment_type"`
	ProductName          *string `json:"product_name"`
	ActiveIngredient     *string `json:"active_ingredient"`
	Dosage               *string `json:"dosage"`
	ApplicationMethod    *string `json:"application_method"`
	SafetyIntervalDays   *uint   `json:"safety_interval_days"`
	ReentryIntervalHours *uint   `json:"reentry_interval_hours"`
	Note                 *string `json:"note"`
}

// --- Time helpers ---

func timeToStr(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}

func timePtrToStr(t *time.Time) *string {
	if t == nil || t.IsZero() {
		return nil
	}
	s := t.Format(time.RFC3339)
	return &s
}
