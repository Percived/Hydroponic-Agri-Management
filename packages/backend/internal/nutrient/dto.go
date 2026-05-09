package nutrient

// ---------- NutrientTank ----------

type CreateNutrientTankRequest struct {
	GrowingZoneID        uint64   `json:"growing_zone_id" binding:"required"`
	Code                 string   `json:"code" binding:"required,min=1,max=32"`
	TotalVolumeLiter     float64  `json:"total_volume_liter" binding:"required,gt=0"`
	CurrentVolumeLiter   *float64 `json:"current_volume_liter"`
	Status               string   `json:"status"`
	ECSensorChannelID    *uint64  `json:"ec_sensor_channel_id"`
	PHSensorChannelID    *uint64  `json:"ph_sensor_channel_id"`
	LevelSensorChannelID *uint64  `json:"level_sensor_channel_id"`
	TempSensorChannelID  *uint64  `json:"temp_sensor_channel_id"`
}

type UpdateNutrientTankRequest struct {
	Code                 *string  `json:"code" binding:"omitempty,min=1,max=32"`
	TotalVolumeLiter     *float64 `json:"total_volume_liter" binding:"omitempty,gt=0"`
	CurrentVolumeLiter   *float64 `json:"current_volume_liter"`
	Status               *string  `json:"status" binding:"omitempty,oneof=ACTIVE INACTIVE EMPTY"`
	ECSensorChannelID    *uint64  `json:"ec_sensor_channel_id"`
	PHSensorChannelID    *uint64  `json:"ph_sensor_channel_id"`
	LevelSensorChannelID *uint64  `json:"level_sensor_channel_id"`
	TempSensorChannelID  *uint64  `json:"temp_sensor_channel_id"`
}

type NutrientTankResponse struct {
	ID                   uint64   `json:"id"`
	GrowingZoneID        uint64   `json:"growing_zone_id"`
	Code                 string   `json:"code"`
	TotalVolumeLiter     float64  `json:"total_volume_liter"`
	CurrentVolumeLiter   *float64 `json:"current_volume_liter"`
	Status               string   `json:"status"`
	ECSensorChannelID    *uint64  `json:"ec_sensor_channel_id,omitempty"`
	PHSensorChannelID    *uint64  `json:"ph_sensor_channel_id,omitempty"`
	LevelSensorChannelID *uint64  `json:"level_sensor_channel_id,omitempty"`
	TempSensorChannelID  *uint64  `json:"temp_sensor_channel_id,omitempty"`
	CreatedAt            string   `json:"created_at"`
	UpdatedAt            string   `json:"updated_at"`
}

// ---------- SolutionChangeEvent ----------

type CreateSolutionChangeEventRequest struct {
	TankID              uint64   `json:"tank_id" binding:"required"`
	ChangeType          string   `json:"change_type" binding:"required,oneof=FULL_REPLACE PARTIAL_REFRESH TOP_UP"`
	VolumeReplacedLiter float64  `json:"volume_replaced_liter" binding:"required,gt=0"`
	SourceWaterEC       *float64 `json:"source_water_ec"`
	SourceWaterPH       *float64 `json:"source_water_ph"`
	BeforeEC            *float64 `json:"before_ec"`
	BeforePH            *float64 `json:"before_ph"`
	AfterEC             *float64 `json:"after_ec"`
	AfterPH             *float64 `json:"after_ph"`
	NutrientAAddedMl    *float64 `json:"nutrient_a_added_ml"`
	NutrientBAddedMl    *float64 `json:"nutrient_b_added_ml"`
	AcidAddedMl         *float64 `json:"acid_added_ml"`
	AlkaliAddedMl       *float64 `json:"alkali_added_ml"`
	Note                string   `json:"note"`
	OperatedAt          string   `json:"operated_at" binding:"required"`
}

