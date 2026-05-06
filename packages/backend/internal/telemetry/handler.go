package telemetry

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"hydroponic-backend/internal/auth"
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

// IngestTelemetry handles POST /telemetry/ingest
// Supports both a single record (top-level fields) and batch (items array).
func (h *Handler) IngestTelemetry(c *gin.Context) {
	var batchReq IngestTelemetryBatchRequest

	// Try batch first
	if err := c.ShouldBindJSON(&batchReq); err != nil {
		// Fall back to single record
		var singleReq IngestTelemetryRequest
		if err2 := c.ShouldBindJSON(&singleReq); err2 != nil {
			response.ValidationError(c, err)
			return
		}
		batchReq.Items = []IngestTelemetryRequest{singleReq}
	}

	records := make([]TelemetryRecord, 0, len(batchReq.Items))
	for _, item := range batchReq.Items {
		collectedAt, err := time.Parse(time.RFC3339, item.CollectedAt)
		if err != nil {
			// Try RFC3339Nano as fallback
			collectedAt, err = time.Parse(time.RFC3339Nano, item.CollectedAt)
			if err != nil {
				response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError,
					"invalid_collected_at", gin.H{"field": "collected_at"})
				return
			}
		}

		qualityFlag := item.QualityFlag
		if qualityFlag == "" {
			qualityFlag = QualityFlagNormal
		}

		records = append(records, TelemetryRecord{
			SensorChannelID: item.SensorChannelID,
			MetricCode:      item.MetricCode,
			Value:           item.Value,
			RawValue:        item.RawValue,
			QualityFlag:     qualityFlag,
			CollectedAt:     collectedAt.UTC(),
			BatchID:         item.BatchID,
		})
	}

	if err := h.db.Create(&records).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "ingest_failed", nil)
		return
	}

	response.Success(c, gin.H{"accepted": len(records)})
}

// QueryTelemetry handles GET /telemetry/query
func (h *Handler) QueryTelemetry(c *gin.Context) {
	page, pageSize := parsePageQuery(c)

	q := h.db.Model(&TelemetryRecord{})

	if v := strings.TrimSpace(c.Query("sensor_channel_id")); v != "" {
		id, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_sensor_channel_id", nil)
			return
		}
		q = q.Where("sensor_channel_id = ?", id)
	}

	if v := strings.TrimSpace(c.Query("metric_code")); v != "" {
		q = q.Where("metric_code = ?", v)
	}

	if v := strings.TrimSpace(c.Query("start_time")); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_start_time", nil)
			return
		}
		q = q.Where("collected_at >= ?", t)
	}

	if v := strings.TrimSpace(c.Query("end_time")); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_end_time", nil)
			return
		}
		q = q.Where("collected_at <= ?", t)
	}

	if v := strings.TrimSpace(c.Query("batch_id")); v != "" {
		id, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_batch_id", nil)
			return
		}
		q = q.Where("batch_id = ?", id)
	}

	if v := strings.TrimSpace(c.Query("quality_flag")); v != "" {
		q = q.Where("quality_flag = ?", v)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	var records []TelemetryRecord
	if total > 0 {
		if err := q.Order("collected_at DESC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&records).Error; err != nil {
			response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
			return
		}
	}

	items := make([]TelemetryRecordResponse, 0, len(records))
	for _, r := range records {
		items = append(items, toRecordResponse(r))
	}

	response.Success(c, TelemetryListResponse{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

// GetLatestByChannel handles GET /telemetry/channels/:channelId/latest
func (h *Handler) GetLatestByChannel(c *gin.Context) {
	channelID, err := strconv.ParseUint(c.Param("channelId"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_channel_id", nil)
		return
	}

	var record TelemetryRecord
	err = h.db.Where("sensor_channel_id = ?", channelID).
		Order("collected_at DESC").
		First(&record).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "no_data", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	response.Success(c, toRecordResponse(record))
}

// GetChannelHistory handles GET /telemetry/channels/:channelId/history
func (h *Handler) GetChannelHistory(c *gin.Context) {
	channelID, err := strconv.ParseUint(c.Param("channelId"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_channel_id", nil)
		return
	}

	page, pageSize := parsePageQuery(c)

	q := h.db.Model(&TelemetryRecord{}).Where("sensor_channel_id = ?", channelID)

	if v := strings.TrimSpace(c.Query("metric_code")); v != "" {
		q = q.Where("metric_code = ?", v)
	}

	if v := strings.TrimSpace(c.Query("start_time")); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_start_time", nil)
			return
		}
		q = q.Where("collected_at >= ?", t)
	}

	if v := strings.TrimSpace(c.Query("end_time")); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_end_time", nil)
			return
		}
		q = q.Where("collected_at <= ?", t)
	}

	if v := strings.TrimSpace(c.Query("batch_id")); v != "" {
		id, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_batch_id", nil)
			return
		}
		q = q.Where("batch_id = ?", id)
	}

	if v := strings.TrimSpace(c.Query("quality_flag")); v != "" {
		q = q.Where("quality_flag = ?", v)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	var records []TelemetryRecord
	if total > 0 {
		if err := q.Order("collected_at DESC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&records).Error; err != nil {
			response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
			return
		}
	}

	items := make([]TelemetryRecordResponse, 0, len(records))
	for _, r := range records {
		items = append(items, toRecordResponse(r))
	}

	response.Success(c, TelemetryListResponse{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

// DeleteTelemetry handles DELETE /telemetry (admin only)
func (h *Handler) DeleteTelemetry(c *gin.Context) {
	startTimeStr := strings.TrimSpace(c.Query("start_time"))
	endTimeStr := strings.TrimSpace(c.Query("end_time"))

	if startTimeStr == "" && endTimeStr == "" {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "time_range_required", nil)
		return
	}

	q := h.db.Model(&TelemetryRecord{})

	if startTimeStr != "" {
		t, err := time.Parse(time.RFC3339, startTimeStr)
		if err != nil {
			response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_start_time", nil)
			return
		}
		q = q.Where("collected_at >= ?", t)
	}

	if endTimeStr != "" {
		t, err := time.Parse(time.RFC3339, endTimeStr)
		if err != nil {
			response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_end_time", nil)
			return
		}
		q = q.Where("collected_at <= ?", t)
	}

	result := q.Delete(&TelemetryRecord{})
	if result.Error != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "delete_failed", nil)
		return
	}

	response.Success(c, gin.H{"deleted": result.RowsAffected})
}

// toRecordResponse converts a TelemetryRecord to a response DTO.
func toRecordResponse(r TelemetryRecord) TelemetryRecordResponse {
	return TelemetryRecordResponse{
		ID:              r.ID,
		SensorChannelID: r.SensorChannelID,
		MetricCode:      r.MetricCode,
		Value:           r.Value,
		RawValue:        r.RawValue,
		QualityFlag:     r.QualityFlag,
		CollectedAt:     r.CollectedAt.Format(time.RFC3339),
		IngestedAt:      r.IngestedAt.Format(time.RFC3339),
		BatchID:         r.BatchID,
		CreatedAt:       r.CreatedAt.Format(time.RFC3339),
	}
}

// currentUserID extracts the current user's ID from the gin context.
func currentUserID(c *gin.Context) uint64 {
	v, ok := c.Get(auth.CtxUserID)
	if !ok {
		return 0
	}
	id, ok := v.(uint64)
	if !ok {
		return 0
	}
	return id
}

// parsePageQuery parses page and page_size from query params.
func parsePageQuery(c *gin.Context) (int, int) {
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
