package crop

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

// ======================== CropVariety Handlers ========================

func (h *Handler) CreateVariety(c *gin.Context) {
	var req CreateCropVarietyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	variety := CropVariety{
		Code:             req.Code,
		Name:             req.Name,
		Description:      req.Description,
		DefaultCycleDays: req.DefaultCycleDays,
	}

	if err := h.db.Create(&variety).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "create_failed", nil)
		return
	}

	response.Success(c, toVarietyResponse(variety))
}

func (h *Handler) GetVariety(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var variety CropVariety
	if err := h.db.First(&variety, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	response.Success(c, toVarietyResponse(variety))
}

func (h *Handler) UpdateVariety(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var req UpdateCropVarietyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.DefaultCycleDays != nil {
		updates["default_cycle_days"] = *req.DefaultCycleDays
	}

	if len(updates) == 0 {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "no_fields_to_update", nil)
		return
	}

	if err := h.db.Model(&CropVariety{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "update_failed", nil)
		return
	}

	var variety CropVariety
	h.db.First(&variety, id)
	response.Success(c, toVarietyResponse(variety))
}

func (h *Handler) DeleteVariety(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	if err := h.db.Delete(&CropVariety{}, id).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "delete_failed", nil)
		return
	}

	response.Success(c, gin.H{})
}

func (h *Handler) ListVarieties(c *gin.Context) {
	page, pageSize := parsePageQuery(c)

	q := h.db.Model(&CropVariety{})

	if v := strings.TrimSpace(c.Query("code")); v != "" {
		q = q.Where("code LIKE ?", "%"+v+"%")
	}
	if v := strings.TrimSpace(c.Query("name")); v != "" {
		q = q.Where("name LIKE ?", "%"+v+"%")
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	var varieties []CropVariety
	if total > 0 {
		if err := q.Order("code ASC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&varieties).Error; err != nil {
			response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
			return
		}
	}

	items := make([]CropVarietyResponse, 0, len(varieties))
	for _, v := range varieties {
		items = append(items, toVarietyResponse(v))
	}

	response.Success(c, CropListResponse{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

// ======================== GrowthStage Handlers ========================

func (h *Handler) CreateStage(c *gin.Context) {
	var req CreateGrowthStageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	stage := GrowthStage{
		Code:                req.Code,
		Name:                req.Name,
		SortOrder:           req.SortOrder,
		DefaultDurationDays: req.DefaultDurationDays,
	}

	if err := h.db.Create(&stage).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "create_failed", nil)
		return
	}

	response.Success(c, toStageResponse(stage))
}

func (h *Handler) GetStage(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var stage GrowthStage
	if err := h.db.First(&stage, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	response.Success(c, toStageResponse(stage))
}

func (h *Handler) UpdateStage(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var req UpdateGrowthStageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.SortOrder != nil {
		updates["sort_order"] = *req.SortOrder
	}
	if req.DefaultDurationDays != nil {
		updates["default_duration_days"] = *req.DefaultDurationDays
	}

	if len(updates) == 0 {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "no_fields_to_update", nil)
		return
	}

	if err := h.db.Model(&GrowthStage{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "update_failed", nil)
		return
	}

	var stage GrowthStage
	h.db.First(&stage, id)
	response.Success(c, toStageResponse(stage))
}

func (h *Handler) DeleteStage(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	if err := h.db.Delete(&GrowthStage{}, id).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "delete_failed", nil)
		return
	}

	response.Success(c, gin.H{})
}

func (h *Handler) ListStages(c *gin.Context) {
	var stages []GrowthStage
	if err := h.db.Order("sort_order ASC, id ASC").Find(&stages).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	items := make([]GrowthStageResponse, 0, len(stages))
	for _, s := range stages {
		items = append(items, toStageResponse(s))
	}

	response.Success(c, CropListResponse{
		Items:    items,
		Total:    int64(len(items)),
		Page:     1,
		PageSize: len(items),
	})
}

// ======================== CropBatch Handlers ========================

func (h *Handler) CreateBatch(c *gin.Context) {
	var req CreateCropBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	status := req.Status
	if status == "" {
		status = BatchStatusPlanned
	}

	userID := currentUserID(c)

	batch := CropBatch{
		BatchNo:           req.BatchNo,
		GreenhouseID:      req.GreenhouseID,
		GrowingZoneID:     req.GrowingZoneID,
		CropVarietyID:     req.CropVarietyID,
		Status:            status,
		PlantingDensity:   req.PlantingDensity,
		TotalPlants:       req.TotalPlants,
		StartedAt:         parseTimePtr(req.StartedAt),
		EndedAt:           parseTimePtr(req.EndedAt),
		ExpectedHarvestAt: parseTimePtr(req.ExpectedHarvestAt),
		RecipeVersion:     req.RecipeVersion,
		PolicyVersion:     req.PolicyVersion,
		Note:              req.Note,
		CreatedBy:         &userID,
	}

	if err := h.db.Create(&batch).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "create_failed", nil)
		return
	}

	resp := h.toBatchResponse(batch)
	response.Success(c, resp)
}

