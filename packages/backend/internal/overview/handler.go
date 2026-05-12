package overview

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	alertpkg "hydroponic-backend/internal/alert"
	platformErrors "hydroponic-backend/internal/platform/errors"
	"hydroponic-backend/internal/platform/response"

	"github.com/gin-gonic/gin"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"gorm.io/gorm"
)

type Handler struct {
	db     *gorm.DB
	influx influxdb2.Client
	org    string
	bucket string
}

func NewHandler(db *gorm.DB, influx influxdb2.Client, org, bucket string) *Handler {
	return &Handler{db: db, influx: influx, org: org, bucket: bucket}
}

func (h *Handler) Dashboard(c *gin.Context) {
	var (
		errCh              = make(chan error, 12)
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
		energyKwhToday     float64
		waterLToday        float64
	)

	var wg sync.WaitGroup

	// Devices count
	wg.Add(1)
	go func() {
		defer wg.Done()
		if h.db.Migrator().HasTable("sensor_devices") {
			h.db.Table("sensor_devices").Count(&totalSensors)
			h.db.Table("sensor_devices").Where("status = ?", "ONLINE").Count(&sensorsOnline)
		}
		if h.db.Migrator().HasTable("actuator_devices") {
			h.db.Table("actuator_devices").Count(&totalActuators)
			h.db.Table("actuator_devices").Where("status = ?", "ONLINE").Count(&actuatorsOnline)
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

	// Energy and Water consumption for today
	wg.Add(1)
	go func() {
		defer wg.Done()
		if h.db.Migrator().HasTable("energy_consumption_records") {
			todayStart := time.Now().Truncate(24 * time.Hour)
			var totalEnergy struct {
				Total float64
			}
			h.db.Table("energy_consumption_records").
				Select("COALESCE(SUM(consumption_value), 0) as total").
				Where("record_type = ? AND record_period_start >= ?", "ELECTRICITY", todayStart).
				Scan(&totalEnergy)
			energyKwhToday = totalEnergy.Total
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if h.db.Migrator().HasTable("energy_consumption_records") {
			todayStart := time.Now().Truncate(24 * time.Hour)
			var totalWater struct {
				Total float64
			}
			h.db.Table("energy_consumption_records").
				Select("COALESCE(SUM(consumption_value), 0) as total").
				Where("record_type = ? AND record_period_start >= ?", "WATER", todayStart).
				Scan(&totalWater)
			waterLToday = totalWater.Total
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
					cv.name as crop_name,
					COALESCE(gs.name, cb.status) as stage,
					DATEDIFF(NOW(), cb.started_at) as day,
					CAST(cb.greenhouse_id AS CHAR) as greenhouse_id
				FROM crop_batches cb
				JOIN crop_varieties cv ON cv.id = cb.crop_variety_id
				LEFT JOIN batch_stage_runtime bsr ON bsr.batch_id = cb.id
				LEFT JOIN growth_stages gs ON gs.id = bsr.current_growth_stage_id
				WHERE cb.status = 'RUNNING'
				LIMIT 5
			`).Scan(&activeBatches)
		}
	}()

	// Recent alerts
	wg.Add(1)
	go func() {
		defer wg.Done()
		if h.db.Migrator().HasTable("alerts") {
			h.db.Raw(`
				SELECT
					CAST(a.id AS CHAR) as alert_id,
					a.level as severity,
					a.message,
					a.triggered_at as timestamp,
					COALESCE(g_sensor.name, g_actuator.name, 'System') as greenhouse_name
				FROM alerts a
				LEFT JOIN sensor_channels sc ON sc.id = a.sensor_channel_id
				LEFT JOIN sensor_devices sd ON sd.id = sc.sensor_device_id
				LEFT JOIN greenhouses g_sensor ON g_sensor.id = sd.greenhouse_id
				LEFT JOIN actuator_channels ac ON ac.id = a.actuator_channel_id
				LEFT JOIN actuator_devices ad ON ad.id = ac.actuator_device_id
				LEFT JOIN greenhouses g_actuator ON g_actuator.id = ad.greenhouse_id
				WHERE a.status = 'OPEN'
				ORDER BY a.triggered_at DESC
				LIMIT 5
			`).Scan(&recentAlerts)
		}
	}()

	// Recent commands
	wg.Add(1)
	go func() {
		defer wg.Done()
		if h.db.Migrator().HasTable("commands") {
			h.db.Raw(`
				SELECT 
					CAST(c.id AS CHAR) as command_id, 
					c.command_type, 
					c.status, 
					c.dispatched_at as timestamp, 
					g.name as greenhouse_name
				FROM commands c
				JOIN actuator_devices ad ON ad.id = c.actuator_device_id
				JOIN greenhouses g ON g.id = ad.greenhouse_id
				ORDER BY c.dispatched_at DESC
				LIMIT 5
			`).Scan(&recentCmds)
		}
	}()

	// Query InfluxDB for 24h Trends
	now := time.Now().Truncate(time.Hour)
	trends := DashboardTrends{
		Timestamps: make([]string, 24),
		ECAvg:      make([]float64, 24),
		PHAvg:      make([]float64, 24),
	}

	// Pre-fill timestamps
	for i := 0; i < 24; i++ {
		t := now.Add(-time.Duration(23-i) * time.Hour)
		trends.Timestamps[i] = t.Format("15:04")
	}

	if h.influx != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// Flux query to aggregate mean by 1 hour for EC and PH
			fluxQuery := fmt.Sprintf(`
				from(bucket: "%s")
					|> range(start: -24h)
					|> filter(fn: (r) => r["_measurement"] == "telemetry")
					|> filter(fn: (r) => r["metric_code"] == "EC" or r["metric_code"] == "PH")
					|> filter(fn: (r) => r["_field"] == "value")
					|> aggregateWindow(every: 1h, fn: mean, createEmpty: false)
					|> yield(name: "mean")
			`, h.bucket)

			queryAPI := h.influx.QueryAPI(h.org)
			result, err := queryAPI.Query(context.Background(), fluxQuery)
			if err != nil {
				errCh <- err
				return
			}

			// We have 24 buckets, need to map result timestamps to our pre-filled buckets
			for result.Next() {
				record := result.Record()
				val, ok := record.Value().(float64)
				if !ok {
					continue
				}

				metricCode, _ := record.ValueByKey("metric_code").(string)
				t := record.Time().Truncate(time.Hour)

				// Find index in our 24h array
				idx := -1
				for i := 0; i < 24; i++ {
					bucketTime := now.Add(-time.Duration(23-i) * time.Hour)
					if t.Equal(bucketTime) {
						idx = i
						break
					}
				}

				if idx >= 0 && idx < 24 {
					if metricCode == "EC" {
						trends.ECAvg[idx] = val
					} else if metricCode == "PH" {
						trends.PHAvg[idx] = val
					}
				}
			}
			if result.Err() != nil {
				errCh <- result.Err()
			}
		}()
	}

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
	stats.EnergyKwhToday = energyKwhToday
	stats.WaterLToday = waterLToday

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
		ID             uint64  `gorm:"column:id"`
		Name           string  `gorm:"column:name"`
		Temp           float64 `gorm:"column:avg_temp"`
		Humidity       float64 `gorm:"column:avg_humidity"`
		EC             float64 `gorm:"column:avg_ec"`
		PH             float64 `gorm:"column:avg_ph"`
		DO             float64 `gorm:"column:avg_do"`
		CO2            float64 `gorm:"column:avg_co2"`
		Lux            float64 `gorm:"column:avg_lux"`
		CriticalAlerts int64   `gorm:"column:critical_alerts"`
		WarningAlerts  int64   `gorm:"column:warning_alerts"`
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
			COALESCE(AVG(t_lux.value), 0) AS avg_lux,
			(SELECT COUNT(*) FROM alerts a WHERE a.status = 'OPEN' AND a.level = 'CRITICAL' AND (
				a.sensor_channel_id IN (SELECT sc.id FROM sensor_channels sc JOIN sensor_devices sd2 ON sd2.id = sc.sensor_device_id WHERE sd2.greenhouse_id = g.id)
				OR a.actuator_channel_id IN (SELECT ac.id FROM actuator_channels ac JOIN actuator_devices ad2 ON ad2.id = ac.actuator_device_id WHERE ad2.greenhouse_id = g.id)
			)) as critical_alerts,
			(SELECT COUNT(*) FROM alerts a WHERE a.status = 'OPEN' AND a.level = 'WARNING' AND (
				a.sensor_channel_id IN (SELECT sc.id FROM sensor_channels sc JOIN sensor_devices sd2 ON sd2.id = sc.sensor_device_id WHERE sd2.greenhouse_id = g.id)
				OR a.actuator_channel_id IN (SELECT ac.id FROM actuator_channels ac JOIN actuator_devices ad2 ON ad2.id = ac.actuator_device_id WHERE ad2.greenhouse_id = g.id)
			)) as warning_alerts
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
		// Calculate Health Score
		healthScore := "good"
		if r.CriticalAlerts > 0 {
			healthScore = "critical"
		} else if r.WarningAlerts > 0 {
			healthScore = "warning"
		}

		strategies, err := loadGreenhouseActiveStrategies(db, r.ID)
		if err != nil {
			return err
		}

		gh := DashboardGreenhouse{
			ID:               strconv.FormatUint(r.ID, 10),
			Name:             r.Name,
			HealthScore:      healthScore,
			ActiveStrategies: strategies,
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

func loadGreenhouseActiveStrategies(db *gorm.DB, greenhouseID uint64) ([]string, error) {
	strategies := make([]string, 0, 4)

	if db.Migrator().HasTable("climate_profiles") {
		var climateNames []string
		if err := db.Table("climate_profiles").
			Distinct("name").
			Where("greenhouse_id = ? AND enabled = ?", greenhouseID, true).
			Order("name ASC").
			Pluck("name", &climateNames).Error; err != nil {
			return nil, err
		}
		strategies = append(strategies, climateNames...)
	}

	if db.Migrator().HasTable("crop_batches") && db.Migrator().HasTable("nutrient_recipes") {
		var recipeNames []string
		if err := db.Table("crop_batches AS cb").
			Joins("JOIN nutrient_recipes nr ON nr.id = cb.active_recipe_id").
			Distinct("nr.name").
			Where("cb.greenhouse_id = ? AND cb.status = ?", greenhouseID, "RUNNING").
			Where("nr.status = ?", "ACTIVE").
			Order("nr.name ASC").
			Pluck("nr.name", &recipeNames).Error; err != nil {
			return nil, err
		}
		strategies = append(strategies, recipeNames...)
	}

	return strategies, nil
}
