package crop

import (
	"hydroponic-backend/internal/auth"
	"hydroponic-backend/internal/platform/di"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup, deps di.Deps) {
	h := NewHandler(deps.MySQL)

	// CropVariety routes
	varieties := r.Group("/crop-varieties")
	varieties.GET("", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator, auth.RoleViewer), h.ListVarieties)
	varieties.POST("", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin), h.CreateVariety)
	varieties.GET("/:id", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator, auth.RoleViewer), h.GetVariety)
	varieties.PUT("/:id", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin), h.UpdateVariety)
	varieties.DELETE("/:id", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin), h.DeleteVariety)

	// GrowthStage routes
	stages := r.Group("/growth-stages")
	stages.GET("", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator, auth.RoleViewer), h.ListStages)
	stages.POST("", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin), h.CreateStage)
	stages.GET("/:id", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator, auth.RoleViewer), h.GetStage)
	stages.PUT("/:id", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin), h.UpdateStage)
	stages.DELETE("/:id", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin), h.DeleteStage)

	// CropBatch routes
	batches := r.Group("/batches")
	batches.GET("", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator, auth.RoleViewer), h.ListBatches)
	batches.POST("", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator), h.CreateBatch)
	batches.GET("/:id", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator, auth.RoleViewer), h.GetBatch)
	batches.PUT("/:id", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator), h.UpdateBatch)
	batches.DELETE("/:id", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin), h.DeleteBatch)
	batches.POST("/:id/transition", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator), h.TransitionBatchStatus)

	// BatchStagePlan routes
	stagePlans := r.Group("/batch-stage-plans")
	stagePlans.GET("", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator, auth.RoleViewer), h.ListStagePlans)
	stagePlans.POST("", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator), h.CreateStagePlan)
	stagePlans.GET("/:id", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator, auth.RoleViewer), h.GetStagePlan)
	stagePlans.PUT("/:id", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator), h.UpdateStagePlan)
	stagePlans.DELETE("/:id", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin), h.DeleteStagePlan)

	// HarvestRecord routes
	harvests := r.Group("/harvests")
	harvests.GET("", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator, auth.RoleViewer), h.ListHarvestsByBatch)
	harvests.POST("", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator), h.CreateHarvest)
	harvests.GET("/summary/:batchId", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator, auth.RoleViewer), h.GetHarvestSummary)
}
