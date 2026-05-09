package crop

import (
	"net/http"
	"sync"
	"time"

	platformErrors "hydroponic-backend/internal/platform/errors"
	"hydroponic-backend/internal/platform/response"

	"github.com/gin-gonic/gin"
)

// ======================== Batch Dashboard Handler ========================

// GetBatchDashboard aggregates all batch-related data into a single dashboard response.
// Uses goroutines for concurrent sub-queries.
func (h *Handler) GetBatchDashboard(c *gin.Context) {
	batchID, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	// Load batch first (needed to verify existence)
	var batch CropBatch
	if err := h.db.First(&batch, batchID).Error; err != nil {
		response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
		return
	}

	batchResp := h.toBatchResponse(batch)

	// Result holders
	type varietyInfo struct {
		Code        string `json:"code"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	type nameResult struct {
		GreenhouseName string `json:"greenhouse_name"`
		ZoneName       string `json:"zone_name"`
	}

	var (
		wg sync.WaitGroup

		varietyInfoRes  *varietyInfo
		nameRes         *nameResult
		plantingRes     *PlantingRecordResponse
		stageProgress   *StageProgressResponse
		devices         []BatchDeviceResponse
		latestTelemetry []gin.H
		recentAlerts    []gin.H
		recentCommands  []gin.H
		harvestSummary  *gin.H

		mu sync.Mutex // protects concurrent writes to slices/maps above
	)

	// 1. Variety info
	wg.Add(1)
	go func() {
		defer wg.Done()
		var v varietyInfo
		if err := h.db.Table("crop_varieties").
			Select("code, name, description").
			Where("id = ?", batch.CropVarietyID).
			First(&v).Error; err == nil {
			varietyInfoRes = &v
		}
	}()

	// 2. Greenhouse and zone names
	wg.Add(1)
	go func() {
		defer wg.Done()
		var ghName, zoneName string
		h.db.Table("greenhouses").Select("name").Where("id = ?", batch.GreenhouseID).Scan(&ghName)
		if batch.GrowingZoneID != nil {
			h.db.Table("growing_zones").Select("name").Where("id = ?", *batch.GrowingZoneID).Scan(&zoneName)
		}
		nameRes = &nameResult{GreenhouseName: ghName, ZoneName: zoneName}
	}()

	// 3. Planting record
	wg.Add(1)
	go func() {
		defer wg.Done()
		var pr PlantingRecord
		if err := h.db.Where("batch_id = ?", batchID).First(&pr).Error; err == nil {
			resp := toPlantingResponse(pr)
			plantingRes = &resp
		}
	}()

	// 4. Stage progress
	wg.Add(1)
	go func() {
		defer wg.Done()
		sp := h.queryStageProgress(batchID)
		stageProgress = sp
	}()

	// 5. Devices
	wg.Add(1)
	go func() {
		defer wg.Done()
		devices = h.queryDashboardDevices(batchID)
	}()

	// 6. Latest telemetry (10 most recent records)
	wg.Add(1)
	go func() {
		defer wg.Done()
		latestTelemetry = h.queryLatestTelemetry(batchID, 10)
	}()

	// 7. Recent open alerts (5 most recent)
	wg.Add(1)
	go func() {
		defer wg.Done()
		recentAlerts = h.queryRecentAlerts(batchID, 5)
	}()

	// 8. Recent commands (5 most recent)
	wg.Add(1)
	go func() {
		defer wg.Done()
		recentCommands = h.queryRecentCommands(batchID, 5)
	}()

	// 9. Harvest summary
	wg.Add(1)
	go func() {
		defer wg.Done()
		harvestSummary = h.queryHarvestSummaryForDashboard(batchID)
	}()

	// 10. Stage plans
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = mu // not needed for single write
	}()

	wg.Wait()

	dashboard := gin.H{
		"batch":            batchResp,
		"variety":          varietyInfoRes,
		"greenhouse_name":  nameRes.GreenhouseName,
		"zone_name":        nameRes.ZoneName,
		"planting_record":  plantingRes,
		"stage_progress":   stageProgress,
		"devices":          devices,
		"latest_telemetry": latestTelemetry,
		"recent_alerts":    recentAlerts,
		"recent_commands":  recentCommands,
		"harvest_summary":  harvestSummary,
	}

	response.Success(c, dashboard)
}

// --- Sub-query helpers (called from goroutines) ---

func (h *Handler) queryStageProgress(batchID uint64) *StageProgressResponse {
	now := time.Now().UTC()
	var currentPlan struct {
		BatchStagePlan
		StageName string `gorm:"column:stage_name"`
		StageCode string `gorm:"column:stage_code"`
	}
	err := h.db.Table("batch_stage_plans bsp").
		Select("bsp.*, gs.name AS stage_name, gs.code AS stage_code").
		Joins("JOIN growth_stages gs ON gs.id = bsp.growth_stage_id").
		Where("bsp.batch_id = ? AND bsp.stage_start_at <= ? AND bsp.stage_end_at > ?", batchID, now, now).
		Order("bsp.stage_start_at ASC").
		Limit(1).
		Scan(&currentPlan).Error

	if err != nil || currentPlan.ID == 0 {
		return &StageProgressResponse{BatchID: batchID}
	}

	totalDuration := currentPlan.StageEndAt.Sub(currentPlan.StageStartAt).Hours()
	elapsed := now.Sub(currentPlan.StageStartAt).Hours()
	progress := (elapsed / totalDuration) * 100
	if progress > 100 {
		progress = 100
	}
	if progress < 0 {
		progress = 0
	}

	return &StageProgressResponse{
		BatchID:          batchID,
		CurrentStageID:   &currentPlan.ID,
		CurrentStageName: currentPlan.StageName,
		CurrentStageCode: currentPlan.StageCode,
		ProgressPercent:  float64(int(progress*10)) / 10,
		DaysElapsed:      int(elapsed / 24),
		DaysRemaining:    int((totalDuration - elapsed) / 24),
		TargetECMin:      currentPlan.TargetECMin,
		TargetECMax:      currentPlan.TargetECMax,
		TargetPHMin:      currentPlan.TargetPHMin,
		TargetPHMax:      currentPlan.TargetPHMax,
	}
}

func (h *Handler) queryDashboardDevices(batchID uint64) []BatchDeviceResponse {
	var bindings []BatchDevice
	h.db.Where("batch_id = ? AND is_active = 1", batchID).Order("bound_at DESC").Find(&bindings)

	items := make([]BatchDeviceResponse, 0, len(bindings))
	for _, bd := range bindings {
		resp := BatchDeviceResponse{
			ID:         bd.ID,
			BatchID:    bd.BatchID,
			DeviceType: bd.DeviceType,
			DeviceID:   bd.DeviceID,
			IsActive:   bd.IsActive,
			BoundAt:    bd.BoundAt.Format(time.RFC3339),
		}
		if bd.UnboundAt != nil {
			s := bd.UnboundAt.Format(time.RFC3339)
			resp.UnboundAt = &s
		}

		if bd.DeviceType == DeviceTypeSensor {
			var dev struct {
				Name   string `gorm:"column:name"`
				Code   string `gorm:"column:device_code"`
				Status string `gorm:"column:status"`
			}
			if err := h.db.Table("sensor_devices").Select("name, device_code, status").Where("id = ?", bd.DeviceID).First(&dev).Error; err == nil {
				resp.DeviceName = dev.Name
				resp.DeviceCode = dev.Code
			}
		} else if bd.DeviceType == DeviceTypeActuator {
			var dev struct {
				Name   string `gorm:"column:name"`
				Code   string `gorm:"column:device_code"`
				Status string `gorm:"column:status"`
			}
			if err := h.db.Table("actuator_devices").Select("name, device_code, status").Where("id = ?", bd.DeviceID).First(&dev).Error; err == nil {
				resp.DeviceName = dev.Name
				resp.DeviceCode = dev.Code
			}
		}

		items = append(items, resp)
	}
	return items
}

func (h *Handler) queryLatestTelemetry(batchID uint64, limit int) []gin.H {
	type telemetryRow struct {
		MetricCode  string    `gorm:"column:metric_code"`
		Value       float64   `gorm:"column:value"`
		CollectedAt time.Time `gorm:"column:collected_at"`
	}
	var rows []telemetryRow
	h.db.Table("telemetry_records tr").
		Select("tr.metric_code, tr.value, tr.collected_at").
		Joins("JOIN sensor_channels sc ON sc.id = tr.sensor_channel_id").
		Joins("JOIN sensor_devices sd ON sd.id = sc.sensor_device_id").
		Where("sd.id IN (SELECT device_id FROM batch_devices WHERE batch_id = ? AND device_type = 'sensor' AND is_active = 1)", batchID).
		Order("tr.collected_at DESC").
		Limit(limit).
		Scan(&rows)

	// Lookup metric names
	metricNames := map[string]struct {
		Name string `gorm:"column:name"`
		Unit string `gorm:"column:unit"`
	}{}
	var metrics []struct {
		Code string `gorm:"column:code"`
		Name string `gorm:"column:name"`
		Unit string `gorm:"column:unit"`
	}
	h.db.Table("metric_definitions").Select("code, name, unit").Find(&metrics)
	for _, m := range metrics {
		metricNames[m.Code] = struct {
			Name string `gorm:"column:name"`
			Unit string `gorm:"column:unit"`
		}{m.Name, m.Unit}
	}

	items := make([]gin.H, 0, len(rows))
	for _, r := range rows {
		name := r.MetricCode
		unit := ""
		if m, ok := metricNames[r.MetricCode]; ok {
			name = m.Name
			unit = m.Unit
		}
		items = append(items, gin.H{
			"metric_code":  r.MetricCode,
			"metric_name":  name,
			"value":        r.Value,
			"unit":         unit,
			"collected_at": r.CollectedAt.Format(time.RFC3339),
		})
	}
	return items
}

func (h *Handler) queryRecentAlerts(batchID uint64, limit int) []gin.H {
	var alerts []struct {
		ID          uint64    `gorm:"column:id"`
		Type        string    `gorm:"column:type"`
		Level       string    `gorm:"column:level"`
		Message     string    `gorm:"column:message"`
		Status      string    `gorm:"column:status"`
		TriggeredAt time.Time `gorm:"column:triggered_at"`
	}
	h.db.Table("alerts").
		Select("id, type, level, message, status, triggered_at").
		Where("batch_id = ?", batchID).
		Where("status IN ('OPEN', 'ACKNOWLEDGED')").
		Order("triggered_at DESC").
		Limit(limit).
		Scan(&alerts)

	items := make([]gin.H, 0, len(alerts))
	for _, a := range alerts {
		items = append(items, gin.H{
			"id":           a.ID,
			"type":         a.Type,
			"level":        a.Level,
			"message":      a.Message,
			"status":       a.Status,
			"triggered_at": a.TriggeredAt.Format(time.RFC3339),
		})
	}
	return items
}

func (h *Handler) queryRecentCommands(batchID uint64, limit int) []gin.H {
	var cmds []struct {
		ID          uint64    `gorm:"column:id"`
		CommandType string    `gorm:"column:command_type"`
		Status      string    `gorm:"column:status"`
		CreatedAt   time.Time `gorm:"column:created_at"`
	}
	h.db.Table("control_commands").
		Select("id, command_type, status, created_at").
		Where("batch_id = ?", batchID).
		Order("created_at DESC").
		Limit(limit).
		Scan(&cmds)

	items := make([]gin.H, 0, len(cmds))
	for _, cmd := range cmds {
		items = append(items, gin.H{
			"id":           cmd.ID,
			"command_type": cmd.CommandType,
			"status":       cmd.Status,
			"created_at":   cmd.CreatedAt.Format(time.RFC3339),
		})
	}
	return items
}

func (h *Handler) queryHarvestSummaryForDashboard(batchID uint64) *gin.H {
	type gradeRow struct {
		Grade    string  `gorm:"column:grade"`
		WeightKg float64 `gorm:"column:total_weight"`
		Count    int64   `gorm:"column:cnt"`
	}
	var grades []gradeRow
	h.db.Table("harvest_records").
		Select("grade, SUM(grade_weight_kg) AS total_weight, COUNT(*) AS cnt").
		Where("batch_id = ?", batchID).
		Group("grade").
		Scan(&grades)

	if len(grades) == 0 {
		return nil
	}

	totalWeight := 0.0
	gradeItems := make([]gin.H, 0, len(grades))
	for _, g := range grades {
		totalWeight += g.WeightKg
		gradeItems = append(gradeItems, gin.H{
			"grade":     g.Grade,
			"weight_kg": g.WeightKg,
			"count":     g.Count,
		})
	}

	summary := gin.H{
		"total_weight_kg": totalWeight,
		"grades":          gradeItems,
	}
	return &summary
}
