package energy

import (
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

// CreateEnergyRecord creates a new energy consumption record.
func (h *Handler) CreateEnergyRecord(c *gin.Context) {
	var req CreateEnergyRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	record := EnergyConsumptionRecord{
		GreenhouseID:      req.GreenhouseID,
		RecordType:        req.RecordType,
		ConsumptionValue:  req.ConsumptionValue,
		Unit:              req.Unit,
		RecordPeriodStart: req.RecordPeriodStart,
		RecordPeriodEnd:   req.RecordPeriodEnd,
		MeterReadingStart: req.MeterReadingStart,
		MeterReadingEnd:   req.MeterReadingEnd,
		BatchID:           req.BatchID,
		RecordedBy:        req.RecordedBy,
	}

	if err := h.db.Create(&record).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "create_failed", nil)
		return
	}

	response.Success(c, gin.H{"id": record.ID})
}

// UpdateEnergyRecord updates an existing energy consumption record.
func (h *Handler) UpdateEnergyRecord(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var req UpdateEnergyRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	updates := map[string]interface{}{}
	if req.ConsumptionValue != nil {
		updates["consumption_value"] = *req.ConsumptionValue
	}
	if req.Unit != nil {
		updates["unit"] = *req.Unit
	}
	if req.MeterReadingStart != nil {
		updates["meter_reading_start"] = *req.MeterReadingStart
	}
	if req.MeterReadingEnd != nil {
		updates["meter_reading_end"] = *req.MeterReadingEnd
	}
	if req.BatchID != nil {
		updates["batch_id"] = *req.BatchID
	}

	if len(updates) == 0 {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "no_fields", nil)
		return
	}

	result := h.db.Model(&EnergyConsumptionRecord{}).Where("id = ?", id).Updates(updates)
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

