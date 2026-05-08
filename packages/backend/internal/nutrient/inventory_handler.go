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

// ======================== NutrientConcentrateInventory Handlers ========================

func (h *Handler) CreateConcentrateInventory(c *gin.Context) {
	var req CreateConcentrateInventoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	status := req.Status
	if status == "" {
		status = InventoryStatusInUse
	}

	var expiredAt *time.Time
	if req.ExpiredAt != nil && *req.ExpiredAt != "" {
		t, err := time.Parse("2006-01-02", *req.ExpiredAt)
		if err != nil {
			response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_expired_at", nil)
			return
		}
		t = t.UTC()
		expiredAt = &t
	}

	inventory := NutrientConcentrateInventory{
		GreenhouseID:      req.GreenhouseID,
		ConcentrateType:   req.ConcentrateType,
		Brand:             req.Brand,
		ProductName:       req.ProductName,
		TotalVolumeMl:     req.TotalVolumeMl,
		RemainingVolumeMl: req.RemainingVolumeMl,
		UnitPrice:         req.UnitPrice,
		BatchNo:           req.BatchNo,
		ExpiredAt:         expiredAt,
		Status:            status,
	}

	if err := h.db.Create(&inventory).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "create_failed", nil)
		return
	}

	response.Success(c, toConcentrateInventoryResponse(inventory))
}

