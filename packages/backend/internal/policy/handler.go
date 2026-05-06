package policy

import (
	"encoding/json"
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

// Handler holds dependencies for policy HTTP handlers.
type Handler struct {
	db *gorm.DB
}

// NewHandler creates a new policy Handler.
func NewHandler(db *gorm.DB) *Handler {
	return &Handler{db: db}
}

// --- ControlPolicy handlers ---

// CreatePolicy creates a new control policy.
func (h *Handler) CreatePolicy(c *gin.Context) {
	var req CreatePolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	userID := currentUserID(c)
	policy := ControlPolicy{
		PolicyCode:    req.PolicyCode,
		Name:          req.Name,
		PolicyType:    req.PolicyType,
		GreenhouseID:  req.GreenhouseID,
		GrowingZoneID: req.GrowingZoneID,
		CreatedBy:     &userID,
	}
	if req.Priority != nil {
		policy.Priority = *req.Priority
	}
	if req.RetryLimit != nil {
		policy.RetryLimit = *req.RetryLimit
	}
	if req.TimeoutSec != nil {
		policy.TimeoutSec = *req.TimeoutSec
	}
	if req.Enabled != nil {
		policy.Enabled = *req.Enabled
	}
	if req.Version != nil {
		policy.Version = *req.Version
	}
	if req.EffectiveFrom != nil {
		policy.EffectiveFrom = req.EffectiveFrom
	}
	if req.EffectiveTo != nil {
		policy.EffectiveTo = req.EffectiveTo
	}

	if err := h.db.Create(&policy).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "create_failed", nil)
		return
	}
	response.Success(c, gin.H{"id": policy.ID})
}

// CreatePolicyWithNested creates a policy with nested conditions and targets in one transaction.
func (h *Handler) CreatePolicyWithNested(c *gin.Context) {
	var req CreatePolicyWithNestedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	userID := currentUserID(c)
	var policyID uint64

	err := h.db.Transaction(func(tx *gorm.DB) error {
		policy := ControlPolicy{
			PolicyCode:    req.PolicyCode,
			Name:          req.Name,
			PolicyType:    req.PolicyType,
			GreenhouseID:  req.GreenhouseID,
			GrowingZoneID: req.GrowingZoneID,
			CreatedBy:     &userID,
		}
		if req.Priority != nil {
			policy.Priority = *req.Priority
		}
		if req.RetryLimit != nil {
			policy.RetryLimit = *req.RetryLimit
		}
		if req.TimeoutSec != nil {
			policy.TimeoutSec = *req.TimeoutSec
		}
		if req.Enabled != nil {
			policy.Enabled = *req.Enabled
		}
		if req.Version != nil {
			policy.Version = *req.Version
		}
		if req.EffectiveFrom != nil {
			policy.EffectiveFrom = req.EffectiveFrom
		}
		if req.EffectiveTo != nil {
			policy.EffectiveTo = req.EffectiveTo
		}
		if err := tx.Create(&policy).Error; err != nil {
			return err
		}
		policyID = policy.ID

		for _, condReq := range req.Conditions {
			condition := PolicyCondition{
				PolicyID:            policy.ID,
				MetricCode:          condReq.MetricCode,
				Operator:            condReq.Operator,
				ThresholdValue:      condReq.ThresholdValue,
				Hysteresis:          condReq.Hysteresis,
				WindowSec:           condReq.WindowSec,
				RequiredDurationSec: condReq.RequiredDurationSec,
			}
			if condReq.Aggregation != nil {
				condition.Aggregation = *condReq.Aggregation
			}
			if condReq.Enabled != nil {
				condition.Enabled = *condReq.Enabled
			}
			if err := tx.Create(&condition).Error; err != nil {
				return err
			}

			for _, targetReq := range condReq.Targets {
				payloadBytes, err := json.Marshal(targetReq.CommandPayload)
				if err != nil {
					return err
				}
				target := PolicyTarget{
					PolicyID:          policy.ID,
					ActuatorChannelID: targetReq.ActuatorChannelID,
					CommandType:       targetReq.CommandType,
					CommandPayload:    string(payloadBytes),
				}
				if targetReq.ExecutionOrder != nil {
					target.ExecutionOrder = *targetReq.ExecutionOrder
				}
				if targetReq.Enabled != nil {
					target.Enabled = *targetReq.Enabled
				}
				if err := tx.Create(&target).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})

	if err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "create_failed", nil)
		return
	}
	response.Success(c, gin.H{"id": policyID})
}

