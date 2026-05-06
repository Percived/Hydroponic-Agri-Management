package greenhouse

import "time"

const (
	StatusEnabled  = "ENABLED"
	StatusDisabled = "DISABLED"
	SystemTypeDWC  = "DWC"
	SystemTypeNFT  = "NFT"
)

type Greenhouse struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement"`
	Code        string    `gorm:"size:32;uniqueIndex;not null"`
	Name        string    `gorm:"size:64;not null"`
	Location    string    `gorm:"size:128"`
	AreaSqm     float64   `gorm:"column:area_sqm;type:decimal(10,2)"`
	Description string    `gorm:"size:255"`
	Status      string    `gorm:"size:16;default:ENABLED"`
	CreatedAt   time.Time `gorm:"autoCreateTime:milli"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime:milli"`
}

func (Greenhouse) TableName() string { return "greenhouses" }

type GrowingZone struct {
	ID                    uint64    `gorm:"primaryKey;autoIncrement"`
	GreenhouseID          uint64    `gorm:"not null"`
	Code                  string    `gorm:"size:32;not null"`
	Name                  string    `gorm:"size:64;not null"`
	SystemType            string    `gorm:"size:16;default:DWC"`
	TankVolumeLiter       float64   `gorm:"column:tank_volume_liter;type:decimal(10,2)"`
	PlantingDensityPerSqm float64   `gorm:"column:planting_density_per_sqm;type:decimal(8,2)"`
	Status                string    `gorm:"size:16;default:ENABLED"`
	CreatedAt             time.Time `gorm:"autoCreateTime:milli"`
	UpdatedAt             time.Time `gorm:"autoUpdateTime:milli"`
}

func (GrowingZone) TableName() string { return "growing_zones" }
