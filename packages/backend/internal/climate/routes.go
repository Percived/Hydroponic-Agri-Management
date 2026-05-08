package climate

import (
	"hydroponic-backend/internal/auth"
	"hydroponic-backend/internal/platform/di"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers all climate module routes.
func RegisterRoutes(r *gin.RouterGroup, deps di.Deps) {
	h := NewHandler(deps.MySQL, deps.MQTT, deps.EventHub, deps.Log)

	// Start climate profile auto-scheduler (event-driven)
	NewProfileScheduler(deps.MySQL, deps.EventHub, deps.MQTT, deps.Log).Start()

	cp := r.Group("/climate-profiles")
	// ClimateProfile CRUD - Admin/Operator for writes, all roles for reads
	cp.POST("", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator), h.CreateProfile)
	cp.POST("/full", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator), h.CreateProfileWithStages)
	cp.GET("", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator, auth.RoleViewer), h.ListProfiles)
	cp.GET("/:id", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator, auth.RoleViewer), h.GetProfile)
	cp.PUT("/:id", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator), h.UpdateProfile)
	cp.DELETE("/:id", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin), h.DeleteProfile)

	// ClimateStage nested resources
	cp.POST("/:id/stages", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator), h.CreateStage)
	cp.GET("/:id/stages", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator, auth.RoleViewer), h.ListStages)
	cp.GET("/:id/stages/:stageId", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator, auth.RoleViewer), h.GetStage)
	cp.PUT("/:id/stages/:stageId", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator), h.UpdateStage)
	cp.DELETE("/:id/stages/:stageId", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator), h.DeleteStage)

	// ClimateStageAction nested resources
	cp.POST("/:id/stages/:stageId/actions", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator), h.CreateAction)
	cp.GET("/:id/stages/:stageId/actions", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator, auth.RoleViewer), h.ListActions)
	cp.PUT("/:id/stages/:stageId/actions/:actionId", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator), h.UpdateAction)
	cp.DELETE("/:id/stages/:stageId/actions/:actionId", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator), h.DeleteAction)

	// ClimateExecutionLog - profile-scoped
	cp.GET("/:id/execution-logs", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator, auth.RoleViewer), h.ListExecutionLogs)
	cp.POST("/:id/execute", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator), h.ExecuteProfile)

	// ClimateExecutionLog - global list
	r.GET("/climate-execution-logs", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator, auth.RoleViewer), h.ListAllExecutionLogs)
}
