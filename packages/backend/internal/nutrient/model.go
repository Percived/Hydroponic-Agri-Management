package nutrient

import "time"

const (
	TankStatusActive   = "ACTIVE"
	TankStatusInactive = "INACTIVE"
	TankStatusEmpty    = "EMPTY"

	ChangeTypeFullReplace    = "FULL_REPLACE"
	ChangeTypePartialRefresh = "PARTIAL_REFRESH"
	ChangeTypeTopUp          = "TOP_UP"

	TestMethodLab   = "LAB"
	TestMethodStrip = "STRIP"
	TestMethodMeter = "METER"

	ConcentrateTypeA      = "A"
	ConcentrateTypeB      = "B"
	ConcentrateTypeAcid   = "ACID"
	ConcentrateTypeAlkali = "ALKALI"

	InventoryStatusInUse   = "IN_USE"
	InventoryStatusEmpty   = "EMPTY"
	InventoryStatusExpired = "EXPIRED"
)

type NutrientTank struct {
	ID                   uint64    `gorm:"primaryKey;autoIncrement"`
	GrowingZoneID        uint64    `gorm:"column:growing_zone_id;not null"`
	Code                 string    `gorm:"size:32;not null"`
	TotalVolumeLiter     float64   `gorm:"column:total_volume_liter;type:decimal(10,2);not null"`
	CurrentVolumeLiter   *float64  `gorm:"column:current_volume_liter;type:decimal(10,2)"`
	Status               string    `gorm:"size:16;default:ACTIVE"`
	ECSensorChannelID    *uint64   `gorm:"column:ec_sensor_channel_id"`
	PHSensorChannelID    *uint64   `gorm:"column:ph_sensor_channel_id"`
	LevelSensorChannelID *uint64   `gorm:"column:level_sensor_channel_id"`
	TempSensorChannelID  *uint64   `gorm:"column:temp_sensor_channel_id"`
	CreatedAt            time.Time `gorm:"autoCreateTime:milli"`
	UpdatedAt            time.Time `gorm:"autoUpdateTime:milli"`
}

func (NutrientTank) TableName() string { return "nutrient_tanks" }

type SolutionChangeEvent struct {
	ID                  uint64    `gorm:"primaryKey;autoIncrement"`
	TankID              uint64    `gorm:"column:tank_id;not null"`
	ChangeType          string    `gorm:"column:change_type;size:16;not null"`
	VolumeReplacedLiter float64   `gorm:"column:volume_replaced_liter;type:decimal(10,2);not null"`
	SourceWaterEC       *float64  `gorm:"column:source_water_ec;type:decimal(12,4)"`
	SourceWaterPH       *float64  `gorm:"column:source_water_ph;type:decimal(12,4)"`
	BeforeEC            *float64  `gorm:"column:before_ec;type:decimal(12,4)"`
	BeforePH            *float64  `gorm:"column:before_ph;type:decimal(12,4)"`
	AfterEC             *float64  `gorm:"column:after_ec;type:decimal(12,4)"`
	AfterPH             *float64  `gorm:"column:after_ph;type:decimal(12,4)"`
	NutrientAAddedMl    *float64  `gorm:"column:nutrient_a_added_ml;type:decimal(10,2)"`
	NutrientBAddedMl    *float64  `gorm:"column:nutrient_b_added_ml;type:decimal(10,2)"`
	AcidAddedMl         *float64  `gorm:"column:acid_added_ml;type:decimal(10,2)"`
	AlkaliAddedMl       *float64  `gorm:"column:alkali_added_ml;type:decimal(10,2)"`
	Note                string    `gorm:"size:255"`
	OperatedBy          *uint64   `gorm:"column:operated_by"`
	OperatedAt          time.Time `gorm:"column:operated_at;not null"`
	CreatedAt           time.Time `gorm:"autoCreateTime:milli"`
}

func (SolutionChangeEvent) TableName() string { return "solution_change_events" }

