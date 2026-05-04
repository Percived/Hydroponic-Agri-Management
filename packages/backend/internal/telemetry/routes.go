package telemetry

import (
	"hydroponic-backend/internal/auth"
	"hydroponic-backend/internal/platform/di"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup, deps di.Deps) {
	h := NewHandler(deps.MySQL, deps.Influx, deps.MQTT, deps.Config.Influx, deps.Log, deps.EventHub)

	telemetry := r.Group("/telemetry")
	telemetry.POST("", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator), h.Ingest)
	telemetry.GET("/latest", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator, auth.RoleViewer), h.Latest)
	telemetry.GET("/history", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator, auth.RoleViewer), h.History)
	telemetry.GET("/stats", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator, auth.RoleViewer), h.Stats)
	telemetry.POST("/retention", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin), h.SetRetention)
	telemetry.GET("/subscribe", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator, auth.RoleViewer), h.Subscribe)
	telemetry.GET("/system-configs", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin), h.GetConfigs)
	telemetry.PUT("/system-configs", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin), h.UpdateConfig)
}
