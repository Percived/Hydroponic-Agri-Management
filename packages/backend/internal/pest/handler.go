package pest

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

// =============================================
// Pest/Disease Observation Handlers
// =============================================

// CreateObservation creates a new pest/disease observation.
func (h *Handler) CreateObservation(c *gin.Context) {
	var req CreateObservationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	obs := PestDiseaseObservation{
		GreenhouseID:       req.GreenhouseID,
		GrowingZoneID:      req.GrowingZoneID,
		BatchID:            req.BatchID,
		ObservedAt:         req.ObservedAt,
		PestOrDisease:      req.PestOrDisease,
		Severity:           req.Severity,
		AffectedAreaPct:    req.AffectedAreaPct,
		AffectedPlantCount: req.AffectedPlantCount,
		Symptoms:           req.Symptoms,
		PhotoUrls:          req.PhotoUrls,
		ObservedBy:         req.ObservedBy,
	}

	if err := h.db.Create(&obs).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "create_failed", nil)
		return
	}

	response.Success(c, gin.H{"id": obs.ID})
}

// UpdateObservation updates an existing observation.
func (h *Handler) UpdateObservation(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var req UpdateObservationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	updates := map[string]interface{}{}
	if req.PestOrDisease != nil {
		updates["pest_or_disease"] = *req.PestOrDisease
	}
	if req.Severity != nil {
		updates["severity"] = *req.Severity
	}
	if req.AffectedAreaPct != nil {
		updates["affected_area_pct"] = *req.AffectedAreaPct
	}
	if req.AffectedPlantCount != nil {
		updates["affected_plant_count"] = *req.AffectedPlantCount
	}
	if req.Symptoms != nil {
		updates["symptoms"] = *req.Symptoms
	}
	if req.PhotoUrls != nil {
		updates["photo_urls"] = *req.PhotoUrls
	}

	if len(updates) == 0 {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "no_fields", nil)
		return
	}

	result := h.db.Model(&PestDiseaseObservation{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "update_failed", nil)
		return
	}
	if result.RowsAffected == 0 {
		response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
		return
	}

	response.Success(c, gin.H{})
}

