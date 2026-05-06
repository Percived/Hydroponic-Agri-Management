package metric

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

// ListMetrics returns all metric definitions with optional filtering and pagination.
func (h *Handler) ListMetrics(c *gin.Context) {
	page, pageSize := parsePage(c)

	query := h.db.Model(&MetricDefinition{})
	if v := c.Query("status"); v != "" {
		query = query.Where("status = ?", v)
	}
	if v := c.Query("is_core"); v != "" {
		query = query.Where("is_core = ?", v)
	}
	if v := c.Query("keyword"); v != "" {
		like := "%" + v + "%"
		query = query.Where("code LIKE ? OR name LIKE ?", like, like)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	var metrics []MetricDefinition
	if total > 0 {
		if err := query.Order("id asc").Limit(pageSize).Offset((page - 1) * pageSize).Find(&metrics).Error; err != nil {
			response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
			return
		}
	}

	items := make([]MetricDefinitionResponse, 0, len(metrics))
	for _, m := range metrics {
		items = append(items, metricToResponse(m))
	}

	response.Success(c, gin.H{
		"page":      page,
		"page_size": pageSize,
		"total":     total,
		"items":     items,
	})
}

// GetMetric returns a single metric definition by ID.
func (h *Handler) GetMetric(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var m MetricDefinition
	if err := h.db.First(&m, id).Error; err != nil {
		response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
		return
	}

	response.Success(c, metricToResponse(m))
}

// metricToResponse converts a MetricDefinition to its response DTO.
func metricToResponse(m MetricDefinition) MetricDefinitionResponse {
	return MetricDefinitionResponse{
		ID:              m.ID,
		Code:            m.Code,
		Name:            m.Name,
		Unit:            m.Unit,
		PrecisionDigits: m.PrecisionDigits,
		NormalRangeMin:  m.NormalRangeMin,
		NormalRangeMax:  m.NormalRangeMax,
		IsCore:          m.IsCore,
		Status:          m.Status,
		CreatedAt:       m.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       m.UpdatedAt.Format(time.RFC3339),
	}
}

// ---- Helpers ----

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
