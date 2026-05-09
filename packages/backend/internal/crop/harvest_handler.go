package crop

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	platformErrors "hydroponic-backend/internal/platform/errors"
	"hydroponic-backend/internal/platform/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ======================== BatchStagePlan Handlers ========================

func (h *Handler) CreateStagePlan(c *gin.Context) {
	var req CreateBatchStagePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	startAt, err := time.Parse(time.RFC3339, req.StageStartAt)
	if err != nil {
		startAt, err = time.Parse(time.RFC3339Nano, req.StageStartAt)
		if err != nil {
			response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_stage_start_at", nil)
			return
		}
	}

	endAt, err := time.Parse(time.RFC3339, req.StageEndAt)
	if err != nil {
		endAt, err = time.Parse(time.RFC3339Nano, req.StageEndAt)
		if err != nil {
			response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_stage_end_at", nil)
			return
		}
	}

	plan := BatchStagePlan{
		BatchID:       req.BatchID,
		GrowthStageID: req.GrowthStageID,
		StageStartAt:  startAt.UTC(),
		StageEndAt:    endAt.UTC(),
		TargetECMin:   req.TargetECMin,
		TargetECMax:   req.TargetECMax,
		TargetPHMin:   req.TargetPHMin,
		TargetPHMax:   req.TargetPHMax,
	}

	if err := h.db.Create(&plan).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "create_failed", nil)
		return
	}

	response.Success(c, toStagePlanResponse(plan))
}

func (h *Handler) GetStagePlan(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var plan BatchStagePlan
	if err := h.db.First(&plan, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	response.Success(c, toStagePlanResponse(plan))
}

func (h *Handler) UpdateStagePlan(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var req UpdateBatchStagePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	updates := map[string]interface{}{}
	if req.StageStartAt != nil {
		updates["stage_start_at"] = parseTimePtr(req.StageStartAt)
	}
	if req.StageEndAt != nil {
		updates["stage_end_at"] = parseTimePtr(req.StageEndAt)
	}
	if req.TargetECMin != nil {
		updates["target_ec_min"] = *req.TargetECMin
	}
	if req.TargetECMax != nil {
		updates["target_ec_max"] = *req.TargetECMax
	}
	if req.TargetPHMin != nil {
		updates["target_ph_min"] = *req.TargetPHMin
	}
	if req.TargetPHMax != nil {
		updates["target_ph_max"] = *req.TargetPHMax
	}

	if len(updates) == 0 {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "no_fields_to_update", nil)
		return
	}

	result := h.db.Model(&BatchStagePlan{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "update_failed", nil)
		return
	}
	if result.RowsAffected == 0 {
		response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
		return
	}

	var plan BatchStagePlan
	h.db.First(&plan, id)
	response.Success(c, toStagePlanResponse(plan))
}

func (h *Handler) DeleteStagePlan(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	result := h.db.Delete(&BatchStagePlan{}, id)
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

func (h *Handler) ListStagePlans(c *gin.Context) {
	q := h.db.Model(&BatchStagePlan{})

	if v := strings.TrimSpace(c.Query("batch_id")); v != "" {
		id, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_batch_id", nil)
			return
		}
		q = q.Where("batch_id = ?", id)
	}

	var plans []BatchStagePlan
	if err := q.Order("stage_start_at ASC").Find(&plans).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	items := make([]BatchStagePlanResponse, 0, len(plans))
	for _, p := range plans {
		items = append(items, toStagePlanResponse(p))
	}

	response.Success(c, CropListResponse{
		Items:    items,
		Total:    int64(len(items)),
		Page:     1,
		PageSize: len(items),
	})
}

// ======================== HarvestRecord Handlers ========================

func (h *Handler) CreateHarvest(c *gin.Context) {
	var req CreateHarvestRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	harvestedAt, err := time.Parse(time.RFC3339, req.HarvestedAt)
	if err != nil {
		harvestedAt, err = time.Parse(time.RFC3339Nano, req.HarvestedAt)
		if err != nil {
			response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_harvested_at", nil)
			return
		}
	}

	grade := req.Grade
	if grade == "" {
		grade = HarvestGradeA
	}

	userID := currentUserID(c)

	record := HarvestRecord{
		BatchID:         req.BatchID,
		HarvestedAt:     harvestedAt.UTC(),
		HarvestWeightKg: req.HarvestWeightKg,
		Grade:           grade,
		GradeWeightKg:   req.GradeWeightKg,
		Note:            req.Note,
		HarvestedBy:     &userID,
	}

	if err := h.db.Create(&record).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "create_failed", nil)
		return
	}

	response.Success(c, toHarvestResponse(record))
}

func (h *Handler) ListHarvestsByBatch(c *gin.Context) {
	page, pageSize := parsePageQuery(c)

	q := h.db.Model(&HarvestRecord{})

	if v := strings.TrimSpace(c.Query("batch_id")); v != "" {
		id, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_batch_id", nil)
			return
		}
		q = q.Where("batch_id = ?", id)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	var records []HarvestRecord
	if total > 0 {
		if err := q.Order("harvested_at DESC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&records).Error; err != nil {
			response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
			return
		}
	}

	items := make([]HarvestRecordResponse, 0, len(records))
	for _, r := range records {
		items = append(items, toHarvestResponse(r))
	}

	response.Success(c, CropListResponse{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

func (h *Handler) GetHarvestSummary(c *gin.Context) {
	batchID, err := parseID(c.Param("batchId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_batch_id", nil)
		return
	}

	// Load batch info for yield estimation
	var batch CropBatch
	if err := h.db.First(&batch, batchID).Error; err != nil {
		response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "batch_not_found", nil)
		return
	}

	type gradeSummary struct {
		Grade   string  `json:"grade"`
		TotalKg float64 `json:"total_kg"`
		Count   int64   `json:"count"`
	}
	var summaries []gradeSummary
	if err := h.db.Model(&HarvestRecord{}).
		Select("grade, SUM(grade_weight_kg) as total_kg, COUNT(*) as count").
		Where("batch_id = ?", batchID).
		Group("grade").
		Order("grade ASC").
		Scan(&summaries).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	var totalWeight float64
	grades := make([]HarvestGradeSummary, 0, len(summaries))
	for _, s := range summaries {
		totalWeight += s.TotalKg
		grades = append(grades, HarvestGradeSummary{
			Grade:    s.Grade,
			WeightKg: s.TotalKg,
			Count:    s.Count,
		})
	}

	// Estimate yield: planting_density × zone_area (m²)
	var estimatedYield *float64
	if batch.PlantingDensity != nil && batch.GrowingZoneID != nil {
		var zoneArea float64
		if err := h.db.Table("growing_zones").
			Select("COALESCE(area_sqm, 0)").
			Where("id = ?", *batch.GrowingZoneID).
			Scan(&zoneArea).Error; err == nil && zoneArea > 0 {
			est := *batch.PlantingDensity * zoneArea
			estimatedYield = &est
		}
	}

	resp := gin.H{
		"batch_id":        batchID,
		"total_weight_kg": totalWeight,
		"grades":          grades,
	}
	if estimatedYield != nil {
		resp["estimated_yield_kg"] = *estimatedYield
		resp["yield_rate"] = 0.0
		if *estimatedYield > 0 {
			resp["yield_rate"] = float64(int(totalWeight / *estimatedYield * 1000)) / 10 // percentage with 1 decimal
		}
	}

	response.Success(c, resp)
}
