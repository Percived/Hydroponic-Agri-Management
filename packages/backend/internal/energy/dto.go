package energy

import "time"

// --- Request DTOs ---

type CreateEnergyRecordRequest struct {
	GreenhouseID      uint64    `json:"greenhouse_id" binding:"required"`
	RecordType        string    `json:"record_type" binding:"required,oneof=ELECTRICITY WATER CO2_GAS"`
	ConsumptionValue  float64   `json:"consumption_value" binding:"required"`
	Unit              string    `json:"unit" binding:"required,min=1,max=16"`
	RecordPeriodStart time.Time `json:"record_period_start" binding:"required"`
	RecordPeriodEnd   time.Time `json:"record_period_end" binding:"required"`
	MeterReadingStart *float64  `json:"meter_reading_start"`
	MeterReadingEnd   *float64  `json:"meter_reading_end"`
	BatchID           *uint64   `json:"batch_id"`
	RecordedBy        *uint64   `json:"recorded_by"`
}

type UpdateEnergyRecordRequest struct {
	ConsumptionValue  *float64 `json:"consumption_value"`
	Unit              *string  `json:"unit"`
	MeterReadingStart *float64 `json:"meter_reading_start"`
	MeterReadingEnd   *float64 `json:"meter_reading_end"`
	BatchID           *uint64  `json:"batch_id"`
}

// --- Energy Summary Response ---

type EnergySummaryItem struct {
	RecordType string  `json:"record_type"`
	TotalValue float64 `json:"total_value"`
	Unit       string  `json:"unit"`
}

// --- Time format helpers ---

func timeToStr(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}
