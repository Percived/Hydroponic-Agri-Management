package control

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"hydroponic-backend/internal/audit"
	"hydroponic-backend/internal/auth"
	"hydroponic-backend/internal/device"
	platformErrors "hydroponic-backend/internal/platform/errors"
	"hydroponic-backend/internal/platform/response"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	db   *gorm.DB
	mqtt mqtt.Client
	log  *slog.Logger
}

func NewHandler(db *gorm.DB, mqttClient mqtt.Client, log *slog.Logger) *Handler {
	return &Handler{db: db, mqtt: mqttClient, log: log}
}

func (h *Handler) CreateCommand(c *gin.Context) {
	var req CreateCommandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	var dev device.Device
	if err := h.db.Select("id", "device_code", "status").First(&dev, req.DeviceID).Error; err != nil {
		response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "device_not_found", nil)
		return
	}

	payloadBytes, err := json.Marshal(req.Payload)
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_payload", nil)
		return
	}

	userID := currentUserID(c)
	cmd := ControlCommand{
		DeviceID:    req.DeviceID,
		CommandType: req.CommandType,
		Payload:     payloadBytes,
		Status:      CommandStatusPending,
		CreatedBy:   userID,
	}
	if err := h.db.Create(&cmd).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "create_failed", nil)
		return
	}

	publishStatus := cmd.Status
	if dev.Status == device.DeviceStatusEnabled {
		if h.publishCommand(dev.DeviceCode, cmd.ID, req.CommandType, req.Payload) {
			now := time.Now().UTC()
			publishStatus = CommandStatusSent
			_ = h.db.Model(&cmd).Updates(map[string]interface{}{"status": CommandStatusSent, "sent_at": now}).Error
		} else {
			publishStatus = CommandStatusFailed
			_ = h.db.Model(&cmd).Update("status", CommandStatusFailed).Error
		}
	} else {
		publishStatus = CommandStatusFailed
		_ = h.db.Model(&cmd).Update("status", CommandStatusFailed).Error
	}

	cmdID := cmd.ID
	audit.Write(h.db, userID, "CONTROL_COMMAND", "control_commands", &cmdID, req)
	response.Success(c, gin.H{"id": cmd.ID, "status": publishStatus})
}

func (h *Handler) GetCommand(c *gin.Context) {
	id, err := parseID(c.Param("commandId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var cmd ControlCommand
	if err := h.db.First(&cmd, id).Error; err != nil {
		response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
		return
	}

	response.Success(c, gin.H{
		"id":           cmd.ID,
		"device_id":    cmd.DeviceID,
		"command_type": cmd.CommandType,
		"payload":      json.RawMessage(cmd.Payload),
		"status":       cmd.Status,
		"created_by":   cmd.CreatedBy,
		"created_at":   cmd.CreatedAt,
		"sent_at":      cmd.SentAt,
		"executed_at":  cmd.ExecutedAt,
	})
}

func (h *Handler) ListCommands(c *gin.Context) {
	page, size := parsePage(c)
	q := h.db.Model(&ControlCommand{})
	if v := c.Query("device_id"); v != "" {
		q = q.Where("device_id = ?", v)
	}
	if v := c.Query("status"); v != "" {
		q = q.Where("status = ?", v)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	items := []ControlCommand{}
	if total > 0 {
		if err := q.Order("id desc").Limit(size).Offset((page - 1) * size).Find(&items).Error; err != nil {
			response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
			return
		}
	}

	rows := make([]gin.H, 0, len(items))
	for _, it := range items {
		rows = append(rows, gin.H{"id": it.ID, "status": it.Status, "device_id": it.DeviceID, "command_type": it.CommandType, "created_at": it.CreatedAt})
	}
	response.Success(c, gin.H{"page": page, "page_size": size, "total": total, "items": rows})
}

func (h *Handler) CreateRule(c *gin.Context) {
	var req CreateRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	metricID, err := h.metricIDByCode(req.MetricCode)
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_metric_code", nil)
		return
	}

	actionBytes, err := json.Marshal(req.Action)
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_action", nil)
		return
	}

	userID := currentUserID(c)
	rule := ControlRule{
		Name:           req.Name,
		MetricID:       metricID,
		Operator:       req.Operator,
		Threshold:      req.Threshold,
		Action:         actionBytes,
		TargetDeviceID: req.TargetDeviceID,
		Enabled:        *req.Enabled,
		CreatedBy:      userID,
	}
	if err := h.db.Create(&rule).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "create_failed", nil)
		return
	}

	ruleID := rule.ID
	audit.Write(h.db, userID, "CREATE_RULE", "control_rules", &ruleID, req)
	response.Success(c, gin.H{"id": rule.ID})
}

