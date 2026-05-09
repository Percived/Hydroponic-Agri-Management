package review

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

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

// CreateSnapshot creates a new batch review snapshot.
func (h *Handler) CreateSnapshot(c *gin.Context) {
	var req CreateSnapshotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	snapshot := BatchReviewSnapshot{
		BatchID:      req.BatchID,
		SnapshotType: req.SnapshotType,
		WindowStart:  req.WindowStart,
		WindowEnd:    req.WindowEnd,
		Summary:      req.Summary,
		GeneratedAt:  req.GeneratedAt,
	}

	if err := h.db.Create(&snapshot).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "create_failed", nil)
		return
	}

	response.Success(c, gin.H{"id": snapshot.ID})
}

// UpdateSnapshot updates an existing snapshot.
func (h *Handler) UpdateSnapshot(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var req UpdateSnapshotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	updates := map[string]interface{}{}
	if req.SnapshotType != nil {
		updates["snapshot_type"] = *req.SnapshotType
	}
	if req.Summary != nil {
		updates["summary"] = *req.Summary
	}

	if len(updates) == 0 {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "no_fields", nil)
		return
	}

	result := h.db.Model(&BatchReviewSnapshot{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "update_failed", nil)
		return
	}
	if result.RowsAffected == 0 {
		response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
		return
	}

	response.Success(c, gin.H{})
}

// GetSnapshot returns a single snapshot.
func (h *Handler) GetSnapshot(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var snapshot BatchReviewSnapshot
	if err := h.db.First(&snapshot, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	response.Success(c, snapshotToItem(snapshot))
}

// DeleteSnapshot deletes a snapshot.
func (h *Handler) DeleteSnapshot(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	result := h.db.Delete(&BatchReviewSnapshot{}, id)
	if result.Error != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "delete_failed", nil)
		return
	}
	if result.RowsAffected == 0 {
		response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
		return
	}

	response.Success(c, gin.H{})
}

