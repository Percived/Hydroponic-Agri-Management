package device

import (
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

// ---- SensorDevice CRUD ----

func (h *Handler) CreateSensorDevice(c *gin.Context) {
	var req CreateSensorDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	var count int64
	if err := h.db.Model(&SensorDevice{}).Where("device_code = ?", req.DeviceCode).Count(&count).Error; err == nil && count > 0 {
		response.Error(c, http.StatusConflict, platformErrors.CodeDeviceCodeRepeat, "device_code_exists", nil)
		return
	}

	dev := SensorDevice{
		DeviceCode:      req.DeviceCode,
		Name:            req.Name,
		Model:           req.Model,
		FirmwareVersion: req.FirmwareVersion,
		GreenhouseID:    req.GreenhouseID,
		GrowingZoneID:   req.GrowingZoneID,
		Protocol:        req.Protocol,
		Metadata:        req.Metadata,
		Status:          StatusOnline,
	}
	if dev.Protocol == "" {
		dev.Protocol = ProtocolMQTT
	}

	if err := h.db.Create(&dev).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "create_failed", nil)
		return
	}

	response.Success(c, gin.H{"id": dev.ID})
}

func (h *Handler) UpdateSensorDevice(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var req UpdateSensorDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Model != nil {
		updates["model"] = *req.Model
	}
	if req.FirmwareVersion != nil {
		updates["firmware_version"] = *req.FirmwareVersion
	}
	if req.GrowingZoneID != nil {
		updates["growing_zone_id"] = *req.GrowingZoneID
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.Metadata != nil {
		updates["metadata"] = *req.Metadata
	}

	if len(updates) > 0 {
		result := h.db.Model(&SensorDevice{}).Where("id = ?", id).Updates(updates)
		if result.Error != nil {
			response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "update_failed", nil)
			return
		}
		if result.RowsAffected == 0 {
			response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
			return
		}
	}

	response.Success(c, gin.H{})
}

func (h *Handler) ListSensorDevices(c *gin.Context) {
	page, pageSize := parsePage(c)

	query := h.db.Model(&SensorDevice{})
	if v := c.Query("greenhouse_id"); v != "" {
		query = query.Where("greenhouse_id = ?", v)
	}
	if v := c.Query("growing_zone_id"); v != "" {
		query = query.Where("growing_zone_id = ?", v)
	}
	if v := c.Query("status"); v != "" {
		query = query.Where("status = ?", v)
	}
	if v := c.Query("keyword"); v != "" {
		like := "%" + v + "%"
		query = query.Where("name LIKE ? OR device_code LIKE ?", like, like)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	var devs []SensorDevice
	if total > 0 {
		if err := query.Order("id desc").Limit(pageSize).Offset((page - 1) * pageSize).Find(&devs).Error; err != nil {
			response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
			return
		}
	}

	items := make([]SensorDeviceResponse, 0, len(devs))
	for _, d := range devs {
		items = append(items, deviceToSensorResponse(d))
	}

	response.Success(c, gin.H{
		"page":      page,
		"page_size": pageSize,
		"total":     total,
		"items":     items,
	})
}

func (h *Handler) GetSensorDevice(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var dev SensorDevice
	if err := h.db.First(&dev, id).Error; err != nil {
		response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
		return
	}

	response.Success(c, deviceToSensorResponse(dev))
}

func (h *Handler) DeleteSensorDevice(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	result := h.db.Where("id = ?", id).Delete(&SensorDevice{})
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

// ---- SensorChannel CRUD ----

func (h *Handler) CreateSensorChannel(c *gin.Context) {
	var req CreateSensorChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	ch := SensorChannel{
		SensorDeviceID:      req.SensorDeviceID,
		ChannelCode:         req.ChannelCode,
		MetricCode:          req.MetricCode,
		Unit:                req.Unit,
		PrecisionDigits:     req.PrecisionDigits,
		RangeMin:            req.RangeMin,
		RangeMax:            req.RangeMax,
		SamplingIntervalSec: req.SamplingIntervalSec,
		Metadata:            req.Metadata,
		Enabled:             1,
	}
	if ch.PrecisionDigits == 0 {
		ch.PrecisionDigits = 2
	}
	if ch.SamplingIntervalSec == 0 {
		ch.SamplingIntervalSec = 60
	}

	if err := h.db.Create(&ch).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "create_failed", nil)
		return
	}

	response.Success(c, gin.H{"id": ch.ID})
}

