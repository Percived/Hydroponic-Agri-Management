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

// ======================== SolutionChangeEvent Handlers ========================

func (h *Handler) CreateSolutionChange(c *gin.Context) {
	var req CreateSolutionChangeEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	operatedAt, err := time.Parse(time.RFC3339, req.OperatedAt)
	if err != nil {
		operatedAt, err = time.Parse(time.RFC3339Nano, req.OperatedAt)
		if err != nil {
			response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_operated_at", nil)
			return
		}
	}

	userID := currentUserID(c)

	event := SolutionChangeEvent{
		TankID:              req.TankID,
		ChangeType:          req.ChangeType,
		VolumeReplacedLiter: req.VolumeReplacedLiter,
		SourceWaterEC:       req.SourceWaterEC,
		SourceWaterPH:       req.SourceWaterPH,
		BeforeEC:            req.BeforeEC,
		BeforePH:            req.BeforePH,
		AfterEC:             req.AfterEC,
		AfterPH:             req.AfterPH,
		NutrientAAddedMl:    req.NutrientAAddedMl,
		NutrientBAddedMl:    req.NutrientBAddedMl,
		AcidAddedMl:         req.AcidAddedMl,
		AlkaliAddedMl:       req.AlkaliAddedMl,
		Note:                req.Note,
		OperatedBy:          &userID,
		OperatedAt:          operatedAt.UTC(),
	}

	if err := h.db.Create(&event).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "create_failed", nil)
		return
	}

	response.Success(c, toSolutionChangeResponse(event))
}

func (h *Handler) GetSolutionChange(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var event SolutionChangeEvent
	if err := h.db.First(&event, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	response.Success(c, toSolutionChangeResponse(event))
}

func (h *Handler) ListSolutionChanges(c *gin.Context) {
	page, pageSize := parsePageParam(c)

	q := h.db.Model(&SolutionChangeEvent{})

	if v := strings.TrimSpace(c.Query("tank_id")); v != "" {
		id, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_tank_id", nil)
			return
		}
		q = q.Where("tank_id = ?", id)
	}

	if v := strings.TrimSpace(c.Query("change_type")); v != "" {
		q = q.Where("change_type = ?", v)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	var events []SolutionChangeEvent
	if total > 0 {
		if err := q.Order("operated_at DESC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&events).Error; err != nil {
			response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
			return
		}
	}

	items := make([]SolutionChangeEventResponse, 0, len(events))
	for _, e := range events {
		items = append(items, toSolutionChangeResponse(e))
	}

	response.Success(c, NutrientListResponse{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

func toSolutionChangeResponse(e SolutionChangeEvent) SolutionChangeEventResponse {
	return SolutionChangeEventResponse{
		ID:                  e.ID,
		TankID:              e.TankID,
		ChangeType:          e.ChangeType,
		VolumeReplacedLiter: e.VolumeReplacedLiter,
		SourceWaterEC:       e.SourceWaterEC,
		SourceWaterPH:       e.SourceWaterPH,
		BeforeEC:            e.BeforeEC,
		BeforePH:            e.BeforePH,
		AfterEC:             e.AfterEC,
		AfterPH:             e.AfterPH,
		NutrientAAddedMl:    e.NutrientAAddedMl,
		NutrientBAddedMl:    e.NutrientBAddedMl,
		AcidAddedMl:         e.AcidAddedMl,
		AlkaliAddedMl:       e.AlkaliAddedMl,
		Note:                e.Note,
		OperatedBy:          e.OperatedBy,
		OperatedAt:          e.OperatedAt.Format(time.RFC3339),
		CreatedAt:           e.CreatedAt.Format(time.RFC3339),
	}
}
