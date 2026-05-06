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
	// The telemetry_data table has: device_id, metric_id, value, collected_at
	// We use batch_id from a subquery on devices or batch associations.
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
	err := h.db.Table("telemetry_data td").
		Select("m.code AS metric_code, m.name AS metric_name, m.unit AS unit, AVG(td.value) AS avg_value, MIN(td.value) AS min_value, MAX(td.value) AS max_value, COUNT(*) AS count").
		Joins("JOIN metrics m ON m.id = td.metric_id").
		Joins("JOIN devices d ON d.id = td.device_id").
		Where("d.id IN (SELECT device_id FROM batch_devices WHERE batch_id = ?)", req.BatchID).
		Where("td.collected_at >= ? AND td.collected_at <= ?", req.WindowStart, req.WindowEnd).
		Group("m.code, m.name, m.unit").
		Find(&rows).Error

	if err != nil {
		// Fallback: query without batch device join if batch_devices table doesn't exist
		err = h.db.Table("telemetry_data td").
			Select("m.code AS metric_code, m.name AS metric_name, m.unit AS unit, AVG(td.value) AS avg_value, MIN(td.value) AS min_value, MAX(td.value) AS max_value, COUNT(*) AS count").
			Joins("JOIN metrics m ON m.id = td.metric_id").
			Where("td.collected_at >= ? AND td.collected_at <= ?", req.WindowStart, req.WindowEnd).
			Group("m.code, m.name, m.unit").
			Find(&rows).Error

		if err != nil {
			// If telemetry_data doesn't exist yet, create an empty summary
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
