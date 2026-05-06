package recipe

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

// ======================== NutrientRecipe Handlers ========================

func (h *Handler) CreateRecipe(c *gin.Context) {
	var req CreateRecipeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	status := req.Status
	if status == "" {
		status = RecipeStatusDraft
	}
	version := req.Version
	if version == "" {
		version = "v1"
	}

	userID := currentUserID(c)

	recipe := NutrientRecipe{
		RecipeCode:    req.RecipeCode,
		Name:          req.Name,
		CropVarietyID: req.CropVarietyID,
		Description:   req.Description,
		Version:       version,
		Status:        status,
		EffectiveFrom: parseTimePtr(req.EffectiveFrom),
		EffectiveTo:   parseTimePtr(req.EffectiveTo),
		CreatedBy:     &userID,
	}

	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&recipe).Error; err != nil {
			return err
		}

		// Create stage targets
		for _, st := range req.StageTargets {
			enabled := uint8(1)
			if st.Enabled != nil {
				enabled = *st.Enabled
			}
			target := RecipeStageTarget{
				RecipeID:      recipe.ID,
				GrowthStageID: st.GrowthStageID,
				MetricCode:    st.MetricCode,
				TargetMin:     st.TargetMin,
				TargetMax:     st.TargetMax,
				Tolerance:     st.Tolerance,
				Unit:          st.Unit,
				Enabled:       enabled,
			}
			if err := tx.Create(&target).Error; err != nil {
				return err
			}
		}

		// Create ion targets
		for _, it := range req.IonTargets {
			enabled := uint8(1)
			if it.Enabled != nil {
				enabled = *it.Enabled
			}
			target := RecipeIonTarget{
				RecipeID:      recipe.ID,
				GrowthStageID: it.GrowthStageID,
				IonCode:       it.IonCode,
				TargetMinMgL:  it.TargetMinMgL,
				TargetMaxMgL:  it.TargetMaxMgL,
				Enabled:       enabled,
			}
			if err := tx.Create(&target).Error; err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "create_failed", nil)
		return
	}

	response.Success(c, toRecipeResponse(recipe))
}