func (h *Handler) UpdateSensorChannel(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var req UpdateSensorChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	updates := map[string]interface{}{}
	if req.ChannelCode != nil {
		updates["channel_code"] = *req.ChannelCode
	}
	if req.MetricCode != nil {
		updates["metric_code"] = *req.MetricCode
	}
	if req.Unit != nil {
		updates["unit"] = *req.Unit
	}
	if req.PrecisionDigits != nil {
		updates["precision_digits"] = *req.PrecisionDigits
	}
	if req.RangeMin != nil {
		updates["range_min"] = *req.RangeMin
	}
	if req.RangeMax != nil {
		updates["range_max"] = *req.RangeMax
	}
	if req.SamplingIntervalSec != nil {
		updates["sampling_interval_sec"] = *req.SamplingIntervalSec
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}
	if req.Metadata != nil {
		updates["metadata"] = *req.Metadata
	}

	if len(updates) > 0 {
		result := h.db.Model(&SensorChannel{}).Where("id = ?", id).Updates(updates)
		if result.Error != nil {
			response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "update_failed", nil)
			return
		}
		if result.RowsAffected == 0 {
			response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
			return
		}
	}

	response.Success(c, gin.H{})
}

func (h *Handler) ListSensorChannels(c *gin.Context) {
	page, pageSize := parsePage(c)

	query := h.db.Model(&SensorChannel{})
	if v := c.Query("sensor_device_id"); v != "" {
		query = query.Where("sensor_device_id = ?", v)
	}
	if v := c.Query("enabled"); v != "" {
		query = query.Where("enabled = ?", v)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	var channels []SensorChannel
	if total > 0 {
		if err := query.Order("id desc").Limit(pageSize).Offset((page - 1) * pageSize).Find(&channels).Error; err != nil {
			response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
			return
		}
	}

	items := make([]SensorChannelResponse, 0, len(channels))
	for _, ch := range channels {
		items = append(items, channelToSensorResponse(ch))
	}

	response.Success(c, gin.H{
		"page":      page,
		"page_size": pageSize,
		"total":     total,
		"items":     items,
	})
}

func (h *Handler) GetSensorChannel(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var ch SensorChannel
	if err := h.db.First(&ch, id).Error; err != nil {
		response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
		return
	}

	response.Success(c, channelToSensorResponse(ch))
}

func (h *Handler) DeleteSensorChannel(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	result := h.db.Where("id = ?", id).Delete(&SensorChannel{})
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

// ---- ActuatorDevice CRUD ----

func (h *Handler) CreateActuatorDevice(c *gin.Context) {
	var req CreateActuatorDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	var count int64
	if err := h.db.Model(&ActuatorDevice{}).Where("device_code = ?", req.DeviceCode).Count(&count).Error; err == nil && count > 0 {
		response.Error(c, http.StatusConflict, platformErrors.CodeDeviceCodeRepeat, "device_code_exists", nil)
		return
	}

	dev := ActuatorDevice{
		DeviceCode:      req.DeviceCode,
		Name:            req.Name,
		Model:           req.Model,
		FirmwareVersion: req.FirmwareVersion,
		GreenhouseID:    req.GreenhouseID,
		GrowingZoneID:   req.GrowingZoneID,
		Protocol:        req.Protocol,
		Metadata:        req.Metadata,
		Status:          StatusOnline,
	}
	if dev.Protocol == "" {
		dev.Protocol = ProtocolMQTT
	}

	if err := h.db.Create(&dev).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "create_failed", nil)
		return
	}

	response.Success(c, gin.H{"id": dev.ID})
}

