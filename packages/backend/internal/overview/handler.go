package overview

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	alertpkg "hydroponic-backend/internal/alert"
	platformErrors "hydroponic-backend/internal/platform/errors"
	"hydroponic-backend/internal/platform/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	db *gorm.DB
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{db: db}
}

func (h *Handler) Dashboard(c *gin.Context) {
	var (
		errCh              = make(chan error, 10)
		stats              DashboardStats
		greenhouses        []DashboardGreenhouse
		activeBatches      []DashboardActiveBatch
		recentAlerts       []DashboardRecentAlert
		recentCmds         []RecentCommand
		sensorsOnline      int64
		actuatorsOnline    int64
		totalSensors       int64
		totalActuators     int64
		activeBatchesCount int64
		alertsOpen         int64
	)

	var wg sync.WaitGroup

	// Devices count
	wg.Add(1)
	go func() {
		defer wg.Done()
		if e := h.db.Table("sensor_devices").Where("status = ?", "ONLINE").Count(&sensorsOnline).Error; e != nil {
			errCh <- e
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		if e := h.db.Table("sensor_devices").Count(&totalSensors).Error; e != nil {
			errCh <- e
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		if e := h.db.Table("actuator_devices").Where("status = ?", "ONLINE").Count(&actuatorsOnline).Error; e != nil {
			errCh <- e
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		if e := h.db.Table("actuator_devices").Count(&totalActuators).Error; e != nil {
			errCh <- e
		}
	}()

	// Alerts open count
	wg.Add(1)
	go func() {
		defer wg.Done()
		if e := h.db.Table("alerts").Where("status = ?", alertpkg.StatusOpen).Count(&alertsOpen).Error; e != nil {
			errCh <- e
		}
	}()

	// Active batches count
	wg.Add(1)
	go func() {
		defer wg.Done()
		if h.db.Migrator().HasTable("crop_batches") {
			h.db.Table("crop_batches").Where("status = ?", "ACTIVE").Count(&activeBatchesCount)
		}
	}()

	// Greenhouse metrics
	wg.Add(1)
	go func() {
		defer wg.Done()
		if e := queryGreenhouseMetrics(h.db, &greenhouses); e != nil {
			errCh <- e
		}
	}()

	// Active batches list
	wg.Add(1)
	go func() {
		defer wg.Done()
		if h.db.Migrator().HasTable("crop_batches") {
			h.db.Raw(`
				SELECT 
					CAST(cb.id AS CHAR) as batch_id, 
					c.name as crop_name, 
					'GROWING' as stage, 
					DATEDIFF(NOW(), cb.start_date) as day, 
					CAST(cb.greenhouse_id AS CHAR) as greenhouse_id
				FROM crop_batches cb
				JOIN crops c ON c.id = cb.crop_id
				WHERE cb.status = 'ACTIVE'
				LIMIT 5
			`).Scan(&activeBatches)
		}
	}()

	// Recent alerts
	wg.Add(1)
	go func() {
		defer wg.Done()
		h.db.Raw(`
			SELECT 
				CAST(a.id AS CHAR) as alert_id, 
				a.level as severity, 
				a.message, 
				a.triggered_at as timestamp, 
				COALESCE(g.name, 'System') as greenhouse_name
			FROM alerts a
			LEFT JOIN sensor_devices sd ON sd.id = a.device_id AND a.device_type = 'SENSOR'
			LEFT JOIN actuator_devices ad ON ad.id = a.device_id AND a.device_type = 'ACTUATOR'
			LEFT JOIN greenhouses g ON g.id = sd.greenhouse_id OR g.id = ad.greenhouse_id
			WHERE a.status = 'OPEN'
			ORDER BY a.triggered_at DESC
			LIMIT 5
		`).Scan(&recentAlerts)
	}()

	// Recent commands
	wg.Add(1)
	go func() {
		defer wg.Done()
		if e := h.db.Raw(`
			SELECT cc.id, cc.command_type, ad.name AS device_name, cc.status, CAST(cc.created_at AS CHAR) as created_at
			FROM control_commands cc
			JOIN actuator_channels ac ON ac.id = cc.actuator_channel_id
			JOIN actuator_devices ad ON ad.id = ac.actuator_device_id
			ORDER BY cc.created_at DESC
			LIMIT 5
		`).Scan(&recentCmds).Error; e != nil {
			errCh <- e
		}
	}()

	wg.Wait()
	close(errCh)

	for e := range errCh {
		if e != nil {
			response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
			return
		}
	}

	sensorsOffline := totalSensors - sensorsOnline
	if sensorsOffline < 0 {
		sensorsOffline = 0
	}
	actuatorsOffline := totalActuators - actuatorsOnline
	if actuatorsOffline < 0 {
		actuatorsOffline = 0
	}

	stats.ActiveBatchesCount = int(activeBatchesCount)
	stats.UnresolvedAlerts = int(alertsOpen)
	stats.DevicesOnline = int(sensorsOnline + actuatorsOnline)
	stats.DevicesOffline = int(sensorsOffline + actuatorsOffline)
	stats.EnergyKwhToday = 145.0 // Mock data
	stats.WaterLToday = 2000.0   // Mock data

	// Mock Trends for now
	now := time.Now().Truncate(time.Hour)
	trends := DashboardTrends{
		Timestamps: make([]string, 24),
		ECAvg:      make([]float64, 24),
		PHAvg:      make([]float64, 24),
	}
	for i := 0; i < 24; i++ {
		t := now.Add(-time.Duration(23-i) * time.Hour)
		trends.Timestamps[i] = t.Format("15:04")
		trends.ECAvg[i] = 1.5 + float64(i%5)*0.1
		trends.PHAvg[i] = 5.8 + float64(i%3)*0.1
	}

	resp := DashboardResponse{
		Stats:          stats,
		Greenhouses:    greenhouses,
		Trends:         trends,
		ActiveBatches:  activeBatches,
		RecentAlerts:   recentAlerts,
		RecentCommands: recentCmds,
	}

	// Make sure slices are not nil
	if resp.Greenhouses == nil {
		resp.Greenhouses = []DashboardGreenhouse{}
	}
	if resp.ActiveBatches == nil {
		resp.ActiveBatches = []DashboardActiveBatch{}
	}
	if resp.RecentAlerts == nil {
		resp.RecentAlerts = []DashboardRecentAlert{}
	}
	if resp.RecentCommands == nil {
		resp.RecentCommands = []RecentCommand{}
	}

	response.Success(c, resp)
}

func queryGreenhouseMetrics(db *gorm.DB, out *[]DashboardGreenhouse) error {
	var rows []struct {
		ID       uint64  `gorm:"column:id"`
		Name     string  `gorm:"column:name"`
		Temp     float64 `gorm:"column:avg_temp"`
		Humidity float64 `gorm:"column:avg_humidity"`
		EC       float64 `gorm:"column:avg_ec"`
		PH       float64 `gorm:"column:avg_ph"`
		DO       float64 `gorm:"column:avg_do"`
		CO2      float64 `gorm:"column:avg_co2"`
		Lux      float64 `gorm:"column:avg_lux"`
	}

	err := db.Raw(`
		SELECT
			g.id,
			g.name,
			COALESCE(AVG(t_temp.value), 0) AS avg_temp,
			COALESCE(AVG(t_hum.value), 0) AS avg_humidity,
			COALESCE(AVG(t_ec.value), 0) AS avg_ec,
			COALESCE(AVG(t_ph.value), 0) AS avg_ph,
			COALESCE(AVG(t_do.value), 0) AS avg_do,
			COALESCE(AVG(t_co2.value), 0) AS avg_co2,
			COALESCE(AVG(t_lux.value), 0) AS avg_lux
		FROM greenhouses g
		LEFT JOIN LATERAL (
			SELECT tr.value FROM telemetry_records tr
			JOIN sensor_channels sc ON sc.id = tr.sensor_channel_id
			JOIN sensor_devices sd2 ON sd2.id = sc.sensor_device_id
			WHERE sd2.greenhouse_id = g.id AND tr.metric_code = 'TEMP'
			ORDER BY tr.collected_at DESC LIMIT 1
		) t_temp ON true
		LEFT JOIN LATERAL (
			SELECT tr.value FROM telemetry_records tr
			JOIN sensor_channels sc ON sc.id = tr.sensor_channel_id
			JOIN sensor_devices sd2 ON sd2.id = sc.sensor_device_id
			WHERE sd2.greenhouse_id = g.id AND tr.metric_code = 'HUMIDITY'
			ORDER BY tr.collected_at DESC LIMIT 1
		) t_hum ON true
		LEFT JOIN LATERAL (
			SELECT tr.value FROM telemetry_records tr
			JOIN sensor_channels sc ON sc.id = tr.sensor_channel_id
			JOIN sensor_devices sd2 ON sd2.id = sc.sensor_device_id
			WHERE sd2.greenhouse_id = g.id AND tr.metric_code = 'EC'
			ORDER BY tr.collected_at DESC LIMIT 1
		) t_ec ON true
		LEFT JOIN LATERAL (
			SELECT tr.value FROM telemetry_records tr
			JOIN sensor_channels sc ON sc.id = tr.sensor_channel_id
			JOIN sensor_devices sd2 ON sd2.id = sc.sensor_device_id
			WHERE sd2.greenhouse_id = g.id AND tr.metric_code = 'PH'
			ORDER BY tr.collected_at DESC LIMIT 1
		) t_ph ON true
		LEFT JOIN LATERAL (
			SELECT tr.value FROM telemetry_records tr
			JOIN sensor_channels sc ON sc.id = tr.sensor_channel_id
			JOIN sensor_devices sd2 ON sd2.id = sc.sensor_device_id
			WHERE sd2.greenhouse_id = g.id AND tr.metric_code = 'DO'
			ORDER BY tr.collected_at DESC LIMIT 1
		) t_do ON true
		LEFT JOIN LATERAL (
			SELECT tr.value FROM telemetry_records tr
			JOIN sensor_channels sc ON sc.id = tr.sensor_channel_id
			JOIN sensor_devices sd2 ON sd2.id = sc.sensor_device_id
			WHERE sd2.greenhouse_id = g.id AND tr.metric_code = 'CO2'
			ORDER BY tr.collected_at DESC LIMIT 1
		) t_co2 ON true
		LEFT JOIN LATERAL (
			SELECT tr.value FROM telemetry_records tr
			JOIN sensor_channels sc ON sc.id = tr.sensor_channel_id
			JOIN sensor_devices sd2 ON sd2.id = sc.sensor_device_id
			WHERE sd2.greenhouse_id = g.id AND tr.metric_code = 'ILLUMINATION'
			ORDER BY tr.collected_at DESC LIMIT 1
		) t_lux ON true
		GROUP BY g.id, g.name
		ORDER BY g.id
	`).Scan(&rows).Error
	if err != nil {
		return err
	}

	*out = make([]DashboardGreenhouse, 0, len(rows))
	for _, r := range rows {
		gh := DashboardGreenhouse{
			ID:               strconv.FormatUint(r.ID, 10),
			Name:             r.Name,
			HealthScore:      "good", // Default for now
			ActiveStrategies: []string{"Default Strategy"},
			Metrics: GreenhouseMetrics{
				Temperature: r.Temp,
				Humidity:    r.Humidity,
				EC:          r.EC,
				PH:          r.PH,
				DO:          r.DO,
				CO2:         r.CO2,
				Lux:         r.Lux,
			},
		}
		*out = append(*out, gh)
	}
	return nil
}
