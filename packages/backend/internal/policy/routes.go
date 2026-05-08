package policy

import (
	"hydroponic-backend/internal/auth"
	"hydroponic-backend/internal/platform/di"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers all policy module routes.
func RegisterRoutes(r *gin.RouterGroup, deps di.Deps) {
	h := NewHandler(deps.MySQL)

	// Start policy auto-scheduler (event-driven + timer-scan)
	NewScheduler(deps.MySQL, deps.EventHub, deps.MQTT, deps.Log).Start()

	pol := r.Group("/policies")
	// ControlPolicy CRUD - Admin/Operator for writes, all roles for reads
	pol.POST("", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator), h.CreatePolicy)
	pol.POST("/full", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator), h.CreatePolicyWithNested)
	pol.GET("", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator, auth.RoleViewer), h.ListPolicies)
	pol.GET("/:id", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator, auth.RoleViewer), h.GetPolicy)
	pol.PUT("/:id", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator), h.UpdatePolicy)
	pol.DELETE("/:id", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin), h.DeletePolicy)

	// Publish and Archive actions
	pol.POST("/:id/publish", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin), h.PublishPolicy)
	pol.POST("/:id/archive", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin), h.ArchivePolicy)

	// Policy manual execution
	pol.POST("/:id/execute", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator), h.ExecutePolicy)
	pol.GET("/:id/executions", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator, auth.RoleViewer), h.ListExecutionsByPolicy)

	// PolicyCondition nested resources
	pol.GET("/:id/conditions", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator, auth.RoleViewer), h.ListConditions)
	pol.POST("/:id/conditions", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator), h.CreateCondition)
	pol.GET("/:id/conditions/:conditionId", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator, auth.RoleViewer), h.GetCondition)
	pol.PUT("/:id/conditions/:conditionId", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator), h.UpdateCondition)
	pol.DELETE("/:id/conditions/:conditionId", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator), h.DeleteCondition)

	// PolicyTarget nested resources
	pol.GET("/:id/targets", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator, auth.RoleViewer), h.ListTargets)
	pol.POST("/:id/targets", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator), h.CreateTarget)
	pol.GET("/:id/targets/:targetId", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator, auth.RoleViewer), h.GetTarget)
	pol.PUT("/:id/targets/:targetId", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator), h.UpdateTarget)
	pol.DELETE("/:id/targets/:targetId", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator), h.DeleteTarget)

	// PolicyExecution - global list and detail
	r.GET("/policy-executions", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator, auth.RoleViewer), h.ListExecutions)
	r.GET("/policy-executions/:executionId", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator, auth.RoleViewer), h.GetExecution)
}
