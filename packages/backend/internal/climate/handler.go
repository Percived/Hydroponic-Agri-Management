package climate

import (
	"encoding/json"
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

// Handler holds dependencies for climate HTTP handlers.
type Handler struct {
	db *gorm.DB
}

// NewHandler creates a new climate Handler.
func NewHandler(db *gorm.DB) *Handler {
	return &Handler{db: db}
}

// --- ClimateProfile handlers ---

// CreateProfile creates a new climate profile.
func (h *Handler) CreateProfile(c *gin.Context) {
	var req CreateClimateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	profile := ClimateProfile{
		GreenhouseID:      req.GreenhouseID,
		Code:              req.Code,
		Name:              req.Name,
		Description:       req.Description,
		TriggerMetricCode: req.TriggerMetricCode,
	}
	if req.Enabled != nil {
		profile.Enabled = *req.Enabled
	}

	if err := h.db.Create(&profile).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "create_failed", nil)
		return
	}
	response.Success(c, gin.H{"id": profile.ID})
}

// CreateProfileWithStages creates a full climate profile with nested stages and actions.
func (h *Handler) CreateProfileWithStages(c *gin.Context) {
	var req CreateClimateProfileWithStagesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	var profileID uint64
	err := h.db.Transaction(func(tx *gorm.DB) error {
		profile := ClimateProfile{
			GreenhouseID:      req.GreenhouseID,
			Code:              req.Code,
			Name:              req.Name,
			Description:       req.Description,
			TriggerMetricCode: req.TriggerMetricCode,
		}
		if req.Enabled != nil {
			profile.Enabled = *req.Enabled
		}
		if err := tx.Create(&profile).Error; err != nil {
			return err
		}
		profileID = profile.ID

		for _, stageReq := range req.Stages {
			hysteresis := 1.0
			if stageReq.Hysteresis != nil {
				hysteresis = *stageReq.Hysteresis
			}
			stage := ClimateStage{
				ProfileID:        profile.ID,
				StageLevel:       stageReq.StageLevel,
				Name:             stageReq.Name,
				TriggerOperator:  stageReq.TriggerOperator,
				TriggerThreshold: stageReq.TriggerThreshold,
				Hysteresis:       hysteresis,
			}
			if err := tx.Create(&stage).Error; err != nil {
				return err
			}

			for _, actionReq := range stageReq.Actions {
				payloadBytes, err := json.Marshal(actionReq.CommandPayload)
				if err != nil {
					return err
				}
				action := ClimateStageAction{
					StageID:           stage.ID,
					ActuatorChannelID: actionReq.ActuatorChannelID,
					CommandType:       actionReq.CommandType,
					CommandPayload:    string(payloadBytes),
				}
				if actionReq.ExecutionOrder != nil {
					action.ExecutionOrder = *actionReq.ExecutionOrder
				}
				if actionReq.Enabled != nil {
					action.Enabled = *actionReq.Enabled
				}
				if err := tx.Create(&action).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})

	if err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "create_failed", nil)
		return
	}
	response.Success(c, gin.H{"id": profileID})
}

// UpdateProfile updates an existing climate profile.
func (h *Handler) UpdateProfile(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var req UpdateClimateProfileRequest
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
	if req.TriggerMetricCode != nil {
		updates["trigger_metric_code"] = *req.TriggerMetricCode
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}

	if len(updates) == 0 {
		response.Success(c, gin.H{})
		return
	}

	result := h.db.Model(&ClimateProfile{}).Where("id = ?", id).Updates(updates)
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

// DeleteProfile soft-deletes (or hard-deletes) a climate profile and its stages/actions.
func (h *Handler) DeleteProfile(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	err = h.db.Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&ClimateProfile{}).Where("id = ?", id).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return gorm.ErrRecordNotFound
		}

		// Delete actions for all stages of this profile
		var stageIDs []uint64
		if err := tx.Model(&ClimateStage{}).Where("profile_id = ?", id).Pluck("id", &stageIDs).Error; err != nil {
			return err
		}
		if len(stageIDs) > 0 {
			if err := tx.Where("stage_id IN ?", stageIDs).Delete(&ClimateStageAction{}).Error; err != nil {
				return err
			}
		}

		// Delete stages
		if err := tx.Where("profile_id = ?", id).Delete(&ClimateStage{}).Error; err != nil {
			return err
		}

		// Delete execution logs
		if err := tx.Where("profile_id = ?", id).Delete(&ClimateExecutionLog{}).Error; err != nil {
			return err
		}

		// Delete profile
		return tx.Where("id = ?", id).Delete(&ClimateProfile{}).Error
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

// GetProfile retrieves a climate profile with its stages and actions.
func (h *Handler) GetProfile(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var profile ClimateProfile
	if err := h.db.Preload("Stages.Actions").First(&profile, id).Error; err != nil {
		response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
		return
	}

	resp := toProfileResponse(profile)
	response.Success(c, resp)
}

