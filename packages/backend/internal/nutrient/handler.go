package nutrient

import (
	"net/http"
	"strconv"
	"strings"
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
		GrowingZoneID:      req.GrowingZoneID,
		Code:               req.Code,
		TotalVolumeLiter:   req.TotalVolumeLiter,
		CurrentVolumeLiter: req.CurrentVolumeLiter,
		Status:             status,
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

	if len(updates) == 0 {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "no_fields_to_update", nil)
		return
	}

	if err := h.db.Model(&NutrientTank{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "update_failed", nil)
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

	if err := h.db.Delete(&NutrientTank{}, id).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "delete_failed", nil)
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

// ======================== IonTestRecord Handlers ========================

func (h *Handler) CreateIonTest(c *gin.Context) {
	var req CreateIonTestRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	sampledAt, err := time.Parse(time.RFC3339, req.SampledAt)
	if err != nil {
		sampledAt, err = time.Parse(time.RFC3339Nano, req.SampledAt)
		if err != nil {
			response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_sampled_at", nil)
			return
		}
	}

	testMethod := req.TestMethod
	if testMethod == "" {
		testMethod = TestMethodLab
	}

	userID := currentUserID(c)

	var testedAt *time.Time
	if req.TestedAt != nil && *req.TestedAt != "" {
		t, err := time.Parse(time.RFC3339, *req.TestedAt)
		if err != nil {
			t, err = time.Parse(time.RFC3339Nano, *req.TestedAt)
			if err != nil {
				response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_tested_at", nil)
				return
			}
		}
		t = t.UTC()
		testedAt = &t
	}

	record := IonTestRecord{
		TankID:     req.TankID,
		BatchID:    req.BatchID,
		SampleCode: req.SampleCode,
		SampledAt:  sampledAt.UTC(),
		TestedAt:   testedAt,
		TestMethod: testMethod,
		NO3N:       req.NO3N,
		NH4N:       req.NH4N,
		P:          req.P,
		K:          req.K,
		Ca:         req.Ca,
		Mg:         req.Mg,
		S:          req.S,
		Fe:         req.Fe,
		Mn:         req.Mn,
		Zn:         req.Zn,
		B:          req.B,
		Cu:         req.Cu,
		Mo:         req.Mo,
		ECAtSample: req.ECAtSample,
		PHAtSample: req.PHAtSample,
		LabName:    req.LabName,
		ReportURL:  req.ReportURL,
		Note:       req.Note,
		CreatedBy:  &userID,
	}

	if err := h.db.Create(&record).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "create_failed", nil)
		return
	}

	response.Success(c, toIonTestResponse(record))
}

func (h *Handler) GetIonTest(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var record IonTestRecord
	if err := h.db.First(&record, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	response.Success(c, toIonTestResponse(record))
}

func (h *Handler) UpdateIonTest(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var req UpdateIonTestRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	updates := map[string]interface{}{}
	if req.SampleCode != nil {
		updates["sample_code"] = *req.SampleCode
	}
	if req.TestMethod != nil {
		updates["test_method"] = *req.TestMethod
	}
	setField(updates, "tested_at", parseTimePtr(req.TestedAt))
	setField(updates, "no3_n", req.NO3N)
	setField(updates, "nh4_n", req.NH4N)
	setField(updates, "p", req.P)
	setField(updates, "k", req.K)
	setField(updates, "ca", req.Ca)
	setField(updates, "mg", req.Mg)
	setField(updates, "s", req.S)
	setField(updates, "fe", req.Fe)
	setField(updates, "mn", req.Mn)
	setField(updates, "zn", req.Zn)
	setField(updates, "b", req.B)
	setField(updates, "cu", req.Cu)
	setField(updates, "mo", req.Mo)
	setField(updates, "ec_at_sample", req.ECAtSample)
	setField(updates, "ph_at_sample", req.PHAtSample)
	setField(updates, "lab_name", req.LabName)
	setField(updates, "report_url", req.ReportURL)
	setField(updates, "note", req.Note)

	if len(updates) == 0 {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "no_fields_to_update", nil)
		return
	}

	if err := h.db.Model(&IonTestRecord{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "update_failed", nil)
		return
	}

	var record IonTestRecord
	h.db.First(&record, id)
	response.Success(c, toIonTestResponse(record))
}

