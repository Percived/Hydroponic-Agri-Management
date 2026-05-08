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

	result := h.db.Model(&CropBatch{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "update_failed", nil)
		return
	}
	if result.RowsAffected == 0 {
		response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
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

	result := h.db.Delete(&CropBatch{}, id)
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

	if v := strings.TrimSpace(c.Query("start_time")); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			t, err = time.Parse(time.RFC3339Nano, v)
			if err != nil {
				response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_start_time", nil)
				return
			}
		}
		q = q.Where("started_at >= ?", t.UTC())
	}

	if v := strings.TrimSpace(c.Query("end_time")); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			t, err = time.Parse(time.RFC3339Nano, v)
			if err != nil {
				response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_end_time", nil)
				return
			}
		}
		q = q.Where("started_at <= ?", t.UTC())
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

	result := h.db.Model(&CropBatch{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "update_failed", nil)
		return
	}
	if result.RowsAffected == 0 {
		response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
		return
	}

	var batch CropBatch
	h.db.First(&batch, id)
	resp := h.toBatchResponse(batch)
	response.Success(c, resp)
}
