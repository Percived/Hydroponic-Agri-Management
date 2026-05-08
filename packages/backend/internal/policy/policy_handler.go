package policy

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	platformErrors "hydroponic-backend/internal/platform/errors"
	"hydroponic-backend/internal/platform/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

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

// ArchivePolicy disables a policy (sets enabled=false).
func (h *Handler) ArchivePolicy(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	result := h.db.Model(&ControlPolicy{}).Where("id = ?", id).Update("enabled", false)
	if result.Error != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "archive_failed", nil)
		return
	}
	if result.RowsAffected == 0 {
		response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
		return
	}
	response.Success(c, gin.H{"id": id, "enabled": false})
}

// --- ControlPolicy helpers ---

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
