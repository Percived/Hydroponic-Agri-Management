package alert

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"hydroponic-backend/internal/audit"
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

func (h *Handler) List(c *gin.Context) {
	page, size := parsePage(c)
	q := h.db.Model(&Alert{})
	if v := strings.TrimSpace(c.Query("type")); v != "" {
		q = q.Where("type = ?", v)
	}
	if v := strings.TrimSpace(c.Query("level")); v != "" {
		q = q.Where("level = ?", v)
	}
	if v := strings.TrimSpace(c.Query("status")); v != "" {
		q = q.Where("status = ?", v)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	alerts := []Alert{}
	if total > 0 {
		if err := q.Order("triggered_at desc").Limit(size).Offset((page - 1) * size).Find(&alerts).Error; err != nil {
			response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
			return
		}
	}

	items := make([]gin.H, 0, len(alerts))
	for _, a := range alerts {
		items = append(items, toItem(a))
	}
	response.Success(c, gin.H{"page": page, "page_size": size, "total": total, "items": items})
}

func (h *Handler) Get(c *gin.Context) {
	id, err := parseID(c.Param("alertId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}
	var a Alert
	if err := h.db.First(&a, id).Error; err != nil {
		response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
		return
	}
	response.Success(c, toItem(a))
}

func (h *Handler) UpdateStatus(c *gin.Context) {
	id, err := parseID(c.Param("alertId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}
	var req UpdateAlertStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	updates := map[string]interface{}{"status": req.Status}
	if req.Status == StatusClosed {
		now := time.Now().UTC()
		updates["resolved_at"] = now
	} else {
		updates["resolved_at"] = nil
	}

	if err := h.db.Model(&Alert{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "update_failed", nil)
		return
	}
	audit.Write(h.db, currentUserID(c), "UPDATE_ALERT_STATUS", "alerts", &id, req)
	response.Success(c, gin.H{})
}

func (h *Handler) Stats(c *gin.Context) {
	var openCount, ackCount, closedCount int64
	if err := h.db.Model(&Alert{}).Where("status = ?", StatusOpen).Count(&openCount).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}
	if err := h.db.Model(&Alert{}).Where("status = ?", StatusAck).Count(&ackCount).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}
	if err := h.db.Model(&Alert{}).Where("status = ?", StatusClosed).Count(&closedCount).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}
	response.Success(c, gin.H{"open": openCount, "ack": ackCount, "closed": closedCount})
}

func (h *Handler) Subscribe(c *gin.Context) {
	response.Success(c, gin.H{"url": "/api/alerts/subscribe"})
}

func toItem(a Alert) gin.H {
	return gin.H{
		"id":           a.ID,
		"type":         a.Type,
		"level":        a.Level,
		"metric_id":    a.MetricID,
		"device_id":    a.DeviceID,
		"value":        a.Value,
		"message":      a.Message,
		"status":       a.Status,
		"triggered_at": a.TriggeredAt,
		"resolved_at":  a.ResolvedAt,
	}
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
