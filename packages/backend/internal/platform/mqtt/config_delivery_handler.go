package mqtt

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	platformErrors "hydroponic-backend/internal/platform/errors"
	"hydroponic-backend/internal/platform/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ConfigDeliveryHandler struct {
	db *gorm.DB
}

func NewConfigDeliveryHandler(db *gorm.DB) *ConfigDeliveryHandler {
	return &ConfigDeliveryHandler{db: db}
}

type ConfigDeliveryResponse struct {
	ID             uint64  `json:"id"`
	MsgID          string  `json:"msg_id"`
	TraceID        string  `json:"trace_id"`
	DeviceCode     string  `json:"device_code"`
	ConfigType     string  `json:"config_type"`
	Action         string  `json:"action"`
	EntityID       uint64  `json:"entity_id"`
	EntityRev      uint64  `json:"entity_rev"`
	SchemaVersion  int     `json:"schema_version"`
	IssuedAtMS     uint64  `json:"issued_at_ms"`
	TTLsec         int     `json:"ttl_sec"`
	RequireAck     bool    `json:"require_ack"`
	Status         string  `json:"status"`
	RetryCount     int     `json:"retry_count"`
	NextRetryAt    *string `json:"next_retry_at"`
	SentAt         *string `json:"sent_at"`
	AckedAt        *string `json:"acked_at"`
	LastErrorCode  string  `json:"last_error_code"`
	LastErrorMsg   string  `json:"last_error_message"`
	AppliedHash    string  `json:"applied_hash"`
	DeviceFWVer    string  `json:"device_fw_version"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
	RequestPayload string  `json:"request_payload,omitempty"`
	AckPayload     *string `json:"ack_payload,omitempty"`
}

func toConfigDeliveryResponse(d ConfigDelivery, withPayload bool) ConfigDeliveryResponse {
	var nextRetryAt *string
	if d.NextRetryAt != nil {
		s := d.NextRetryAt.UTC().Format(time.RFC3339)
		nextRetryAt = &s
	}
	var sentAt *string
	if d.SentAt != nil {
		s := d.SentAt.UTC().Format(time.RFC3339)
		sentAt = &s
	}
	var ackedAt *string
	if d.AckedAt != nil {
		s := d.AckedAt.UTC().Format(time.RFC3339)
		ackedAt = &s
	}
	out := ConfigDeliveryResponse{
		ID:            d.ID,
		MsgID:         d.MsgID,
		TraceID:       d.TraceID,
		DeviceCode:    d.DeviceCode,
		ConfigType:    d.ConfigType,
		Action:        d.Action,
		EntityID:      d.EntityID,
		EntityRev:     d.EntityRev,
		SchemaVersion: d.SchemaVersion,
		IssuedAtMS:    d.IssuedAtMS,
		TTLsec:        d.TTLsec,
		RequireAck:    d.RequireAck,
		Status:        d.Status,
		RetryCount:    d.RetryCount,
		NextRetryAt:   nextRetryAt,
		SentAt:        sentAt,
		AckedAt:       ackedAt,
		LastErrorCode: d.LastErrorCode,
		LastErrorMsg:  d.LastErrorMsg,
		AppliedHash:   d.AppliedHash,
		DeviceFWVer:   d.DeviceFWVer,
		CreatedAt:     d.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:     d.UpdatedAt.UTC().Format(time.RFC3339),
	}
	if withPayload {
		out.RequestPayload = d.RequestPayload
		out.AckPayload = d.AckPayload
	}
	return out
}

func (h *ConfigDeliveryHandler) List(c *gin.Context) {
	page, pageSize := parsePageQS(c.Query("page"), c.Query("page_size"))
	q := h.db.Model(&ConfigDelivery{})

	if v := strings.TrimSpace(c.Query("device_code")); v != "" {
		q = q.Where("device_code = ?", v)
	}
	if v := strings.TrimSpace(c.Query("config_type")); v != "" {
		q = q.Where("config_type = ?", v)
	}
	if v := strings.TrimSpace(c.Query("status")); v != "" {
		q = q.Where("status = ?", v)
	}
	if v := strings.TrimSpace(c.Query("msg_id")); v != "" {
		q = q.Where("msg_id = ?", v)
	}
	if v := strings.TrimSpace(c.Query("entity_id")); v != "" {
		if id, err := strconv.ParseUint(v, 10, 64); err == nil {
			q = q.Where("entity_id = ?", id)
		}
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	var rows []ConfigDelivery
	if total > 0 {
		if err := q.Order("id desc").Limit(pageSize).Offset((page - 1) * pageSize).Find(&rows).Error; err != nil {
			response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
			return
		}
	}

	items := make([]ConfigDeliveryResponse, 0, len(rows))
	for _, r := range rows {
		items = append(items, toConfigDeliveryResponse(r, false))
	}
	response.Success(c, gin.H{
		"page":      page,
		"page_size": pageSize,
		"total":     total,
		"items":     items,
	})
}

func (h *ConfigDeliveryHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var d ConfigDelivery
	if err := h.db.First(&d, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	response.Success(c, toConfigDeliveryResponse(d, true))
}

func parsePageQS(pageStr string, pageSizeStr string) (int, int) {
	page := 1
	pageSize := 20
	if v := strings.TrimSpace(pageStr); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			page = n
		}
	}
	if v := strings.TrimSpace(pageSizeStr); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			pageSize = n
		}
	}
	if pageSize > 200 {
		pageSize = 200
	}
	return page, pageSize
}
