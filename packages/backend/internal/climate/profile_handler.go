package climate

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"hydroponic-backend/internal/auth"
	platformErrors "hydroponic-backend/internal/platform/errors"
	"hydroponic-backend/internal/platform/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// --- ClimateProfile handlers ---

// CreateProfile creates a new climate profile.
func (h *Handler) CreateProfile(c *gin.Context) {
	var req CreateClimateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	if err := validateTriggerSource(h.db, req.GreenhouseID, req.TriggerMetricCode, req.TriggerSensorChannelID); err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_trigger_source", nil)
		return
	}

	profile := ClimateProfile{
		GreenhouseID:      req.GreenhouseID,
		Code:              req.Code,
		Name:              req.Name,
		Description:       req.Description,
		TriggerMetricCode: req.TriggerMetricCode,
	}
	profile.TriggerSensorChannelID = &req.TriggerSensorChannelID
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
		if err := validateTriggerSource(tx, req.GreenhouseID, req.TriggerMetricCode, req.TriggerSensorChannelID); err != nil {
			return err
		}
		if len(req.Stages) > 0 {
			toValidate := make([]ClimateStage, 0, len(req.Stages))
			for _, s := range req.Stages {
				hysteresis := 1.0
				if s.Hysteresis != nil {
					hysteresis = *s.Hysteresis
				}
				toValidate = append(toValidate, ClimateStage{
					StageLevel:       s.StageLevel,
					TriggerOperator:  s.TriggerOperator,
					TriggerThreshold: s.TriggerThreshold,
					Hysteresis:       hysteresis,
				})
			}
			if err := validateStageSet(toValidate); err != nil {
				return err
			}
		}

		profile := ClimateProfile{
			GreenhouseID:      req.GreenhouseID,
			Code:              req.Code,
			Name:              req.Name,
			Description:       req.Description,
			TriggerMetricCode: req.TriggerMetricCode,
		}
		profile.TriggerSensorChannelID = &req.TriggerSensorChannelID
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
	h.pushProfileConfig(profileID, "create")
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

	var existing ClimateProfile
	if err := h.db.First(&existing, id).Error; err != nil {
		response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
		return
	}

	nextMetric := existing.TriggerMetricCode
	if req.TriggerMetricCode != nil {
		nextMetric = *req.TriggerMetricCode
	}
	nextChannelID := existing.TriggerSensorChannelID
	if req.TriggerSensorChannelID != nil {
		nextChannelID = req.TriggerSensorChannelID
	}
	enableRequested := req.Enabled != nil && *req.Enabled
	triggerChanged := req.TriggerMetricCode != nil || req.TriggerSensorChannelID != nil
	if enableRequested || triggerChanged {
		if nextChannelID == nil {
			response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_trigger_source", nil)
			return
		}
		if err := validateTriggerSource(h.db, existing.GreenhouseID, nextMetric, *nextChannelID); err != nil {
			response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_trigger_source", nil)
			return
		}
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
	if req.TriggerSensorChannelID != nil {
		updates["trigger_sensor_channel_id"] = *req.TriggerSensorChannelID
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
	h.pushProfileConfig(id, "update")
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
	h.pushProfileConfig(id, "delete")
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

	resp := BuildProfileConfigPayload(profile)
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

// toProfileResponse converts a ClimateProfile with stages to a response struct.
func toProfileResponse(p ClimateProfile) ClimateProfileResponse {
	return BuildProfileConfigPayload(p)
}

// toProfileSummary converts a ClimateProfile to a summary response (without stages).
func toProfileSummary(p ClimateProfile, stagesCount int) ClimateProfileResponse {
	return ClimateProfileResponse{
		ID:                     p.ID,
		GreenhouseID:           p.GreenhouseID,
		Code:                   p.Code,
		Name:                   p.Name,
		Description:            p.Description,
		TriggerMetricCode:      p.TriggerMetricCode,
		TriggerSensorChannelID: p.TriggerSensorChannelID,
		Enabled:                p.Enabled,
		StagesCount:            stagesCount,
		CreatedAt:              p.CreatedAt,
		UpdatedAt:              p.UpdatedAt,
	}
}

func validateTriggerSource(db *gorm.DB, greenhouseID uint64, triggerMetricCode string, triggerSensorChannelID uint64) error {
	var row struct {
		MetricCode   string `gorm:"column:metric_code"`
		GreenhouseID uint64 `gorm:"column:greenhouse_id"`
		Enabled      bool   `gorm:"column:enabled"`
	}
	if err := db.Table("sensor_channels sc").
		Select("sc.metric_code, sd.greenhouse_id, sc.enabled").
		Joins("JOIN sensor_devices sd ON sd.id = sc.sensor_device_id").
		Where("sc.id = ?", triggerSensorChannelID).
		Scan(&row).Error; err != nil {
		return err
	}
	if row.MetricCode == "" || row.GreenhouseID == 0 {
		return gorm.ErrRecordNotFound
	}
	if !row.Enabled {
		return gorm.ErrInvalidData
	}
	if row.GreenhouseID != greenhouseID {
		return gorm.ErrInvalidData
	}
	if row.MetricCode != triggerMetricCode {
		return gorm.ErrInvalidData
	}
	return nil
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

func (h *Handler) pushProfileConfig(profileID uint64, action string) {
	if h.configPusher == nil {
		return
	}
	go func() {
		var profile ClimateProfile
		if err := h.db.Preload("Stages.Actions").First(&profile, profileID).Error; err != nil {
			h.log.Warn("climate: failed to load profile for config push", "profile_id", profileID, "error", err)
			return
		}
		payload := BuildProfileConfigPayload(profile)

		type channelRow struct {
			ActuatorChannelID uint64
		}
		var channels []channelRow
		if err := h.db.Table("climate_stage_actions").
			Select("actuator_channel_id").
			Joins("JOIN climate_stages ON climate_stages.id = climate_stage_actions.stage_id").
			Where("climate_stages.profile_id = ?", profileID).
			Group("actuator_channel_id").
			Find(&channels).Error; err != nil {
			h.log.Warn("climate: failed to load channels for config push", "profile_id", profileID, "error", err)
			return
		}
		for _, ch := range channels {
			_ = h.configPusher.PushToActuatorChannel(ch.ActuatorChannelID, "climate_profile", action, profileID, payload)
		}
	}()
}