// ListProfiles lists all climate profiles, optionally filtered by greenhouse_id.
func (h *Handler) ListProfiles(c *gin.Context) {
	page, size := parsePage(c)
	q := h.db.Model(&ClimateProfile{})

	if v := c.Query("greenhouse_id"); v != "" {
		q = q.Where("greenhouse_id = ?", v)
	}
	if v := strings.TrimSpace(c.Query("code")); v != "" {
		q = q.Where("code LIKE ?", "%"+v+"%")
	}
	if v := strings.TrimSpace(c.Query("enabled")); v != "" {
		q = q.Where("enabled = ?", v)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	var profiles []ClimateProfile
	if total > 0 {
		if err := q.Order("id desc").Limit(size).Offset((page - 1) * size).Find(&profiles).Error; err != nil {
			response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
			return
		}
	}

	// Count stages per profile
	items := make([]ClimateProfileResponse, 0, len(profiles))
	if len(profiles) > 0 {
		profileIDs := make([]uint64, len(profiles))
		for i, p := range profiles {
			profileIDs[i] = p.ID
		}
		var stageCounts []struct {
			ProfileID uint64 `gorm:"column:profile_id"`
			Count     int    `gorm:"column:count"`
		}
		h.db.Model(&ClimateStage{}).
			Select("profile_id, COUNT(*) as count").
			Where("profile_id IN ?", profileIDs).
			Group("profile_id").
			Scan(&stageCounts)
		countMap := make(map[uint64]int)
		for _, sc := range stageCounts {
			countMap[sc.ProfileID] = sc.Count
		}
		for _, p := range profiles {
			items = append(items, toProfileSummary(p, countMap[p.ID]))
		}
	}

	response.Success(c, gin.H{
		"page":      page,
		"page_size": size,
		"total":     total,
		"items":     items,
	})
}

// ListByGreenhouse is an alias that lists profiles filtered by greenhouse_id.
func (h *Handler) ListByGreenhouse(c *gin.Context) {
	h.ListProfiles(c)
}

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

// --- ClimateExecutionLog handlers ---

// ListExecutionLogs lists execution logs for a profile with time range filter and pagination.
func (h *Handler) ListExecutionLogs(c *gin.Context) {
	profileID, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	page, size := parsePage(c)
	q := h.db.Model(&ClimateExecutionLog{}).Where("profile_id = ?", profileID)

	if v := c.Query("from"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_from", nil)
			return
		}
		q = q.Where("executed_at >= ?", t.UTC())
	}
	if v := c.Query("to"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_to", nil)
			return
		}
		q = q.Where("executed_at <= ?", t.UTC())
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	var logs []ClimateExecutionLog
	if total > 0 {
		if err := q.Order("executed_at desc").Limit(size).Offset((page - 1) * size).Find(&logs).Error; err != nil {
			response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
			return
		}
	}

	// Get profile name for context
	var profileName string
	h.db.Model(&ClimateProfile{}).Select("name").Where("id = ?", profileID).Pluck("name", &profileName)

	items := make([]ClimateExecutionLogResponse, 0, len(logs))
	for _, l := range logs {
		resp := ClimateExecutionLogResponse{
			ID:                   l.ID,
			ProfileID:            l.ProfileID,
			ProfileName:          profileName,
			FromStageLevel:       l.FromStageLevel,
			ToStageLevel:         l.ToStageLevel,
			TriggerValue:         l.TriggerValue,
			ExecutedActionsCount: l.ExecutedActionsCount,
			ExecutedAt:           l.ExecutedAt,
			CreatedAt:            l.CreatedAt,
		}
		items = append(items, resp)
	}

	response.Success(c, gin.H{
		"page":      page,
		"page_size": size,
		"total":     total,
		"items":     items,
	})
}

// ExecuteProfile manually triggers a profile execution and creates a log entry.
func (h *Handler) ExecuteProfile(c *gin.Context) {
	profileID, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var req ExecuteClimateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	// Verify profile exists
	var profile ClimateProfile
	if err := h.db.First(&profile, profileID).Error; err != nil {
		response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "profile_not_found", nil)
		return
	}

	// Count actions for the target stage
	var actionCount int64
	h.db.Model(&ClimateStageAction{}).
		Joins("JOIN climate_stages ON climate_stages.id = climate_stage_actions.stage_id").
		Where("climate_stages.profile_id = ? AND climate_stages.stage_level = ?", profileID, req.ToStageLevel).
		Count(&actionCount)

	log := ClimateExecutionLog{
		ProfileID:            profileID,
		FromStageLevel:       req.FromStageLevel,
		ToStageLevel:         req.ToStageLevel,
		TriggerValue:         req.TriggerValue,
		ExecutedActionsCount: uint(actionCount),
		ExecutedAt:           time.Now().UTC(),
	}

	if err := h.db.Create(&log).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "create_failed", nil)
		return
	}

	response.Success(c, gin.H{
		"id":                     log.ID,
		"to_stage_level":         log.ToStageLevel,
		"executed_actions_count": log.ExecutedActionsCount,
		"executed_at":            log.ExecutedAt.Format(time.RFC3339),
	})
}