func (h *Handler) GetBatch(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var batch CropBatch
	if err := h.db.First(&batch, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	resp := h.toBatchResponse(batch)
	response.Success(c, resp)
}

func (h *Handler) UpdateBatch(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var req UpdateCropBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	updates := map[string]interface{}{}
	if req.BatchNo != nil {
		updates["batch_no"] = *req.BatchNo
	}
	if req.GreenhouseID != nil {
		updates["greenhouse_id"] = *req.GreenhouseID
	}
	if req.GrowingZoneID != nil {
		updates["growing_zone_id"] = *req.GrowingZoneID
	}
	if req.CropVarietyID != nil {
		updates["crop_variety_id"] = *req.CropVarietyID
	}
	if req.PlantingDensity != nil {
		updates["planting_density"] = *req.PlantingDensity
	}
	if req.TotalPlants != nil {
		updates["total_plants"] = *req.TotalPlants
	}
	if req.StartedAt != nil {
		updates["started_at"] = parseTimePtr(req.StartedAt)
	}
	if req.EndedAt != nil {
		updates["ended_at"] = parseTimePtr(req.EndedAt)
	}
	if req.ExpectedHarvestAt != nil {
		updates["expected_harvest_at"] = parseTimePtr(req.ExpectedHarvestAt)
	}
	if req.RecipeVersion != nil {
		updates["recipe_version"] = *req.RecipeVersion
	}
	if req.PolicyVersion != nil {
		updates["policy_version"] = *req.PolicyVersion
	}
	if req.Note != nil {
		updates["note"] = *req.Note
	}

	if len(updates) == 0 {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "no_fields_to_update", nil)
		return
	}

	if err := h.db.Model(&CropBatch{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "update_failed", nil)
		return
	}

	var batch CropBatch
	h.db.First(&batch, id)
	resp := h.toBatchResponse(batch)
	response.Success(c, resp)
}

func (h *Handler) DeleteBatch(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	if err := h.db.Delete(&CropBatch{}, id).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "delete_failed", nil)
		return
	}

	response.Success(c, gin.H{})
}

func (h *Handler) ListBatches(c *gin.Context) {
	page, pageSize := parsePageQuery(c)

	q := h.db.Model(&CropBatch{})

	if v := strings.TrimSpace(c.Query("greenhouse_id")); v != "" {
		id, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_greenhouse_id", nil)
			return
		}
		q = q.Where("greenhouse_id = ?", id)
	}

	if v := strings.TrimSpace(c.Query("growing_zone_id")); v != "" {
		id, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_growing_zone_id", nil)
			return
		}
		q = q.Where("growing_zone_id = ?", id)
	}

	if v := strings.TrimSpace(c.Query("status")); v != "" {
		q = q.Where("status = ?", v)
	}

	if v := strings.TrimSpace(c.Query("crop_variety_id")); v != "" {
		id, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_crop_variety_id", nil)
			return
		}
		q = q.Where("crop_variety_id = ?", id)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	var batches []CropBatch
	if total > 0 {
		if err := q.Order("created_at DESC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&batches).Error; err != nil {
			response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
			return
		}
	}

	items := make([]CropBatchResponse, 0, len(batches))
	for _, b := range batches {
		items = append(items, h.toBatchResponse(b))
	}

	response.Success(c, CropListResponse{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

func (h *Handler) TransitionBatchStatus(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var req BatchStatusTransitionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	now := time.Now().UTC()
	updates := map[string]interface{}{
		"status": req.Status,
	}

	switch req.Status {
	case BatchStatusRunning:
		updates["started_at"] = now
	case BatchStatusCompleted:
		updates["ended_at"] = now
	case BatchStatusAborted:
		updates["ended_at"] = now
	}

	if err := h.db.Model(&CropBatch{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "update_failed", nil)
		return
	}

	var batch CropBatch
	h.db.First(&batch, id)
	resp := h.toBatchResponse(batch)
	response.Success(c, resp)
}

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

	if err := h.db.Model(&BatchStagePlan{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "update_failed", nil)
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

	if err := h.db.Delete(&BatchStagePlan{}, id).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "delete_failed", nil)
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
	batchIDStr := strings.TrimSpace(c.Query("batch_id"))
	if batchIDStr == "" {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "batch_id_required", nil)
		return
	}

	batchID, err := strconv.ParseUint(batchIDStr, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_batch_id", nil)
		return
	}

	page, pageSize := parsePageQuery(c)

	var total int64
	q := h.db.Model(&HarvestRecord{}).Where("batch_id = ?", batchID)
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

	response.Success(c, HarvestSummaryResponse{
		BatchID:       batchID,
		TotalWeightKg: totalWeight,
		Grades:        grades,
	})
}

