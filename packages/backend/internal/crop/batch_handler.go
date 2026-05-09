package crop

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

// ======================== CropBatch Handlers ========================

func (h *Handler) CreateBatch(c *gin.Context) {
	var req CreateCropBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	status := req.Status
	if status == "" {
		status = BatchStatusPlanned
	}

	userID := currentUserID(c)

	batch := CropBatch{
		BatchNo:           req.BatchNo,
		GreenhouseID:      req.GreenhouseID,
		GrowingZoneID:     req.GrowingZoneID,
		CropVarietyID:     req.CropVarietyID,
		Status:            status,
		PlantingDensity:   req.PlantingDensity,
		TotalPlants:       req.TotalPlants,
		StartedAt:         parseTimePtr(req.StartedAt),
		EndedAt:           parseTimePtr(req.EndedAt),
		ExpectedHarvestAt: parseTimePtr(req.ExpectedHarvestAt),
		RecipeVersion:     req.RecipeVersion,
		PolicyVersion:     req.PolicyVersion,
		Note:              req.Note,
		CreatedBy:         &userID,
	}

	if err := h.db.Create(&batch).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "create_failed", nil)
		return
	}

	resp := h.toBatchResponse(batch)
	response.Success(c, resp)
}

func (h *Handler) GetBatch(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var batch CropBatch
	if err := h.db.First(&batch, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	resp := h.toBatchResponse(batch)
	response.Success(c, resp)
}

func (h *Handler) UpdateBatch(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var req UpdateCropBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	updates := map[string]interface{}{}
	if req.BatchNo != nil {
		updates["batch_no"] = *req.BatchNo
	}
	if req.GreenhouseID != nil {
		updates["greenhouse_id"] = *req.GreenhouseID
	}
	if req.GrowingZoneID != nil {
		updates["growing_zone_id"] = *req.GrowingZoneID
	}
	if req.CropVarietyID != nil {
		updates["crop_variety_id"] = *req.CropVarietyID
	}
	if req.PlantingDensity != nil {
		updates["planting_density"] = *req.PlantingDensity
	}
	if req.TotalPlants != nil {
		updates["total_plants"] = *req.TotalPlants
	}
	if req.StartedAt != nil {
		updates["started_at"] = parseTimePtr(req.StartedAt)
	}
	if req.EndedAt != nil {
		updates["ended_at"] = parseTimePtr(req.EndedAt)
	}
	if req.ExpectedHarvestAt != nil {
		updates["expected_harvest_at"] = parseTimePtr(req.ExpectedHarvestAt)
	}
	if req.RecipeVersion != nil {
		updates["recipe_version"] = *req.RecipeVersion
	}
	if req.PolicyVersion != nil {
		updates["policy_version"] = *req.PolicyVersion
	}
	if req.Note != nil {
		updates["note"] = *req.Note
	}

	if len(updates) == 0 {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "no_fields_to_update", nil)
		return
	}

	result := h.db.Model(&CropBatch{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "update_failed", nil)
		return
	}
	if result.RowsAffected == 0 {
		response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
		return
	}

	var batch CropBatch
	h.db.First(&batch, id)
	resp := h.toBatchResponse(batch)
	response.Success(c, resp)
}

