package policy

import (
	"net/http"

	platformErrors "hydroponic-backend/internal/platform/errors"
	"hydroponic-backend/internal/platform/response"

	"github.com/gin-gonic/gin"
)

// --- PolicyCondition handlers ---

// CreateCondition creates a condition under a policy.
func (h *Handler) CreateCondition(c *gin.Context) {
	policyID, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var req CreatePolicyConditionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	// Verify policy exists
	var count int64
	if err := h.db.Model(&ControlPolicy{}).Where("id = ?", policyID).Count(&count).Error; err != nil || count == 0 {
		response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "policy_not_found", nil)
		return
	}

	condition := PolicyCondition{
		PolicyID:            policyID,
		MetricCode:          req.MetricCode,
		Operator:            req.Operator,
		ThresholdValue:      req.ThresholdValue,
		Hysteresis:          req.Hysteresis,
		WindowSec:           req.WindowSec,
		RequiredDurationSec: req.RequiredDurationSec,
	}
	if req.Aggregation != nil {
		condition.Aggregation = *req.Aggregation
	}
	if req.Enabled != nil {
		condition.Enabled = *req.Enabled
	}

	if err := h.db.Create(&condition).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "create_failed", nil)
		return
	}
	response.Success(c, gin.H{"id": condition.ID})
}

// UpdateCondition updates an existing policy condition.
func (h *Handler) UpdateCondition(c *gin.Context) {
	policyID, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}
	conditionID, err := parseID(c.Param("conditionId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_condition_id", nil)
		return
	}

	var req UpdatePolicyConditionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	updates := map[string]interface{}{}
	if req.Operator != nil {
		updates["operator"] = *req.Operator
	}
	if req.ThresholdValue != nil {
		updates["threshold_value"] = *req.ThresholdValue
	}
	if req.Hysteresis != nil {
		updates["hysteresis"] = *req.Hysteresis
	}
	if req.WindowSec != nil {
		updates["window_sec"] = *req.WindowSec
	}
	if req.RequiredDurationSec != nil {
		updates["required_duration_sec"] = *req.RequiredDurationSec
	}
	if req.Aggregation != nil {
		updates["aggregation"] = *req.Aggregation
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}

	if len(updates) == 0 {
		response.Success(c, gin.H{})
		return
	}

	result := h.db.Model(&PolicyCondition{}).Where("id = ? AND policy_id = ?", conditionID, policyID).Updates(updates)
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

// DeleteCondition deletes a policy condition.
func (h *Handler) DeleteCondition(c *gin.Context) {
	policyID, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}
	conditionID, err := parseID(c.Param("conditionId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_condition_id", nil)
		return
	}

	result := h.db.Where("id = ? AND policy_id = ?", conditionID, policyID).Delete(&PolicyCondition{})
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

// ListConditions lists all conditions for a policy.
func (h *Handler) ListConditions(c *gin.Context) {
	policyID, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var conditions []PolicyCondition
	if err := h.db.Where("policy_id = ?", policyID).Order("id asc").Find(&conditions).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	items := make([]PolicyConditionResponse, 0, len(conditions))
	for _, c := range conditions {
		items = append(items, toConditionResponse(c))
	}
	response.Success(c, gin.H{"items": items})
}

// GetCondition retrieves a single condition.
func (h *Handler) GetCondition(c *gin.Context) {
	policyID, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}
	conditionID, err := parseID(c.Param("conditionId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_condition_id", nil)
		return
	}

	var condition PolicyCondition
	if err := h.db.Where("id = ? AND policy_id = ?", conditionID, policyID).First(&condition).Error; err != nil {
		response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
		return
	}
	response.Success(c, toConditionResponse(condition))
}

// toConditionResponse converts a PolicyCondition to a response struct.
func toConditionResponse(c PolicyCondition) PolicyConditionResponse {
	return PolicyConditionResponse{
		ID:                  c.ID,
		PolicyID:            c.PolicyID,
		MetricCode:          c.MetricCode,
		Operator:            c.Operator,
		ThresholdValue:      c.ThresholdValue,
		Hysteresis:          c.Hysteresis,
		WindowSec:           c.WindowSec,
		RequiredDurationSec: c.RequiredDurationSec,
		Aggregation:         c.Aggregation,
		Enabled:             c.Enabled,
		CreatedAt:           c.CreatedAt,
		UpdatedAt:           c.UpdatedAt,
	}
}
