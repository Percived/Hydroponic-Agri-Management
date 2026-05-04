package overview

import (
	"net/http"
	"sync"
	"time"

	alertpkg "hydroponic-backend/internal/alert"
	"hydroponic-backend/internal/device"
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
		errCh          = make(chan error, 8)
		devicesOnline  int64
		totalDevices   int64
		alertsOpen     int64
		alertsCritical int64
		alertsToday    int64
		typeDist       []deviceTypeCount
		ghSummaries    []greenhouseSummary
		recentCmds     []recentCommand
	)

	// Run all queries in parallel
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		e := h.db.Model(&device.Device{}).
			Where("status = ?", device.DeviceStatusEnabled).
			Where("last_seen_at IS NOT NULL").
			Where("TIMESTAMPDIFF(SECOND, last_seen_at, UTC_TIMESTAMP()) <= sampling_interval_sec * 3").
			Count(&devicesOnline).Error
		if e != nil {
			errCh <- e
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		e := h.db.Model(&device.Device{}).Count(&totalDevices).Error
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
			Where("status = ? AND level = ?", alertpkg.StatusOpen, "CRITICAL").
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
		e := h.db.Raw("SELECT type, COUNT(*) AS count FROM devices GROUP BY type").Scan(&typeDist).Error
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
			SELECT cc.id, cc.command_type, d.name AS device_name, cc.status, cc.created_at
			FROM control_commands cc
			JOIN devices d ON d.id = cc.device_id
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

	devicesOffline := totalDevices - devicesOnline
	if devicesOffline < 0 {
		devicesOffline = 0
	}

	response.Success(c, gin.H{
		"devices_online":           devicesOnline,
		"devices_offline":          devicesOffline,
		"devices_total":            totalDevices,
		"alerts_open":              alertsOpen,
		"alerts_critical":          alertsCritical,
		"alerts_today":             alertsToday,
		"device_type_distribution": typeDist,
		"greenhouse_summary":       ghSummaries,
		"recent_commands":          recentCmds,
	})
}

type deviceTypeCount struct {
	Type  string `json:"type"`
	Count int64  `json:"count"`
}

type greenhouseSummary struct {
	GreenhouseID uint64   `json:"greenhouse_id"`
	Name         string   `json:"name"`
	DeviceCount  int64    `json:"device_count"`
	AvgTemp      *float64 `json:"avg_temp"`
	AvgHumidity  *float64 `json:"avg_humidity"`
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
		ID          uint64  `gorm:"column:id"`
		Name        string  `gorm:"column:name"`
		DeviceCount int64   `gorm:"column:device_count"`
		AvgTemp     float64 `gorm:"column:avg_temp"`
		AvgHumidity float64 `gorm:"column:avg_humidity"`
	}

	err := db.Raw(`
		SELECT
			g.id,
			g.name,
			COUNT(d.id) AS device_count,
			AVG(t_temp.value) AS avg_temp,
			AVG(t_hum.value) AS avg_humidity
		FROM greenhouses g
		LEFT JOIN devices d ON d.greenhouse_id = g.id
		LEFT JOIN LATERAL (
			SELECT value FROM telemetry_data
			WHERE device_id = d.id
			  AND metric_id = (SELECT id FROM metrics WHERE code = 'TEMP' LIMIT 1)
			ORDER BY collected_at DESC
			LIMIT 1
		) t_temp ON true
		LEFT JOIN LATERAL (
			SELECT value FROM telemetry_data
			WHERE device_id = d.id
			  AND metric_id = (SELECT id FROM metrics WHERE code = 'HUMIDITY' LIMIT 1)
			ORDER BY collected_at DESC
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
			GreenhouseID: r.ID,
			Name:         r.Name,
			DeviceCount:  r.DeviceCount,
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
