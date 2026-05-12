package command

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"hydroponic-backend/internal/auth"
	platformErrors "hydroponic-backend/internal/platform/errors"
	"hydroponic-backend/internal/platform/event"
	mqttpkg "hydroponic-backend/internal/platform/mqtt"
	"hydroponic-backend/internal/platform/response"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Handler holds dependencies for command HTTP handlers.
type Handler struct {
	db         *gorm.DB
	mqttClient mqtt.Client
	eventHub   *event.Hub
	waiter     *CommandWaiter
}

// NewHandler creates a new command Handler.
func NewHandler(db *gorm.DB, mqttClient mqtt.Client, hub *event.Hub) *Handler {
	return &Handler{
		db:         db,
		mqttClient: mqttClient,
		eventHub:   hub,
		waiter:     NewCommandWaiter(hub),
	}
}

// --- ControlCommand handlers ---

// CreateCommand creates a new control command in PENDING status.
func (h *Handler) CreateCommand(c *gin.Context) {
	var req CreateCommandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	payloadBytes, err := json.Marshal(req.Payload)
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_payload", nil)
		return
	}

	userID := currentUserID(c)
	cmd := ControlCommand{
		ActuatorChannelID: req.ActuatorChannelID,
		BatchID:           req.BatchID,
		CommandType:       req.CommandType,
		Payload:           string(payloadBytes),
		Status:            "PENDING",
		RequestID:         req.RequestID,
		CreatedBy:         userID,
	}

	if err := h.db.Create(&cmd).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "create_failed", nil)
		return
	}

	response.Success(c, gin.H{
		"id":     cmd.ID,
		"status": cmd.Status,
	})
}

// SendCommand marks a command as SENT with timestamp and dispatches via MQTT.
func (h *Handler) SendCommand(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var req SendCommandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	// Load command
	var cmd ControlCommand
	if err := h.db.First(&cmd, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	now := time.Now().UTC()

	// Dispatch via MQTT
	deviceCode, err := h.dispatchMQTT(cmd)
	if err != nil {
		h.markFailed(cmd.ID, deviceCode, err)
		response.Error(c, http.StatusServiceUnavailable, platformErrors.CodeDeviceOffline, "mqtt_dispatch_failed", nil)
		return
	}

	// Update DB status
	updates := map[string]interface{}{
		"status":  "SENT",
		"sent_at": now,
	}
	if req.RequestID != "" {
		updates["request_id"] = req.RequestID
	}

	result := h.db.Model(&ControlCommand{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "send_failed", nil)
		return
	}

	h.eventHub.Publish(event.SSEEvent{
		Type: "command:dispatched",
		Data: event.CommandDispatchedSSEDataV1{
			SchemaVersion: 1,
			CommandID:     cmd.ID,
			DeviceCode:    deviceCode,
			Status:        "SENT",
			DispatchedAt:  now.Format(time.RFC3339),
			SourceType:    "MANUAL",
		},
	})

	response.Success(c, gin.H{
		"id":      id,
		"status":  "SENT",
		"sent_at": now.Format(time.RFC3339),
	})
}

// DispatchAndWait creates a command, dispatches via MQTT, and waits for ack (sync mode).
func (h *Handler) DispatchAndWait(c *gin.Context) {
	var req CreateCommandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	payloadBytes, err := json.Marshal(req.Payload)
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_payload", nil)
		return
	}

	userID := currentUserID(c)
	cmd := ControlCommand{
		ActuatorChannelID: req.ActuatorChannelID,
		BatchID:           req.BatchID,
		CommandType:       req.CommandType,
		Payload:           string(payloadBytes),
		Status:            "PENDING",
		RequestID:         req.RequestID,
		CreatedBy:         userID,
	}

	if err := h.db.Create(&cmd).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "create_failed", nil)
		return
	}

	// Register waiter before dispatching
	h.waiter.Register(cmd.ID)

	// Dispatch via MQTT
	deviceCode, err := h.dispatchMQTT(cmd)
	if err != nil {
		h.markFailed(cmd.ID, deviceCode, err)
		response.Error(c, http.StatusServiceUnavailable, platformErrors.CodeDeviceOffline, "mqtt_dispatch_failed", nil)
		return
	}

	// Mark as SENT
	now := time.Now().UTC()
	h.db.Model(&ControlCommand{}).Where("id = ?", cmd.ID).Updates(map[string]interface{}{
		"status":  "SENT",
		"sent_at": now,
	})

	h.eventHub.Publish(event.SSEEvent{
		Type: "command:dispatched",
		Data: event.CommandDispatchedSSEDataV1{
			SchemaVersion: 1,
			CommandID:     cmd.ID,
			DeviceCode:    deviceCode,
			Status:        "SENT",
			DispatchedAt:  now.Format(time.RFC3339),
			SourceType:    "MANUAL",
		},
	})

	// Wait for ack with 10s timeout
	receipt, waitErr := h.waiter.Wait(cmd.ID, 10*time.Second)
	if waitErr != nil {
		// Mark as TIMEOUT
		h.db.Model(&ControlCommand{}).Where("id = ?", cmd.ID).Update("status", "TIMEOUT")
		response.Success(c, gin.H{
			"id":      cmd.ID,
			"status":  "TIMEOUT",
			"message": waitErr.Error(),
		})
		return
	}

	response.Success(c, gin.H{
		"id":          cmd.ID,
		"status":      "ACKED",
		"ack_code":    receipt.AckCode,
		"ack_message": receipt.AckMessage,
	})
}