// UpdatePolicy updates an existing control policy.
func (h *Handler) UpdatePolicy(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var req UpdatePolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.PolicyType != nil {
		updates["policy_type"] = *req.PolicyType
	}
	if req.GrowingZoneID != nil {
		updates["growing_zone_id"] = *req.GrowingZoneID
	}
	if req.Priority != nil {
		updates["priority"] = *req.Priority
	}
	if req.RetryLimit != nil {
		updates["retry_limit"] = *req.RetryLimit
	}
	if req.TimeoutSec != nil {
		updates["timeout_sec"] = *req.TimeoutSec
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}
	if req.Version != nil {
		updates["version"] = *req.Version
	}
	if req.EffectiveFrom != nil {
		updates["effective_from"] = *req.EffectiveFrom
	}
	if req.EffectiveTo != nil {
		updates["effective_to"] = *req.EffectiveTo
	}

	if len(updates) == 0 {
		response.Success(c, gin.H{})
		return
	}

	result := h.db.Model(&ControlPolicy{}).Where("id = ?", id).Updates(updates)
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

// DeletePolicy deletes a control policy and its conditions/targets.
func (h *Handler) DeletePolicy(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	err = h.db.Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&ControlPolicy{}).Where("id = ?", id).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return gorm.ErrRecordNotFound
		}

		if err := tx.Where("policy_id = ?", id).Delete(&PolicyTarget{}).Error; err != nil {
			return err
		}
		if err := tx.Where("policy_id = ?", id).Delete(&PolicyCondition{}).Error; err != nil {
			return err
		}
		if err := tx.Where("policy_id = ?", id).Delete(&PolicyExecution{}).Error; err != nil {
			return err
		}
		return tx.Where("id = ?", id).Delete(&ControlPolicy{}).Error
	})

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "delete_failed", nil)
		return
	}
	response.Success(c, gin.H{})
}

// GetPolicy retrieves a control policy with its conditions and targets.
func (h *Handler) GetPolicy(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var policy ControlPolicy
	if err := h.db.Preload("Conditions").Preload("Targets").First(&policy, id).Error; err != nil {
		response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
		return
	}

	response.Success(c, toPolicyResponse(policy))
}

// ListPolicies lists control policies with optional filters.
func (h *Handler) ListPolicies(c *gin.Context) {
	page, size := parsePage(c)
	q := h.db.Model(&ControlPolicy{})

	if v := c.Query("greenhouse_id"); v != "" {
		q = q.Where("greenhouse_id = ?", v)
	}
	if v := strings.TrimSpace(c.Query("policy_type")); v != "" {
		q = q.Where("policy_type = ?", v)
	}
	if v := strings.TrimSpace(c.Query("enabled")); v != "" {
		q = q.Where("enabled = ?", v)
	}
	if v := strings.TrimSpace(c.Query("keyword")); v != "" {
		q = q.Where("(name LIKE ? OR policy_code LIKE ?)", "%"+v+"%", "%"+v+"%")
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	var policies []ControlPolicy
	if total > 0 {
		if err := q.Order("id desc").Limit(size).Offset((page - 1) * size).Find(&policies).Error; err != nil {
			response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
			return
		}
	}

	items := make([]ControlPolicyResponse, 0, len(policies))
	for _, p := range policies {
		items = append(items, toPolicySummary(p))
	}

	response.Success(c, gin.H{
		"page":      page,
		"page_size": size,
		"total":     total,
		"items":     items,
	})
}

// ListByGreenhouse lists policies filtered by greenhouse_id (convenience method).
func (h *Handler) ListByGreenhouse(c *gin.Context) {
	h.ListPolicies(c)
}

// PublishPolicy publishes a policy (sets published_by and published_at).
func (h *Handler) PublishPolicy(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	userID := currentUserID(c)
	now := time.Now().UTC()
	updates := map[string]interface{}{
		"published_by": userID,
		"published_at": now,
	}

	result := h.db.Model(&ControlPolicy{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "publish_failed", nil)
		return
	}
	if result.RowsAffected == 0 {
		response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
		return
	}

	response.Success(c, gin.H{
		"id":           id,
		"published_by": userID,
		"published_at": now.Format(time.RFC3339),
	})
}

// ArchivePolicy disables a policy (sets enabled=0).
func (h *Handler) ArchivePolicy(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	result := h.db.Model(&ControlPolicy{}).Where("id = ?", id).Update("enabled", 0)
	if result.Error != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "archive_failed", nil)
		return
	}
	if result.RowsAffected == 0 {
		response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
		return
	}
	response.Success(c, gin.H{"id": id, "enabled": 0})
}

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

// --- PolicyExecution handlers ---

