package command

import (
	"hydroponic-backend/internal/auth"
	"hydroponic-backend/internal/platform/di"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers all command module routes.
func RegisterRoutes(r *gin.RouterGroup, deps di.Deps) {
	h := NewHandler(deps.MySQL, deps.MQTT, deps.EventHub)

	cmds := r.Group("/commands")
	// ControlCommand CRUD - Admin/Operator for writes, all roles for reads
	cmds.POST("", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator), h.CreateCommand)
	cmds.GET("", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator, auth.RoleViewer), h.ListCommands)
	cmds.GET("/:id", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator, auth.RoleViewer), h.GetCommand)

	// Send and Ack actions
	cmds.POST("/:id/send", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator), h.SendCommand)
	cmds.POST("/:id/ack", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator), h.AckCommand)

	// MQTT dispatch: sync (wait for ack) and async (fire-and-forget)
	cmds.POST("/dispatch-and-wait", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator), h.DispatchAndWait)
	cmds.POST("/dispatch-async", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator), h.DispatchAsync)

	// ControlCommandReceipt nested resources
	cmds.POST("/:id/receipts", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator), h.CreateReceipt)
	cmds.GET("/:id/receipts", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator, auth.RoleViewer), h.ListReceipts)
	cmds.GET("/:id/receipts/:receiptId", auth.AuthRequired(deps.Config.Auth, deps.MySQL, auth.RoleAdmin, auth.RoleOperator, auth.RoleViewer), h.GetReceipt)
}
