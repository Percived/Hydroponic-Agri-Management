package crop

import (
	"net/http"
	"time"

	platformErrors "hydroponic-backend/internal/platform/errors"
	"hydroponic-backend/internal/platform/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ======================== PlantingRecord Handlers ========================

// CreatePlantingRecord creates a planting record for a batch (1:1).
func (h *Handler) CreatePlantingRecord(c *gin.Context) {
	var req CreatePlantingRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	userID := currentUserID(c)
	record := PlantingRecord{
		BatchID:                 req.BatchID,
		SeedSource:              req.SeedSource,
		SeedBatchNo:             req.SeedBatchNo,
		SeedlingAgeDays:         req.SeedlingAgeDays,
		SeededAt:                parseTimePtr(req.SeededAt),
		PlantedAt:               parseTimePtr(req.PlantedAt),
		ActualPlantCount:        req.ActualPlantCount,
		InitialEC:               req.InitialEC,
		InitialPH:               req.InitialPH,
		InitialWaterTemp:        req.InitialWaterTemp,
		InitialNutrientRecipeID: req.InitialNutrientRecipeID,
		PlantedBy:               &userID,
		Note:                    req.Note,
	}

	if err := h.db.Create(&record).Error; err != nil {
		response.Error(c, http.StatusConflict, platformErrors.CodeConflict, "planting_record_exists", nil)
		return
	}

	response.Success(c, gin.H{"id": record.ID})
}

// GetPlantingRecord returns the planting record for a batch.
func (h *Handler) GetPlantingRecord(c *gin.Context) {
	batchID, err := parseID(c.Param("batchId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var record PlantingRecord
	if err := h.db.Where("batch_id = ?", batchID).First(&record).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	resp := toPlantingResponse(record)
	response.Success(c, resp)
}

// UpdatePlantingRecord updates the planting record for a batch.
func (h *Handler) UpdatePlantingRecord(c *gin.Context) {
	batchID, err := parseID(c.Param("batchId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var req UpdatePlantingRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	updates := map[string]interface{}{}
	if req.SeedSource != nil {
		updates["seed_source"] = *req.SeedSource
	}
	if req.SeedBatchNo != nil {
		updates["seed_batch_no"] = *req.SeedBatchNo
	}
	if req.SeedlingAgeDays != nil {
		updates["seedling_age_days"] = *req.SeedlingAgeDays
	}
	if req.SeededAt != nil {
		updates["seeded_at"] = parseTimePtr(req.SeededAt)
	}
	if req.PlantedAt != nil {
		updates["planted_at"] = parseTimePtr(req.PlantedAt)
	}
	if req.ActualPlantCount != nil {
		updates["actual_plant_count"] = *req.ActualPlantCount
	}
	if req.InitialEC != nil {
		updates["initial_ec"] = *req.InitialEC
	}
	if req.InitialPH != nil {
		updates["initial_ph"] = *req.InitialPH
	}
	if req.InitialWaterTemp != nil {
		updates["initial_water_temp"] = *req.InitialWaterTemp
	}
	if req.InitialNutrientRecipeID != nil {
		updates["initial_nutrient_recipe_id"] = *req.InitialNutrientRecipeID
	}
	if req.Note != nil {
		updates["note"] = *req.Note
	}

	if len(updates) == 0 {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "no_fields_to_update", nil)
		return
	}

	result := h.db.Model(&PlantingRecord{}).Where("batch_id = ?", batchID).Updates(updates)
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

// ======================== Helpers ========================

func toPlantingResponse(r PlantingRecord) PlantingRecordResponse {
	resp := PlantingRecordResponse{
		ID:                      r.ID,
		BatchID:                 r.BatchID,
		SeedSource:              r.SeedSource,
		SeedBatchNo:             r.SeedBatchNo,
		SeedlingAgeDays:         r.SeedlingAgeDays,
		ActualPlantCount:        r.ActualPlantCount,
		InitialEC:               r.InitialEC,
		InitialPH:               r.InitialPH,
		InitialWaterTemp:        r.InitialWaterTemp,
		InitialNutrientRecipeID: r.InitialNutrientRecipeID,
		PlantedBy:               r.PlantedBy,
		Note:                    r.Note,
		CreatedAt:               r.CreatedAt.Format(time.RFC3339),
		UpdatedAt:               r.UpdatedAt.Format(time.RFC3339),
	}

	if r.SeededAt != nil {
		s := r.SeededAt.Format(time.RFC3339)
		resp.SeededAt = &s
	}
	if r.PlantedAt != nil {
		s := r.PlantedAt.Format(time.RFC3339)
		resp.PlantedAt = &s
	}

	return resp
}
