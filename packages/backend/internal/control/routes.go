package control

import (
	"hydroponic-backend/internal/auth"
	"hydroponic-backend/internal/platform/di"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup, deps di.Deps) {
	h := NewHandler(deps.MySQL, deps.MQTT, deps.Log)
	controls := r.Group("/controls")

	controls.POST("/commands", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator), h.CreateCommand)
	controls.GET("/commands/:commandId", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator, auth.RoleViewer), h.GetCommand)
	controls.GET("/commands", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator, auth.RoleViewer), h.ListCommands)

	controls.POST("/rules", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator), h.CreateRule)
	controls.PUT("/rules/:ruleId", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator), h.UpdateRule)
	controls.DELETE("/rules/:ruleId", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator), h.DeleteRule)
	controls.GET("/rules", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator, auth.RoleViewer), h.ListRules)

	controls.POST("/templates", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin), h.CreateTemplate)
	controls.POST("/templates/:templateId/apply", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin), h.ApplyTemplate)
	controls.GET("/templates", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator, auth.RoleViewer), h.ListTemplates)
}