// ListExecutions lists executions for a policy with time range filter and pagination.
func (h *Handler) ListExecutions(c *gin.Context) {
	page, size := parsePage(c)
	q := h.db.Model(&PolicyExecution{})

	if v := c.Query("policy_id"); v != "" {
		q = q.Where("policy_id = ?", v)
	}
	if v := c.Query("decision"); v != "" {
		q = q.Where("decision = ?", v)
	}
	if v := c.Query("trigger_source"); v != "" {
		q = q.Where("trigger_source = ?", v)
	}
	if v := c.Query("from"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_from", nil)
			return
		}
		q = q.Where("created_at >= ?", t.UTC())
	}
	if v := c.Query("to"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_to", nil)
			return
		}
		q = q.Where("created_at <= ?", t.UTC())
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	var executions []PolicyExecution
	if total > 0 {
		if err := q.Order("id desc").Limit(size).Offset((page - 1) * size).Find(&executions).Error; err != nil {
			response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
			return
		}
	}

	// Get policy names
	nameMap := getPolicyNames(h.db, executions)

	items := make([]PolicyExecutionResponse, 0, len(executions))
	for _, e := range executions {
		resp := PolicyExecutionResponse{
			ID:                e.ID,
			PolicyID:          e.PolicyID,
			PolicyName:        nameMap[e.PolicyID],
			TriggerSource:     e.TriggerSource,
			TriggerMetricCode: e.TriggerMetricCode,
			TriggerValue:      e.TriggerValue,
			Decision:          e.Decision,
			DecisionReason:    e.DecisionReason,
			CommandID:         e.CommandID,
			BatchID:           e.BatchID,
			ExecutedAt:        e.ExecutedAt,
			CreatedAt:         e.CreatedAt,
		}
		items = append(items, resp)
	}

	response.Success(c, gin.H{
		"page":      page,
		"page_size": size,
		"total":     total,
		"items":     items,
	})
}

// GetExecution retrieves a single policy execution by ID.
func (h *Handler) GetExecution(c *gin.Context) {
	id, err := parseID(c.Param("executionId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var execution PolicyExecution
	if err := h.db.First(&execution, id).Error; err != nil {
		response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
		return
	}

	var policyName string
	h.db.Model(&ControlPolicy{}).Select("name").Where("id = ?", execution.PolicyID).Pluck("name", &policyName)

	resp := PolicyExecutionResponse{
		ID:                execution.ID,
		PolicyID:          execution.PolicyID,
		PolicyName:        policyName,
		TriggerSource:     execution.TriggerSource,
		TriggerMetricCode: execution.TriggerMetricCode,
		TriggerValue:      execution.TriggerValue,
		Decision:          execution.Decision,
		DecisionReason:    execution.DecisionReason,
		CommandID:         execution.CommandID,
		BatchID:           execution.BatchID,
		ExecutedAt:        execution.ExecutedAt,
		CreatedAt:         execution.CreatedAt,
	}
	response.Success(c, resp)
}

// ListExecutionsByPolicy lists executions for a specific policy.
func (h *Handler) ListExecutionsByPolicy(c *gin.Context) {
	policyID, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	page, size := parsePage(c)
	q := h.db.Model(&PolicyExecution{}).Where("policy_id = ?", policyID)

	if v := c.Query("from"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_from", nil)
			return
		}
		q = q.Where("created_at >= ?", t.UTC())
	}
	if v := c.Query("to"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_to", nil)
			return
		}
		q = q.Where("created_at <= ?", t.UTC())
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	var executions []PolicyExecution
	if total > 0 {
		if err := q.Order("id desc").Limit(size).Offset((page - 1) * size).Find(&executions).Error; err != nil {
			response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
			return
		}
	}

	var policyName string
	h.db.Model(&ControlPolicy{}).Select("name").Where("id = ?", policyID).Pluck("name", &policyName)

	items := make([]PolicyExecutionResponse, 0, len(executions))
	for _, e := range executions {
		resp := PolicyExecutionResponse{
			ID:                e.ID,
			PolicyID:          e.PolicyID,
			PolicyName:        policyName,
			TriggerSource:     e.TriggerSource,
			TriggerMetricCode: e.TriggerMetricCode,
			TriggerValue:      e.TriggerValue,
			Decision:          e.Decision,
			DecisionReason:    e.DecisionReason,
			CommandID:         e.CommandID,
			BatchID:           e.BatchID,
			ExecutedAt:        e.ExecutedAt,
			CreatedAt:         e.CreatedAt,
		}
		items = append(items, resp)
	}

	response.Success(c, gin.H{
		"page":      page,
		"page_size": size,
		"total":     total,
		"items":     items,
	})
}

// ExecutePolicy manually triggers a policy evaluation and creates an execution log.
func (h *Handler) ExecutePolicy(c *gin.Context) {
	policyID, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var req ExecutePolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	// Verify policy exists and is enabled
	var policy ControlPolicy
	if err := h.db.Select("id", "name", "enabled").First(&policy, policyID).Error; err != nil {
		response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "policy_not_found", nil)
		return
	}
	if policy.Enabled == 0 {
		response.Error(c, http.StatusConflict, platformErrors.CodeConflict, "policy_disabled", nil)
		return
	}

	now := time.Now().UTC()
	decision := "EXECUTED"
	decisionReason := "manual_trigger"

	execution := PolicyExecution{
		PolicyID:          policyID,
		TriggerSource:     req.TriggerSource,
		TriggerMetricCode: req.TriggerMetricCode,
		TriggerValue:      req.TriggerValue,
		Decision:          decision,
		DecisionReason:    decisionReason,
		ExecutedAt:        &now,
	}

	if err := h.db.Create(&execution).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "create_failed", nil)
		return
	}

	response.Success(c, gin.H{
		"id":              execution.ID,
		"decision":        execution.Decision,
		"decision_reason": execution.DecisionReason,
		"executed_at":     execution.ExecutedAt.Format(time.RFC3339),
	})
}

