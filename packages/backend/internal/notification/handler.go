package notification

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

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

func (h *Handler) ListChannels(c *gin.Context) {
	userID := currentUserID(c)

	var channels []NotificationChannel
	if err := h.db.Where("user_id = ?", userID).Order("id DESC").Find(&channels).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	type configPayload map[string]interface{}
	items := make([]gin.H, 0, len(channels))
	for _, ch := range channels {
		var cfg configPayload
		if err := json.Unmarshal(ch.Config, &cfg); err != nil {
			cfg = configPayload{}
		}
		items = append(items, gin.H{
			"id":              ch.ID,
			"user_id":         ch.UserID,
			"channel_type":    ch.ChannelType,
			"name":            ch.Name,
			"config":          cfg,
			"min_alert_level": ch.MinAlertLevel,
			"enabled":         ch.Enabled,
			"created_at":      ch.CreatedAt,
			"updated_at":      ch.UpdatedAt,
		})
	}

	response.Success(c, gin.H{"items": items})
}

func (h *Handler) CreateChannel(c *gin.Context) {
	var req CreateChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	configBytes, err := json.Marshal(req.Config)
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_config", nil)
		return
	}

	minLevel := req.MinAlertLevel
	if minLevel == "" {
		minLevel = "WARN"
	}

	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	channel := NotificationChannel{
		UserID:        currentUserID(c),
		ChannelType:   req.ChannelType,
		Name:          req.Name,
		Config:        configBytes,
		MinAlertLevel: minLevel,
		Enabled:       enabled,
	}

	if err := h.db.Create(&channel).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "create_failed", nil)
		return
	}

	response.Success(c, gin.H{"id": channel.ID})
}

func (h *Handler) UpdateChannel(c *gin.Context) {
	id, err := parseID(c.Param("channelId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var req UpdateChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Config != nil {
		b, err := json.Marshal(req.Config)
		if err != nil {
			response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_config", nil)
			return
		}
		updates["config"] = b
	}
	if req.MinAlertLevel != nil {
		updates["min_alert_level"] = *req.MinAlertLevel
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}

	if len(updates) == 0 {
		response.Success(c, gin.H{})
		return
	}

	userID := currentUserID(c)
	if err := h.db.Model(&NotificationChannel{}).Where("id = ? AND user_id = ?", id, userID).Updates(updates).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "update_failed", nil)
		return
	}

	response.Success(c, gin.H{})
}

func (h *Handler) DeleteChannel(c *gin.Context) {
	id, err := parseID(c.Param("channelId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	userID := currentUserID(c)
	if err := h.db.Where("id = ? AND user_id = ?", id, userID).Delete(&NotificationChannel{}).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "delete_failed", nil)
		return
	}

	response.Success(c, gin.H{})
}

func (h *Handler) TestChannel(c *gin.Context) {
	id, err := parseID(c.Param("channelId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var channel NotificationChannel
	userID := currentUserID(c)
	if err := h.db.Where("id = ? AND user_id = ?", id, userID).First(&channel).Error; err != nil {
		response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
		return
	}

	ok := h.sendTestNotification(channel)
	response.Success(c, gin.H{"sent": ok})
}

func (h *Handler) sendTestNotification(channel NotificationChannel) bool {
	if channel.ChannelType != ChannelWebhook {
		return false
	}

	var cfg struct {
		URL    string `json:"url"`
		Secret string `json:"secret"`
	}
	if err := json.Unmarshal(channel.Config, &cfg); err != nil || cfg.URL == "" {
		return false
	}

	body, _ := json.Marshal(gin.H{
		"type":       "test",
		"channel_id": channel.ID,
		"message":    "Test notification from Hydroponic System",
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
	})

	req, err := http.NewRequest("POST", cfg.URL, bytes.NewReader(body))
	if err != nil {
		return false
	}
	req.Header.Set("Content-Type", "application/json")

	if cfg.Secret != "" {
		mac := hmac.New(sha256.New, []byte(cfg.Secret))
		mac.Write(body)
		sig := hex.EncodeToString(mac.Sum(nil))
		req.Header.Set("X-Signature", sig)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode >= 200 && resp.StatusCode < 300
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