func (h *Handler) UpdateRule(c *gin.Context) {
	id, err := parseID(c.Param("ruleId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var req UpdateRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Operator != nil {
		updates["operator"] = *req.Operator
	}
	if req.Threshold != nil {
		updates["threshold"] = *req.Threshold
	}
	if req.TargetDeviceID != nil {
		updates["target_device_id"] = *req.TargetDeviceID
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}
	if req.Action != nil {
		b, err := json.Marshal(req.Action)
		if err != nil {
			response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_action", nil)
			return
		}
		updates["action"] = b
	}
	if len(updates) == 0 {
		response.Success(c, gin.H{})
		return
	}

	if err := h.db.Model(&ControlRule{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "update_failed", nil)
		return
	}

	audit.Write(h.db, currentUserID(c), "UPDATE_RULE", "control_rules", &id, updates)
	response.Success(c, gin.H{})
}

func (h *Handler) DeleteRule(c *gin.Context) {
	id, err := parseID(c.Param("ruleId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}
	if err := h.db.Delete(&ControlRule{}, id).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "delete_failed", nil)
		return
	}
	audit.Write(h.db, currentUserID(c), "DELETE_RULE", "control_rules", &id, nil)
	response.Success(c, gin.H{})
}

func (h *Handler) ListRules(c *gin.Context) {
	page, size := parsePage(c)
	q := h.db.Table("control_rules r").
		Select("r.id, r.name, m.code AS metric_code, r.operator, r.threshold, r.action, r.target_device_id, d.name AS target_device_name, r.enabled, r.created_at, r.updated_at").
		Joins("JOIN metrics m ON m.id = r.metric_id").
		Joins("LEFT JOIN devices d ON d.id = r.target_device_id")
	if v := strings.TrimSpace(c.Query("metric_code")); v != "" {
		q = q.Where("m.code = ?", v)
	}
	if v := strings.TrimSpace(c.Query("enabled")); v != "" {
		b := strings.EqualFold(v, "true")
		q = q.Where("r.enabled = ?", b)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	type item struct {
		ID               uint64          `json:"id"`
		Name             string          `json:"name"`
		MetricCode       string          `json:"metric_code"`
		Operator         string          `json:"operator"`
		Threshold        float64         `json:"threshold"`
		Action           json.RawMessage `json:"action"`
		TargetDeviceID   uint64          `json:"target_device_id"`
		TargetDeviceName string          `json:"target_device_name"`
		Enabled          bool            `json:"enabled"`
		CreatedAt        time.Time       `json:"created_at"`
		UpdatedAt        time.Time       `json:"updated_at"`
	}
	rows := []item{}
	if total > 0 {
		if err := q.Order("r.id desc").Limit(size).Offset((page - 1) * size).Scan(&rows).Error; err != nil {
			response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
			return
		}
	}
	response.Success(c, gin.H{"page": page, "page_size": size, "total": total, "items": rows})
}

func (h *Handler) CreateTemplate(c *gin.Context) {
	var req CreateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}
	content, err := json.Marshal(req.Content)
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_content", nil)
		return
	}

	t := ControlTemplate{Name: req.Name, Description: req.Description, Content: content, CreatedBy: currentUserID(c)}
	if err := h.db.Create(&t).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "create_failed", nil)
		return
	}
	id := t.ID
	audit.Write(h.db, currentUserID(c), "CREATE_TEMPLATE", "control_templates", &id, req)
	response.Success(c, gin.H{"id": t.ID})
}