type SolutionChangeEventResponse struct {
	ID                  uint64   `json:"id"`
	TankID              uint64   `json:"tank_id"`
	ChangeType          string   `json:"change_type"`
	VolumeReplacedLiter float64  `json:"volume_replaced_liter"`
	SourceWaterEC       *float64 `json:"source_water_ec"`
	SourceWaterPH       *float64 `json:"source_water_ph"`
	BeforeEC            *float64 `json:"before_ec"`
	BeforePH            *float64 `json:"before_ph"`
	AfterEC             *float64 `json:"after_ec"`
	AfterPH             *float64 `json:"after_ph"`
	NutrientAAddedMl    *float64 `json:"nutrient_a_added_ml"`
	NutrientBAddedMl    *float64 `json:"nutrient_b_added_ml"`
	AcidAddedMl         *float64 `json:"acid_added_ml"`
	AlkaliAddedMl       *float64 `json:"alkali_added_ml"`
	Note                string   `json:"note"`
	OperatedBy          *uint64  `json:"operated_by"`
	OperatedAt          string   `json:"operated_at"`
	CreatedAt           string   `json:"created_at"`
}

// ---------- IonTestRecord ----------

type CreateIonTestRecordRequest struct {
	TankID     uint64   `json:"tank_id" binding:"required"`
	BatchID    *uint64  `json:"batch_id"`
	SampleCode string   `json:"sample_code" binding:"required,min=1,max=64"`
	SampledAt  string   `json:"sampled_at" binding:"required"`
	TestedAt   *string  `json:"tested_at"`
	TestMethod string   `json:"test_method"`
	NO3N       *float64 `json:"no3_n"`
	NH4N       *float64 `json:"nh4_n"`
	P          *float64 `json:"p"`
	K          *float64 `json:"k"`
	Ca         *float64 `json:"ca"`
	Mg         *float64 `json:"mg"`
	S          *float64 `json:"s"`
	Fe         *float64 `json:"fe"`
	Mn         *float64 `json:"mn"`
	Zn         *float64 `json:"zn"`
	B          *float64 `json:"b"`
	Cu         *float64 `json:"cu"`
	Mo         *float64 `json:"mo"`
	ECAtSample *float64 `json:"ec_at_sample"`
	PHAtSample *float64 `json:"ph_at_sample"`
	LabName    string   `json:"lab_name"`
	ReportURL  string   `json:"report_url"`
	Note       string   `json:"note"`
}

type UpdateIonTestRecordRequest struct {
	SampleCode *string  `json:"sample_code" binding:"omitempty,min=1,max=64"`
	TestedAt   *string  `json:"tested_at"`
	TestMethod *string  `json:"test_method" binding:"omitempty,oneof=LAB STRIP METER"`
	NO3N       *float64 `json:"no3_n"`
	NH4N       *float64 `json:"nh4_n"`
	P          *float64 `json:"p"`
	K          *float64 `json:"k"`
	Ca         *float64 `json:"ca"`
	Mg         *float64 `json:"mg"`
	S          *float64 `json:"s"`
	Fe         *float64 `json:"fe"`
	Mn         *float64 `json:"mn"`
	Zn         *float64 `json:"zn"`
	B          *float64 `json:"b"`
	Cu         *float64 `json:"cu"`
	Mo         *float64 `json:"mo"`
	ECAtSample *float64 `json:"ec_at_sample"`
	PHAtSample *float64 `json:"ph_at_sample"`
	LabName    *string  `json:"lab_name"`
	ReportURL  *string  `json:"report_url"`
	Note       *string  `json:"note"`
}

