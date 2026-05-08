package overview

import "time"

// DashboardResponse is the typed response DTO for the dashboard overview.
type DashboardResponse struct {
	SensorsOnline    int64                `json:"sensors_online"`
	SensorsOffline   int64                `json:"sensors_offline"`
	SensorsTotal     int64                `json:"sensors_total"`
	ActuatorsOnline  int64                `json:"actuators_online"`
	ActuatorsOffline int64                `json:"actuators_offline"`
	ActuatorsTotal   int64                `json:"actuators_total"`
	DevicesOnline    int64                `json:"devices_online"`
	DevicesOffline   int64                `json:"devices_offline"`
	DevicesTotal     int64                `json:"devices_total"`
	AlertsOpen       int64                `json:"alerts_open"`
	AlertsCritical   int64                `json:"alerts_critical"`
	AlertsToday      int64                `json:"alerts_today"`
	DeviceTypeDist   []DeviceTypeDistItem `json:"device_type_distribution"`
	GreenhouseSum    []GreenhouseSummary  `json:"greenhouse_summary"`
	RecentCommands   []RecentCommand      `json:"recent_commands"`
}

type DeviceTypeDistItem struct {
	Type  string `json:"type"`
	Count int64  `json:"count"`
}

type GreenhouseSummary struct {
	GreenhouseID  uint64   `json:"greenhouse_id"`
	Name          string   `json:"name"`
	SensorCount   int64    `json:"sensor_count"`
	ActuatorCount int64    `json:"actuator_count"`
	ZoneCount     int64    `json:"zone_count"`
	AvgTemp       *float64 `json:"avg_temp"`
	AvgHumidity   *float64 `json:"avg_humidity"`
}

type RecentCommand struct {
	ID          uint64 `json:"id"`
	CommandType string `json:"command_type"`
	DeviceName  string `json:"device_name"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
}

// toDashboardResponse assembles the typed dashboard response from raw data.
func toDashboardResponse(
	sensorsOnline, sensorsOffline, totalSensors,
	actuatorsOnline, actuatorsOffline, totalActuators,
	alertsOpen, alertsCritical, alertsToday int64,
	ghSummaries []greenhouseSummary,
	recentCmds []recentCommand,
) DashboardResponse {
	return DashboardResponse{
		SensorsOnline:    sensorsOnline,
		SensorsOffline:   sensorsOffline,
		SensorsTotal:     totalSensors,
		ActuatorsOnline:  actuatorsOnline,
		ActuatorsOffline: actuatorsOffline,
		ActuatorsTotal:   totalActuators,
		DevicesOnline:    sensorsOnline + actuatorsOnline,
		DevicesOffline:   sensorsOffline + actuatorsOffline,
		DevicesTotal:     totalSensors + totalActuators,
		AlertsOpen:       alertsOpen,
		AlertsCritical:   alertsCritical,
		AlertsToday:      alertsToday,
		DeviceTypeDist: []DeviceTypeDistItem{
			{Type: "SENSOR", Count: totalSensors},
			{Type: "ACTUATOR", Count: totalActuators},
		},
		GreenhouseSum:  toGreenhouseSummaries(ghSummaries),
		RecentCommands: toRecentCommands(recentCmds),
	}
}

func toGreenhouseSummaries(src []greenhouseSummary) []GreenhouseSummary {
	out := make([]GreenhouseSummary, 0, len(src))
	for _, s := range src {
		out = append(out, GreenhouseSummary{
			GreenhouseID:  s.GreenhouseID,
			Name:          s.Name,
			SensorCount:   s.SensorCount,
			ActuatorCount: s.ActuatorCount,
			ZoneCount:     s.ZoneCount,
			AvgTemp:       s.AvgTemp,
			AvgHumidity:   s.AvgHumidity,
		})
	}
	return out
}

func toRecentCommands(src []recentCommand) []RecentCommand {
	out := make([]RecentCommand, 0, len(src))
	for _, c := range src {
		out = append(out, RecentCommand{
			ID:          c.ID,
			CommandType: c.CommandType,
			DeviceName:  c.DeviceName,
			Status:      c.Status,
			CreatedAt:   c.CreatedAt.Format(time.RFC3339),
		})
	}
	return out
}
