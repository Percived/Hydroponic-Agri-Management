package policy

import (
	"encoding/json"
	"errors"
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
	if req.ScheduleMode != nil {
		policy.ScheduleMode = req.ScheduleMode
	}
	if req.RunOnceAt != nil {
		policy.RunOnceAt = req.RunOnceAt
	}
	if req.TimeOfDay != nil {
		policy.TimeOfDay = req.TimeOfDay
	}
	if req.WeekdaysMask != nil {
		policy.WeekdaysMask = req.WeekdaysMask
	}
	if req.Timezone != nil {
		policy.Timezone = strings.TrimSpace(*req.Timezone)
	}

	normalizePolicySchedule(&policy)
	if err := validatePolicySchedule(policy); err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, err.Error(), nil)
		return
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
		if req.ScheduleMode != nil {
			policy.ScheduleMode = req.ScheduleMode
		}
		if req.RunOnceAt != nil {
			policy.RunOnceAt = req.RunOnceAt
		}
		if req.TimeOfDay != nil {
			policy.TimeOfDay = req.TimeOfDay
		}
		if req.WeekdaysMask != nil {
			policy.WeekdaysMask = req.WeekdaysMask
		}
		if req.Timezone != nil {
			policy.Timezone = strings.TrimSpace(*req.Timezone)
		}
		normalizePolicySchedule(&policy)
		if err := validatePolicySchedule(policy); err != nil {
			return err
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

	var existing ControlPolicy
	if err := h.db.First(&existing, id).Error; err != nil {
		response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
		return
	}

	candidate := existing
	if req.Name != nil {
		candidate.Name = *req.Name
	}
	if req.PolicyType != nil {
		candidate.PolicyType = *req.PolicyType
	}
	if req.GrowingZoneID != nil {
		candidate.GrowingZoneID = req.GrowingZoneID
	}
	if req.Priority != nil {
		candidate.Priority = *req.Priority
	}
	if req.RetryLimit != nil {
		candidate.RetryLimit = *req.RetryLimit
	}
	if req.TimeoutSec != nil {
		candidate.TimeoutSec = *req.TimeoutSec
	}
	if req.Enabled != nil {
		candidate.Enabled = *req.Enabled
	}
	if req.Version != nil {
		candidate.Version = *req.Version
	}
	if req.EffectiveFrom != nil {
		candidate.EffectiveFrom = req.EffectiveFrom
	}
	if req.EffectiveTo != nil {
		candidate.EffectiveTo = req.EffectiveTo
	}
	if req.ScheduleMode != nil {
		candidate.ScheduleMode = req.ScheduleMode
	}
	if req.RunOnceAt != nil {
		candidate.RunOnceAt = req.RunOnceAt
	}
	if req.TimeOfDay != nil {
		candidate.TimeOfDay = req.TimeOfDay
	}
	if req.WeekdaysMask != nil {
		candidate.WeekdaysMask = req.WeekdaysMask
	}
	if req.Timezone != nil {
		candidate.Timezone = strings.TrimSpace(*req.Timezone)
	}

	normalizePolicySchedule(&candidate)
	if err := validatePolicySchedule(candidate); err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, err.Error(), nil)
		return
	}
	if scheduleConfigChanged(existing, candidate) {
		candidate.LastScheduledFor = nil
	}

	updates := map[string]interface{}{
		"name":               candidate.Name,
		"policy_type":        candidate.PolicyType,
		"growing_zone_id":    candidate.GrowingZoneID,
		"priority":           candidate.Priority,
		"retry_limit":        candidate.RetryLimit,
		"timeout_sec":        candidate.TimeoutSec,
		"enabled":            candidate.Enabled,
		"version":            candidate.Version,
		"effective_from":     candidate.EffectiveFrom,
		"effective_to":       candidate.EffectiveTo,
		"schedule_mode":      candidate.ScheduleMode,
		"run_once_at":        candidate.RunOnceAt,
		"time_of_day":        candidate.TimeOfDay,
		"weekdays_mask":      candidate.WeekdaysMask,
		"timezone":           candidate.Timezone,
		"last_scheduled_for": candidate.LastScheduledFor,
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
		ID:               p.ID,
		PolicyCode:       p.PolicyCode,
		Name:             p.Name,
		PolicyType:       p.PolicyType,
		GreenhouseID:     p.GreenhouseID,
		GrowingZoneID:    p.GrowingZoneID,
		Priority:         p.Priority,
		RetryLimit:       p.RetryLimit,
		TimeoutSec:       p.TimeoutSec,
		Enabled:          p.Enabled,
		Version:          p.Version,
		EffectiveFrom:    p.EffectiveFrom,
		EffectiveTo:      p.EffectiveTo,
		ScheduleMode:     p.ScheduleMode,
		RunOnceAt:        p.RunOnceAt,
		TimeOfDay:        p.TimeOfDay,
		WeekdaysMask:     p.WeekdaysMask,
		Timezone:         p.Timezone,
		LastScheduledFor: p.LastScheduledFor,
		CreatedBy:        p.CreatedBy,
		PublishedBy:      p.PublishedBy,
		PublishedAt:      p.PublishedAt,
		Conditions:       conditions,
		Targets:          targets,
		CreatedAt:        p.CreatedAt,
		UpdatedAt:        p.UpdatedAt,
	}
}