func (h *Handler) DeleteIonTest(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	if err := h.db.Delete(&IonTestRecord{}, id).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "delete_failed", nil)
		return
	}

	response.Success(c, gin.H{})
}

func (h *Handler) ListIonTests(c *gin.Context) {
	page, pageSize := parsePageParam(c)

	q := h.db.Model(&IonTestRecord{})

	if v := strings.TrimSpace(c.Query("tank_id")); v != "" {
		id, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_tank_id", nil)
			return
		}
		q = q.Where("tank_id = ?", id)
	}

	if v := strings.TrimSpace(c.Query("batch_id")); v != "" {
		id, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_batch_id", nil)
			return
		}
		q = q.Where("batch_id = ?", id)
	}

	if v := strings.TrimSpace(c.Query("test_method")); v != "" {
		q = q.Where("test_method = ?", v)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	var records []IonTestRecord
	if total > 0 {
		if err := q.Order("sampled_at DESC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&records).Error; err != nil {
			response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
			return
		}
	}

	items := make([]IonTestRecordResponse, 0, len(records))
	for _, r := range records {
		items = append(items, toIonTestResponse(r))
	}

	response.Success(c, NutrientListResponse{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

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

	if err := h.db.Model(&NutrientConcentrateInventory{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "update_failed", nil)
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

	if err := h.db.Delete(&NutrientConcentrateInventory{}, id).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "delete_failed", nil)
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

// ======================== Helpers ========================

func toTankResponse(t NutrientTank) NutrientTankResponse {
	return NutrientTankResponse{
		ID:                 t.ID,
		GrowingZoneID:      t.GrowingZoneID,
		Code:               t.Code,
		TotalVolumeLiter:   t.TotalVolumeLiter,
		CurrentVolumeLiter: t.CurrentVolumeLiter,
		Status:             t.Status,
		CreatedAt:          t.CreatedAt.Format(time.RFC3339),
		UpdatedAt:          t.UpdatedAt.Format(time.RFC3339),
	}
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

func toIonTestResponse(r IonTestRecord) IonTestRecordResponse {
	resp := IonTestRecordResponse{
		ID:         r.ID,
		TankID:     r.TankID,
		BatchID:    r.BatchID,
		SampleCode: r.SampleCode,
		SampledAt:  r.SampledAt.Format(time.RFC3339),
		TestMethod: r.TestMethod,
		NO3N:       r.NO3N,
		NH4N:       r.NH4N,
		P:          r.P,
		K:          r.K,
		Ca:         r.Ca,
		Mg:         r.Mg,
		S:          r.S,
		Fe:         r.Fe,
		Mn:         r.Mn,
		Zn:         r.Zn,
		B:          r.B,
		Cu:         r.Cu,
		Mo:         r.Mo,
		ECAtSample: r.ECAtSample,
		PHAtSample: r.PHAtSample,
		LabName:    r.LabName,
		ReportURL:  r.ReportURL,
		Note:       r.Note,
		CreatedBy:  r.CreatedBy,
		CreatedAt:  r.CreatedAt.Format(time.RFC3339),
	}
	if r.TestedAt != nil {
		s := r.TestedAt.Format(time.RFC3339)
		resp.TestedAt = &s
	}
	return resp
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

// ---------- Generic helpers ----------

func parseID(v string) (uint64, error) {
	return strconv.ParseUint(v, 10, 64)
}

func parsePageParam(c *gin.Context) (int, int) {
	page := 1
	if v := c.Query("page"); v != "" {
		if i, err := strconv.Atoi(v); err == nil && i > 0 {
			page = i
		}
	}
	pageSize := 20
	if v := c.Query("page_size"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			pageSize = i
		}
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 200 {
		pageSize = 200
	}
	return page, pageSize
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

func parseTimePtr(s *string) *time.Time {
	if s == nil || *s == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339, *s)
	if err != nil {
		t, err = time.Parse(time.RFC3339Nano, *s)
		if err != nil {
			return nil
		}
	}
	t = t.UTC()
	return &t
}

func setField(updates map[string]interface{}, key string, val interface{}) {
	if val != nil {
		// Handle pointer types via reflection-like check
		switch v := val.(type) {
		case *string:
			if v != nil {
				updates[key] = *v
			}
		case *float64:
			if v != nil {
				updates[key] = *v
			}
		case *time.Time:
			if v != nil {
				updates[key] = *v
			}
		default:
			updates[key] = val
		}
	}
}