// DispatchAsync creates a command and dispatches via MQTT without waiting (async mode).
func (h *Handler) DispatchAsync(c *gin.Context) {
	var req CreateCommandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	payloadBytes, err := json.Marshal(req.Payload)
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_payload", nil)
		return
	}

	userID := currentUserID(c)
	cmd := ControlCommand{
		ActuatorChannelID: req.ActuatorChannelID,
		BatchID:           req.BatchID,
		CommandType:       req.CommandType,
		Payload:           string(payloadBytes),
		Status:            "PENDING",
		RequestID:         req.RequestID,
		CreatedBy:         userID,
	}

	if err := h.db.Create(&cmd).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "create_failed", nil)
		return
	}

	// Dispatch via MQTT
	deviceCode, err := h.dispatchMQTT(cmd)
	if err != nil {
		h.markFailed(cmd.ID, deviceCode, err)
		response.Error(c, http.StatusServiceUnavailable, platformErrors.CodeDeviceOffline, "mqtt_dispatch_failed", nil)
		return
	}

	// Mark as SENT
	now := time.Now().UTC()
	h.db.Model(&ControlCommand{}).Where("id = ?", cmd.ID).Updates(map[string]interface{}{
		"status":  "SENT",
		"sent_at": now,
	})

	// Publish event for async status tracking
	h.eventHub.Publish(event.SSEEvent{
		Type: "command:dispatched",
		Data: event.CommandDispatchedSSEDataV1{
			SchemaVersion: 1,
			CommandID:     cmd.ID,
			DeviceCode:    deviceCode,
			Status:        "SENT",
			DispatchedAt:  now.Format(time.RFC3339),
			SourceType:    "MANUAL",
		},
	})

	response.Success(c, gin.H{
		"id":      cmd.ID,
		"status":  "SENT",
		"sent_at": now.Format(time.RFC3339),
	})
}

// dispatchMQTT publishes a command to the device's MQTT command topic.
func (h *Handler) dispatchMQTT(cmd ControlCommand) (string, error) {
	target, err := h.lookupActuatorTarget(cmd.ActuatorChannelID)
	if err != nil {
		return "", fmt.Errorf("device lookup: %w", err)
	}
	if h.mqttClient == nil || !h.mqttClient.IsConnected() {
		return target.DeviceCode, fmt.Errorf("mqtt not connected")
	}

	payload := BuildDeviceCommandPayload(cmd.Payload, DispatchTargetMeta{
		CommandID:         cmd.ID,
		CommandType:       cmd.CommandType,
		ActuatorChannelID: cmd.ActuatorChannelID,
		ChannelCode:       target.ChannelCode,
	})

	topic := fmt.Sprintf("%s/%s/%s/%s", mqttpkg.TopicPrefix, target.DeviceCode, mqttpkg.TopicCmdPrefix, cmd.CommandType)
	token := h.mqttClient.Publish(topic, 1, false, payload)
	if token.Wait() && token.Error() != nil {
		return target.DeviceCode, fmt.Errorf("publish: %w", token.Error())
	}
	return target.DeviceCode, nil
}

