package climate

import (
	"net/http"

	platformErrors "hydroponic-backend/internal/platform/errors"
	"hydroponic-backend/internal/platform/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// --- ClimateStage handlers ---

// CreateStage adds a stage to an existing climate profile.
func (h *Handler) CreateStage(c *gin.Context) {
	profileID, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var req CreateClimateStageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	// Verify profile exists
	var profileCount int64
	if err := h.db.Model(&ClimateProfile{}).Where("id = ?", profileID).Count(&profileCount).Error; err != nil || profileCount == 0 {
		response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "profile_not_found", nil)
		return
	}

	hysteresis := 1.0
	if req.Hysteresis != nil {
		hysteresis = *req.Hysteresis
	}

	stage := ClimateStage{
		ProfileID:        profileID,
		StageLevel:       req.StageLevel,
		Name:             req.Name,
		TriggerOperator:  req.TriggerOperator,
		TriggerThreshold: req.TriggerThreshold,
		Hysteresis:       hysteresis,
	}

	if err := h.db.Create(&stage).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "create_failed", nil)
		return
	}
	response.Success(c, gin.H{"id": stage.ID})
}

// UpdateStage updates an existing climate stage.
func (h *Handler) UpdateStage(c *gin.Context) {
	profileID, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}
	stageID, err := parseID(c.Param("stageId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_stage_id", nil)
		return
	}

	var req UpdateClimateStageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.TriggerOperator != nil {
		updates["trigger_operator"] = *req.TriggerOperator
	}
	if req.TriggerThreshold != nil {
		updates["trigger_threshold"] = *req.TriggerThreshold
	}
	if req.Hysteresis != nil {
		updates["hysteresis"] = *req.Hysteresis
	}

	if len(updates) == 0 {
		response.Success(c, gin.H{})
		return
	}

	result := h.db.Model(&ClimateStage{}).Where("id = ? AND profile_id = ?", stageID, profileID).Updates(updates)
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

// DeleteStage deletes a stage and its actions.
func (h *Handler) DeleteStage(c *gin.Context) {
	profileID, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}
	stageID, err := parseID(c.Param("stageId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_stage_id", nil)
		return
	}

	err = h.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Where("id = ? AND profile_id = ?", stageID, profileID).Delete(&ClimateStage{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return tx.Where("stage_id = ?", stageID).Delete(&ClimateStageAction{}).Error
	})

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "delete_failed", nil)
		return
	}
	response.Success(c, gin.H{})
}

// ListStages lists all stages for a profile.
func (h *Handler) ListStages(c *gin.Context) {
	profileID, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var stages []ClimateStage
	if err := h.db.Preload("Actions").Where("profile_id = ?", profileID).Order("stage_level asc").Find(&stages).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	items := make([]ClimateStageResponse, 0, len(stages))
	for _, s := range stages {
		items = append(items, toStageResponse(s))
	}
	response.Success(c, gin.H{"items": items})
}

// GetStage retrieves a single stage with its actions.
func (h *Handler) GetStage(c *gin.Context) {
	profileID, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}
	stageID, err := parseID(c.Param("stageId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_stage_id", nil)
		return
	}

	var stage ClimateStage
	if err := h.db.Preload("Actions").Where("id = ? AND profile_id = ?", stageID, profileID).First(&stage).Error; err != nil {
		response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
		return
	}
	response.Success(c, toStageResponse(stage))
}

// toStageResponse converts a ClimateStage with actions to a response struct.
func toStageResponse(s ClimateStage) ClimateStageResponse {
	actions := make([]ClimateStageActionResponse, 0, len(s.Actions))
	for _, a := range s.Actions {
		actions = append(actions, ClimateStageActionResponse{
			ID:                a.ID,
			StageID:           a.StageID,
			ActuatorChannelID: a.ActuatorChannelID,
			CommandType:       a.CommandType,
			CommandPayload:    a.CommandPayload,
			ExecutionOrder:    a.ExecutionOrder,
			Enabled:           a.Enabled,
			CreatedAt:         a.CreatedAt,
			UpdatedAt:         a.UpdatedAt,
		})
	}

	return ClimateStageResponse{
		ID:               s.ID,
		ProfileID:        s.ProfileID,
		StageLevel:       s.StageLevel,
		Name:             s.Name,
		TriggerOperator:  s.TriggerOperator,
		TriggerThreshold: s.TriggerThreshold,
		Hysteresis:       s.Hysteresis,
		ActionCount:      len(s.Actions),
		CreatedAt:        s.CreatedAt,
		UpdatedAt:        s.UpdatedAt,
		Actions:          actions,
	}
}