type IonTestRecord struct {
	ID         uint64     `gorm:"primaryKey;autoIncrement"`
	TankID     uint64     `gorm:"column:tank_id;not null"`
	BatchID    *uint64    `gorm:"column:batch_id"`
	SampleCode string     `gorm:"column:sample_code;size:64;not null;uniqueIndex"`
	SampledAt  time.Time  `gorm:"column:sampled_at;not null"`
	TestedAt   *time.Time `gorm:"column:tested_at"`
	TestMethod string     `gorm:"column:test_method;size:16;default:LAB"`
	NO3N       *float64   `gorm:"column:no3_n;type:decimal(10,2)"`
	NH4N       *float64   `gorm:"column:nh4_n;type:decimal(10,2)"`
	P          *float64   `gorm:"type:decimal(10,2)"`
	K          *float64   `gorm:"type:decimal(10,2)"`
	Ca         *float64   `gorm:"type:decimal(10,2)"`
	Mg         *float64   `gorm:"type:decimal(10,2)"`
	S          *float64   `gorm:"type:decimal(10,2)"`
	Fe         *float64   `gorm:"type:decimal(10,4)"`
	Mn         *float64   `gorm:"type:decimal(10,4)"`
	Zn         *float64   `gorm:"type:decimal(10,4)"`
	B          *float64   `gorm:"type:decimal(10,4)"`
	Cu         *float64   `gorm:"type:decimal(10,4)"`
	Mo         *float64   `gorm:"type:decimal(10,4)"`
	ECAtSample *float64   `gorm:"column:ec_at_sample;type:decimal(12,4)"`
	PHAtSample *float64   `gorm:"column:ph_at_sample;type:decimal(12,4)"`
	LabName    string     `gorm:"column:lab_name;size:64"`
	ReportURL  string     `gorm:"column:report_url;size:255"`
	Note       string     `gorm:"size:255"`
	CreatedBy  *uint64    `gorm:"column:created_by"`
	CreatedAt  time.Time  `gorm:"autoCreateTime:milli"`
}

func (IonTestRecord) TableName() string { return "ion_test_records" }

type NutrientConcentrateInventory struct {
	ID                uint64     `gorm:"primaryKey;autoIncrement"`
	GreenhouseID      uint64     `gorm:"column:greenhouse_id;not null"`
	ConcentrateType   string     `gorm:"column:concentrate_type;size:8;not null"`
	Brand             string     `gorm:"size:64"`
	ProductName       string     `gorm:"column:product_name;size:128"`
	TotalVolumeMl     float64    `gorm:"column:total_volume_ml;type:decimal(12,2);not null"`
	RemainingVolumeMl float64    `gorm:"column:remaining_volume_ml;type:decimal(12,2);default:0"`
	UnitPrice         *float64   `gorm:"column:unit_price;type:decimal(10,2)"`
	BatchNo           string     `gorm:"column:batch_no;size:64"`
	ExpiredAt         *time.Time `gorm:"column:expired_at;type:date"`
	Status            string     `gorm:"size:16;default:IN_USE"`
	CreatedAt         time.Time  `gorm:"autoCreateTime:milli"`
	UpdatedAt         time.Time  `gorm:"autoUpdateTime:milli"`
}

func (NutrientConcentrateInventory) TableName() string { return "nutrient_concentrate_inventory" }

type ConcentrateUsageLog struct {
	ID               uint64    `gorm:"primaryKey;autoIncrement"`
	InventoryID      uint64    `gorm:"column:inventory_id;not null"`
	SolutionChangeID *uint64   `gorm:"column:solution_change_id"`
	TankID           *uint64   `gorm:"column:tank_id"`
	VolumeUsedMl     float64   `gorm:"column:volume_used_ml;type:decimal(10,2);not null"`
	UsedBy           *uint64   `gorm:"column:used_by"`
	UsedAt           time.Time `gorm:"column:used_at;not null"`
	CreatedAt        time.Time `gorm:"autoCreateTime:milli"`
}

func (ConcentrateUsageLog) TableName() string { return "concentrate_usage_logs" }
