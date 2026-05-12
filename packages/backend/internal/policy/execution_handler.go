package policy

import (
	"net/http"
	"strconv"
	"time"

	"hydroponic-backend/internal/auth"
	platformErrors "hydroponic-backend/internal/platform/errors"
	"hydroponic-backend/internal/platform/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Handler holds dependencies for policy HTTP handlers.
type Handler struct {
	db        *gorm.DB
	scheduler *Scheduler
}

// NewHandler creates a new policy Handler.
func NewHandler(db *gorm.DB, scheduler *Scheduler) *Handler {
	return &Handler{db: db, scheduler: scheduler}
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
	if !policy.Enabled {
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