// GetEnergyRecord returns a single energy consumption record.
func (h *Handler) GetEnergyRecord(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var record EnergyConsumptionRecord
	if err := h.db.First(&record, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	response.Success(c, recordToItem(record))
}

// DeleteEnergyRecord soft-deletes an energy consumption record.
func (h *Handler) DeleteEnergyRecord(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	result := h.db.Delete(&EnergyConsumptionRecord{}, id)
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

// ListEnergyRecords returns a paginated list of energy records with filters.
func (h *Handler) ListEnergyRecords(c *gin.Context) {
	page, pageSize := parsePage(c)

	query := h.db.Model(&EnergyConsumptionRecord{})
	if v := c.Query("record_type"); v != "" {
		query = query.Where("record_type = ?", v)
	}
	if v := c.Query("unit"); v != "" {
		query = query.Where("unit = ?", v)
	}
	if from := c.Query("period_start_from"); from != "" {
		t, err := time.Parse(time.RFC3339, from)
		if err == nil {
			query = query.Where("record_period_start >= ?", t)
		}
	}
	if to := c.Query("period_end_to"); to != "" {
		t, err := time.Parse(time.RFC3339, to)
		if err == nil {
			query = query.Where("record_period_end <= ?", t)
		}
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	var records []EnergyConsumptionRecord
	if total > 0 {
		if err := query.Order("record_period_start DESC, id DESC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&records).Error; err != nil {
			response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
			return
		}
	}

	items := make([]gin.H, 0, len(records))
	for _, r := range records {
		items = append(items, recordToItem(r))
	}

	response.Success(c, gin.H{
		"page":      page,
		"page_size": pageSize,
		"total":     total,
		"items":     items,
	})
}

// ListByGreenhouse returns energy records for a specific greenhouse.
func (h *Handler) ListByGreenhouse(c *gin.Context) {
	greenhouseID, err := parseID(c.Param("greenhouseId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	page, pageSize := parsePage(c)

	query := h.db.Model(&EnergyConsumptionRecord{}).Where("greenhouse_id = ?", greenhouseID)
	if v := c.Query("record_type"); v != "" {
		query = query.Where("record_type = ?", v)
	}
	if from := c.Query("start"); from != "" {
		t, err := time.Parse(time.RFC3339, from)
		if err == nil {
			query = query.Where("record_period_start >= ?", t)
		}
	}
	if to := c.Query("end"); to != "" {
		t, err := time.Parse(time.RFC3339, to)
		if err == nil {
			query = query.Where("record_period_end <= ?", t)
		}
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	var records []EnergyConsumptionRecord
	if total > 0 {
		if err := query.Order("record_period_start DESC, id DESC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&records).Error; err != nil {
			response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
			return
		}
	}

	items := make([]gin.H, 0, len(records))
	for _, r := range records {
		items = append(items, recordToItem(r))
	}

	response.Success(c, gin.H{
		"page":      page,
		"page_size": pageSize,
		"total":     total,
		"items":     items,
	})
}

// ListByBatch returns energy records for a specific batch.
func (h *Handler) ListByBatch(c *gin.Context) {
	batchID, err := parseID(c.Param("batchId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	page, pageSize := parsePage(c)

	query := h.db.Model(&EnergyConsumptionRecord{}).Where("batch_id = ?", batchID)
	if v := c.Query("record_type"); v != "" {
		query = query.Where("record_type = ?", v)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	var records []EnergyConsumptionRecord
	if total > 0 {
		if err := query.Order("record_period_start DESC, id DESC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&records).Error; err != nil {
			response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
			return
		}
	}

	items := make([]gin.H, 0, len(records))
	for _, r := range records {
		items = append(items, recordToItem(r))
	}

	response.Success(c, gin.H{
		"page":      page,
		"page_size": pageSize,
		"total":     total,
		"items":     items,
	})
}

// EnergySummary returns summarized energy consumption grouped by record_type.
func (h *Handler) EnergySummary(c *gin.Context) {
	query := h.db.Model(&EnergyConsumptionRecord{})

	if v := c.Query("greenhouse_id"); v != "" {
		id, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_greenhouse_id", nil)
			return
		}
		query = query.Where("greenhouse_id = ?", id)
	}
	if start := c.Query("start"); start != "" {
		t, err := time.Parse(time.RFC3339, start)
		if err == nil {
			query = query.Where("record_period_start >= ?", t)
		}
	}
	if end := c.Query("end"); end != "" {
		t, err := time.Parse(time.RFC3339, end)
		if err == nil {
			query = query.Where("record_period_end <= ?", t)
		}
	}

	type summaryRow struct {
		RecordType string  `gorm:"column:record_type"`
		TotalValue float64 `gorm:"column:total_value"`
		Unit       string  `gorm:"column:unit"`
	}

	var rows []summaryRow
	if err := query.Select("record_type, SUM(consumption_value) AS total_value, unit").
		Group("record_type, unit").
		Scan(&rows).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	summaries := make([]EnergySummaryItem, 0, len(rows))
	for _, row := range rows {
		summaries = append(summaries, EnergySummaryItem{
			RecordType:       row.RecordType,
			TotalConsumption: row.TotalValue,
			Unit:             row.Unit,
		})
	}

	response.Success(c, gin.H{
		"items": summaries,
	})
}

// --- Helpers ---

func recordToItem(r EnergyConsumptionRecord) gin.H {
	return gin.H{
		"id":                  r.ID,
		"greenhouse_id":       r.GreenhouseID,
		"record_type":         r.RecordType,
		"consumption_value":   r.ConsumptionValue,
		"unit":                r.Unit,
		"record_period_start": timeToStr(r.RecordPeriodStart),
		"record_period_end":   timeToStr(r.RecordPeriodEnd),
		"meter_reading_start": r.MeterReadingStart,
		"meter_reading_end":   r.MeterReadingEnd,
		"batch_id":            r.BatchID,
		"recorded_by":         r.RecordedBy,
		"created_at":          timeToStr(r.CreatedAt),
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