// GetObservation returns a single observation.
func (h *Handler) GetObservation(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var obs PestDiseaseObservation
	if err := h.db.First(&obs, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	response.Success(c, observationToItem(obs))
}

// DeleteObservation deletes an observation.
func (h *Handler) DeleteObservation(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	result := h.db.Delete(&PestDiseaseObservation{}, id)
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

// ListObservations returns a paginated list with filters.
func (h *Handler) ListObservations(c *gin.Context) {
	page, pageSize := parsePage(c)

	query := h.db.Model(&PestDiseaseObservation{})
	if v := c.Query("severity"); v != "" {
		query = query.Where("severity = ?", v)
	}
	if v := c.Query("pest_or_disease"); v != "" {
		query = query.Where("pest_or_disease LIKE ?", "%"+v+"%")
	}
	if from := c.Query("observed_from"); from != "" {
		t, err := time.Parse(time.RFC3339, from)
		if err == nil {
			query = query.Where("observed_at >= ?", t)
		}
	}
	if to := c.Query("observed_to"); to != "" {
		t, err := time.Parse(time.RFC3339, to)
		if err == nil {
			query = query.Where("observed_at <= ?", t)
		}
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	var obs []PestDiseaseObservation
	if total > 0 {
		if err := query.Order("observed_at DESC, id DESC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&obs).Error; err != nil {
			response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
			return
		}
	}

	items := make([]gin.H, 0, len(obs))
	for _, o := range obs {
		items = append(items, observationToItem(o))
	}

	response.Success(c, gin.H{
		"page":      page,
		"page_size": pageSize,
		"total":     total,
		"items":     items,
	})
}

// ListObservationsByGreenhouse returns observations for a greenhouse.
func (h *Handler) ListObservationsByGreenhouse(c *gin.Context) {
	greenhouseID, err := parseID(c.Param("greenhouseId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	page, pageSize := parsePage(c)

	query := h.db.Model(&PestDiseaseObservation{}).Where("greenhouse_id = ?", greenhouseID)
	if v := c.Query("severity"); v != "" {
		query = query.Where("severity = ?", v)
	}
	if from := c.Query("observed_from"); from != "" {
		t, err := time.Parse(time.RFC3339, from)
		if err == nil {
			query = query.Where("observed_at >= ?", t)
		}
	}
	if to := c.Query("observed_to"); to != "" {
		t, err := time.Parse(time.RFC3339, to)
		if err == nil {
			query = query.Where("observed_at <= ?", t)
		}
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	var obs []PestDiseaseObservation
	if total > 0 {
		if err := query.Order("observed_at DESC, id DESC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&obs).Error; err != nil {
			response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
			return
		}
	}

	items := make([]gin.H, 0, len(obs))
	for _, o := range obs {
		items = append(items, observationToItem(o))
	}

	response.Success(c, gin.H{
		"page":      page,
		"page_size": pageSize,
		"total":     total,
		"items":     items,
	})
}

// ListObservationsByBatch returns observations for a batch.
func (h *Handler) ListObservationsByBatch(c *gin.Context) {
	batchID, err := parseID(c.Param("batchId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	page, pageSize := parsePage(c)

	query := h.db.Model(&PestDiseaseObservation{}).Where("batch_id = ?", batchID)
	if v := c.Query("severity"); v != "" {
		query = query.Where("severity = ?", v)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	var obs []PestDiseaseObservation
	if total > 0 {
		if err := query.Order("observed_at DESC, id DESC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&obs).Error; err != nil {
			response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
			return
		}
	}

	items := make([]gin.H, 0, len(obs))
	for _, o := range obs {
		items = append(items, observationToItem(o))
	}

	response.Success(c, gin.H{
		"page":      page,
		"page_size": pageSize,
		"total":     total,
		"items":     items,
	})
}

// GetTreatmentsForObservation returns all treatments linked to an observation.
func (h *Handler) GetTreatmentsForObservation(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	// Verify observation exists
	var count int64
	if err := h.db.Model(&PestDiseaseObservation{}).Where("id = ?", id).Count(&count).Error; err != nil || count == 0 {
		response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
		return
	}

	var treatments []TreatmentRecord
	if err := h.db.Where("observation_id = ?", id).Order("treated_at DESC, id DESC").Find(&treatments).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	items := make([]gin.H, 0, len(treatments))
	for _, t := range treatments {
		items = append(items, treatmentToItem(t))
	}

	response.Success(c, gin.H{"items": items})
}

// =============================================
// Treatment Record Handlers
// =============================================

// CreateTreatment creates a new treatment record.
func (h *Handler) CreateTreatment(c *gin.Context) {
	var req CreateTreatmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	treatment := TreatmentRecord{
		ObservationID:        req.ObservationID,
		GreenhouseID:         req.GreenhouseID,
		GrowingZoneID:        req.GrowingZoneID,
		BatchID:              req.BatchID,
		TreatmentType:        req.TreatmentType,
		ProductName:          req.ProductName,
		ActiveIngredient:     req.ActiveIngredient,
		Dosage:               req.Dosage,
		ApplicationMethod:    req.ApplicationMethod,
		SafetyIntervalDays:   req.SafetyIntervalDays,
		ReentryIntervalHours: req.ReentryIntervalHours,
		TreatedAt:            req.TreatedAt,
		TreatedBy:            req.TreatedBy,
		Note:                 req.Note,
	}

	if err := h.db.Create(&treatment).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "create_failed", nil)
		return
	}

	response.Success(c, gin.H{"id": treatment.ID})
}

// UpdateTreatment updates an existing treatment record.
func (h *Handler) UpdateTreatment(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var req UpdateTreatmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	updates := map[string]interface{}{}
	if req.TreatmentType != nil {
		updates["treatment_type"] = *req.TreatmentType
	}
	if req.ProductName != nil {
		updates["product_name"] = *req.ProductName
	}
	if req.ActiveIngredient != nil {
		updates["active_ingredient"] = *req.ActiveIngredient
	}
	if req.Dosage != nil {
		updates["dosage"] = *req.Dosage
	}
	if req.ApplicationMethod != nil {
		updates["application_method"] = *req.ApplicationMethod
	}
	if req.SafetyIntervalDays != nil {
		updates["safety_interval_days"] = *req.SafetyIntervalDays
	}
	if req.ReentryIntervalHours != nil {
		updates["reentry_interval_hours"] = *req.ReentryIntervalHours
	}
	if req.Note != nil {
		updates["note"] = *req.Note
	}

	if len(updates) == 0 {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "no_fields", nil)
		return
	}

	result := h.db.Model(&TreatmentRecord{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "update_failed", nil)
		return
	}
	if result.RowsAffected == 0 {
		response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
		return
	}

	response.Success(c, gin.H{})
}

// GetTreatment returns a single treatment record.
func (h *Handler) GetTreatment(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var treatment TreatmentRecord
	if err := h.db.First(&treatment, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	response.Success(c, treatmentToItem(treatment))
}

// DeleteTreatment deletes a treatment record.
func (h *Handler) DeleteTreatment(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	result := h.db.Delete(&TreatmentRecord{}, id)
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

// ListTreatments returns a paginated list of treatments.
func (h *Handler) ListTreatments(c *gin.Context) {
	page, pageSize := parsePage(c)

	query := h.db.Model(&TreatmentRecord{})
	if v := c.Query("treatment_type"); v != "" {
		query = query.Where("treatment_type = ?", v)
	}
	if v := c.Query("application_method"); v != "" {
		query = query.Where("application_method = ?", v)
	}
	if from := c.Query("treated_from"); from != "" {
		t, err := time.Parse(time.RFC3339, from)
		if err == nil {
			query = query.Where("treated_at >= ?", t)
		}
	}
	if to := c.Query("treated_to"); to != "" {
		t, err := time.Parse(time.RFC3339, to)
		if err == nil {
			query = query.Where("treated_at <= ?", t)
		}
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	var treatments []TreatmentRecord
	if total > 0 {
		if err := query.Order("treated_at DESC, id DESC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&treatments).Error; err != nil {
			response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
			return
		}
	}

	items := make([]gin.H, 0, len(treatments))
	for _, t := range treatments {
		items = append(items, treatmentToItem(t))
	}

	response.Success(c, gin.H{
		"page":      page,
		"page_size": pageSize,
		"total":     total,
		"items":     items,
	})
}

// ListTreatmentsByGreenhouse returns treatments for a greenhouse.
func (h *Handler) ListTreatmentsByGreenhouse(c *gin.Context) {
	greenhouseID, err := parseID(c.Param("greenhouseId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	page, pageSize := parsePage(c)

	query := h.db.Model(&TreatmentRecord{}).Where("greenhouse_id = ?", greenhouseID)
	if v := c.Query("treatment_type"); v != "" {
		query = query.Where("treatment_type = ?", v)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	var treatments []TreatmentRecord
	if total > 0 {
		if err := query.Order("treated_at DESC, id DESC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&treatments).Error; err != nil {
			response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
			return
		}
	}

	items := make([]gin.H, 0, len(treatments))
	for _, t := range treatments {
		items = append(items, treatmentToItem(t))
	}

	response.Success(c, gin.H{
		"page":      page,
		"page_size": pageSize,
		"total":     total,
		"items":     items,
	})
}

// ListTreatmentsByBatch returns treatments for a batch.
func (h *Handler) ListTreatmentsByBatch(c *gin.Context) {
	batchID, err := parseID(c.Param("batchId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	page, pageSize := parsePage(c)

	query := h.db.Model(&TreatmentRecord{}).Where("batch_id = ?", batchID)
	if v := c.Query("treatment_type"); v != "" {
		query = query.Where("treatment_type = ?", v)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	var treatments []TreatmentRecord
	if total > 0 {
		if err := query.Order("treated_at DESC, id DESC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&treatments).Error; err != nil {
			response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
			return
		}
	}

	items := make([]gin.H, 0, len(treatments))
	for _, t := range treatments {
		items = append(items, treatmentToItem(t))
	}

	response.Success(c, gin.H{
		"page":      page,
		"page_size": pageSize,
		"total":     total,
		"items":     items,
	})
}

// =============================================
// Helpers
// =============================================

func observationToItem(o PestDiseaseObservation) gin.H {
	return gin.H{
		"id":                   o.ID,
		"greenhouse_id":        o.GreenhouseID,
		"growing_zone_id":      o.GrowingZoneID,
		"batch_id":             o.BatchID,
		"observed_at":          timeToStr(o.ObservedAt),
		"pest_or_disease":      o.PestOrDisease,
		"severity":             o.Severity,
		"affected_area_pct":    o.AffectedAreaPct,
		"affected_plant_count": o.AffectedPlantCount,
		"symptoms":             o.Symptoms,
		"photo_urls":           o.PhotoUrls,
		"observed_by":          o.ObservedBy,
		"created_at":           timeToStr(o.CreatedAt),
	}
}

func treatmentToItem(t TreatmentRecord) gin.H {
	return gin.H{
		"id":                     t.ID,
		"observation_id":         t.ObservationID,
		"greenhouse_id":          t.GreenhouseID,
		"growing_zone_id":        t.GrowingZoneID,
		"batch_id":               t.BatchID,
		"treatment_type":         t.TreatmentType,
		"product_name":           t.ProductName,
		"active_ingredient":      t.ActiveIngredient,
		"dosage":                 t.Dosage,
		"application_method":     t.ApplicationMethod,
		"safety_interval_days":   t.SafetyIntervalDays,
		"reentry_interval_hours": t.ReentryIntervalHours,
		"treated_at":             timeToStr(t.TreatedAt),
		"treated_by":             t.TreatedBy,
		"note":                   t.Note,
		"created_at":             timeToStr(t.CreatedAt),
	}
}

func parseID(v string) (uint64, error) {
	return strconv.ParseUint(v, 10, 64)
}

func parsePage(c *gin.Context) (int, int) {
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