type IonTestRecordResponse struct {
	ID         uint64   `json:"id"`
	TankID     uint64   `json:"tank_id"`
	BatchID    *uint64  `json:"batch_id"`
	SampleCode string   `json:"sample_code"`
	SampledAt  string   `json:"sampled_at"`
	TestedAt   *string  `json:"tested_at"`
	TestMethod string   `json:"test_method"`
	NO3N       *float64 `json:"no3_n"`
	NH4N       *float64 `json:"nh4_n"`
	P          *float64 `json:"p"`
	K          *float64 `json:"k"`
	Ca         *float64 `json:"ca"`
	Mg         *float64 `json:"mg"`
	S          *float64 `json:"s"`
	Fe         *float64 `json:"fe"`
	Mn         *float64 `json:"mn"`
	Zn         *float64 `json:"zn"`
	B          *float64 `json:"b"`
	Cu         *float64 `json:"cu"`
	Mo         *float64 `json:"mo"`
	ECAtSample *float64 `json:"ec_at_sample"`
	PHAtSample *float64 `json:"ph_at_sample"`
	LabName    string   `json:"lab_name"`
	ReportURL  string   `json:"report_url"`
	Note       string   `json:"note"`
	CreatedBy  *uint64  `json:"created_by"`
	CreatedAt  string   `json:"created_at"`
}

// ---------- NutrientConcentrateInventory ----------

type CreateConcentrateInventoryRequest struct {
	GreenhouseID      uint64   `json:"greenhouse_id" binding:"required"`
	ConcentrateType   string   `json:"concentrate_type" binding:"required,oneof=A B ACID ALKALI"`
	Brand             string   `json:"brand"`
	ProductName       string   `json:"product_name"`
	TotalVolumeMl     float64  `json:"total_volume_ml" binding:"required,gt=0"`
	RemainingVolumeMl float64  `json:"remaining_volume_ml"`
	UnitPrice         *float64 `json:"unit_price"`
	BatchNo           string   `json:"batch_no"`
	ExpiredAt         *string  `json:"expired_at"`
	Status            string   `json:"status"`
}

type UpdateConcentrateInventoryRequest struct {
	ConcentrateType   *string  `json:"concentrate_type" binding:"omitempty,oneof=A B ACID ALKALI"`
	Brand             *string  `json:"brand"`
	ProductName       *string  `json:"product_name"`
	TotalVolumeMl     *float64 `json:"total_volume_ml" binding:"omitempty,gt=0"`
	RemainingVolumeMl *float64 `json:"remaining_volume_ml"`
	UnitPrice         *float64 `json:"unit_price"`
	BatchNo           *string  `json:"batch_no"`
	ExpiredAt         *string  `json:"expired_at"`
	Status            *string  `json:"status" binding:"omitempty,oneof=IN_USE EMPTY EXPIRED"`
}

type ConcentrateInventoryResponse struct {
	ID                uint64   `json:"id"`
	GreenhouseID      uint64   `json:"greenhouse_id"`
	ConcentrateType   string   `json:"concentrate_type"`
	Brand             string   `json:"brand"`
	ProductName       string   `json:"product_name"`
	TotalVolumeMl     float64  `json:"total_volume_ml"`
	RemainingVolumeMl float64  `json:"remaining_volume_ml"`
	UnitPrice         *float64 `json:"unit_price"`
	BatchNo           string   `json:"batch_no"`
	ExpiredAt         *string  `json:"expired_at"`
	Status            string   `json:"status"`
	CreatedAt         string   `json:"created_at"`
	UpdatedAt         string   `json:"updated_at"`
}

// ---------- ConcentrateUsageLog ----------

type CreateConcentrateUsageLogRequest struct {
	InventoryID      uint64  `json:"inventory_id" binding:"required"`
	SolutionChangeID *uint64 `json:"solution_change_id"`
	TankID           *uint64 `json:"tank_id"`
	VolumeUsedMl     float64 `json:"volume_used_ml" binding:"required,gt=0"`
	UsedAt           string  `json:"used_at" binding:"required"`
}

type ConcentrateUsageLogResponse struct {
	ID               uint64  `json:"id"`
	InventoryID      uint64  `json:"inventory_id"`
	SolutionChangeID *uint64 `json:"solution_change_id"`
	TankID           *uint64 `json:"tank_id"`
	VolumeUsedMl     float64 `json:"volume_used_ml"`
	UsedBy           *uint64 `json:"used_by"`
	UsedAt           string  `json:"used_at"`
	CreatedAt        string  `json:"created_at"`
}

// ---------- Common List Response ----------

type NutrientListResponse struct {
	Items    interface{} `json:"items"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}
