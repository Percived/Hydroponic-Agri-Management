package greenhouse

// CreateGreenhouseRequest ...
type CreateGreenhouseRequest struct {
	Code        string  `json:"code" binding:"required,max=32"`
	Name        string  `json:"name" binding:"required,max=64"`
	Location    string  `json:"location" binding:"max=128"`
	AreaSqm     float64 `json:"area_sqm"`
	Description string  `json:"description" binding:"max=255"`
}

// UpdateGreenhouseRequest ...
type UpdateGreenhouseRequest struct {
	Name        string  `json:"name" binding:"max=64"`
	Location    string  `json:"location" binding:"max=128"`
	AreaSqm     float64 `json:"area_sqm"`
	Description string  `json:"description" binding:"max=255"`
	Status      string  `json:"status"`
}

// CreateGrowingZoneRequest ...
type CreateGrowingZoneRequest struct {
	GreenhouseID          uint64  `json:"greenhouse_id" binding:"required"`
	Code                  string  `json:"code" binding:"required,max=32"`
	Name                  string  `json:"name" binding:"required,max=64"`
	SystemType            string  `json:"system_type"`
	TankVolumeLiter       float64 `json:"tank_volume_liter"`
	PlantingDensityPerSqm float64 `json:"planting_density_per_sqm"`
}

// UpdateGrowingZoneRequest ...
type UpdateGrowingZoneRequest struct {
	Name                  string  `json:"name" binding:"max=64"`
	SystemType            string  `json:"system_type"`
	TankVolumeLiter       float64 `json:"tank_volume_liter"`
	PlantingDensityPerSqm float64 `json:"planting_density_per_sqm"`
	Status                string  `json:"status"`
}

// GreenhouseResponse ...
type GreenhouseResponse struct {
	ID          uint64  `json:"id"`
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	Location    string  `json:"location"`
	AreaSqm     float64 `json:"area_sqm"`
	Description string  `json:"description"`
	Status      string  `json:"status"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
	ZoneCount   int64   `json:"zone_count,omitempty"`
}

// GrowingZoneResponse ...
type GrowingZoneResponse struct {
	ID                    uint64  `json:"id"`
	GreenhouseID          uint64  `json:"greenhouse_id"`
	Code                  string  `json:"code"`
	Name                  string  `json:"name"`
	SystemType            string  `json:"system_type"`
	TankVolumeLiter       float64 `json:"tank_volume_liter"`
	PlantingDensityPerSqm float64 `json:"planting_density_per_sqm"`
	Status                string  `json:"status"`
	CreatedAt             string  `json:"created_at"`
	UpdatedAt             string  `json:"updated_at"`
}

// GreenhouseListResponse ...
type GreenhouseListResponse struct {
	Items    []GreenhouseResponse `json:"items"`
	Total    int64                `json:"total"`
	Page     int                  `json:"page"`
	PageSize int                  `json:"page_size"`
}

// GrowingZoneListResponse ...
type GrowingZoneListResponse struct {
	Items    []GrowingZoneResponse `json:"items"`
	Total    int64                 `json:"total"`
	Page     int                   `json:"page"`
	PageSize int                   `json:"page_size"`
}
