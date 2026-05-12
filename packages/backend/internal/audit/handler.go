package audit

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	platformErrors "hydroponic-backend/internal/platform/errors"
	"hydroponic-backend/internal/platform/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	db *gorm.DB
}

type listRow struct {
	AuditLog
	Username string `gorm:"column:username"`
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{db: db}
}

func (h *Handler) List(c *gin.Context) {
	page, size := parsePage(c)
	q := h.db.Table("audit_logs").
		Select("audit_logs.*, users.username").
		Joins("LEFT JOIN users ON users.id = audit_logs.user_id")

	if v := strings.TrimSpace(c.Query("user_id")); v != "" {
		q = q.Where("user_id = ?", v)
	}
	if v := strings.TrimSpace(c.Query("action")); v != "" {
		q = q.Where("action = ?", v)
	}

	if v := strings.TrimSpace(c.Query("start_time")); v != "" {
		t, err := parseTime(v)
		if err != nil {
			response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_start_time", nil)
			return
		}
		q = q.Where("created_at >= ?", t)
	}
	if v := strings.TrimSpace(c.Query("end_time")); v != "" {
		t, err := parseTime(v)
		if err != nil {
			response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_end_time", nil)
			return
		}
		q = q.Where("created_at <= ?", t)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	items := []listRow{}
	if total > 0 {
		if err := q.Order("audit_logs.id desc").Limit(size).Offset((page - 1) * size).Scan(&items).Error; err != nil {
			response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
			return
		}
	}

	rows := make([]gin.H, 0, len(items))
	for _, it := range items {
		rows = append(rows, gin.H{
			"id":          it.ID,
			"user_id":     it.UserID,
			"username":    it.Username,
			"action":      it.Action,
			"target_type": it.TargetType,
			"target_id":   it.TargetID,
			"detail":      stringifyJSON(it.Detail),
			"request_id":  it.RequestID,
			"created_at":  it.CreatedAt,
		})
	}

	response.Success(c, gin.H{
		"page":      page,
		"page_size": size,
		"total":     total,
		"items":     rows,
	})
}

func parsePage(c *gin.Context) (int, int) {
	page := 1
	if v := c.Query("page"); v != "" {
		if i, err := strconv.Atoi(v); err == nil && i > 0 {
			page = i
		}
	}

	size := 20
	if v := c.Query("page_size"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			size = i
		}
	}
	if size < 1 {
		size = 20
	}
	if size > 200 {
		size = 200
	}
	return page, size
}

func parseTime(v string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339Nano, v)
	if err != nil {
		return time.Time{}, err
	}
	return t.UTC(), nil
}

func stringifyJSON(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}

	var out bytes.Buffer
	if err := json.Compact(&out, raw); err == nil {
		return out.String()
	}

	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		return text
	}

	return string(raw)
}