func (h *Handler) markFailed(commandID uint64, deviceCode string, cause error) {
	now := time.Now().UTC()
	_ = h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&ControlCommand{}).Where("id = ?", commandID).Updates(map[string]interface{}{
			"status": "FAILED",
		}).Error; err != nil {
			return err
		}
		receipt := ControlCommandReceipt{
			CommandID:     commandID,
			ReceiptSeq:    1,
			ReceiptStatus: "FAILED",
			AckCode:       "MQTT_DISPATCH_FAILED",
			AckMessage:    cause.Error(),
			AckPayload:    "{}",
			AckAt:         &now,
		}
		return tx.Create(&receipt).Error
	})

	h.eventHub.Publish(event.SSEEvent{
		Type: "command:dispatched",
		Data: event.CommandDispatchedSSEDataV1{
			SchemaVersion: 1,
			CommandID:     commandID,
			DeviceCode:    deviceCode,
			Status:        "FAILED",
			DispatchedAt:  now.Format(time.RFC3339),
			SourceType:    "MANUAL",
			ErrorMessage:  cause.Error(),
		},
	})
}

type actuatorTarget struct {
	DeviceCode  string
	ChannelCode string
}

// lookupActuatorTarget finds the device and channel identifiers for an actuator channel.
func (h *Handler) lookupActuatorTarget(actuatorChannelID uint64) (actuatorTarget, error) {
	var result actuatorTarget
	err := h.db.Table("actuator_channels").
		Select("actuator_devices.device_code, actuator_channels.channel_code").
		Joins("JOIN actuator_devices ON actuator_devices.id = actuator_channels.actuator_device_id").
		Where("actuator_channels.id = ?", actuatorChannelID).
		Scan(&result).Error
	if err != nil {
		return actuatorTarget{}, err
	}
	if result.DeviceCode == "" {
		return actuatorTarget{}, fmt.Errorf("device not found for channel %d", actuatorChannelID)
	}
	return result, nil
}

// AckCommand marks a command as ACKED with receipt information.
func (h *Handler) AckCommand(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var req AckCommandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	now := time.Now().UTC()

	err = h.db.Transaction(func(tx *gorm.DB) error {
		// Update command status to ACKED
		result := tx.Model(&ControlCommand{}).Where("id = ?", id).Updates(map[string]interface{}{
			"status":   "ACKED",
			"acked_at": now,
		})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		// Create a receipt entry
		ackPayload := "{}"
		if req.AckPayload != nil {
			b, err := json.Marshal(req.AckPayload)
			if err != nil {
				return err
			}
			ackPayload = string(b)
		}

		receipt := ControlCommandReceipt{
			CommandID:     id,
			ReceiptSeq:    1,
			ReceiptStatus: "ACKED",
			AckCode:       req.AckCode,
			AckMessage:    req.AckMessage,
			AckPayload:    ackPayload,
			AckAt:         &now,
		}
		return tx.Create(&receipt).Error
	})

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "ack_failed", nil)
		return
	}

	response.Success(c, gin.H{
		"id":       id,
		"status":   "ACKED",
		"acked_at": now.Format(time.RFC3339),
	})
}

// GetCommand retrieves a command with its receipts.
func (h *Handler) GetCommand(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var cmd ControlCommand
	if err := h.db.Preload("Receipts").First(&cmd, id).Error; err != nil {
		response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
		return
	}

	response.Success(c, toCommandResponse(cmd))
}

// ListCommands lists commands with optional filters for actuator_channel_id, status, and time range.
func (h *Handler) ListCommands(c *gin.Context) {
	page, size := parsePage(c)
	q := h.db.Model(&ControlCommand{})

	if v := c.Query("actuator_channel_id"); v != "" {
		q = q.Where("actuator_channel_id = ?", v)
	}
	if v := strings.TrimSpace(c.Query("status")); v != "" {
		q = q.Where("status = ?", strings.ToUpper(v))
	}
	if v := c.Query("command_type"); v != "" {
		q = q.Where("command_type = ?", v)
	}
	if v := c.Query("batch_id"); v != "" {
		id, err := strconv.ParseUint(v, 10, 64)
		if err == nil {
			q = q.Where("batch_id = ?", id)
		}
	}
	if v := c.Query("from"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_from", nil)
			return
		}
		q = q.Where("created_at >= ?", t.UTC())
	}
	if v := c.Query("to"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_to", nil)
			return
		}
		q = q.Where("created_at <= ?", t.UTC())
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	var commands []ControlCommand
	if total > 0 {
		if err := q.Order("id desc").Limit(size).Offset((page - 1) * size).Find(&commands).Error; err != nil {
			response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
			return
		}
	}

	items := make([]CommandResponse, 0, len(commands))
	for _, cmd := range commands {
		items = append(items, CommandResponse{
			ID:                cmd.ID,
			ActuatorChannelID: cmd.ActuatorChannelID,
			BatchID:           cmd.BatchID,
			CommandType:       cmd.CommandType,
			Payload:           cmd.Payload,
			Status:            cmd.Status,
			SentAt:            cmd.SentAt,
			AckedAt:           cmd.AckedAt,
			RequestID:         cmd.RequestID,
			CreatedBy:         cmd.CreatedBy,
			CreatedAt:         cmd.CreatedAt,
		})
	}

	response.Success(c, gin.H{
		"page":      page,
		"page_size": size,
		"total":     total,
		"items":     items,
	})
}

