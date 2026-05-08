package crop

import (
	"net/http"
	"strings"

	platformErrors "hydroponic-backend/internal/platform/errors"
	"hydroponic-backend/internal/platform/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ======================== CropVariety Handlers ========================

func (h *Handler) CreateVariety(c *gin.Context) {
	var req CreateCropVarietyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	variety := CropVariety{
		Code:             req.Code,
		Name:             req.Name,
		Description:      req.Description,
		DefaultCycleDays: req.DefaultCycleDays,
	}

	if err := h.db.Create(&variety).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "create_failed", nil)
		return
	}

	response.Success(c, toVarietyResponse(variety))
}

func (h *Handler) GetVariety(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var variety CropVariety
	if err := h.db.First(&variety, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	response.Success(c, toVarietyResponse(variety))
}

func (h *Handler) UpdateVariety(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var req UpdateCropVarietyRequest
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
	if req.DefaultCycleDays != nil {
		updates["default_cycle_days"] = *req.DefaultCycleDays
	}

	if len(updates) == 0 {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "no_fields_to_update", nil)
		return
	}

	result := h.db.Model(&CropVariety{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "update_failed", nil)
		return
	}
	if result.RowsAffected == 0 {
		response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
		return
	}

	var variety CropVariety
	h.db.First(&variety, id)
	response.Success(c, toVarietyResponse(variety))
}

func (h *Handler) DeleteVariety(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	result := h.db.Delete(&CropVariety{}, id)
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

func (h *Handler) ListVarieties(c *gin.Context) {
	page, pageSize := parsePageQuery(c)

	q := h.db.Model(&CropVariety{})

	if v := strings.TrimSpace(c.Query("code")); v != "" {
		q = q.Where("code LIKE ?", "%"+v+"%")
	}
	if v := strings.TrimSpace(c.Query("name")); v != "" {
		q = q.Where("name LIKE ?", "%"+v+"%")
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	var varieties []CropVariety
	if total > 0 {
		if err := q.Order("code ASC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&varieties).Error; err != nil {
			response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
			return
		}
	}

	items := make([]CropVarietyResponse, 0, len(varieties))
	for _, v := range varieties {
		items = append(items, toVarietyResponse(v))
	}

	response.Success(c, CropListResponse{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}
