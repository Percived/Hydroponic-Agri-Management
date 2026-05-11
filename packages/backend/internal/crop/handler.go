package crop

import (
	"strconv"
	"time"

	"hydroponic-backend/internal/auth"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	db *gorm.DB
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{db: db}
}

// ======================== Helpers ========================

// toBatchResponse loads the related CropVariety and builds the response.
func (h *Handler) toBatchResponse(b CropBatch) CropBatchResponse {
	// Load variety info
	var variety CropVariety
	h.db.Select("code", "name").Where("id = ?", b.CropVarietyID).First(&variety)

	resp := CropBatchResponse{
		ID:              b.ID,
		BatchNo:         b.BatchNo,
		GreenhouseID:    b.GreenhouseID,
		GrowingZoneID:   b.GrowingZoneID,
		CropVarietyID:   b.CropVarietyID,
		VarietyCode:     variety.Code,
		VarietyName:     variety.Name,
		Status:          b.Status,
		PlantingDensity: b.PlantingDensity,
		TotalPlants:     b.TotalPlants,
		RecipeVersion:   b.RecipeVersion,
		PolicyVersion:   b.PolicyVersion,
		ActiveRecipeID:  b.ActiveRecipeID,
		ActivePolicyID:  b.ActivePolicyID,
		ActiveClimateID: b.ActiveClimateID,
		Note:            b.Note,
		CreatedBy:       b.CreatedBy,
		CreatedAt:       b.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       b.UpdatedAt.Format(time.RFC3339),
	}

	if b.StartedAt != nil {
		s := b.StartedAt.Format(time.RFC3339)
		resp.StartedAt = &s
	}
	if b.EndedAt != nil {
		s := b.EndedAt.Format(time.RFC3339)
		resp.EndedAt = &s
	}
	if b.ExpectedHarvestAt != nil {
		s := b.ExpectedHarvestAt.Format(time.RFC3339)
		resp.ExpectedHarvestAt = &s
	}

	return resp
}

func toVarietyResponse(v CropVariety) CropVarietyResponse {
	return CropVarietyResponse{
		ID:               v.ID,
		Code:             v.Code,
		Name:             v.Name,
		Description:      v.Description,
		DefaultCycleDays: v.DefaultCycleDays,
		CreatedAt:        v.CreatedAt.Format(time.RFC3339),
		UpdatedAt:        v.UpdatedAt.Format(time.RFC3339),
	}
}

func toStageResponse(s GrowthStage) GrowthStageResponse {
	return GrowthStageResponse{
		ID:                  s.ID,
		Code:                s.Code,
		Name:                s.Name,
		SortOrder:           s.SortOrder,
		DefaultDurationDays: s.DefaultDurationDays,
		CreatedAt:           s.CreatedAt.Format(time.RFC3339),
		UpdatedAt:           s.UpdatedAt.Format(time.RFC3339),
	}
}

func toStagePlanResponse(p BatchStagePlan) BatchStagePlanResponse {
	return BatchStagePlanResponse{
		ID:            p.ID,
		BatchID:       p.BatchID,
		GrowthStageID: p.GrowthStageID,
		RecipeID:      p.RecipeID,
		PolicyID:      p.PolicyID,
		ClimateID:     p.ClimateID,
		StageStartAt:  p.StageStartAt.Format(time.RFC3339),
		StageEndAt:    p.StageEndAt.Format(time.RFC3339),
		TargetECMin:   p.TargetECMin,
		TargetECMax:   p.TargetECMax,
		TargetPHMin:   p.TargetPHMin,
		TargetPHMax:   p.TargetPHMax,
		CreatedAt:     p.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     p.UpdatedAt.Format(time.RFC3339),
	}
}

func toHarvestResponse(r HarvestRecord) HarvestRecordResponse {
	return HarvestRecordResponse{
		ID:              r.ID,
		BatchID:         r.BatchID,
		HarvestedAt:     r.HarvestedAt.Format(time.RFC3339),
		HarvestWeightKg: r.HarvestWeightKg,
		Grade:           r.Grade,
		GradeWeightKg:   r.GradeWeightKg,
		Note:            r.Note,
		HarvestedBy:     r.HarvestedBy,
		CreatedAt:       r.CreatedAt.Format(time.RFC3339),
	}
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
