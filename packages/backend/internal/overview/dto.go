package overview

import "time"

// DashboardStats provides key high-level statistics
type DashboardStats struct {
	ActiveBatchesCount int     `json:"active_batches_count"`
	UnresolvedAlerts   int     `json:"unresolved_alerts"`
	DevicesOnline      int     `json:"devices_online"`
	DevicesOffline     int     `json:"devices_offline"`
	EnergyKwhToday     float64 `json:"energy_kwh_today"`
	WaterLToday        float64 `json:"water_l_today"`
}

// GreenhouseMetrics contains the latest average metrics for a greenhouse
type GreenhouseMetrics struct {
	Temperature float64 `json:"temperature"`
	Humidity    float64 `json:"humidity"`
	EC          float64 `json:"ec"`
	PH          float64 `json:"ph"`
	DO          float64 `json:"do"`
	CO2         float64 `json:"co2"`
	Lux         float64 `json:"lux"`
}

// DashboardGreenhouse provides a snapshot of a greenhouse's current state
type DashboardGreenhouse struct {
	ID               string            `json:"id"`
	Name             string            `json:"name"`
	HealthScore      string            `json:"health_score"`
	LastCollectedAt  *time.Time        `json:"last_collected_at"`
	Metrics          GreenhouseMetrics `json:"metrics"`
	ActiveStrategies []string          `json:"active_strategies"`
}

// DashboardTrends provides 24h trend data
type DashboardTrends struct {
	Timestamps []string  `json:"timestamps"`
	ECAvg      []float64 `json:"ec_avg"`
	PHAvg      []float64 `json:"ph_avg"`
}

// DashboardActiveBatch provides info about an active crop batch
type DashboardActiveBatch struct {
	BatchID      string `json:"batch_id"`
	CropName     string `json:"crop_name"`
	Stage        string `json:"stage"`
	Day          int    `json:"day"`
	GreenhouseID string `json:"greenhouse_id"`
}

// DashboardRecentAlert provides info about a recent critical alert
type DashboardRecentAlert struct {
	AlertID        string    `json:"alert_id"`
	Severity       string    `json:"severity"`
	Message        string    `json:"message"`
	Timestamp      time.Time `json:"timestamp"`
	GreenhouseName string    `json:"greenhouse_name"`
}

// RecentCommand provides info about recently dispatched commands
type RecentCommand struct {
	ID          uint64 `json:"id"`
	CommandType string `json:"command_type"`
	DeviceName  string `json:"device_name"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
}

// DashboardResponse is the typed response DTO for the dashboard overview.
type DashboardResponse struct {
	Stats          DashboardStats         `json:"stats"`
	Greenhouses    []DashboardGreenhouse  `json:"greenhouses"`
	Trends         DashboardTrends        `json:"trends"`
	ActiveBatches  []DashboardActiveBatch `json:"active_batches"`
	RecentAlerts   []DashboardRecentAlert `json:"recent_alerts"`
	RecentCommands []RecentCommand        `json:"recent_commands"`
}
