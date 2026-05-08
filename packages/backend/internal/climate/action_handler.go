package climate

import (
	"encoding/json"
	"net/http"

	platformErrors "hydroponic-backend/internal/platform/errors"
	"hydroponic-backend/internal/platform/response"

	"github.com/gin-gonic/gin"
)

// --- ClimateStageAction handlers ---

// CreateAction adds an action to a stage.
func (h *Handler) CreateAction(c *gin.Context) {
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

	var req CreateClimateStageActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	// Verify stage exists and belongs to profile
	var stageCount int64
	if err := h.db.Model(&ClimateStage{}).Where("id = ? AND profile_id = ?", stageID, profileID).Count(&stageCount).Error; err != nil || stageCount == 0 {
		response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "stage_not_found", nil)
		return
	}

	payloadBytes, err := json.Marshal(req.CommandPayload)
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_payload", nil)
		return
	}

	action := ClimateStageAction{
		StageID:           stageID,
		ActuatorChannelID: req.ActuatorChannelID,
		CommandType:       req.CommandType,
		CommandPayload:    string(payloadBytes),
	}
	if req.ExecutionOrder != nil {
		action.ExecutionOrder = *req.ExecutionOrder
	}
	if req.Enabled != nil {
		action.Enabled = *req.Enabled
	}

	if err := h.db.Create(&action).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "create_failed", nil)
		return
	}
	response.Success(c, gin.H{"id": action.ID})
	h.pushProfileConfig(profileID, "update")
}

// UpdateAction updates an existing stage action.
func (h *Handler) UpdateAction(c *gin.Context) {
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
	actionID, err := parseID(c.Param("actionId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_action_id", nil)
		return
	}

	var req UpdateClimateStageActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	// Verify action exists and belongs to the stage/profile
	var actionCount int64
	if err := h.db.Model(&ClimateStageAction{}).
		Joins("JOIN climate_stages ON climate_stages.id = climate_stage_actions.stage_id").
		Where("climate_stage_actions.id = ? AND climate_stage_actions.stage_id = ? AND climate_stages.profile_id = ?", actionID, stageID, profileID).
		Count(&actionCount).Error; err != nil || actionCount == 0 {
		response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
		return
	}

	updates := map[string]interface{}{}
	if req.ActuatorChannelID != nil {
		updates["actuator_channel_id"] = *req.ActuatorChannelID
	}
	if req.CommandType != nil {
		updates["command_type"] = *req.CommandType
	}
	if req.CommandPayload != nil {
		payloadBytes, err := json.Marshal(req.CommandPayload)
		if err != nil {
			response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_payload", nil)
			return
		}
		updates["command_payload"] = string(payloadBytes)
	}
	if req.ExecutionOrder != nil {
		updates["execution_order"] = *req.ExecutionOrder
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}

	if len(updates) == 0 {
		response.Success(c, gin.H{})
		return
	}

	if err := h.db.Model(&ClimateStageAction{}).Where("id = ?", actionID).Updates(updates).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "update_failed", nil)
		return
	}
	response.Success(c, gin.H{})
	h.pushProfileConfig(profileID, "update")
}

// DeleteAction deletes a stage action.
func (h *Handler) DeleteAction(c *gin.Context) {
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
	actionID, err := parseID(c.Param("actionId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_action_id", nil)
		return
	}

	// Verify action belongs to the correct stage under the correct profile
	var count int64
	if err := h.db.Model(&ClimateStageAction{}).
		Joins("JOIN climate_stages ON climate_stages.id = climate_stage_actions.stage_id").
		Where("climate_stage_actions.id = ? AND climate_stage_actions.stage_id = ? AND climate_stages.profile_id = ?", actionID, stageID, profileID).
		Count(&count).Error; err != nil || count == 0 {
		response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
		return
	}

	result := h.db.Where("id = ? AND stage_id = ?", actionID, stageID).Delete(&ClimateStageAction{})
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

// ListActions lists all actions for a stage.
func (h *Handler) ListActions(c *gin.Context) {
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

	// Verify stage belongs to profile
	var stageCount int64
	if err := h.db.Model(&ClimateStage{}).Where("id = ? AND profile_id = ?", stageID, profileID).Count(&stageCount).Error; err != nil || stageCount == 0 {
		response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "stage_not_found", nil)
		return
	}

	var actions []ClimateStageAction
	if err := h.db.Where("stage_id = ?", stageID).Order("execution_order ASC").Find(&actions).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	items := make([]ClimateStageActionResponse, 0, len(actions))
	for _, a := range actions {
		items = append(items, ClimateStageActionResponse{
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

	response.Success(c, gin.H{"items": items})
}