func (h *Handler) GetRecipe(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var recipe NutrientRecipe
	if err := h.db.First(&recipe, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	response.Success(c, toRecipeResponse(recipe))
}

func (h *Handler) UpdateRecipe(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var req UpdateRecipeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.CropVarietyID != nil {
		updates["crop_variety_id"] = *req.CropVarietyID
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Version != nil {
		updates["version"] = *req.Version
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.EffectiveFrom != nil {
		updates["effective_from"] = parseTimePtr(req.EffectiveFrom)
	}
	if req.EffectiveTo != nil {
		updates["effective_to"] = parseTimePtr(req.EffectiveTo)
	}

	if len(updates) == 0 {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "no_fields_to_update", nil)
		return
	}

	if err := h.db.Model(&NutrientRecipe{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "update_failed", nil)
		return
	}

	var recipe NutrientRecipe
	h.db.First(&recipe, id)
	response.Success(c, toRecipeResponse(recipe))
}

func (h *Handler) DeleteRecipe(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	if err := h.db.Transaction(func(tx *gorm.DB) error {
		// Delete associated targets
		if err := tx.Where("recipe_id = ?", id).Delete(&RecipeStageTarget{}).Error; err != nil {
			return err
		}
		if err := tx.Where("recipe_id = ?", id).Delete(&RecipeIonTarget{}).Error; err != nil {
			return err
		}
		// Delete bindings
		if err := tx.Where("recipe_id = ?", id).Delete(&BatchRecipeBinding{}).Error; err != nil {
			return err
		}
		// Delete the recipe itself
		if err := tx.Delete(&NutrientRecipe{}, id).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "delete_failed", nil)
		return
	}

	response.Success(c, gin.H{})
}

func (h *Handler) ListRecipes(c *gin.Context) {
	page, pageSize := parsePageQuery(c)

	q := h.db.Model(&NutrientRecipe{})

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
	if v := strings.TrimSpace(c.Query("search")); v != "" {
		q = q.Where("recipe_code LIKE ? OR name LIKE ?", "%"+v+"%", "%"+v+"%")
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	var recipes []NutrientRecipe
	if total > 0 {
		if err := q.Order("created_at DESC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&recipes).Error; err != nil {
			response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
			return
		}
	}

	items := make([]NutrientRecipeResponse, 0, len(recipes))
	for _, r := range recipes {
		items = append(items, toRecipeResponse(r))
	}

	response.Success(c, RecipeListResponse{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

func (h *Handler) PublishRecipe(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var req PublishRecipeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	now := time.Now().UTC()
	userID := currentUserID(c)

	updates := map[string]interface{}{
		"status":       RecipeStatusActive,
		"version":      req.Version,
		"published_by": userID,
		"published_at": now,
	}

	if err := h.db.Model(&NutrientRecipe{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "publish_failed", nil)
		return
	}

	var recipe NutrientRecipe
	h.db.First(&recipe, id)
	response.Success(c, toRecipeResponse(recipe))
}

// ======================== Recipe Targets Handlers ========================

func (h *Handler) GetRecipeTargets(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var recipe NutrientRecipe
	if err := h.db.First(&recipe, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	var stageTargets []RecipeStageTarget
	if err := h.db.Where("recipe_id = ?", id).Order("id ASC").Find(&stageTargets).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	var ionTargets []RecipeIonTarget
	if err := h.db.Where("recipe_id = ?", id).Order("id ASC").Find(&ionTargets).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	stageResp := make([]RecipeStageTargetResponse, 0, len(stageTargets))
	for _, st := range stageTargets {
		stageResp = append(stageResp, toStageTargetResponse(st))
	}

	ionResp := make([]RecipeIonTargetResponse, 0, len(ionTargets))
	for _, it := range ionTargets {
		ionResp = append(ionResp, toIonTargetResponse(it))
	}

	response.Success(c, RecipeTargetsResponse{
		Recipe:       toRecipeResponse(recipe),
		StageTargets: stageResp,
		IonTargets:   ionResp,
	})
}

func (h *Handler) UpdateRecipeTargets(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var req UpdateRecipeTargetsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	if err := h.db.Transaction(func(tx *gorm.DB) error {
		// Delete existing targets
		if err := tx.Where("recipe_id = ?", id).Delete(&RecipeStageTarget{}).Error; err != nil {
			return err
		}
		if err := tx.Where("recipe_id = ?", id).Delete(&RecipeIonTarget{}).Error; err != nil {
			return err
		}

		// Recreate stage targets
		for _, st := range req.StageTargets {
			enabled := uint8(1)
			if st.Enabled != nil {
				enabled = *st.Enabled
			}
			target := RecipeStageTarget{
				RecipeID:      id,
				GrowthStageID: st.GrowthStageID,
				MetricCode:    st.MetricCode,
				TargetMin:     st.TargetMin,
				TargetMax:     st.TargetMax,
				Tolerance:     st.Tolerance,
				Unit:          st.Unit,
				Enabled:       enabled,
			}
			if err := tx.Create(&target).Error; err != nil {
				return err
			}
		}

		// Recreate ion targets
		for _, it := range req.IonTargets {
			enabled := uint8(1)
			if it.Enabled != nil {
				enabled = *it.Enabled
			}
			target := RecipeIonTarget{
				RecipeID:      id,
				GrowthStageID: it.GrowthStageID,
				IonCode:       it.IonCode,
				TargetMinMgL:  it.TargetMinMgL,
				TargetMaxMgL:  it.TargetMaxMgL,
				Enabled:       enabled,
			}
			if err := tx.Create(&target).Error; err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "update_failed", nil)
		return
	}

	// Return the updated targets
	var stageTargets []RecipeStageTarget
	h.db.Where("recipe_id = ?", id).Order("id ASC").Find(&stageTargets)
	var ionTargets []RecipeIonTarget
	h.db.Where("recipe_id = ?", id).Order("id ASC").Find(&ionTargets)

	stageResp := make([]RecipeStageTargetResponse, 0, len(stageTargets))
	for _, st := range stageTargets {
		stageResp = append(stageResp, toStageTargetResponse(st))
	}
	ionResp := make([]RecipeIonTargetResponse, 0, len(ionTargets))
	for _, it := range ionTargets {
		ionResp = append(ionResp, toIonTargetResponse(it))
	}

	response.Success(c, RecipeTargetsResponse{
		StageTargets: stageResp,
		IonTargets:   ionResp,
	})
}

// ======================== BatchRecipeBinding Handlers ========================

func (h *Handler) CreateBinding(c *gin.Context) {
	var req CreateBindingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	bindingType := req.BindingType
	if bindingType == "" {
		bindingType = BindingTypePrimary
	}
	status := req.Status
	if status == "" {
		status = BindingStatusActive
	}
	version := req.Version
	if version == "" {
		version = "v1"
	}

	effectiveFrom, err := time.Parse(time.RFC3339, req.EffectiveFrom)
	if err != nil {
		effectiveFrom, err = time.Parse(time.RFC3339Nano, req.EffectiveFrom)
		if err != nil {
			response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_effective_from", nil)
			return
		}
	}

	userID := currentUserID(c)

	binding := BatchRecipeBinding{
		BatchID:       req.BatchID,
		RecipeID:      req.RecipeID,
		BindingType:   bindingType,
		Version:       version,
		EffectiveFrom: effectiveFrom.UTC(),
		EffectiveTo:   parseTimePtr(req.EffectiveTo),
		Status:        status,
		CreatedBy:     &userID,
	}

	if err := h.db.Create(&binding).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "create_failed", nil)
		return
	}

	response.Success(c, toBindingResponse(binding))
}

func (h *Handler) GetBinding(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var binding BatchRecipeBinding
	if err := h.db.First(&binding, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(c, http.StatusNotFound, platformErrors.CodeNotFound, "not_found", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	response.Success(c, toBindingResponse(binding))
}

func (h *Handler) UpdateBinding(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var req UpdateBindingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	updates := map[string]interface{}{}
	if req.BindingType != nil {
		updates["binding_type"] = *req.BindingType
	}
	if req.Version != nil {
		updates["version"] = *req.Version
	}
	if req.EffectiveFrom != nil {
		updates["effective_from"] = parseTimePtr(req.EffectiveFrom)
	}
	if req.EffectiveTo != nil {
		updates["effective_to"] = parseTimePtr(req.EffectiveTo)
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	if len(updates) == 0 {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "no_fields_to_update", nil)
		return
	}

	if err := h.db.Model(&BatchRecipeBinding{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "update_failed", nil)
		return
	}

	var binding BatchRecipeBinding
	h.db.First(&binding, id)
	response.Success(c, toBindingResponse(binding))
}

func (h *Handler) DeleteBinding(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	if err := h.db.Delete(&BatchRecipeBinding{}, id).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "delete_failed", nil)
		return
	}

	response.Success(c, gin.H{})
}

func (h *Handler) ListBindings(c *gin.Context) {
	page, pageSize := parsePageQuery(c)

	q := h.db.Model(&BatchRecipeBinding{})

	if v := strings.TrimSpace(c.Query("batch_id")); v != "" {
		id, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_batch_id", nil)
			return
		}
		q = q.Where("batch_id = ?", id)
	}

	if v := strings.TrimSpace(c.Query("recipe_id")); v != "" {
		id, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_recipe_id", nil)
			return
		}
		q = q.Where("recipe_id = ?", id)
	}

	if v := strings.TrimSpace(c.Query("status")); v != "" {
		q = q.Where("status = ?", v)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
		return
	}

	var bindings []BatchRecipeBinding
	if total > 0 {
		if err := q.Order("created_at DESC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&bindings).Error; err != nil {
			response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "query_failed", nil)
			return
		}
	}

	items := make([]BatchRecipeBindingResponse, 0, len(bindings))
	for _, b := range bindings {
		items = append(items, toBindingResponse(b))
	}

	response.Success(c, RecipeListResponse{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

// BindRecipeToBatch handles POST /recipes/:id/bind
func (h *Handler) BindRecipeToBatch(c *gin.Context) {
	recipeID, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_id", nil)
		return
	}

	var req struct {
		BatchID       uint64  `json:"batch_id" binding:"required"`
		BindingType   string  `json:"binding_type"`
		Version       string  `json:"version"`
		EffectiveFrom string  `json:"effective_from" binding:"required"`
		EffectiveTo   *string `json:"effective_to"`
		Status        string  `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	bindingType := req.BindingType
	if bindingType == "" {
		bindingType = BindingTypePrimary
	}
	status := req.Status
	if status == "" {
		status = BindingStatusActive
	}
	version := req.Version
	if version == "" {
		version = "v1"
	}

	effectiveFrom, err := time.Parse(time.RFC3339, req.EffectiveFrom)
	if err != nil {
		effectiveFrom, err = time.Parse(time.RFC3339Nano, req.EffectiveFrom)
		if err != nil {
			response.Error(c, http.StatusBadRequest, platformErrors.CodeValidationError, "invalid_effective_from", nil)
			return
		}
	}

	userID := currentUserID(c)

	binding := BatchRecipeBinding{
		BatchID:       req.BatchID,
		RecipeID:      recipeID,
		BindingType:   bindingType,
		Version:       version,
		EffectiveFrom: effectiveFrom.UTC(),
		EffectiveTo:   parseTimePtr(req.EffectiveTo),
		Status:        status,
		CreatedBy:     &userID,
	}

	if err := h.db.Create(&binding).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, platformErrors.CodeConflict, "create_failed", nil)
		return
	}

	response.Success(c, toBindingResponse(binding))
}

// ======================== Helpers ========================

func toRecipeResponse(r NutrientRecipe) NutrientRecipeResponse {
	resp := NutrientRecipeResponse{
		ID:            r.ID,
		RecipeCode:    r.RecipeCode,
		Name:          r.Name,
		CropVarietyID: r.CropVarietyID,
		Description:   r.Description,
		Version:       r.Version,
		Status:        r.Status,
		CreatedBy:     r.CreatedBy,
		PublishedBy:   r.PublishedBy,
		CreatedAt:     r.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     r.UpdatedAt.Format(time.RFC3339),
	}
	if r.EffectiveFrom != nil {
		s := r.EffectiveFrom.Format(time.RFC3339)
		resp.EffectiveFrom = &s
	}
	if r.EffectiveTo != nil {
		s := r.EffectiveTo.Format(time.RFC3339)
		resp.EffectiveTo = &s
	}
	if r.PublishedAt != nil {
		s := r.PublishedAt.Format(time.RFC3339)
		resp.PublishedAt = &s
	}
	return resp
}

func toStageTargetResponse(st RecipeStageTarget) RecipeStageTargetResponse {
	return RecipeStageTargetResponse{
		ID:            st.ID,
		RecipeID:      st.RecipeID,
		GrowthStageID: st.GrowthStageID,
		MetricCode:    st.MetricCode,
		TargetMin:     st.TargetMin,
		TargetMax:     st.TargetMax,
		Tolerance:     st.Tolerance,
		Unit:          st.Unit,
		Enabled:       st.Enabled,
		CreatedAt:     st.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     st.UpdatedAt.Format(time.RFC3339),
	}
}

func toIonTargetResponse(it RecipeIonTarget) RecipeIonTargetResponse {
	return RecipeIonTargetResponse{
		ID:            it.ID,
		RecipeID:      it.RecipeID,
		GrowthStageID: it.GrowthStageID,
		IonCode:       it.IonCode,
		TargetMinMgL:  it.TargetMinMgL,
		TargetMaxMgL:  it.TargetMaxMgL,
		Enabled:       it.Enabled,
		CreatedAt:     it.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     it.UpdatedAt.Format(time.RFC3339),
	}
}

func toBindingResponse(b BatchRecipeBinding) BatchRecipeBindingResponse {
	resp := BatchRecipeBindingResponse{
		ID:            b.ID,
		BatchID:       b.BatchID,
		RecipeID:      b.RecipeID,
		BindingType:   b.BindingType,
		Version:       b.Version,
		EffectiveFrom: b.EffectiveFrom.Format(time.RFC3339),
		Status:        b.Status,
		CreatedBy:     b.CreatedBy,
		CreatedAt:     b.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     b.UpdatedAt.Format(time.RFC3339),
	}
	if b.EffectiveTo != nil {
		s := b.EffectiveTo.Format(time.RFC3339)
		resp.EffectiveTo = &s
	}
	return resp
}

// ---------- Generic helpers ----------

func parseID(v string) (uint64, error) {
	return strconv.ParseUint(v, 10, 64)
}

func parsePageQuery(c *gin.Context) (int, int) {
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