func (h *Handler) DeleteBatch(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	result := h.db.Delete(&CropBatch{}, id)
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

func (h *Handler) ListBatches(c *gin.Context) {
	page, pageSize := parsePageQuery(c)

	q := h.db.Model(&CropBatch{})

	if v := strings.TrimSpace(c.Query("greenhouse_id")); v != "" {
		id, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_greenhouse_id", nil)
			return
		}
		q = q.Where("greenhouse_id = ?", id)
	}

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

	if v := strings.TrimSpace(c.Query("crop_variety_id")); v != "" {
		id, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_crop_variety_id", nil)
			return
		}
		q = q.Where("crop_variety_id = ?", id)
	}

	if v := strings.TrimSpace(c.Query("start_time")); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			t, err = time.Parse(time.RFC3339Nano, v)
			if err != nil {
				response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_start_time", nil)
				return
			}
		}
		q = q.Where("started_at >= ?", t.UTC())
	}

	if v := strings.TrimSpace(c.Query("end_time")); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			t, err = time.Parse(time.RFC3339Nano, v)
			if err != nil {
				response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_end_time", nil)
				return
			}
		}
		q = q.Where("started_at <= ?", t.UTC())
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	var batches []CropBatch
	if total > 0 {
		if err := q.Order("created_at DESC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&batches).Error; err != nil {
			response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
			return
		}
	}

	items := make([]CropBatchResponse, 0, len(batches))
	for _, b := range batches {
		items = append(items, h.toBatchResponse(b))
	}

	response.Success(c, CropListResponse{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

func (h *Handler) TransitionBatchStatus(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var req BatchStatusTransitionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	// Fetch current batch to validate state transition
	var batch CropBatch
	if err := h.db.First(&batch, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	// Validate state transition
	allowed, ok := BatchStatusTransitions[batch.Status]
	if !ok {
		response.Error(c, http.StatusConflict, platformErrors.CodeValidationError, "invalid_current_status", nil)
		return
	}

	legal := false
	for _, s := range allowed {
		if s == req.Status {
			legal = true
			break
		}
	}
	if !legal {
		response.Error(c, http.StatusConflict, platformErrors.CodeValidationError, "illegal_status_transition", nil)
		return
	}

	now := time.Now().UTC()
	updates := map[string]interface{}{
		"status": req.Status,
	}

	switch req.Status {
	case BatchStatusRunning:
		updates["started_at"] = now
	case BatchStatusCompleted:
		updates["ended_at"] = now
	case BatchStatusAborted:
		updates["ended_at"] = now
	}

	result := h.db.Model(&CropBatch{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "update_failed", nil)
		return
	}
	if result.RowsAffected == 0 {
		response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
		return
	}

	h.db.First(&batch, id)
	resp := h.toBatchResponse(batch)
	response.Success(c, resp)
}

// ======================== BatchDevice Handlers ========================

// BindDevice binds a sensor or actuator device to a batch.
func (h *Handler) BindDevice(c *gin.Context) {
	batchID, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var req BindDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	// Verify batch exists
	var count int64
	if err := h.db.Model(&CropBatch{}).Where("id = ?", batchID).Count(&count).Error; err != nil || count == 0 {
		response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "batch_not_found", nil)
		return
	}

	now := time.Now().UTC()

	// Check existing binding (including inactive ones from previous unbind)
	var existing BatchDevice
	err = h.db.Where("batch_id = ? AND device_type = ? AND device_id = ?", batchID, req.DeviceType, req.DeviceID).First(&existing).Error
	if err == nil {
		// Row exists — check if already active
		if existing.IsActive {
			response.Error(c, http.StatusConflict, platformErrors.CodeConflict, "device_already_bound", nil)
			return
		}
		// Re-activate previous binding
		h.db.Model(&existing).Updates(map[string]interface{}{
			"is_active":  true,
			"bound_at":   now,
			"unbound_at": nil,
		})
		response.Success(c, gin.H{"id": existing.ID})
		return
	}

	// No existing row — create new
	bd := BatchDevice{
		BatchID:    batchID,
		DeviceType: req.DeviceType,
		DeviceID:   req.DeviceID,
		BoundAt:    now,
		IsActive:   true,
	}

	if err := h.db.Create(&bd).Error; err != nil {
		response.Error(c, http.StatusConflict, platformErrors.CodeConflict, "device_already_bound", nil)
		return
	}

	response.Success(c, gin.H{"id": bd.ID})
}

// UnbindDevice unbinds a device from a batch.
func (h *Handler) UnbindDevice(c *gin.Context) {
	batchID, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	deviceID, err := parseID(c.Param("deviceId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_device_id", nil)
		return
	}

	deviceType := c.Query("device_type")
	if deviceType == "" {
		deviceType = "sensor"
	}

	now := time.Now().UTC()
	result := h.db.Model(&BatchDevice{}).
		Where("batch_id = ? AND device_type = ? AND device_id = ? AND is_active = 1", batchID, deviceType, deviceID).
		Updates(map[string]interface{}{
			"is_active":  false,
			"unbound_at": now,
		})

	if result.Error != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "unbind_failed", nil)
		return
	}
	if result.RowsAffected == 0 {
		response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "binding_not_found", nil)
		return
	}

	response.Success(c, gin.H{})
}

// ListBatchDevices returns all devices bound to a batch.
func (h *Handler) ListBatchDevices(c *gin.Context) {
	batchID, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	deviceType := c.Query("device_type")

	query := h.db.Model(&BatchDevice{}).Where("batch_id = ? AND is_active = 1", batchID)
	if deviceType != "" {
		query = query.Where("device_type = ?", deviceType)
	}

	var bindings []BatchDevice
	if err := query.Order("bound_at DESC").Find(&bindings).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	items := make([]BatchDeviceResponse, 0, len(bindings))
	for _, bd := range bindings {
		resp := BatchDeviceResponse{
			ID:         bd.ID,
			BatchID:    bd.BatchID,
			DeviceType: bd.DeviceType,
			DeviceID:   bd.DeviceID,
			IsActive:   bd.IsActive,
			BoundAt:    bd.BoundAt.Format(time.RFC3339),
		}
		if bd.UnboundAt != nil {
			s := bd.UnboundAt.Format(time.RFC3339)
			resp.UnboundAt = &s
		}

		// Load device name/code based on type
		if bd.DeviceType == DeviceTypeSensor {
			var dev struct {
				Name string `gorm:"column:name"`
				Code string `gorm:"column:device_code"`
			}
			if err := h.db.Table("sensor_devices").Select("name, device_code").Where("id = ?", bd.DeviceID).First(&dev).Error; err == nil {
				resp.DeviceName = dev.Name
				resp.DeviceCode = dev.Code
			}
		} else if bd.DeviceType == DeviceTypeActuator {
			var dev struct {
				Name string `gorm:"column:name"`
				Code string `gorm:"column:device_code"`
			}
			if err := h.db.Table("actuator_devices").Select("name, device_code").Where("id = ?", bd.DeviceID).First(&dev).Error; err == nil {
				resp.DeviceName = dev.Name
				resp.DeviceCode = dev.Code
			}
		}

		items = append(items, resp)
	}

	response.Success(c, gin.H{"items": items})
}

