package telemetry

import (
	"hydroponic-backend/internal/auth"
	"hydroponic-backend/internal/platform/di"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup, deps di.Deps) {
	RegisterRoutesWithCache(r, deps, NewSensorStatusCache())
}

func RegisterRoutesWithCache(r *gin.RouterGroup, deps di.Deps, cache *SensorStatusCache) {
	h := NewHandler(deps.MySQL, cache, deps.Influx, deps.Config.Influx)

	telemetry := r.Group("/telemetry")
	telemetry.POST("/ingest", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator), h.IngestTelemetry)
	telemetry.GET("/query", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator, auth.RoleViewer), h.QueryTelemetry)

	channels := r.Group("/telemetry/channels")
	channels.GET("/latest", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator, auth.RoleViewer), h.GetLatestBatch)
	channels.GET("/:channelId/latest", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator, auth.RoleViewer), h.GetLatestByChannel)
	channels.GET("/:channelId/history", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator, auth.RoleViewer), h.GetChannelHistory)

	telemetry.DELETE("", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin), h.DeleteTelemetry)
}
