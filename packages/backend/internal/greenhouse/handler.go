package greenhouse

import (
	"errors"
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

// ---- Greenhouse CRUD ----

func (h *Handler) CreateGreenhouse(c *gin.Context) {
	var req CreateGreenhouseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	var existing int64
	if err := h.db.Model(&Greenhouse{}).Where("code = ?", req.Code).Count(&existing).Error; err == nil && existing > 0 {
		response.Error(c, http.StatusConflict, platformErrors.CodeConflict, "code_exists", nil)
		return
	}

	greenhouse := Greenhouse{
		Code:        req.Code,
		Name:        req.Name,
		Location:    req.Location,
		AreaSqm:     req.AreaSqm,
		Description: req.Description,
		Status:      StatusEnabled,
	}

	if err := h.db.Create(&greenhouse).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "create_failed", nil)
		return
	}

	response.Success(c, gin.H{"id": greenhouse.ID})
}

func (h *Handler) UpdateGreenhouse(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var req UpdateGreenhouseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Location != "" {
		updates["location"] = req.Location
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.AreaSqm != 0 {
		updates["area_sqm"] = req.AreaSqm
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}

	if len(updates) > 0 {
		result := h.db.Model(&Greenhouse{}).Where("id = ?", id).Updates(updates)
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

func (h *Handler) ListGreenhouses(c *gin.Context) {
	page, pageSize := parsePage(c)

	query := h.db.Model(&Greenhouse{})
	if v := c.Query("status"); v != "" {
		query = query.Where("status = ?", v)
	}
	if v := c.Query("keyword"); v != "" {
		like := "%" + v + "%"
		query = query.Where("name LIKE ? OR code LIKE ?", like, like)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	var greenhouses []Greenhouse
	if total > 0 {
		if err := query.Order("id desc").Limit(pageSize).Offset((page - 1) * pageSize).Find(&greenhouses).Error; err != nil {
			response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
			return
		}
	}

	greenhouseIDs := make([]uint64, 0, len(greenhouses))
	for _, g := range greenhouses {
		greenhouseIDs = append(greenhouseIDs, g.ID)
	}

	zoneCounts := make(map[uint64]int64)
	if len(greenhouseIDs) > 0 {
		var rows []struct {
			GreenhouseID uint64 `gorm:"column:greenhouse_id"`
			Count        int64  `gorm:"column:count"`
		}
		h.db.Model(&GrowingZone{}).
			Select("greenhouse_id, COUNT(*) as count").
			Where("greenhouse_id IN ?", greenhouseIDs).
			Group("greenhouse_id").
			Scan(&rows)
		for _, r := range rows {
			zoneCounts[r.GreenhouseID] = r.Count
		}
	}

	items := make([]GreenhouseResponse, 0, len(greenhouses))
	for _, g := range greenhouses {
		items = append(items, GreenhouseResponse{
			ID:          g.ID,
			Code:        g.Code,
			Name:        g.Name,
			Location:    g.Location,
			AreaSqm:     g.AreaSqm,
			Description: g.Description,
			Status:      g.Status,
			CreatedAt:   g.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   g.UpdatedAt.Format(time.RFC3339),
			ZoneCount:   zoneCounts[g.ID],
		})
	}

	response.Success(c, gin.H{
		"page":      page,
		"page_size": pageSize,
		"total":     total,
		"items":     items,
	})
}

func (h *Handler) GetGreenhouse(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var g Greenhouse
	if err := h.db.First(&g, id).Error; err != nil {
		response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
		return
	}

	var zoneCount int64
	h.db.Model(&GrowingZone{}).Where("greenhouse_id = ?", id).Count(&zoneCount)

	response.Success(c, GreenhouseResponse{
		ID:          g.ID,
		Code:        g.Code,
		Name:        g.Name,
		Location:    g.Location,
		AreaSqm:     g.AreaSqm,
		Description: g.Description,
		Status:      g.Status,
		CreatedAt:   g.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   g.UpdatedAt.Format(time.RFC3339),
		ZoneCount:   zoneCount,
	})
}

func (h *Handler) DeleteGreenhouse(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	notFoundErr := errors.New("greenhouse_not_found")
	err = h.db.Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&Greenhouse{}).Where("id = ?", id).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return notFoundErr
		}

		// Nullify zones' greenhouse reference
		if err := tx.Model(&GrowingZone{}).Where("greenhouse_id = ?", id).Update("greenhouse_id", 0).Error; err != nil {
			return err
		}

		result := tx.Where("id = ?", id).Delete(&Greenhouse{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return notFoundErr
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, notFoundErr) {
			response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "delete_failed", nil)
		return
	}

	response.Success(c, gin.H{})
}

// ---- GrowingZone CRUD ----

func (h *Handler) CreateGrowingZone(c *gin.Context) {
	var req CreateGrowingZoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	// Verify greenhouse exists
	var ghCount int64
	if err := h.db.Model(&Greenhouse{}).Where("id = ?", req.GreenhouseID).Count(&ghCount).Error; err != nil || ghCount == 0 {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "greenhouse_not_found", nil)
		return
	}

	zone := GrowingZone{
		GreenhouseID:          req.GreenhouseID,
		Code:                  req.Code,
		Name:                  req.Name,
		SystemType:            req.SystemType,
		TankVolumeLiter:       req.TankVolumeLiter,
		PlantingDensityPerSqm: req.PlantingDensityPerSqm,
		Status:                StatusEnabled,
	}
	if zone.SystemType == "" {
		zone.SystemType = SystemTypeDWC
	}

	if err := h.db.Create(&zone).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "create_failed", nil)
		return
	}

	response.Success(c, gin.H{"id": zone.ID})
}

