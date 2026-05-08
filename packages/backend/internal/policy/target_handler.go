package policy

import (
	"encoding/json"
	"net/http"

	platformErrors "hydroponic-backend/internal/platform/errors"
	"hydroponic-backend/internal/platform/response"

	"github.com/gin-gonic/gin"
)

// --- PolicyTarget handlers ---

// CreateTarget creates a target under a policy.
func (h *Handler) CreateTarget(c *gin.Context) {
	policyID, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var req CreatePolicyTargetRequest
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

	payloadBytes, err := json.Marshal(req.CommandPayload)
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_payload", nil)
		return
	}

	target := PolicyTarget{
		PolicyID:          policyID,
		ActuatorChannelID: req.ActuatorChannelID,
		CommandType:       req.CommandType,
		CommandPayload:    string(payloadBytes),
	}
	if req.ExecutionOrder != nil {
		target.ExecutionOrder = *req.ExecutionOrder
	}
	if req.Enabled != nil {
		target.Enabled = *req.Enabled
	}

	if err := h.db.Create(&target).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "create_failed", nil)
		return
	}
	response.Success(c, gin.H{"id": target.ID})
}

// UpdateTarget updates an existing policy target.
func (h *Handler) UpdateTarget(c *gin.Context) {
	policyID, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}
	targetID, err := parseID(c.Param("targetId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_target_id", nil)
		return
	}

	var req UpdatePolicyTargetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	updates := map[string]interface{}{}
	if req.ActuatorChannelID != nil {
		updates["actuator_channel_id"] = *req.ActuatorChannelID
	}
	if req.CommandType != nil {
		updates["command_type"] = *req.CommandType
	}
	if req.CommandPayload != nil {
		payloadBytes, err := json.Marshal(req.CommandPayload)
		if err != nil {
			response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_payload", nil)
			return
		}
		updates["command_payload"] = string(payloadBytes)
	}
	if req.ExecutionOrder != nil {
		updates["execution_order"] = *req.ExecutionOrder
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}

	if len(updates) == 0 {
		response.Success(c, gin.H{})
		return
	}

	result := h.db.Model(&PolicyTarget{}).Where("id = ? AND policy_id = ?", targetID, policyID).Updates(updates)
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

// DeleteTarget deletes a policy target.
func (h *Handler) DeleteTarget(c *gin.Context) {
	policyID, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}
	targetID, err := parseID(c.Param("targetId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_target_id", nil)
		return
	}

	result := h.db.Where("id = ? AND policy_id = ?", targetID, policyID).Delete(&PolicyTarget{})
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

// ListTargets lists all targets for a policy.
func (h *Handler) ListTargets(c *gin.Context) {
	policyID, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var targets []PolicyTarget
	if err := h.db.Where("policy_id = ?", policyID).Order("execution_order asc").Find(&targets).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	items := make([]PolicyTargetResponse, 0, len(targets))
	for _, t := range targets {
		items = append(items, toTargetResponse(t))
	}
	response.Success(c, gin.H{"items": items})
}

// GetTarget retrieves a single target.
func (h *Handler) GetTarget(c *gin.Context) {
	policyID, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}
	targetID, err := parseID(c.Param("targetId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_target_id", nil)
		return
	}

	var target PolicyTarget
	if err := h.db.Where("id = ? AND policy_id = ?", targetID, policyID).First(&target).Error; err != nil {
		response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
		return
	}
	response.Success(c, toTargetResponse(target))
}

// toTargetResponse converts a PolicyTarget to a response struct.
func toTargetResponse(t PolicyTarget) PolicyTargetResponse {
	return PolicyTargetResponse{
		ID:                t.ID,
		PolicyID:          t.PolicyID,
		ActuatorChannelID: t.ActuatorChannelID,
		CommandType:       t.CommandType,
		CommandPayload:    t.CommandPayload,
		ExecutionOrder:    t.ExecutionOrder,
		Enabled:           t.Enabled,
		CreatedAt:         t.CreatedAt,
		UpdatedAt:         t.UpdatedAt,
	}
}
