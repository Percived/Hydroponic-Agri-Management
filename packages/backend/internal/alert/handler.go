package alert

import (
	"net/http"
	"strconv"
	"time"

	"hydroponic-backend/internal/auth"
	platformErrors "hydroponic-backend/internal/platform/errors"
	"hydroponic-backend/internal/platform/event"
	"hydroponic-backend/internal/platform/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	db       *gorm.DB
	eventHub *event.Hub
}

func NewHandler(db *gorm.DB, hub *event.Hub) *Handler {
	return &Handler{db: db, eventHub: hub}
}

// CreateAlert creates a new alert.
func (h *Handler) CreateAlert(c *gin.Context) {
	var req CreateAlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	alert := Alert{
		Type:              req.Type,
		Level:             req.Level,
		MetricCode:        req.MetricCode,
		SensorChannelID:   req.SensorChannelID,
		ActuatorChannelID: req.ActuatorChannelID,
		TriggerValue:      req.TriggerValue,
		Message:           req.Message,
		Status:            StatusOpen,
		TriggeredAt:       req.TriggeredAt,
	}

	if err := h.db.Create(&alert).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "create_failed", nil)
		return
	}

	// Auto-create a "TRIGGERED" timeline event
	timeline := AlertTimelineEvent{
		AlertID:     alert.ID,
		EventType:   EventTriggered,
		EventSource: SourceSystem,
		EventTime:   req.TriggeredAt,
	}
	h.db.Create(&timeline)

	// Publish event for real-time notification
	if h.eventHub != nil {
		h.eventHub.Publish(event.SSEEvent{Type: "alert:created", Data: BuildAlertSSEDataV1(alert, "", 1)})
	}

	response.Success(c, gin.H{"id": alert.ID})
}

// ListAlerts returns a paginated list of alerts with filters.
func (h *Handler) ListAlerts(c *gin.Context) {
	page, pageSize := parsePage(c)

	query := h.db.Model(&Alert{})
	if v := c.Query("status"); v != "" {
		query = query.Where("status = ?", v)
	}
	if v := c.Query("type"); v != "" {
		query = query.Where("type = ?", v)
	}
	if v := c.Query("level"); v != "" {
		query = query.Where("level = ?", v)
	}
	if v := c.Query("sensor_channel_id"); v != "" {
		id, err := strconv.ParseUint(v, 10, 64)
		if err == nil {
			query = query.Where("sensor_channel_id = ?", id)
		}
	}
	if v := c.Query("actuator_channel_id"); v != "" {
		id, err := strconv.ParseUint(v, 10, 64)
		if err == nil {
			query = query.Where("actuator_channel_id = ?", id)
		}
	}
	if v := c.Query("batch_id"); v != "" {
		id, err := strconv.ParseUint(v, 10, 64)
		if err == nil {
			query = query.Where("batch_id = ?", id)
		}
	}
	if from := c.Query("triggered_from"); from != "" {
		t, err := time.Parse(time.RFC3339, from)
		if err == nil {
			query = query.Where("triggered_at >= ?", t)
		}
	}
	if to := c.Query("triggered_to"); to != "" {
		t, err := time.Parse(time.RFC3339, to)
		if err == nil {
			query = query.Where("triggered_at <= ?", t)
		}
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	var alerts []Alert
	if total > 0 {
		if err := query.Order("triggered_at DESC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&alerts).Error; err != nil {
			response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
			return
		}
	}

	// Gather alert IDs to batch-fetch timeline counts
	alertIDs := make([]uint64, len(alerts))
	for i, a := range alerts {
		alertIDs[i] = a.ID
	}

	var timelineCounts []struct {
		AlertID uint64 `gorm:"column:alert_id"`
		Count   int64  `gorm:"column:count"`
	}
	timelineCountMap := map[uint64]int64{}
	if len(alertIDs) > 0 {
		h.db.Model(&AlertTimelineEvent{}).
			Select("alert_id, COUNT(*) as count").
			Where("alert_id IN ?", alertIDs).
			Group("alert_id").
			Scan(&timelineCounts)
		for _, tc := range timelineCounts {
			timelineCountMap[tc.AlertID] = tc.Count
		}
	}

	items := make([]gin.H, 0, len(alerts))
	for _, a := range alerts {
		items = append(items, alertToItem(a, timelineCountMap[a.ID]))
	}

	response.Success(c, gin.H{
		"page":      page,
		"page_size": pageSize,
		"total":     total,
		"items":     items,
	})
}

// GetAlert returns a single alert with its timeline events.
func (h *Handler) GetAlert(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var alert Alert
	if err := h.db.Preload("TimelineEvents", func(db *gorm.DB) *gorm.DB {
		return db.Order("event_time ASC")
	}).First(&alert, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	timelineCount := int64(len(alert.TimelineEvents))
	item := alertToItem(alert, timelineCount)
	item["timeline_events"] = timelineEventsToItems(alert.TimelineEvents)

	response.Success(c, item)
}

// UpdateAlertStatus updates an alert's status.
func (h *Handler) UpdateAlertStatus(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var req UpdateAlertStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	now := time.Now().UTC()
	updates := map[string]interface{}{
		"status": req.Status,
	}

	if req.Status == StatusResolved {
		if req.ResolvedAt != nil {
			updates["resolved_at"] = *req.ResolvedAt
		} else {
			updates["resolved_at"] = now
		}
		updates["resolved_by"] = currentUserID(c)
	} else {
		updates["resolved_at"] = nil
		updates["resolved_by"] = nil
	}

	err = h.db.Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&Alert{}).Where("id = ?", id).Updates(updates)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		// Determine event type based on status change
		eventType := ""
		switch req.Status {
		case StatusAcknowledged:
			eventType = EventAcknowledged
		case StatusResolved:
			eventType = EventResolved
		default:
			eventType = EventComment
		}

		operatorID := currentUserID(c)
		timeline := AlertTimelineEvent{
			AlertID:     id,
			EventType:   eventType,
			EventSource: SourceManual,
			OperatorID:  &operatorID,
			Comment:     req.Comment,
			EventTime:   now,
		}
		return tx.Create(&timeline).Error
	})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "update_failed", nil)
		return
	}

	response.Success(c, gin.H{})
}