// ListSnapshots returns a paginated list of snapshots.
func (h *Handler) ListSnapshots(c *gin.Context) {
	page, pageSize := parsePage(c)

	query := h.db.Model(&BatchReviewSnapshot{})
	if v := c.Query("snapshot_type"); v != "" {
		query = query.Where("snapshot_type = ?", v)
	}
	if from := c.Query("window_start_from"); from != "" {
		t, err := time.Parse(time.RFC3339, from)
		if err == nil {
			query = query.Where("window_start >= ?", t)
		}
	}
	if to := c.Query("window_end_to"); to != "" {
		t, err := time.Parse(time.RFC3339, to)
		if err == nil {
			query = query.Where("window_end <= ?", t)
		}
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	var snapshots []BatchReviewSnapshot
	if total > 0 {
		if err := query.Order("generated_at DESC, id DESC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&snapshots).Error; err != nil {
			response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
			return
		}
	}

	items := make([]gin.H, 0, len(snapshots))
	for _, s := range snapshots {
		items = append(items, snapshotToItem(s))
	}

	response.Success(c, gin.H{
		"page":      page,
		"page_size": pageSize,
		"total":     total,
		"items":     items,
	})
}

// GetSnapshotsByBatch returns all snapshots for a specific batch.
func (h *Handler) GetSnapshotsByBatch(c *gin.Context) {
	batchID, err := parseID(c.Param("batchId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	page, pageSize := parsePage(c)

	query := h.db.Model(&BatchReviewSnapshot{}).Where("batch_id = ?", batchID)
	if v := c.Query("snapshot_type"); v != "" {
		query = query.Where("snapshot_type = ?", v)
	}
	if from := c.Query("window_start_from"); from != "" {
		t, err := time.Parse(time.RFC3339, from)
		if err == nil {
			query = query.Where("window_start >= ?", t)
		}
	}
	if to := c.Query("window_end_to"); to != "" {
		t, err := time.Parse(time.RFC3339, to)
		if err == nil {
			query = query.Where("window_end <= ?", t)
		}
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	var snapshots []BatchReviewSnapshot
	if total > 0 {
		if err := query.Order("generated_at DESC, id DESC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&snapshots).Error; err != nil {
			response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
			return
		}
	}

	items := make([]gin.H, 0, len(snapshots))
	for _, s := range snapshots {
		items = append(items, snapshotToItem(s))
	}

	response.Success(c, gin.H{
		"page":      page,
		"page_size": pageSize,
		"total":     total,
		"items":     items,
	})
}

// GenerateReview generates a review snapshot by querying telemetry data and computing averages.
func (h *Handler) GenerateReview(c *gin.Context) {
	var req GenerateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	now := time.Now().UTC()

	// Query telemetry data within the window for the given batch.
	// telemetry_records has: sensor_channel_id, metric_code, value, collected_at
	// Batch filtering via batch_devices -> sensor_devices -> sensor_channels chain.
	type telemetryRow struct {
		MetricCode string  `gorm:"column:metric_code"`
		MetricName string  `gorm:"column:metric_name"`
		Unit       string  `gorm:"column:unit"`
		AvgValue   float64 `gorm:"column:avg_value"`
		MinValue   float64 `gorm:"column:min_value"`
		MaxValue   float64 `gorm:"column:max_value"`
		Count      int64   `gorm:"column:count"`
	}

	var rows []telemetryRow
	err := h.db.Table("telemetry_records tr").
		Select("md.code AS metric_code, md.name AS metric_name, md.unit AS unit, AVG(tr.value) AS avg_value, MIN(tr.value) AS min_value, MAX(tr.value) AS max_value, COUNT(*) AS count").
		Joins("JOIN metric_definitions md ON md.code = tr.metric_code").
		Joins("JOIN sensor_channels sc ON sc.id = tr.sensor_channel_id").
		Joins("JOIN sensor_devices sd ON sd.id = sc.sensor_device_id").
		Where("sd.id IN (SELECT device_id FROM batch_devices WHERE batch_id = ? AND device_type = 'sensor' AND is_active = 1)", req.BatchID).
		Where("tr.collected_at >= ? AND tr.collected_at <= ?", req.WindowStart, req.WindowEnd).
		Group("md.code, md.name, md.unit").
		Find(&rows).Error

	if err != nil {
		// Fallback: query all telemetry within window (no batch filter)
		err = h.db.Table("telemetry_records tr").
			Select("md.code AS metric_code, md.name AS metric_name, md.unit AS unit, AVG(tr.value) AS avg_value, MIN(tr.value) AS min_value, MAX(tr.value) AS max_value, COUNT(*) AS count").
			Joins("JOIN metric_definitions md ON md.code = tr.metric_code").
			Where("tr.collected_at >= ? AND tr.collected_at <= ?", req.WindowStart, req.WindowEnd).
			Group("md.code, md.name, md.unit").
			Find(&rows).Error

		if err != nil {
			rows = []telemetryRow{}
		}
	}

	// Build summary from telemetry rows
	metricSummaries := make([]gin.H, 0, len(rows))
	for _, r := range rows {
		metricSummaries = append(metricSummaries, gin.H{
			"metric_code":  r.MetricCode,
			"metric_name":  r.MetricName,
			"unit":         r.Unit,
			"avg":          truncateDecimals(r.AvgValue, 4),
			"min":          truncateDecimals(r.MinValue, 4),
			"max":          truncateDecimals(r.MaxValue, 4),
			"sample_count": r.Count,
		})
	}

	// Count alerts within the window for this batch
	var alertCount int64
	h.db.Table("alerts").
		Where("batch_id = ?", req.BatchID).
		Where("triggered_at >= ? AND triggered_at <= ?", req.WindowStart, req.WindowEnd).
		Count(&alertCount)

	summary := gin.H{
		"batch_id":      req.BatchID,
		"snapshot_type": req.SnapshotType,
		"window_start":  timeToStr(req.WindowStart),
		"window_end":    timeToStr(req.WindowEnd),
		"metrics":       metricSummaries,
		"alert_count":   alertCount,
	}

	// For FINAL snapshots, include energy and pest statistics
	if req.SnapshotType == SnapshotFinal {
		// Energy consumption total
		var energyTotal float64
		h.db.Table("energy_consumption_records").
			Select("COALESCE(SUM(consumption_value), 0)").
			Where("batch_id = ? AND recorded_at >= ? AND recorded_at <= ?", req.BatchID, req.WindowStart, req.WindowEnd).
			Scan(&energyTotal)
		summary["energy_consumption"] = gin.H{
			"total":            energyTotal,
			"consumption_unit": "kWh",
		}

		// Pest observations count
		var pestCount int64
		h.db.Table("pest_disease_observations").
			Where("batch_id = ? AND observed_at >= ? AND observed_at <= ?", req.BatchID, req.WindowStart, req.WindowEnd).
			Count(&pestCount)
		summary["pest_observations"] = gin.H{"count": pestCount}

		// Treatment records count
		var treatmentCount int64
		h.db.Table("treatment_records").
			Where("batch_id = ? AND treated_at >= ? AND treated_at <= ?", req.BatchID, req.WindowStart, req.WindowEnd).
			Count(&treatmentCount)
		summary["treatment_records"] = gin.H{"count": treatmentCount}

		// Total commands count
		var commandCount int64
		h.db.Table("control_commands").
			Where("batch_id = ? AND created_at >= ? AND created_at <= ?", req.BatchID, req.WindowStart, req.WindowEnd).
			Count(&commandCount)
		summary["command_count"] = commandCount
	}

	summaryJSON, err := json.Marshal(summary)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "generate_failed", nil)
		return
	}

	snapshot := BatchReviewSnapshot{
		BatchID:      req.BatchID,
		SnapshotType: req.SnapshotType,
		WindowStart:  req.WindowStart,
		WindowEnd:    req.WindowEnd,
		Summary:      summaryJSON,
		GeneratedAt:  now,
	}

	if err := h.db.Create(&snapshot).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "create_failed", nil)
		return
	}

	response.Success(c, gin.H{
		"id":      snapshot.ID,
		"summary": summary,
	})
}