func (h *Handler) UpdateActuatorDevice(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var req UpdateActuatorDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Model != nil {
		updates["model"] = *req.Model
	}
	if req.FirmwareVersion != nil {
		updates["firmware_version"] = *req.FirmwareVersion
	}
	if req.GrowingZoneID != nil {
		updates["growing_zone_id"] = *req.GrowingZoneID
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.Metadata != nil {
		updates["metadata"] = *req.Metadata
	}

	if len(updates) > 0 {
		result := h.db.Model(&ActuatorDevice{}).Where("id = ?", id).Updates(updates)
		if result.Error != nil {
			response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "update_failed", nil)
			return
		}
		if result.RowsAffected == 0 {
			response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
			return
		}
	}

	response.Success(c, gin.H{})
}

func (h *Handler) ListActuatorDevices(c *gin.Context) {
	page, pageSize := parsePage(c)

	query := h.db.Model(&ActuatorDevice{})
	if v := c.Query("greenhouse_id"); v != "" {
		query = query.Where("greenhouse_id = ?", v)
	}
	if v := c.Query("growing_zone_id"); v != "" {
		query = query.Where("growing_zone_id = ?", v)
	}
	if v := c.Query("status"); v != "" {
		query = query.Where("status = ?", v)
	}
	if v := c.Query("keyword"); v != "" {
		like := "%" + v + "%"
		query = query.Where("name LIKE ? OR device_code LIKE ?", like, like)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	var devs []ActuatorDevice
	if total > 0 {
		if err := query.Order("id desc").Limit(pageSize).Offset((page - 1) * pageSize).Find(&devs).Error; err != nil {
			response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
			return
		}
	}

	items := make([]ActuatorDeviceResponse, 0, len(devs))
	for _, d := range devs {
		items = append(items, deviceToActuatorResponse(d))
	}

	response.Success(c, gin.H{
		"page":      page,
		"page_size": pageSize,
		"total":     total,
		"items":     items,
	})
}

func (h *Handler) GetActuatorDevice(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var dev ActuatorDevice
	if err := h.db.First(&dev, id).Error; err != nil {
		response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
		return
	}

	response.Success(c, deviceToActuatorResponse(dev))
}

func (h *Handler) DeleteActuatorDevice(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	result := h.db.Where("id = ?", id).Delete(&ActuatorDevice{})
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

// ---- ActuatorChannel CRUD ----

func (h *Handler) CreateActuatorChannel(c *gin.Context) {
	var req CreateActuatorChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	ch := ActuatorChannel{
		ActuatorDeviceID: req.ActuatorDeviceID,
		ChannelCode:      req.ChannelCode,
		ActuatorType:     req.ActuatorType,
		RatedPowerWatt:   req.RatedPowerWatt,
		Metadata:         req.Metadata,
		CurrentState:     "OFF",
		Enabled:          1,
	}

	if err := h.db.Create(&ch).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "create_failed", nil)
		return
	}

	response.Success(c, gin.H{"id": ch.ID})
}

func (h *Handler) UpdateActuatorChannel(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var req UpdateActuatorChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	updates := map[string]interface{}{}
	if req.ChannelCode != nil {
		updates["channel_code"] = *req.ChannelCode
	}
	if req.ActuatorType != nil {
		updates["actuator_type"] = *req.ActuatorType
	}
	if req.CurrentState != nil {
		updates["current_state"] = *req.CurrentState
	}
	if req.RatedPowerWatt != nil {
		updates["rated_power_watt"] = *req.RatedPowerWatt
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}
	if req.Metadata != nil {
		updates["metadata"] = *req.Metadata
	}

	if len(updates) > 0 {
		result := h.db.Model(&ActuatorChannel{}).Where("id = ?", id).Updates(updates)
		if result.Error != nil {
			response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "update_failed", nil)
			return
		}
		if result.RowsAffected == 0 {
			response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
			return
		}
	}

	response.Success(c, gin.H{})
}