// toPolicySummary converts a ControlPolicy to a summary (without nested relations).
func toPolicySummary(p ControlPolicy) ControlPolicyResponse {
	return ControlPolicyResponse{
		ID:               p.ID,
		PolicyCode:       p.PolicyCode,
		Name:             p.Name,
		PolicyType:       p.PolicyType,
		GreenhouseID:     p.GreenhouseID,
		GrowingZoneID:    p.GrowingZoneID,
		Priority:         p.Priority,
		RetryLimit:       p.RetryLimit,
		TimeoutSec:       p.TimeoutSec,
		Enabled:          p.Enabled,
		Version:          p.Version,
		EffectiveFrom:    p.EffectiveFrom,
		EffectiveTo:      p.EffectiveTo,
		ScheduleMode:     p.ScheduleMode,
		RunOnceAt:        p.RunOnceAt,
		TimeOfDay:        p.TimeOfDay,
		WeekdaysMask:     p.WeekdaysMask,
		Timezone:         p.Timezone,
		LastScheduledFor: p.LastScheduledFor,
		CreatedBy:        p.CreatedBy,
		PublishedBy:      p.PublishedBy,
		PublishedAt:      p.PublishedAt,
		CreatedAt:        p.CreatedAt,
		UpdatedAt:        p.UpdatedAt,
	}
}

func normalizePolicySchedule(p *ControlPolicy) {
	if p.PolicyType != "SCHEDULE" {
		p.ScheduleMode = nil
		p.RunOnceAt = nil
		p.TimeOfDay = nil
		p.WeekdaysMask = nil
		return
	}
	if strings.TrimSpace(p.Timezone) == "" {
		p.Timezone = "Asia/Shanghai"
	}
	if p.ScheduleMode == nil {
		return
	}
	switch *p.ScheduleMode {
	case "ONCE":
		p.TimeOfDay = nil
		p.WeekdaysMask = nil
	case "DAILY":
		p.RunOnceAt = nil
		p.WeekdaysMask = nil
	case "WEEKLY":
		p.RunOnceAt = nil
	}
}

type policyValidationError string

func (e policyValidationError) Error() string { return string(e) }

func validatePolicySchedule(p ControlPolicy) error {
	if p.PolicyType != "SCHEDULE" {
		if p.ScheduleMode != nil || p.RunOnceAt != nil || p.TimeOfDay != nil || p.WeekdaysMask != nil {
			return policyValidationError("schedule_fields_not_allowed")
		}
		return nil
	}

	if p.ScheduleMode == nil {
		return policyValidationError("schedule_mode_required")
	}

	switch *p.ScheduleMode {
	case "ONCE":
		if p.RunOnceAt == nil {
			return policyValidationError("run_once_at_required")
		}
	case "DAILY":
		if p.TimeOfDay == nil || strings.TrimSpace(*p.TimeOfDay) == "" {
			return policyValidationError("time_of_day_required")
		}
	case "WEEKLY":
		if p.TimeOfDay == nil || strings.TrimSpace(*p.TimeOfDay) == "" || p.WeekdaysMask == nil || *p.WeekdaysMask == 0 {
			return policyValidationError("weekly_schedule_fields_required")
		}
	default:
		return policyValidationError("schedule_mode_invalid")
	}

	return nil
}

func scheduleConfigChanged(before, after ControlPolicy) bool {
	if before.PolicyType != after.PolicyType {
		return true
	}
	if !equalStringPtr(before.ScheduleMode, after.ScheduleMode) {
		return true
	}
	if !equalTimePtr(before.RunOnceAt, after.RunOnceAt) {
		return true
	}
	if !equalStringPtr(before.TimeOfDay, after.TimeOfDay) {
		return true
	}
	if !equalUint8Ptr(before.WeekdaysMask, after.WeekdaysMask) {
		return true
	}
	return strings.TrimSpace(before.Timezone) != strings.TrimSpace(after.Timezone)
}

func equalStringPtr(a, b *string) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return *a == *b
}

func equalTimePtr(a, b *time.Time) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return a.Equal(*b)
}

func equalUint8Ptr(a, b *uint8) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return *a == *b
}

func isPolicyValidationError(err error) bool {
	var validationErr policyValidationError
	return errors.As(err, &validationErr)
}