func (h *Handler) UpdateGrowingZone(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var req UpdateGrowingZoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.SystemType != "" {
		updates["system_type"] = req.SystemType
	}
	if req.TankVolumeLiter != 0 {
		updates["tank_volume_liter"] = req.TankVolumeLiter
	}
	if req.PlantingDensityPerSqm != 0 {
		updates["planting_density_per_sqm"] = req.PlantingDensityPerSqm
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}

	if len(updates) > 0 {
		result := h.db.Model(&GrowingZone{}).Where("id = ?", id).Updates(updates)
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

func (h *Handler) ListGrowingZones(c *gin.Context) {
	page, pageSize := parsePage(c)

	query := h.db.Model(&GrowingZone{})
	if v := c.Query("greenhouse_id"); v != "" {
		query = query.Where("greenhouse_id = ?", v)
	}
	if v := c.Query("status"); v != "" {
		query = query.Where("status = ?", v)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	var zones []GrowingZone
	if total > 0 {
		if err := query.Order("id desc").Limit(pageSize).Offset((page - 1) * pageSize).Find(&zones).Error; err != nil {
			response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
			return
		}
	}

	items := make([]GrowingZoneResponse, 0, len(zones))
	for _, z := range zones {
		items = append(items, GrowingZoneResponse{
			ID:                    z.ID,
			GreenhouseID:          z.GreenhouseID,
			Code:                  z.Code,
			Name:                  z.Name,
			SystemType:            z.SystemType,
			TankVolumeLiter:       z.TankVolumeLiter,
			PlantingDensityPerSqm: z.PlantingDensityPerSqm,
			Status:                z.Status,
			CreatedAt:             z.CreatedAt.Format(time.RFC3339),
			UpdatedAt:             z.UpdatedAt.Format(time.RFC3339),
		})
	}

	response.Success(c, gin.H{
		"page":      page,
		"page_size": pageSize,
		"total":     total,
		"items":     items,
	})
}

func (h *Handler) GetGrowingZone(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var z GrowingZone
	if err := h.db.First(&z, id).Error; err != nil {
		response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
		return
	}

	response.Success(c, GrowingZoneResponse{
		ID:                    z.ID,
		GreenhouseID:          z.GreenhouseID,
		Code:                  z.Code,
		Name:                  z.Name,
		SystemType:            z.SystemType,
		TankVolumeLiter:       z.TankVolumeLiter,
		PlantingDensityPerSqm: z.PlantingDensityPerSqm,
		Status:                z.Status,
		CreatedAt:             z.CreatedAt.Format(time.RFC3339),
		UpdatedAt:             z.UpdatedAt.Format(time.RFC3339),
	})
}

func (h *Handler) DeleteGrowingZone(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	notFoundErr := errors.New("zone_not_found")
	err = h.db.Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&GrowingZone{}).Where("id = ?", id).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return notFoundErr
		}

		result := tx.Where("id = ?", id).Delete(&GrowingZone{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return notFoundErr
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, notFoundErr) {
			response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "delete_failed", nil)
		return
	}

	response.Success(c, gin.H{})
}

// ListZonesByGreenhouse lists growing zones under a specific greenhouse
func (h *Handler) ListZonesByGreenhouse(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	// Verify greenhouse exists
	var ghCount int64
	if err := h.db.Model(&Greenhouse{}).Where("id = ?", id).Count(&ghCount).Error; err != nil || ghCount == 0 {
		response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "greenhouse_not_found", nil)
		return
	}

	page, pageSize := parsePage(c)
	query := h.db.Model(&GrowingZone{}).Where("greenhouse_id = ?", id)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	var zones []GrowingZone
	if total > 0 {
		if err := query.Order("id desc").Limit(pageSize).Offset((page - 1) * pageSize).Find(&zones).Error; err != nil {
			response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
			return
		}
	}

	items := make([]GrowingZoneResponse, 0, len(zones))
	for _, z := range zones {
		items = append(items, GrowingZoneResponse{
			ID:                    z.ID,
			GreenhouseID:          z.GreenhouseID,
			Code:                  z.Code,
			Name:                  z.Name,
			SystemType:            z.SystemType,
			TankVolumeLiter:       z.TankVolumeLiter,
			PlantingDensityPerSqm: z.PlantingDensityPerSqm,
			Status:                z.Status,
			CreatedAt:             z.CreatedAt.Format(time.RFC3339),
			UpdatedAt:             z.UpdatedAt.Format(time.RFC3339),
		})
	}

	response.Success(c, gin.H{
		"page":      page,
		"page_size": pageSize,
		"total":     total,
		"items":     items,
	})
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