// ======================== Helpers ========================

// toBatchResponse loads the related CropVariety and builds the response.
func (h *Handler) toBatchResponse(b CropBatch) CropBatchResponse {
	// Load variety info
	var variety CropVariety
	h.db.Select("code", "name").Where("id = ?", b.CropVarietyID).First(&variety)

	resp := CropBatchResponse{
		ID:              b.ID,
		BatchNo:         b.BatchNo,
		GreenhouseID:    b.GreenhouseID,
		GrowingZoneID:   b.GrowingZoneID,
		CropVarietyID:   b.CropVarietyID,
		VarietyCode:     variety.Code,
		VarietyName:     variety.Name,
		Status:          b.Status,
		PlantingDensity: b.PlantingDensity,
		TotalPlants:     b.TotalPlants,
		RecipeVersion:   b.RecipeVersion,
		PolicyVersion:   b.PolicyVersion,
		Note:            b.Note,
		CreatedBy:       b.CreatedBy,
		CreatedAt:       b.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       b.UpdatedAt.Format(time.RFC3339),
	}

	if b.StartedAt != nil {
		s := b.StartedAt.Format(time.RFC3339)
		resp.StartedAt = &s
	}
	if b.EndedAt != nil {
		s := b.EndedAt.Format(time.RFC3339)
		resp.EndedAt = &s
	}
	if b.ExpectedHarvestAt != nil {
		s := b.ExpectedHarvestAt.Format(time.RFC3339)
		resp.ExpectedHarvestAt = &s
	}

	return resp
}

func toVarietyResponse(v CropVariety) CropVarietyResponse {
	return CropVarietyResponse{
		ID:               v.ID,
		Code:             v.Code,
		Name:             v.Name,
		Description:      v.Description,
		DefaultCycleDays: v.DefaultCycleDays,
		CreatedAt:        v.CreatedAt.Format(time.RFC3339),
		UpdatedAt:        v.UpdatedAt.Format(time.RFC3339),
	}
}

func toStageResponse(s GrowthStage) GrowthStageResponse {
	return GrowthStageResponse{
		ID:                  s.ID,
		Code:                s.Code,
		Name:                s.Name,
		SortOrder:           s.SortOrder,
		DefaultDurationDays: s.DefaultDurationDays,
		CreatedAt:           s.CreatedAt.Format(time.RFC3339),
		UpdatedAt:           s.UpdatedAt.Format(time.RFC3339),
	}
}

func toStagePlanResponse(p BatchStagePlan) BatchStagePlanResponse {
	return BatchStagePlanResponse{
		ID:            p.ID,
		BatchID:       p.BatchID,
		GrowthStageID: p.GrowthStageID,
		StageStartAt:  p.StageStartAt.Format(time.RFC3339),
		StageEndAt:    p.StageEndAt.Format(time.RFC3339),
		TargetECMin:   p.TargetECMin,
		TargetECMax:   p.TargetECMax,
		TargetPHMin:   p.TargetPHMin,
		TargetPHMax:   p.TargetPHMax,
		CreatedAt:     p.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     p.UpdatedAt.Format(time.RFC3339),
	}
}

func toHarvestResponse(r HarvestRecord) HarvestRecordResponse {
	return HarvestRecordResponse{
		ID:              r.ID,
		BatchID:         r.BatchID,
		HarvestedAt:     r.HarvestedAt.Format(time.RFC3339),
		HarvestWeightKg: r.HarvestWeightKg,
		Grade:           r.Grade,
		GradeWeightKg:   r.GradeWeightKg,
		Note:            r.Note,
		HarvestedBy:     r.HarvestedBy,
		CreatedAt:       r.CreatedAt.Format(time.RFC3339),
	}
}

// ---------- Generic helpers ----------

func parseID(v string) (uint64, error) {
	return strconv.ParseUint(v, 10, 64)
}

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

func parseTimePtr(s *string) *time.Time {
	if s == nil || *s == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339, *s)
	if err != nil {
		t, err = time.Parse(time.RFC3339Nano, *s)
		if err != nil {
			return nil
		}
	}
	t = t.UTC()
	return &t
}