func (h *Handler) ApplyTemplate(c *gin.Context) {
	templateID, err := parseID(c.Param("templateId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}
	var req ApplyTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	var t ControlTemplate
	if err := h.db.First(&t, templateID).Error; err != nil {
		response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
		return
	}

	audit.Write(h.db, currentUserID(c), "APPLY_TEMPLATE", "control_templates", &templateID, gin.H{"target_group_id": req.TargetGroupID})
	response.Success(c, gin.H{})
}

func (h *Handler) ListTemplates(c *gin.Context) {
	var templates []ControlTemplate
	if err := h.db.Order("id desc").Find(&templates).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}
	items := make([]gin.H, 0, len(templates))
	for _, t := range templates {
		items = append(items, gin.H{"id": t.ID, "name": t.Name, "description": t.Description, "created_at": t.CreatedAt})
	}
	response.Success(c, gin.H{"items": items})
}

func (h *Handler) BatchCommands(c *gin.Context) {
	var req struct {
		TargetType  string                 `json:"target_type" binding:"required,oneof=greenhouse device_group devices"`
		TargetIDs   []uint64               `json:"target_ids" binding:"required,min=1,max=100"`
		CommandType string                 `json:"command_type" binding:"required,min=1,max=32"`
		Payload     map[string]interface{} `json:"payload" binding:"required"`
		Remark      string                 `json:"remark" binding:"max=255"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	var deviceIDs []uint64
	switch req.TargetType {
	case "greenhouse":
		if err := h.db.Model(&device.Device{}).Where("greenhouse_id IN ?", req.TargetIDs).Pluck("id", &deviceIDs).Error; err != nil {
			response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
			return
		}
	case "device_group":
		if err := h.db.Model(&device.Device{}).Where("group_id IN ?", req.TargetIDs).Pluck("id", &deviceIDs).Error; err != nil {
			response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
			return
		}
	case "devices":
		deviceIDs = req.TargetIDs
	}

	if len(deviceIDs) == 0 {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeNotFound, "no_devices_found", nil)
		return
	}

	var devices []device.Device
	if err := h.db.Select("id", "device_code", "status").Where("id IN ?", deviceIDs).Find(&devices).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	userID := currentUserID(c)
	results := make([]gin.H, 0, len(devices))

	for _, dev := range devices {
		payloadBytes, _ := json.Marshal(req.Payload)
		cmd := ControlCommand{
			DeviceID:    dev.ID,
			CommandType: req.CommandType,
			Payload:     payloadBytes,
			Status:      CommandStatusPending,
			CreatedBy:   userID,
		}
		if err := h.db.Create(&cmd).Error; err != nil {
			results = append(results, gin.H{"device_id": dev.ID, "status": "FAILED", "message": "create_failed"})
			continue
		}

		publishStatus := cmd.Status
		if dev.Status == device.DeviceStatusEnabled {
			if h.publishCommand(dev.DeviceCode, cmd.ID, req.CommandType, req.Payload) {
				now := time.Now().UTC()
				publishStatus = CommandStatusSent
				_ = h.db.Model(&cmd).Updates(map[string]interface{}{"status": CommandStatusSent, "sent_at": now}).Error
			} else {
				publishStatus = CommandStatusFailed
				_ = h.db.Model(&cmd).Update("status", CommandStatusFailed).Error
			}
		} else {
			publishStatus = CommandStatusFailed
			_ = h.db.Model(&cmd).Update("status", CommandStatusFailed).Error
		}

		results = append(results, gin.H{"device_id": dev.ID, "command_id": cmd.ID, "status": publishStatus})
	}

	audit.Write(h.db, userID, "BATCH_COMMAND", "control_commands", nil, gin.H{"target_type": req.TargetType, "target_ids": req.TargetIDs, "results": results})
	response.Success(c, gin.H{"results": results})
}

func (h *Handler) metricIDByCode(code string) (uint64, error) {
	var m metricRef
	if err := h.db.Table("metrics").Select("id", "code").Where("code = ?", code).First(&m).Error; err != nil {
		return 0, err
	}
	return m.ID, nil
}

func (h *Handler) publishCommand(deviceCode string, commandID uint64, commandType string, payload map[string]interface{}) bool {
	if h.mqtt == nil || !h.mqtt.IsConnectionOpen() {
		h.log.Warn("mqtt not connected, skip command publish", "device_code", deviceCode, "command_id", commandID)
		return false
	}
	topic := fmt.Sprintf("hydroponic/v1/command/%s", deviceCode)
	msg := map[string]interface{}{
		"command_id":   commandID,
		"command_type": commandType,
		"payload":      payload,
		"created_at":   time.Now().UTC().Format(time.RFC3339Nano),
	}
	b, _ := json.Marshal(msg)
	token := h.mqtt.Publish(topic, 1, false, b)
	if !token.WaitTimeout(3 * time.Second) {
		return false
	}
	return token.Error() == nil
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
