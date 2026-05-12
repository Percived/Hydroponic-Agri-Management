package nutrient

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

// ======================== NutrientTank Handlers ========================

func (h *Handler) CreateTank(c *gin.Context) {
	var req CreateNutrientTankRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	status := req.Status
	if status == "" {
		status = TankStatusActive
	}

	tank := NutrientTank{
		GrowingZoneID:        req.GrowingZoneID,
		Code:                 req.Code,
		TotalVolumeLiter:     req.TotalVolumeLiter,
		CurrentVolumeLiter:   req.CurrentVolumeLiter,
		Status:               status,
		ECSensorChannelID:    req.ECSensorChannelID,
		PHSensorChannelID:    req.PHSensorChannelID,
		LevelSensorChannelID: req.LevelSensorChannelID,
		TempSensorChannelID:  req.TempSensorChannelID,
	}

	if err := h.db.Create(&tank).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "create_failed", nil)
		return
	}

	response.Success(c, toTankResponse(tank))
}

func (h *Handler) GetTank(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var tank NutrientTank
	if err := h.db.First(&tank, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	response.Success(c, toTankResponse(tank))
}

func (h *Handler) UpdateTank(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var req UpdateNutrientTankRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	updates := map[string]interface{}{}
	if req.Code != nil {
		updates["code"] = *req.Code
	}
	if req.TotalVolumeLiter != nil {
		updates["total_volume_liter"] = *req.TotalVolumeLiter
	}
	if req.CurrentVolumeLiter != nil {
		updates["current_volume_liter"] = *req.CurrentVolumeLiter
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.ECSensorChannelID.Set {
		updates["ec_sensor_channel_id"] = req.ECSensorChannelID.Value
	}
	if req.PHSensorChannelID.Set {
		updates["ph_sensor_channel_id"] = req.PHSensorChannelID.Value
	}
	if req.LevelSensorChannelID.Set {
		updates["level_sensor_channel_id"] = req.LevelSensorChannelID.Value
	}
	if req.TempSensorChannelID.Set {
		updates["temp_sensor_channel_id"] = req.TempSensorChannelID.Value
	}

	if len(updates) == 0 {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "no_fields_to_update", nil)
		return
	}

	result := h.db.Model(&NutrientTank{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "update_failed", nil)
		return
	}
	if result.RowsAffected == 0 {
		response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
		return
	}

	var tank NutrientTank
	h.db.First(&tank, id)
	response.Success(c, toTankResponse(tank))
}

func (h *Handler) DeleteTank(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	result := h.db.Delete(&NutrientTank{}, id)
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

func (h *Handler) ListTanks(c *gin.Context) {
	page, pageSize := parsePageParam(c)

	q := h.db.Model(&NutrientTank{})

	if v := strings.TrimSpace(c.Query("growing_zone_id")); v != "" {
		id, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_growing_zone_id", nil)
			return
		}
		q = q.Where("growing_zone_id = ?", id)
	}

	if v := strings.TrimSpace(c.Query("status")); v != "" {
		q = q.Where("status = ?", v)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	var tanks []NutrientTank
	if total > 0 {
		if err := q.Order("created_at DESC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&tanks).Error; err != nil {
			response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
			return
		}
	}

	items := make([]NutrientTankResponse, 0, len(tanks))
	for _, tank := range tanks {
		items = append(items, toTankResponse(tank))
	}

	response.Success(c, NutrientListResponse{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

func toTankResponse(t NutrientTank) NutrientTankResponse {
	return NutrientTankResponse{
		ID:                   t.ID,
		GrowingZoneID:        t.GrowingZoneID,
		Code:                 t.Code,
		TotalVolumeLiter:     t.TotalVolumeLiter,
		CurrentVolumeLiter:   t.CurrentVolumeLiter,
		Status:               t.Status,
		ECSensorChannelID:    t.ECSensorChannelID,
		PHSensorChannelID:    t.PHSensorChannelID,
		LevelSensorChannelID: t.LevelSensorChannelID,
		TempSensorChannelID:  t.TempSensorChannelID,
		CreatedAt:            t.CreatedAt.Format(time.RFC3339),
		UpdatedAt:            t.UpdatedAt.Format(time.RFC3339),
	}
}