// GetBatchReview returns live review data (trends, alerts, controls) for a batch.
func (h *Handler) GetBatchReview(c *gin.Context) {
	batchID, err := parseID(c.Param("batchId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	snapshotType := c.Query("snapshot_type")
	windowStart := c.Query("from")
	windowEnd := c.Query("to")

	// Query environment trends via batch_devices chain
	type trendRow struct {
		MetricCode string  `json:"metric_code"`
		Time       string  `json:"time"`
		Value      float64 `json:"value"`
	}
	var trends []trendRow
	h.db.Table("telemetry_records tr").
		Select("md.code AS metric_code, tr.collected_at AS time, tr.value").
		Joins("JOIN metric_definitions md ON md.code = tr.metric_code").
		Joins("JOIN sensor_channels sc ON sc.id = tr.sensor_channel_id").
		Joins("JOIN sensor_devices sd ON sd.id = sc.sensor_device_id").
		Where("sd.id IN (SELECT device_id FROM batch_devices WHERE batch_id = ? AND device_type = 'sensor' AND is_active = 1)", batchID).
		Where("tr.collected_at >= ? AND tr.collected_at <= ?", windowStart, windowEnd).
		Order("tr.collected_at ASC").
		Find(&trends)

	// Query alerts
	type alertRow struct {
		TriggeredAt string `json:"triggered_at"`
		Level       string `json:"level"`
		Message     string `json:"message"`
	}
	var alerts []alertRow
	h.db.Table("alerts").
		Select("triggered_at, level, message").
		Where("batch_id = ? AND triggered_at >= ? AND triggered_at <= ?", batchID, windowStart, windowEnd).
		Order("triggered_at ASC").
		Find(&alerts)

	// Query control commands
	type controlRow struct {
		CreatedAt   string `json:"created_at"`
		CommandType string `json:"command_type"`
		Status      string `json:"status"`
	}
	var controls []controlRow
	h.db.Table("control_commands").
		Select("created_at, command_type, status").
		Where("batch_id = ? AND created_at >= ? AND created_at <= ?", batchID, windowStart, windowEnd).
		Order("created_at ASC").
		Find(&controls)

	// Query latest matching snapshot
	type snapshotRow struct {
		SnapshotType string          `json:"snapshot_type"`
		AlertCount   int64           `json:"alert_count"`
		ControlCount int64           `json:"control_count"`
		FailureCount int64           `json:"failure_count"`
		Summary      json.RawMessage `json:"summary"`
	}
	var snapshots []snapshotRow
	q := h.db.Table("batch_review_snapshots").
		Select("snapshot_type, summary").
		Where("batch_id = ?", batchID)
	if snapshotType != "" {
		q = q.Where("snapshot_type = ?", snapshotType)
	}
	q.Order("generated_at DESC").Limit(1).Find(&snapshots)

	// Parse summary for alert_count / control_count / failure_count
	type parsedSnapshot struct {
		SnapshotType string                 `json:"snapshot_type"`
		AlertCount   int64                  `json:"alert_count"`
		ControlCount int64                  `json:"control_count"`
		FailureCount int64                  `json:"failure_count"`
		Summary      map[string]interface{} `json:"summary"`
	}
	parsedSnapshots := make([]parsedSnapshot, 0, len(snapshots))
	for _, s := range snapshots {
		var m map[string]interface{}
		if err := json.Unmarshal(s.Summary, &m); err != nil {
			m = map[string]interface{}{}
		}
		ps := parsedSnapshot{
			SnapshotType: s.SnapshotType,
			Summary:      m,
		}
		if ac, ok := m["alert_count"].(float64); ok {
			ps.AlertCount = int64(ac)
		}
		if cc, ok := m["command_count"].(float64); ok {
			ps.ControlCount = int64(cc)
		}
		if fc, ok := m["failure_count"].(float64); ok {
			ps.FailureCount = int64(fc)
		}
		parsedSnapshots = append(parsedSnapshots, ps)
	}

	response.Success(c, gin.H{
		"environment_trends": trends,
		"alerts":             alerts,
		"controls":           controls,
		"snapshots":          parsedSnapshots,
		"summary":            gin.H{},
	})
}

// --- Helpers ---

func snapshotToItem(s BatchReviewSnapshot) gin.H {
	var summaryData interface{}
	if s.Summary != nil {
		json.Unmarshal(s.Summary, &summaryData)
	}
	return gin.H{
		"id":            s.ID,
		"batch_id":      s.BatchID,
		"snapshot_type": s.SnapshotType,
		"window_start":  timeToStr(s.WindowStart),
		"window_end":    timeToStr(s.WindowEnd),
		"summary":       summaryData,
		"generated_at":  timeToStr(s.GeneratedAt),
		"created_at":    timeToStr(s.CreatedAt),
	}
}

func parseID(v string) (uint64, error) {
	return strconv.ParseUint(v, 10, 64)
}

func parsePage(c *gin.Context) (int, int) {
	page := 1
	if v := c.Query("page"); v != "" {
		if i, err := strconv.Atoi(v); err == nil && i > 0 {
			page = i
		}
	}
	pageSize := 20
	if v := c.Query("page_size"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			pageSize = i
		}
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 200 {
		pageSize = 200
	}
	return page, pageSize
}

func truncateDecimals(v float64, precision int) float64 {
	format := strconv.FormatFloat(v, 'f', precision, 64)
	result, _ := strconv.ParseFloat(format, 64)
	return result
}
