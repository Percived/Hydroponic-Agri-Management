package device

import (
	"hydroponic-backend/internal/auth"
	"hydroponic-backend/internal/platform/di"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup, deps di.Deps) {
	h := NewHandler(deps.MySQL)

	devices := r.Group("/devices")
	devices.POST("", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin), h.CreateDevice)
	devices.PUT("/:deviceId", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin), h.UpdateDevice)
	devices.PATCH("/:deviceId/status", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin), h.UpdateDeviceStatus)
	devices.GET("", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator, auth.RoleViewer), h.ListDevices)
	devices.GET("/:deviceId", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator, auth.RoleViewer), h.GetDevice)
	devices.GET("/:deviceId/health", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator, auth.RoleViewer), h.DeviceHealth)
	devices.GET("/:deviceId/telemetry-summary", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator, auth.RoleViewer), h.TelemetrySummary)
	devices.POST("/batch-update", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin), h.BatchUpdate)
	devices.DELETE("/batch", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin), h.BatchDelete)

	greenhouses := r.Group("/devices/greenhouses")
	greenhouses.POST("", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin), h.CreateGreenhouse)
	greenhouses.PUT("/:greenhouseId", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin), h.UpdateGreenhouse)
	greenhouses.GET("", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator, auth.RoleViewer), h.ListGreenhouses)
	greenhouses.GET("/:greenhouseId", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator, auth.RoleViewer), h.GetGreenhouse)
	greenhouses.DELETE("/:greenhouseId", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin), h.DeleteGreenhouse)

	groups := r.Group("/device-groups")
	groups.POST("", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin), h.CreateGroup)
	groups.PUT("/:groupId", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin), h.UpdateGroup)
	groups.GET("", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator, auth.RoleViewer), h.ListGroups)
	groups.DELETE("/:groupId", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin), h.DeleteGroup)
	groups.POST("/:groupId/devices/:deviceId", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin), h.BindDeviceGroup)
	groups.DELETE("/:groupId/devices/:deviceId", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin), h.UnbindDeviceGroup)
}
