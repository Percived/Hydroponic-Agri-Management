package device

import (
	"errors"
	"net/http"
	"strconv"
	"time"

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

func (h *Handler) CreateDevice(c *gin.Context) {
	var req CreateDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	var count int64
	if err := h.db.Model(&Device{}).Where("device_code = ?", req.DeviceCode).Count(&count).Error; err == nil && count > 0 {
		response.Error(c, http.StatusConflict, platformErrors.CodeDeviceCodeRepeat, "device_code_exists", nil)
		return
	}

	device := Device{
		DeviceCode: req.DeviceCode,
		Name:       req.Name,
		Type:       req.Type,
		Category:   req.Category,
		Protocol:   req.Protocol,
		Status:     DeviceStatusEnabled,
	}
	if req.GreenhouseID != nil {
		device.GreenhouseID = req.GreenhouseID
	}
	if req.GroupID != nil {
		device.GroupID = req.GroupID
	}
	if req.SamplingIntervalSec != nil {
		device.SamplingIntervalSec = *req.SamplingIntervalSec
	}

	if err := h.db.Create(&device).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "create_failed", nil)
		return
	}
	response.Success(c, gin.H{"id": device.ID})
}

func (h *Handler) UpdateDevice(c *gin.Context) {
	deviceID, err := parseID(c.Param("deviceId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var req UpdateDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Category != nil {
		updates["category"] = *req.Category
	}
	if req.GreenhouseID != nil {
		updates["greenhouse_id"] = *req.GreenhouseID
	}
	if req.GroupID != nil {
		updates["group_id"] = *req.GroupID
	}
	if req.SamplingIntervalSec != nil {
		updates["sampling_interval_sec"] = *req.SamplingIntervalSec
	}

	if len(updates) > 0 {
		if err := h.db.Model(&Device{}).Where("id = ?", deviceID).Updates(updates).Error; err != nil {
			response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "update_failed", nil)
			return
		}
	}
	response.Success(c, gin.H{})
}

