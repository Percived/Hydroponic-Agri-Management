package energy

import "time"

const (
	TypeElectricity = "ELECTRICITY"
	TypeWater       = "WATER"
	TypeCO2Gas      = "CO2_GAS"
)

type EnergyConsumptionRecord struct {
	ID                uint64    `gorm:"primaryKey;autoIncrement"`
	GreenhouseID      uint64    `gorm:"column:greenhouse_id;not null"`
	RecordType        string    `gorm:"column:record_type;size:16;not null"`
	ConsumptionValue  float64   `gorm:"column:consumption_value;type:decimal(12,4);not null"`
	Unit              string    `gorm:"size:16;not null"`
	RecordPeriodStart time.Time `gorm:"column:record_period_start;not null"`
	RecordPeriodEnd   time.Time `gorm:"column:record_period_end;not null"`
	MeterReadingStart *float64  `gorm:"column:meter_reading_start;type:decimal(12,4)"`
	MeterReadingEnd   *float64  `gorm:"column:meter_reading_end;type:decimal(12,4)"`
	BatchID           *uint64   `gorm:"column:batch_id"`
	RecordedBy        *uint64   `gorm:"column:recorded_by"`
	CreatedAt         time.Time `gorm:"autoCreateTime:milli"`
}

func (EnergyConsumptionRecord) TableName() string { return "energy_consumption_records" }