// ListAllExecutionLogs lists all execution logs across profiles, with optional filters.
func (h *Handler) ListAllExecutionLogs(c *gin.Context) {
	page, size := parsePage(c)
	q := h.db.Model(&ClimateExecutionLog{})

	if v := c.Query("profile_id"); v != "" {
		q = q.Where("profile_id = ?", v)
	}
	if v := c.Query("greenhouse_id"); v != "" {
		// Subquery to filter by greenhouse
		q = q.Where("profile_id IN (SELECT id FROM climate_profiles WHERE greenhouse_id = ?)", v)
	}
	if v := c.Query("from"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_from", nil)
			return
		}
		q = q.Where("executed_at >= ?", t.UTC())
	}
	if v := c.Query("to"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_to", nil)
			return
		}
		q = q.Where("executed_at <= ?", t.UTC())
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	var logs []ClimateExecutionLog
	if total > 0 {
		if err := q.Order("executed_at desc").Limit(size).Offset((page - 1) * size).Find(&logs).Error; err != nil {
			response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
			return
		}
	}

	// Get profile names
	nameMap := getProfileNames(h.db, logs)

	items := make([]ClimateExecutionLogResponse, 0, len(logs))
	for _, l := range logs {
		resp := ClimateExecutionLogResponse{
			ID:                   l.ID,
			ProfileID:            l.ProfileID,
			ProfileName:          nameMap[l.ProfileID],
			FromStageLevel:       l.FromStageLevel,
			ToStageLevel:         l.ToStageLevel,
			TriggerValue:         l.TriggerValue,
			ExecutedActionsCount: l.ExecutedActionsCount,
			ExecutedAt:           l.ExecutedAt,
			CreatedAt:            l.CreatedAt,
		}
		items = append(items, resp)
	}

	response.Success(c, gin.H{
		"page":      page,
		"page_size": size,
		"total":     total,
		"items":     items,
	})
}

// --- Helper functions ---

// toProfileResponse converts a ClimateProfile with stages to a response struct.
func toProfileResponse(p ClimateProfile) ClimateProfileResponse {
	stages := make([]ClimateStageResponse, 0, len(p.Stages))
	for _, s := range p.Stages {
		stages = append(stages, toStageResponse(s))
	}

	return ClimateProfileResponse{
		ID:                p.ID,
		GreenhouseID:      p.GreenhouseID,
		Code:              p.Code,
		Name:              p.Name,
		Description:       p.Description,
		TriggerMetricCode: p.TriggerMetricCode,
		Enabled:           p.Enabled,
		StagesCount:       len(p.Stages),
		CreatedAt:         p.CreatedAt,
		UpdatedAt:         p.UpdatedAt,
		Stages:            stages,
	}
}

// toProfileSummary converts a ClimateProfile to a summary response (without stages).
func toProfileSummary(p ClimateProfile, stagesCount int) ClimateProfileResponse {
	return ClimateProfileResponse{
		ID:                p.ID,
		GreenhouseID:      p.GreenhouseID,
		Code:              p.Code,
		Name:              p.Name,
		Description:       p.Description,
		TriggerMetricCode: p.TriggerMetricCode,
		Enabled:           p.Enabled,
		StagesCount:       stagesCount,
		CreatedAt:         p.CreatedAt,
		UpdatedAt:         p.UpdatedAt,
	}
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

// getProfileNames returns a map of profile IDs to names.
func getProfileNames(db *gorm.DB, logs []ClimateExecutionLog) map[uint64]string {
	result := make(map[uint64]string)
	if len(logs) == 0 {
		return result
	}

	ids := make([]uint64, 0, len(logs))
	seen := make(map[uint64]bool)
	for _, l := range logs {
		if !seen[l.ProfileID] {
			ids = append(ids, l.ProfileID)
			seen[l.ProfileID] = true
		}
	}

	var profiles []ClimateProfile
	db.Select("id", "name").Where("id IN ?", ids).Find(&profiles)
	for _, p := range profiles {
		result[p.ID] = p.Name
	}
	return result
}

// parseID parses a uint64 from a string parameter.
func parseID(v string) (uint64, error) {
	return strconv.ParseUint(v, 10, 64)
}

// parsePage extracts page and page_size from query parameters.
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

// currentUserID extracts the user ID from the gin context.
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