// GetAlertTimeline returns all timeline events for an alert.
func (h *Handler) GetAlertTimeline(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	// Verify alert exists
	var count int64
	if err := h.db.Model(&Alert{}).Where("id = ?", id).Count(&count).Error; err != nil || count == 0 {
		response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
		return
	}

	var events []AlertTimelineEvent
	if err := h.db.Where("alert_id = ?", id).Order("event_time ASC, id ASC").Find(&events).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	items := timelineEventsToItems(events)
	response.Success(c, gin.H{"items": items})
}

// AddTimelineEvent adds a timeline event to an alert.
func (h *Handler) AddTimelineEvent(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var req CreateTimelineEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	// Verify alert exists
	var count int64
	if err := h.db.Model(&Alert{}).Where("id = ?", id).Count(&count).Error; err != nil || count == 0 {
		response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
		return
	}

	event := AlertTimelineEvent{
		AlertID:      id,
		EventType:    req.EventType,
		EventSource:  req.EventSource,
		OperatorID:   req.OperatorID,
		Comment:      req.Comment,
		EventPayload: req.EventPayload,
		EventTime:    req.EventTime,
	}

	if err := h.db.Create(&event).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "create_failed", nil)
		return
	}

	response.Success(c, gin.H{"id": event.ID})
}

// GetAlertStats returns aggregated alert statistics by status and level.
func (h *Handler) GetAlertStats(c *gin.Context) {
	var stats struct {
		OpenCount         int64 `gorm:"column:open_count"`
		AcknowledgedCount int64 `gorm:"column:acknowledged_count"`
		ResolvedCount     int64 `gorm:"column:resolved_count"`
		IgnoredCount      int64 `gorm:"column:ignored_count"`
		InfoCount         int64 `gorm:"column:info_count"`
		WarnCount         int64 `gorm:"column:warn_count"`
		CriticalCount     int64 `gorm:"column:critical_count"`
	}

	// Count by status
	if err := h.db.Model(&Alert{}).Where("status = ?", StatusOpen).Count(&stats.OpenCount).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}
	if err := h.db.Model(&Alert{}).Where("status = ?", StatusAcknowledged).Count(&stats.AcknowledgedCount).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}
	if err := h.db.Model(&Alert{}).Where("status = ?", StatusResolved).Count(&stats.ResolvedCount).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}
	if err := h.db.Model(&Alert{}).Where("status = ?", StatusIgnored).Count(&stats.IgnoredCount).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	// Count by level
	if err := h.db.Model(&Alert{}).Where("level = ?", LevelInfo).Count(&stats.InfoCount).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}
	if err := h.db.Model(&Alert{}).Where("level = ?", LevelWarn).Count(&stats.WarnCount).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}
	if err := h.db.Model(&Alert{}).Where("level = ?", LevelCritical).Count(&stats.CriticalCount).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	response.Success(c, AlertStatsResponse{
		OpenCount:         stats.OpenCount,
		AcknowledgedCount: stats.AcknowledgedCount,
		ResolvedCount:     stats.ResolvedCount,
		IgnoredCount:      stats.IgnoredCount,
		InfoCount:         stats.InfoCount,
		WarnCount:         stats.WarnCount,
		CriticalCount:     stats.CriticalCount,
	})
}

// --- Helpers ---

func alertToItem(a Alert, timelineCount int64) gin.H {
	return gin.H{
		"id":                  a.ID,
		"type":                a.Type,
		"level":               a.Level,
		"metric_code":         a.MetricCode,
		"sensor_channel_id":   a.SensorChannelID,
		"actuator_channel_id": a.ActuatorChannelID,
		"batch_id":            a.BatchID,
		"trigger_value":       a.TriggerValue,
		"message":             a.Message,
		"status":              a.Status,
		"triggered_at":        timeToStr(a.TriggeredAt),
		"resolved_at":         timePtrToStr(a.ResolvedAt),
		"resolved_by":         a.ResolvedBy,
		"timeline_count":      timelineCount,
		"created_at":          timeToStr(a.CreatedAt),
		"updated_at":          timeToStr(a.UpdatedAt),
	}
}

func timelineEventsToItems(events []AlertTimelineEvent) []gin.H {
	items := make([]gin.H, 0, len(events))
	for _, e := range events {
		items = append(items, gin.H{
			"id":            e.ID,
			"alert_id":      e.AlertID,
			"event_type":    e.EventType,
			"event_source":  e.EventSource,
			"operator_id":   e.OperatorID,
			"comment":       e.Comment,
			"event_payload": e.EventPayload,
			"event_time":    timeToStr(e.EventTime),
			"created_at":    timeToStr(e.CreatedAt),
		})
	}
	return items
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
