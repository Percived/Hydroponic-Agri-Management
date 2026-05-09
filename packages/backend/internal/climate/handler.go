package climate

import (
	"log/slog"
	"net/http"
	"time"

	platformErrors "hydroponic-backend/internal/platform/errors"
	"hydroponic-backend/internal/platform/event"
	"hydroponic-backend/internal/platform/mqtt"
	"hydroponic-backend/internal/platform/response"

	mqttlib "github.com/eclipse/paho.mqtt.golang"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Handler holds dependencies for climate HTTP handlers.
type Handler struct {
	db           *gorm.DB
	configPusher *mqtt.ConfigPusher
	hub          *event.Hub
	log          *slog.Logger
}

// NewHandler creates a new climate Handler.
func NewHandler(db *gorm.DB, mqttClient mqttlib.Client, hub *event.Hub, log *slog.Logger) *Handler {
	return &Handler{
		db:           db,
		configPusher: mqtt.NewConfigPusher(db, mqttClient, log),
		hub:          hub,
		log:          log,
	}
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
			ID:                     l.ID,
			ProfileID:              l.ProfileID,
			ProfileName:            profileName,
			FromStageLevel:         l.FromStageLevel,
			ToStageLevel:           l.ToStageLevel,
			TriggerValue:           l.TriggerValue,
			TriggerSensorChannelID: l.TriggerSensorChannelID,
			TriggerMetricCode:      l.TriggerMetricCode,
			CollectedAt:            l.CollectedAt,
			ExecutedActionsCount:   l.ExecutedActionsCount,
			ExecutedAt:             l.ExecutedAt,
			CreatedAt:              l.CreatedAt,
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
		ProfileID:              profileID,
		FromStageLevel:         req.FromStageLevel,
		ToStageLevel:           req.ToStageLevel,
		TriggerValue:           req.TriggerValue,
		TriggerSensorChannelID: profile.TriggerSensorChannelID,
		TriggerMetricCode:      &profile.TriggerMetricCode,
		ExecutedActionsCount:   uint(actionCount),
		ExecutedAt:             time.Now().UTC(),
	}
	log.CollectedAt = &log.ExecutedAt

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
			ID:                     l.ID,
			ProfileID:              l.ProfileID,
			ProfileName:            nameMap[l.ProfileID],
			FromStageLevel:         l.FromStageLevel,
			ToStageLevel:           l.ToStageLevel,
			TriggerValue:           l.TriggerValue,
			TriggerSensorChannelID: l.TriggerSensorChannelID,
			TriggerMetricCode:      l.TriggerMetricCode,
			CollectedAt:            l.CollectedAt,
			ExecutedActionsCount:   l.ExecutedActionsCount,
			ExecutedAt:             l.ExecutedAt,
			CreatedAt:              l.CreatedAt,
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