func (h *Handler) GetConcentrateInventory(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var inventory NutrientConcentrateInventory
	if err := h.db.First(&inventory, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	response.Success(c, toConcentrateInventoryResponse(inventory))
}

func (h *Handler) UpdateConcentrateInventory(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var req UpdateConcentrateInventoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	updates := map[string]interface{}{}
	if req.ConcentrateType != nil {
		updates["concentrate_type"] = *req.ConcentrateType
	}
	if req.Brand != nil {
		updates["brand"] = *req.Brand
	}
	if req.ProductName != nil {
		updates["product_name"] = *req.ProductName
	}
	if req.TotalVolumeMl != nil {
		updates["total_volume_ml"] = *req.TotalVolumeMl
	}
	if req.RemainingVolumeMl != nil {
		updates["remaining_volume_ml"] = *req.RemainingVolumeMl
	}
	if req.UnitPrice != nil {
		updates["unit_price"] = *req.UnitPrice
	}
	if req.BatchNo != nil {
		updates["batch_no"] = *req.BatchNo
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.ExpiredAt != nil {
		if *req.ExpiredAt != "" {
			t, err := time.Parse("2006-01-02", *req.ExpiredAt)
			if err != nil {
				response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_expired_at", nil)
				return
			}
			t = t.UTC()
			updates["expired_at"] = t
		} else {
			updates["expired_at"] = nil
		}
	}

	if len(updates) == 0 {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "no_fields_to_update", nil)
		return
	}

	result := h.db.Model(&NutrientConcentrateInventory{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "update_failed", nil)
		return
	}
	if result.RowsAffected == 0 {
		response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
		return
	}

	var inventory NutrientConcentrateInventory
	h.db.First(&inventory, id)
	response.Success(c, toConcentrateInventoryResponse(inventory))
}

func (h *Handler) DeleteConcentrateInventory(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	result := h.db.Delete(&NutrientConcentrateInventory{}, id)
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

func (h *Handler) ListConcentrateInventory(c *gin.Context) {
	page, pageSize := parsePageParam(c)

	q := h.db.Model(&NutrientConcentrateInventory{})

	if v := strings.TrimSpace(c.Query("greenhouse_id")); v != "" {
		id, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_greenhouse_id", nil)
			return
		}
		q = q.Where("greenhouse_id = ?", id)
	}

	if v := strings.TrimSpace(c.Query("concentrate_type")); v != "" {
		q = q.Where("concentrate_type = ?", v)
	}

	if v := strings.TrimSpace(c.Query("status")); v != "" {
		q = q.Where("status = ?", v)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	var inventories []NutrientConcentrateInventory
	if total > 0 {
		if err := q.Order("created_at DESC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&inventories).Error; err != nil {
			response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
			return
		}
	}

	items := make([]ConcentrateInventoryResponse, 0, len(inventories))
	for _, inv := range inventories {
		items = append(items, toConcentrateInventoryResponse(inv))
	}

	response.Success(c, NutrientListResponse{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

// ======================== ConcentrateUsageLog Handlers ========================

func (h *Handler) CreateUsageLog(c *gin.Context) {
	var req CreateConcentrateUsageLogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	usedAt, err := time.Parse(time.RFC3339, req.UsedAt)
	if err != nil {
		usedAt, err = time.Parse(time.RFC3339Nano, req.UsedAt)
		if err != nil {
			response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_used_at", nil)
			return
		}
	}

	userID := currentUserID(c)

	log := ConcentrateUsageLog{
		InventoryID:      req.InventoryID,
		SolutionChangeID: req.SolutionChangeID,
		TankID:           req.TankID,
		VolumeUsedMl:     req.VolumeUsedMl,
		UsedBy:           &userID,
		UsedAt:           usedAt.UTC(),
	}

	if err := h.db.Create(&log).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "create_failed", nil)
		return
	}

	// Also decrement the inventory remaining volume
	h.db.Model(&NutrientConcentrateInventory{}).
		Where("id = ?", req.InventoryID).
		UpdateColumn("remaining_volume_ml", gorm.Expr("remaining_volume_ml - ?", req.VolumeUsedMl))

	response.Success(c, toUsageLogResponse(log))
}

func (h *Handler) ListUsageLogs(c *gin.Context) {
	page, pageSize := parsePageParam(c)

	q := h.db.Model(&ConcentrateUsageLog{})

	if v := strings.TrimSpace(c.Query("inventory_id")); v != "" {
		id, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_inventory_id", nil)
			return
		}
		q = q.Where("inventory_id = ?", id)
	}

	if v := strings.TrimSpace(c.Query("tank_id")); v != "" {
		id, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_tank_id", nil)
			return
		}
		q = q.Where("tank_id = ?", id)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	var logs []ConcentrateUsageLog
	if total > 0 {
		if err := q.Order("used_at DESC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&logs).Error; err != nil {
			response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
			return
		}
	}

	items := make([]ConcentrateUsageLogResponse, 0, len(logs))
	for _, l := range logs {
		items = append(items, toUsageLogResponse(l))
	}

	response.Success(c, NutrientListResponse{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

func toConcentrateInventoryResponse(inv NutrientConcentrateInventory) ConcentrateInventoryResponse {
	resp := ConcentrateInventoryResponse{
		ID:                inv.ID,
		GreenhouseID:      inv.GreenhouseID,
		ConcentrateType:   inv.ConcentrateType,
		Brand:             inv.Brand,
		ProductName:       inv.ProductName,
		TotalVolumeMl:     inv.TotalVolumeMl,
		RemainingVolumeMl: inv.RemainingVolumeMl,
		UnitPrice:         inv.UnitPrice,
		BatchNo:           inv.BatchNo,
		Status:            inv.Status,
		CreatedAt:         inv.CreatedAt.Format(time.RFC3339),
		UpdatedAt:         inv.UpdatedAt.Format(time.RFC3339),
	}
	if inv.ExpiredAt != nil {
		s := inv.ExpiredAt.Format("2006-01-02")
		resp.ExpiredAt = &s
	}
	return resp
}

func toUsageLogResponse(l ConcentrateUsageLog) ConcentrateUsageLogResponse {
	return ConcentrateUsageLogResponse{
		ID:               l.ID,
		InventoryID:      l.InventoryID,
		SolutionChangeID: l.SolutionChangeID,
		TankID:           l.TankID,
		VolumeUsedMl:     l.VolumeUsedMl,
		UsedBy:           l.UsedBy,
		UsedAt:           l.UsedAt.Format(time.RFC3339),
		CreatedAt:        l.CreatedAt.Format(time.RFC3339),
	}
}