// --- Helper functions ---

// toPolicyResponse converts a ControlPolicy with conditions and targets to a response struct.
func toPolicyResponse(p ControlPolicy) ControlPolicyResponse {
	conditions := make([]PolicyConditionResponse, 0, len(p.Conditions))
	for _, c := range p.Conditions {
		conditions = append(conditions, toConditionResponse(c))
	}
	targets := make([]PolicyTargetResponse, 0, len(p.Targets))
	for _, t := range p.Targets {
		targets = append(targets, toTargetResponse(t))
	}

	return ControlPolicyResponse{
		ID:            p.ID,
		PolicyCode:    p.PolicyCode,
		Name:          p.Name,
		PolicyType:    p.PolicyType,
		GreenhouseID:  p.GreenhouseID,
		GrowingZoneID: p.GrowingZoneID,
		Priority:      p.Priority,
		RetryLimit:    p.RetryLimit,
		TimeoutSec:    p.TimeoutSec,
		Enabled:       p.Enabled,
		Version:       p.Version,
		EffectiveFrom: p.EffectiveFrom,
		EffectiveTo:   p.EffectiveTo,
		CreatedBy:     p.CreatedBy,
		PublishedBy:   p.PublishedBy,
		PublishedAt:   p.PublishedAt,
		Conditions:    conditions,
		Targets:       targets,
		CreatedAt:     p.CreatedAt,
		UpdatedAt:     p.UpdatedAt,
	}
}

// toPolicySummary converts a ControlPolicy to a summary (without nested relations).
func toPolicySummary(p ControlPolicy) ControlPolicyResponse {
	return ControlPolicyResponse{
		ID:            p.ID,
		PolicyCode:    p.PolicyCode,
		Name:          p.Name,
		PolicyType:    p.PolicyType,
		GreenhouseID:  p.GreenhouseID,
		GrowingZoneID: p.GrowingZoneID,
		Priority:      p.Priority,
		RetryLimit:    p.RetryLimit,
		TimeoutSec:    p.TimeoutSec,
		Enabled:       p.Enabled,
		Version:       p.Version,
		EffectiveFrom: p.EffectiveFrom,
		EffectiveTo:   p.EffectiveTo,
		CreatedBy:     p.CreatedBy,
		PublishedBy:   p.PublishedBy,
		PublishedAt:   p.PublishedAt,
		CreatedAt:     p.CreatedAt,
		UpdatedAt:     p.UpdatedAt,
	}
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

// getPolicyNames returns a map of policy IDs to names.
func getPolicyNames(db *gorm.DB, executions []PolicyExecution) map[uint64]string {
	result := make(map[uint64]string)
	if len(executions) == 0 {
		return result
	}

	ids := make([]uint64, 0, len(executions))
	seen := make(map[uint64]bool)
	for _, e := range executions {
		if !seen[e.PolicyID] {
			ids = append(ids, e.PolicyID)
			seen[e.PolicyID] = true
		}
	}

	var policies []ControlPolicy
	db.Select("id", "name").Where("id IN ?", ids).Find(&policies)
	for _, p := range policies {
		result[p.ID] = p.Name
	}
	return result
}

// parseID parses a uint64 from a string parameter.
func parseID(v string) (uint64, error) {
	return strconv.ParseUint(v, 10, 64)
}

// parsePage extracts page and page_size from query parameters.
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

// currentUserID extracts the user ID from the gin context.
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
