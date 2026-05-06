package overview

import (
	"net/http"
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
		errCh           = make(chan error, 8)
		sensorsOnline   int64
		actuatorsOnline int64
		totalSensors    int64
		totalActuators  int64
		alertsOpen      int64
		alertsCritical  int64
		alertsToday     int64
		ghSummaries     []greenhouseSummary
		recentCmds      []recentCommand
	)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		e := h.db.Model(&sensorDevice{}).
			Where("status = ?", "ONLINE").
			Count(&sensorsOnline).Error
		if e != nil {
			errCh <- e
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		e := h.db.Model(&sensorDevice{}).Count(&totalSensors).Error
		if e != nil {
			errCh <- e
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		e := h.db.Model(&actuatorDevice{}).
			Where("status = ?", "ONLINE").
			Count(&actuatorsOnline).Error
		if e != nil {
			errCh <- e
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		e := h.db.Model(&actuatorDevice{}).Count(&totalActuators).Error
		if e != nil {
			errCh <- e
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		e := h.db.Model(&alertpkg.Alert{}).Where("status = ?", alertpkg.StatusOpen).Count(&alertsOpen).Error
		if e != nil {
			errCh <- e
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		e := h.db.Model(&alertpkg.Alert{}).
			Where("status = ? AND level = ?", alertpkg.StatusOpen, alertpkg.LevelCritical).
			Count(&alertsCritical).Error
		if e != nil {
			errCh <- e
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		e := h.db.Model(&alertpkg.Alert{}).
			Where("triggered_at >= ?", time.Now().UTC().Truncate(24*time.Hour)).
			Count(&alertsToday).Error
		if e != nil {
			errCh <- e
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		e := queryGreenhouseSummary(h.db, &ghSummaries)
		if e != nil {
			errCh <- e
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		e := h.db.Raw(`
			SELECT cc.id, cc.command_type, ad.name AS device_name, cc.status, cc.created_at
			FROM control_commands cc
			JOIN actuator_channels ac ON ac.id = cc.actuator_channel_id
			JOIN actuator_devices ad ON ad.id = ac.actuator_device_id
			ORDER BY cc.created_at DESC
			LIMIT 5
		`).Scan(&recentCmds).Error
		if e != nil {
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

	response.Success(c, gin.H{
		"sensors_online":     sensorsOnline,
		"sensors_offline":    sensorsOffline,
		"sensors_total":      totalSensors,
		"actuators_online":   actuatorsOnline,
		"actuators_offline":  actuatorsOffline,
		"actuators_total":    totalActuators,
		"alerts_open":        alertsOpen,
		"alerts_critical":    alertsCritical,
		"alerts_today":       alertsToday,
		"greenhouse_summary": ghSummaries,
		"recent_commands":    recentCmds,
	})
}

type sensorDevice struct {
	ID     uint64 `gorm:"primaryKey"`
	Status string `gorm:"size:16"`
}

func (sensorDevice) TableName() string { return "sensor_devices" }

type actuatorDevice struct {
	ID     uint64 `gorm:"primaryKey"`
	Status string `gorm:"size:16"`
}

func (actuatorDevice) TableName() string { return "actuator_devices" }

type greenhouseSummary struct {
	GreenhouseID  uint64   `json:"greenhouse_id"`
	Name          string   `json:"name"`
	SensorCount   int64    `json:"sensor_count"`
	ActuatorCount int64    `json:"actuator_count"`
	ZoneCount     int64    `json:"zone_count"`
	AvgTemp       *float64 `json:"avg_temp"`
	AvgHumidity   *float64 `json:"avg_humidity"`
}

type recentCommand struct {
	ID          uint64    `json:"id"`
	CommandType string    `json:"command_type"`
	DeviceName  string    `json:"device_name"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

func queryGreenhouseSummary(db *gorm.DB, out *[]greenhouseSummary) error {
	var rows []struct {
		ID            uint64  `gorm:"column:id"`
		Name          string  `gorm:"column:name"`
		SensorCount   int64   `gorm:"column:sensor_count"`
		ActuatorCount int64   `gorm:"column:actuator_count"`
		ZoneCount     int64   `gorm:"column:zone_count"`
		AvgTemp       float64 `gorm:"column:avg_temp"`
		AvgHumidity   float64 `gorm:"column:avg_humidity"`
	}

	err := db.Raw(`
		SELECT
			g.id,
			g.name,
			COUNT(DISTINCT sd.id) AS sensor_count,
			COUNT(DISTINCT ad.id) AS actuator_count,
			COUNT(DISTINCT gz.id) AS zone_count,
			COALESCE(AVG(t_temp.value), 0) AS avg_temp,
			COALESCE(AVG(t_hum.value), 0) AS avg_humidity
		FROM greenhouses g
		LEFT JOIN sensor_devices sd ON sd.greenhouse_id = g.id
		LEFT JOIN actuator_devices ad ON ad.greenhouse_id = g.id
		LEFT JOIN growing_zones gz ON gz.greenhouse_id = g.id
		LEFT JOIN LATERAL (
			SELECT tr.value FROM telemetry_records tr
			JOIN sensor_channels sc ON sc.id = tr.sensor_channel_id
			JOIN sensor_devices sd2 ON sd2.id = sc.sensor_device_id
			WHERE sd2.greenhouse_id = g.id AND tr.metric_code = 'TEMP'
			ORDER BY tr.collected_at DESC
			LIMIT 1
		) t_temp ON true
		LEFT JOIN LATERAL (
			SELECT tr.value FROM telemetry_records tr
			JOIN sensor_channels sc ON sc.id = tr.sensor_channel_id
			JOIN sensor_devices sd2 ON sd2.id = sc.sensor_device_id
			WHERE sd2.greenhouse_id = g.id AND tr.metric_code = 'HUMIDITY'
			ORDER BY tr.collected_at DESC
			LIMIT 1
		) t_hum ON true
		GROUP BY g.id, g.name
		ORDER BY g.id
	`).Scan(&rows).Error
	if err != nil {
		return err
	}

	*out = make([]greenhouseSummary, 0, len(rows))
	for _, r := range rows {
		gh := greenhouseSummary{
			GreenhouseID:  r.ID,
			Name:          r.Name,
			SensorCount:   r.SensorCount,
			ActuatorCount: r.ActuatorCount,
			ZoneCount:     r.ZoneCount,
		}
		if r.AvgTemp != 0 {
			v := r.AvgTemp
			gh.AvgTemp = &v
		}
		if r.AvgHumidity != 0 {
			v := r.AvgHumidity
			gh.AvgHumidity = &v
		}
		*out = append(*out, gh)
	}
	return nil
}