// ======================== Stage Progress ========================

// GetBatchStageProgress calculates the current stage and progress percentage for a batch.
func (h *Handler) GetBatchStageProgress(c *gin.Context) {
	batchID, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	// Verify batch exists
	var batch CropBatch
	if err := h.db.First(&batch, batchID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	now := time.Now().UTC()

	// Find the current stage plan (stage_start_at <= now AND stage_end_at > now)
	var currentPlan struct {
		BatchStagePlan
		StageName string `gorm:"column:stage_name"`
		StageCode string `gorm:"column:stage_code"`
	}
	err = h.db.Table("batch_stage_plans bsp").
		Select("bsp.*, gs.name AS stage_name, gs.code AS stage_code").
		Joins("JOIN growth_stages gs ON gs.id = bsp.growth_stage_id").
		Where("bsp.batch_id = ? AND bsp.stage_start_at <= ? AND bsp.stage_end_at > ?", batchID, now, now).
		Order("bsp.stage_start_at ASC").
		Limit(1).
		Scan(&currentPlan).Error

	if err != nil || currentPlan.ID == 0 {
		// No current stage — check if there's a future stage
		var nextPlan struct {
			BatchStagePlan
			StageName string `gorm:"column:stage_name"`
			StageCode string `gorm:"column:stage_code"`
		}
		err = h.db.Table("batch_stage_plans bsp").
			Select("bsp.*, gs.name AS stage_name, gs.code AS stage_code").
			Joins("JOIN growth_stages gs ON gs.id = bsp.growth_stage_id").
			Where("bsp.batch_id = ? AND bsp.stage_start_at > ?", batchID, now).
			Order("bsp.stage_start_at ASC").
			Limit(1).
			Scan(&nextPlan).Error

		if err != nil || nextPlan.ID == 0 {
			response.Success(c, StageProgressResponse{
				BatchID:          batchID,
				CurrentStageID:   nil,
				CurrentStageName: "",
				CurrentStageCode: "",
				ProgressPercent:  0,
			})
			return
		}

		// Future stage — 0% progress
		resp := StageProgressResponse{
			BatchID:          batchID,
			CurrentStageID:   &nextPlan.ID,
			CurrentStageName: nextPlan.StageName,
			CurrentStageCode: nextPlan.StageCode,
			ProgressPercent:  0,
			DaysRemaining:    int(nextPlan.StageEndAt.Sub(now).Hours() / 24),
			TargetECMin:      nextPlan.TargetECMin,
			TargetECMax:      nextPlan.TargetECMax,
			TargetPHMin:      nextPlan.TargetPHMin,
			TargetPHMax:      nextPlan.TargetPHMax,
		}
		response.Success(c, resp)
		return
	}

	// Calculate progress within the current stage
	totalDuration := currentPlan.StageEndAt.Sub(currentPlan.StageStartAt).Hours()
	elapsed := now.Sub(currentPlan.StageStartAt).Hours()
	progress := (elapsed / totalDuration) * 100
	if progress > 100 {
		progress = 100
	}
	if progress < 0 {
		progress = 0
	}

	daysElapsed := int(elapsed / 24)
	daysRemaining := int((totalDuration - elapsed) / 24)
	if daysRemaining < 0 {
		daysRemaining = 0
	}

	resp := StageProgressResponse{
		BatchID:          batchID,
		CurrentStageID:   &currentPlan.ID,
		CurrentStageName: currentPlan.StageName,
		CurrentStageCode: currentPlan.StageCode,
		ProgressPercent:  float64(int(progress*10)) / 10, // 1 decimal place
		DaysElapsed:      daysElapsed,
		DaysRemaining:    daysRemaining,
		TargetECMin:      currentPlan.TargetECMin,
		TargetECMax:      currentPlan.TargetECMax,
		TargetPHMin:      currentPlan.TargetPHMin,
		TargetPHMax:      currentPlan.TargetPHMax,
	}

	response.Success(c, resp)
}