func (h *Handler) ListActuatorChannels(c *gin.Context) {
	page, pageSize := parsePage(c)

	query := h.db.Model(&ActuatorChannel{})
	if v := c.Query("actuator_device_id"); v != "" {
		query = query.Where("actuator_device_id = ?", v)
	}
	if v := c.Query("actuator_type"); v != "" {
		query = query.Where("actuator_type = ?", v)
	}
	if v := c.Query("enabled"); v != "" {
		query = query.Where("enabled = ?", v)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	var channels []ActuatorChannel
	if total > 0 {
		if err := query.Order("id desc").Limit(pageSize).Offset((page - 1) * pageSize).Find(&channels).Error; err != nil {
			response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
			return
		}
	}

	items := make([]ActuatorChannelResponse, 0, len(channels))
	for _, ch := range channels {
		items = append(items, channelToActuatorResponse(ch))
	}

	response.Success(c, gin.H{
		"page":      page,
		"page_size": pageSize,
		"total":     total,
		"items":     items,
	})
}

func (h *Handler) GetActuatorChannel(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var ch ActuatorChannel
	if err := h.db.First(&ch, id).Error; err != nil {
		response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
		return
	}

	response.Success(c, channelToActuatorResponse(ch))
}

func (h *Handler) DeleteActuatorChannel(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	result := h.db.Where("id = ?", id).Delete(&ActuatorChannel{})
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

// ---- Helpers ----

func deviceToSensorResponse(d SensorDevice) SensorDeviceResponse {
	var lastSeen *string
	if d.LastSeenAt != nil {
		formatted := d.LastSeenAt.Format(time.RFC3339)
		lastSeen = &formatted
	}
	return SensorDeviceResponse{
		ID:              d.ID,
		GreenhouseID:    d.GreenhouseID,
		GrowingZoneID:   d.GrowingZoneID,
		DeviceCode:      d.DeviceCode,
		Name:            d.Name,
		Model:           d.Model,
		FirmwareVersion: d.FirmwareVersion,
		Status:          d.Status,
		LastSeenAt:      lastSeen,
		Protocol:        d.Protocol,
		Metadata:        d.Metadata,
		CreatedAt:       d.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       d.UpdatedAt.Format(time.RFC3339),
	}
}

func deviceToActuatorResponse(d ActuatorDevice) ActuatorDeviceResponse {
	var lastSeen *string
	if d.LastSeenAt != nil {
		formatted := d.LastSeenAt.Format(time.RFC3339)
		lastSeen = &formatted
	}
	return ActuatorDeviceResponse{
		ID:              d.ID,
		GreenhouseID:    d.GreenhouseID,
		GrowingZoneID:   d.GrowingZoneID,
		DeviceCode:      d.DeviceCode,
		Name:            d.Name,
		Model:           d.Model,
		FirmwareVersion: d.FirmwareVersion,
		Status:          d.Status,
		LastSeenAt:      lastSeen,
		Protocol:        d.Protocol,
		Metadata:        d.Metadata,
		CreatedAt:       d.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       d.UpdatedAt.Format(time.RFC3339),
	}
}

func channelToSensorResponse(ch SensorChannel) SensorChannelResponse {
	var lastReported *string
	if ch.LastReportedAt != nil {
		formatted := ch.LastReportedAt.Format(time.RFC3339)
		lastReported = &formatted
	}
	return SensorChannelResponse{
		ID:                  ch.ID,
		SensorDeviceID:      ch.SensorDeviceID,
		ChannelCode:         ch.ChannelCode,
		MetricCode:          ch.MetricCode,
		Unit:                ch.Unit,
		PrecisionDigits:     ch.PrecisionDigits,
		RangeMin:            ch.RangeMin,
		RangeMax:            ch.RangeMax,
		SamplingIntervalSec: ch.SamplingIntervalSec,
		Enabled:             ch.Enabled,
		LastReportedAt:      lastReported,
		Metadata:            ch.Metadata,
		CreatedAt:           ch.CreatedAt.Format(time.RFC3339),
		UpdatedAt:           ch.UpdatedAt.Format(time.RFC3339),
	}
}

func channelToActuatorResponse(ch ActuatorChannel) ActuatorChannelResponse {
	return ActuatorChannelResponse{
		ID:               ch.ID,
		ActuatorDeviceID: ch.ActuatorDeviceID,
		ChannelCode:      ch.ChannelCode,
		ActuatorType:     ch.ActuatorType,
		CurrentState:     ch.CurrentState,
		RatedPowerWatt:   ch.RatedPowerWatt,
		Enabled:          ch.Enabled,
		Metadata:         ch.Metadata,
		CreatedAt:        ch.CreatedAt.Format(time.RFC3339),
		UpdatedAt:        ch.UpdatedAt.Format(time.RFC3339),
	}
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