func (h *Handler) UpdateDeviceStatus(c *gin.Context) {
	deviceID, err := parseID(c.Param("deviceId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var req UpdateDeviceStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	if err := h.db.Model(&Device{}).Where("id = ?", deviceID).Update("status", req.Status).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "update_failed", nil)
		return
	}
	response.Success(c, gin.H{})
}

func (h *Handler) ListDevices(c *gin.Context) {
	page, pageSize := parsePage(c)

	query := h.db.Model(&Device{})
	if v := c.Query("type"); v != "" {
		query = query.Where("type = ?", v)
	}
	if v := c.Query("category"); v != "" {
		query = query.Where("category = ?", v)
	}
	if v := c.Query("group_id"); v != "" {
		query = query.Where("group_id = ?", v)
	}
	if v := c.Query("greenhouse_id"); v != "" {
		query = query.Where("greenhouse_id = ?", v)
	}
	if v := c.Query("status"); v != "" {
		query = query.Where("status = ?", v)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	var devices []Device
	if total > 0 {
		if err := query.Order("id desc").Limit(pageSize).Offset((page - 1) * pageSize).Find(&devices).Error; err != nil {
			response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
			return
		}
	}

	items := make([]gin.H, 0, len(devices))
	for _, d := range devices {
		items = append(items, gin.H{
			"id":                    d.ID,
			"device_code":           d.DeviceCode,
			"name":                  d.Name,
			"type":                  d.Type,
			"category":              d.Category,
			"status":                d.Status,
			"protocol":              d.Protocol,
			"greenhouse_id":         d.GreenhouseID,
			"group_id":              d.GroupID,
			"sampling_interval_sec": d.SamplingIntervalSec,
			"last_seen_at":          d.LastSeenAt,
			"created_at":            d.CreatedAt,
			"updated_at":            d.UpdatedAt,
		})
	}

	response.Success(c, gin.H{
		"page":      page,
		"page_size": pageSize,
		"total":     total,
		"items":     items,
	})
}

func (h *Handler) GetDevice(c *gin.Context) {
	deviceID, err := parseID(c.Param("deviceId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var device Device
	if err := h.db.First(&device, deviceID).Error; err != nil {
		response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
		return
	}

	response.Success(c, gin.H{
		"id":                    device.ID,
		"device_code":           device.DeviceCode,
		"name":                  device.Name,
		"type":                  device.Type,
		"category":              device.Category,
		"greenhouse_id":         device.GreenhouseID,
		"group_id":              device.GroupID,
		"status":                device.Status,
		"protocol":              device.Protocol,
		"sampling_interval_sec": device.SamplingIntervalSec,
		"last_seen_at":          device.LastSeenAt,
		"created_at":            device.CreatedAt,
		"updated_at":            device.UpdatedAt,
	})
}

func (h *Handler) DeviceHealth(c *gin.Context) {
	deviceID, err := parseID(c.Param("deviceId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var device Device
	if err := h.db.First(&device, deviceID).Error; err != nil {
		response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
		return
	}

	online := false
	if device.LastSeenAt != nil {
		threshold := time.Duration(device.SamplingIntervalSec*3) * time.Second
		online = time.Since(*device.LastSeenAt) <= threshold
	}

	response.Success(c, gin.H{
		"device_id":    device.ID,
		"online":       online,
		"last_seen_at": device.LastSeenAt,
	})
}

func (h *Handler) CreateGreenhouse(c *gin.Context) {
	var req CreateGreenhouseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	greenhouse := Greenhouse{
		Name: req.Name,
	}
	if req.Location != nil {
		greenhouse.Location = *req.Location
	}
	if req.Description != nil {
		greenhouse.Description = *req.Description
	}

	if err := h.db.Create(&greenhouse).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "create_failed", nil)
		return
	}
	response.Success(c, gin.H{"id": greenhouse.ID})
}

func (h *Handler) UpdateGreenhouse(c *gin.Context) {
	greenhouseID, err := parseID(c.Param("greenhouseId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var req UpdateGreenhouseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Location != nil {
		updates["location"] = *req.Location
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if len(updates) > 0 {
		if err := h.db.Model(&Greenhouse{}).Where("id = ?", greenhouseID).Updates(updates).Error; err != nil {
			response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "update_failed", nil)
			return
		}
	}
	response.Success(c, gin.H{})
}

func (h *Handler) ListGreenhouses(c *gin.Context) {
	page, pageSize := parsePage(c)

	query := h.db.Model(&Greenhouse{})
	if v := c.Query("keyword"); v != "" {
		query = query.Where("name LIKE ?", "%"+v+"%")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	var greenhouses []Greenhouse
	if total > 0 {
		if err := query.Order("id desc").Limit(pageSize).Offset((page - 1) * pageSize).Find(&greenhouses).Error; err != nil {
			response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
			return
		}
	}

	items := make([]gin.H, 0, len(greenhouses))
	for _, g := range greenhouses {
		items = append(items, gin.H{
			"id":          g.ID,
			"name":        g.Name,
			"location":    g.Location,
			"description": g.Description,
			"created_at":  g.CreatedAt,
			"updated_at":  g.UpdatedAt,
		})
	}

	response.Success(c, gin.H{
		"page":      page,
		"page_size": pageSize,
		"total":     total,
		"items":     items,
	})
}

func (h *Handler) GetGreenhouse(c *gin.Context) {
	greenhouseID, err := parseID(c.Param("greenhouseId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var greenhouse Greenhouse
	if err := h.db.First(&greenhouse, greenhouseID).Error; err != nil {
		response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
		return
	}

	response.Success(c, gin.H{
		"id":          greenhouse.ID,
		"name":        greenhouse.Name,
		"location":    greenhouse.Location,
		"description": greenhouse.Description,
		"created_at":  greenhouse.CreatedAt,
		"updated_at":  greenhouse.UpdatedAt,
	})
}

func (h *Handler) DeleteGreenhouse(c *gin.Context) {
	greenhouseID, err := parseID(c.Param("greenhouseId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	notFoundErr := errors.New("greenhouse_not_found")
	err = h.db.Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&Greenhouse{}).Where("id = ?", greenhouseID).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return notFoundErr
		}

		var groupIDs []uint64
		if err := tx.Model(&DeviceGroup{}).Where("greenhouse_id = ?", greenhouseID).Pluck("id", &groupIDs).Error; err != nil {
			return err
		}

		if err := tx.Model(&Device{}).Where("greenhouse_id = ?", greenhouseID).Update("greenhouse_id", nil).Error; err != nil {
			return err
		}

		if len(groupIDs) > 0 {
			if err := tx.Model(&Device{}).Where("group_id IN ?", groupIDs).Update("group_id", nil).Error; err != nil {
				return err
			}
		}

		if err := tx.Where("greenhouse_id = ?", greenhouseID).Delete(&DeviceGroup{}).Error; err != nil {
			return err
		}

		result := tx.Where("id = ?", greenhouseID).Delete(&Greenhouse{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return notFoundErr
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, notFoundErr) {
			response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "delete_failed", nil)
		return
	}

	response.Success(c, gin.H{})
}

func (h *Handler) CreateGroup(c *gin.Context) {
	var req CreateDeviceGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	group := DeviceGroup{
		GreenhouseID: req.GreenhouseID,
		Name:         req.Name,
		Description:  req.Description,
	}
	if err := h.db.Create(&group).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "create_failed", nil)
		return
	}
	response.Success(c, gin.H{"id": group.ID})
}

func (h *Handler) UpdateGroup(c *gin.Context) {
	groupID, err := parseID(c.Param("groupId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var req UpdateDeviceGroupRequest
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
	if len(updates) > 0 {
		if err := h.db.Model(&DeviceGroup{}).Where("id = ?", groupID).Updates(updates).Error; err != nil {
			response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "update_failed", nil)
			return
		}
	}
	response.Success(c, gin.H{})
}

func (h *Handler) ListGroups(c *gin.Context) {
	query := h.db.Model(&DeviceGroup{})
	if v := c.Query("greenhouse_id"); v != "" {
		query = query.Where("greenhouse_id = ?", v)
	}

	var groups []DeviceGroup
	if err := query.Order("id desc").Find(&groups).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	deviceCountByGroup := map[uint64]int64{}
	if len(groups) > 0 {
		groupIDs := make([]uint64, 0, len(groups))
		for _, g := range groups {
			groupIDs = append(groupIDs, g.ID)
		}

		var rows []struct {
			GroupID uint64 `gorm:"column:group_id"`
			Count   int64  `gorm:"column:count"`
		}
		if err := h.db.Model(&Device{}).
			Select("group_id, COUNT(*) as count").
			Where("group_id IN ?", groupIDs).
			Group("group_id").
			Scan(&rows).Error; err != nil {
			response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
			return
		}
		for _, row := range rows {
			deviceCountByGroup[row.GroupID] = row.Count
		}
	}

	items := make([]gin.H, 0, len(groups))
	for _, g := range groups {
		items = append(items, gin.H{
			"id":            g.ID,
			"name":          g.Name,
			"description":   g.Description,
			"greenhouse_id": g.GreenhouseID,
			"device_count":  deviceCountByGroup[g.ID],
		})
	}
	response.Success(c, gin.H{"items": items})
}

func (h *Handler) BindDeviceGroup(c *gin.Context) {
	groupID, err := parseID(c.Param("groupId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}
	deviceID, err := parseID(c.Param("deviceId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	if err := h.db.Model(&Device{}).Where("id = ?", deviceID).Update("group_id", groupID).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "update_failed", nil)
		return
	}
	response.Success(c, gin.H{})
}

func (h *Handler) UnbindDeviceGroup(c *gin.Context) {
	deviceID, err := parseID(c.Param("deviceId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	if err := h.db.Model(&Device{}).Where("id = ?", deviceID).Update("group_id", nil).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "update_failed", nil)
		return
	}
	response.Success(c, gin.H{})
}

func (h *Handler) DeleteGroup(c *gin.Context) {
	groupID, err := parseID(c.Param("groupId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	notFoundErr := errors.New("group_not_found")
	err = h.db.Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&DeviceGroup{}).Where("id = ?", groupID).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return notFoundErr
		}

		if err := tx.Model(&Device{}).Where("group_id = ?", groupID).Update("group_id", nil).Error; err != nil {
			return err
		}

		result := tx.Where("id = ?", groupID).Delete(&DeviceGroup{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return notFoundErr
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, notFoundErr) {
			response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "delete_failed", nil)
		return
	}

	response.Success(c, gin.H{})
}

func (h *Handler) TelemetrySummary(c *gin.Context) {
	deviceID, err := parseID(c.Param("deviceId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	from := time.Now().Add(-24 * time.Hour)
	if v := c.Query("from"); v != "" {
		t, err := time.Parse(time.RFC3339Nano, v)
		if err != nil {
			response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_from", nil)
			return
		}
		from = t.UTC()
	}

	to := time.Now()
	if v := c.Query("to"); v != "" {
		t, err := time.Parse(time.RFC3339Nano, v)
		if err != nil {
			response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_to", nil)
			return
		}
		to = t.UTC()
	}

	var dev Device
	if err := h.db.Select("id", "sampling_interval_sec").First(&dev, deviceID).Error; err != nil {
		response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
		return
	}

	type metricRow struct {
		ID   uint64 `gorm:"column:id"`
		Code string `gorm:"column:code"`
		Name string `gorm:"column:name"`
		Unit string `gorm:"column:unit"`
	}
	var metrics []metricRow
	if err := h.db.Table("metrics").Select("id", "code", "name", "unit").Find(&metrics).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	metricSummaries := make(map[string]gin.H)
	for _, m := range metrics {
		type hourlyRow struct {
			Hour string  `gorm:"column:hour"`
			Avg  float64 `gorm:"column:avg"`
		}
		var hourly []hourlyRow
		h.db.Table("telemetry_data").
			Select("DATE_FORMAT(collected_at, '%Y-%m-%dT%H:00:00Z') AS hour, AVG(value) AS avg").
			Where("device_id = ? AND metric_id = ? AND collected_at >= ? AND collected_at <= ?", deviceID, m.ID, from, to).
			Group("hour").
			Order("hour ASC").
			Scan(&hourly)

		type agg struct {
			Avg *float64 `gorm:"column:avg"`
			Max *float64 `gorm:"column:max"`
			Min *float64 `gorm:"column:min"`
		}
		var stat agg
		h.db.Table("telemetry_data").
			Select("AVG(value) AS avg, MAX(value) AS max, MIN(value) AS min").
			Where("device_id = ? AND metric_id = ? AND collected_at >= ? AND collected_at <= ?", deviceID, m.ID, from, to).
			Scan(&stat)

		var alertCount int64
		h.db.Table("alerts").
			Where("device_id = ? AND metric_id = ? AND triggered_at >= ? AND triggered_at <= ?", deviceID, m.ID, from, to).
			Count(&alertCount)

		hourlyItems := make([]gin.H, 0, len(hourly))
		for _, r := range hourly {
			hourlyItems = append(hourlyItems, gin.H{"hour": r.Hour, "avg": r.Avg})
		}

		metricSummaries[m.Code] = gin.H{
			"code":   m.Code,
			"name":   m.Name,
			"unit":   m.Unit,
			"avg":    stat.Avg,
			"max":    stat.Max,
			"min":    stat.Min,
			"alerts": alertCount,
			"hourly": hourlyItems,
		}
	}

	duration := to.Sub(from)
	totalHours := int(duration.Hours())
	if totalHours < 1 {
		totalHours = 1
	}

	type distinctHour struct {
		Hour string `gorm:"column:hour"`
	}
	var distinctHours []distinctHour
	h.db.Table("telemetry_data").
		Select("DISTINCT DATE_FORMAT(collected_at, '%Y-%m-%dT%H:00:00Z') AS hour").
		Where("device_id = ? AND collected_at >= ? AND collected_at <= ?", deviceID, from, to).
		Scan(&distinctHours)

	onlineRate := float64(len(distinctHours)) / float64(totalHours)
	if onlineRate > 1 {
		onlineRate = 1
	}

	type alertRow struct {
		ID          uint64    `gorm:"column:id"`
		Type        string    `gorm:"column:type"`
		Level       string    `gorm:"column:level"`
		Message     string    `gorm:"column:message"`
		Status      string    `gorm:"column:status"`
		TriggeredAt time.Time `gorm:"column:triggered_at"`
	}
	var alerts []alertRow
	h.db.Table("alerts").
		Select("id", "type", "level", "message", "status", "triggered_at").
		Where("device_id = ? AND triggered_at >= ? AND triggered_at <= ?", deviceID, from, to).
		Order("triggered_at DESC").
		Limit(50).
		Scan(&alerts)

	alertItems := make([]gin.H, 0, len(alerts))
	for _, a := range alerts {
		alertItems = append(alertItems, gin.H{
			"id":           a.ID,
			"type":         a.Type,
			"level":        a.Level,
			"message":      a.Message,
			"status":       a.Status,
			"triggered_at": a.TriggeredAt,
		})
	}

	response.Success(c, gin.H{
		"metrics":      metricSummaries,
		"online_rate":  onlineRate,
		"alert_events": alertItems,
	})
}

func (h *Handler) BatchUpdate(c *gin.Context) {
	var req struct {
		DeviceIDs []uint64               `json:"device_ids" binding:"required,min=1,max=100"`
		Updates   map[string]interface{} `json:"updates" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	allowedFields := map[string]bool{"group_id": true, "sampling_interval_sec": true, "status": true}
	updates := map[string]interface{}{}
	for k, v := range req.Updates {
		if allowedFields[k] {
			updates[k] = v
		}
	}
	if len(updates) == 0 {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "no_valid_fields", nil)
		return
	}

	result := h.db.Model(&Device{}).Where("id IN ?", req.DeviceIDs).Updates(updates)
	if result.Error != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "update_failed", nil)
		return
	}

	response.Success(c, gin.H{"affected": result.RowsAffected})
}

func (h *Handler) BatchDelete(c *gin.Context) {
	var req struct {
		DeviceIDs []uint64 `json:"device_ids" binding:"required,min=1,max=100"`
		Reason    string   `json:"reason" binding:"max=255"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	result := h.db.Where("id IN ?", req.DeviceIDs).Delete(&Device{})
	if result.Error != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "delete_failed", nil)
		return
	}

	response.Success(c, gin.H{"deleted": result.RowsAffected})
}

func parsePage(c *gin.Context) (int, int) {
	page := parseInt(c.Query("page"), 1)
	pageSize := parseInt(c.Query("page_size"), 20)
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 200 {
		pageSize = 200
	}
	if page < 1 {
		page = 1
	}
	return page, pageSize
}

func parseInt(v string, def int) int {
	if v == "" {
		return def
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return i
}

func parseID(v string) (uint64, error) {
	return strconv.ParseUint(v, 10, 64)
}