// --- ControlCommandReceipt handlers ---

// CreateReceipt adds a receipt for a command.
func (h *Handler) CreateReceipt(c *gin.Context) {
	commandID, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var req CreateReceiptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	// Verify command exists
	var count int64
	if err := h.db.Model(&ControlCommand{}).Where("id = ?", commandID).Count(&count).Error; err != nil || count == 0 {
		response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "command_not_found", nil)
		return
	}

	now := time.Now().UTC()
	ackPayload := "{}"
	if req.AckPayload != nil {
		b, err := json.Marshal(req.AckPayload)
		if err != nil {
			response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_ack_payload", nil)
			return
		}
		ackPayload = string(b)
	}

	receipt := ControlCommandReceipt{
		CommandID:     commandID,
		ReceiptSeq:    req.ReceiptSeq,
		ReceiptStatus: req.ReceiptStatus,
		AckCode:       req.AckCode,
		AckMessage:    req.AckMessage,
		AckPayload:    ackPayload,
		AckAt:         &now,
	}

	if err := h.db.Create(&receipt).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "create_failed", nil)
		return
	}

	response.Success(c, gin.H{"id": receipt.ID})
}

// ListReceipts lists all receipts for a command.
func (h *Handler) ListReceipts(c *gin.Context) {
	commandID, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var receipts []ControlCommandReceipt
	if err := h.db.Where("command_id = ?", commandID).Order("receipt_seq asc").Find(&receipts).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	items := make([]CommandReceiptResponse, 0, len(receipts))
	for _, r := range receipts {
		items = append(items, toReceiptResponse(r))
	}

	response.Success(c, gin.H{"items": items})
}

// GetReceipt retrieves a single receipt.
func (h *Handler) GetReceipt(c *gin.Context) {
	commandID, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}
	receiptID, err := parseID(c.Param("receiptId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_receipt_id", nil)
		return
	}

	var receipt ControlCommandReceipt
	if err := h.db.Where("id = ? AND command_id = ?", receiptID, commandID).First(&receipt).Error; err != nil {
		response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
		return
	}

	response.Success(c, toReceiptResponse(receipt))
}

// --- Helper functions ---

// toCommandResponse converts a ControlCommand with receipts to a response struct.
func toCommandResponse(cmd ControlCommand) CommandResponse {
	receipts := make([]CommandReceiptResponse, 0, len(cmd.Receipts))
	for _, r := range cmd.Receipts {
		receipts = append(receipts, toReceiptResponse(r))
	}

	return CommandResponse{
		ID:                cmd.ID,
		ActuatorChannelID: cmd.ActuatorChannelID,
		BatchID:           cmd.BatchID,
		CommandType:       cmd.CommandType,
		Payload:           cmd.Payload,
		Status:            cmd.Status,
		SentAt:            cmd.SentAt,
		AckedAt:           cmd.AckedAt,
		RequestID:         cmd.RequestID,
		CreatedBy:         cmd.CreatedBy,
		CreatedAt:         cmd.CreatedAt,
		Receipts:          receipts,
	}
}

// toReceiptResponse converts a ControlCommandReceipt to a response struct.
func toReceiptResponse(r ControlCommandReceipt) CommandReceiptResponse {
	return CommandReceiptResponse{
		ID:            r.ID,
		CommandID:     r.CommandID,
		ReceiptSeq:    r.ReceiptSeq,
		ReceiptStatus: r.ReceiptStatus,
		AckCode:       r.AckCode,
		AckMessage:    r.AckMessage,
		AckPayload:    r.AckPayload,
		AckAt:         r.AckAt,
		CreatedAt:     r.CreatedAt,
	}
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
