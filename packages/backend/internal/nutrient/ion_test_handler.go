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

	result := h.db.Model(&IonTestRecord{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "update_failed", nil)
		return
	}
	if result.RowsAffected == 0 {
		response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
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

	result := h.db.Delete(&IonTestRecord{}, id)
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
